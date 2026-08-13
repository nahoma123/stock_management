package handlers

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"saas-superadmin-backend/database"
	"saas-superadmin-backend/models"
	"saas-superadmin-backend/services"
)

func GetTenantMonitoring(c *gin.Context) {
	tenant, ok := tenantFromParam(c)
	if !ok {
		return
	}
	monitoring, err := services.FetchTenantMonitoring(tenant)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "status": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, monitoring)
}

func ValidateTenantCustomization(c *gin.Context) {
	if !authorizeCustomization(c) {
		return
	}
	if _, ok := tenantFromParam(c); !ok {
		return
	}
	data, ok := customizationUpload(c)
	if !ok {
		return
	}
	report, err := services.ValidateCustomizationArchive(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func DeployTenantCustomization(c *gin.Context) {
	if !authorizeCustomization(c) {
		return
	}
	tenant, ok := tenantFromParam(c)
	if !ok {
		return
	}
	data, ok := customizationUpload(c)
	if !ok {
		return
	}
	var latest models.CustomizationRelease
	database.DB.Where("tenant_id = ?", tenant.ID).Order("version desc").First(&latest)
	release := models.CustomizationRelease{
		TenantID:    tenant.ID,
		Version:     latest.Version + 1,
		State:       "staging",
		ArchiveHash: services.CustomizationArchiveHash(data),
	}
	if err := database.DB.Create(&release).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create release record"})
		return
	}
	report, err := services.StageCustomizationRelease(tenant.Subdomain, release.ID, data)
	if err != nil {
		database.DB.Model(&release).Updates(map[string]interface{}{"state": "failed", "failure": err.Error()})
		services.RecordAudit(tenant.ID, "release.stage", "customization", "failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&release).Updates(map[string]interface{}{
		"module_name": report.Module,
		"file_count":  report.Files,
		"size_bytes":  report.Size,
		"state":       "staged",
	})
	services.RecordAudit(tenant.ID, "release.stage", report.Module, "success", fmt.Sprintf("release v%d", release.Version))
	services.UpdateLogAndBroadcast(tenant.ID, fmt.Sprintf("Customization module %s staged as release v%d.\n", report.Module, release.Version))
	database.DB.First(&release, release.ID)
	c.JSON(http.StatusOK, gin.H{"report": report, "release": release})
}

func GetCustomizationReleases(c *gin.Context) {
	tenant, ok := tenantFromParam(c)
	if !ok {
		return
	}
	var releases []models.CustomizationRelease
	if err := database.DB.Where("tenant_id = ?", tenant.ID).Order("version desc").Find(&releases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load releases"})
		return
	}
	c.JSON(http.StatusOK, releases)
}

func ActivateCustomizationRelease(c *gin.Context) {
	changeCustomizationRelease(c, false)
}

func RollbackCustomizationRelease(c *gin.Context) {
	changeCustomizationRelease(c, true)
}

func changeCustomizationRelease(c *gin.Context, rollback bool) {
	if !authorizeCustomization(c) {
		return
	}
	tenant, ok := tenantFromParam(c)
	if !ok {
		return
	}
	releaseID, err := strconv.Atoi(c.Param("release_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid release ID"})
		return
	}
	var release models.CustomizationRelease
	if err := database.DB.Where("id = ? AND tenant_id = ?", releaseID, tenant.ID).First(&release).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		return
	}
	if release.State != "staged" && release.State != "superseded" && release.State != "failed" {
		c.JSON(http.StatusConflict, gin.H{"error": "Release is not available for activation"})
		return
	}
	action := "release.activate"
	if rollback {
		action = "release.rollback"
	}
	if err := services.DockerStopTenant(tenant.ID); err != nil {
		services.RecordAudit(tenant.ID, action, release.ModuleName, "failed", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	activation, err := services.ActivateCustomizationRelease(tenant.Subdomain, release.ModuleName, release.ID)
	if err != nil {
		_ = restartTenantContainer(tenant)
		services.RecordAudit(tenant.ID, action, release.ModuleName, "failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = services.DockerRunModuleMaintenance(tenant, release.ModuleName)
	if err != nil {
		recoverCustomizationActivation(tenant, activation)
		database.DB.Model(&release).Updates(map[string]interface{}{"state": "failed", "failure": err.Error()})
		services.RecordAudit(tenant.ID, action, release.ModuleName, "failed", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if err := restartTenantContainer(tenant); err != nil {
		recoverCustomizationActivation(tenant, activation)
		database.DB.Model(&release).Updates(map[string]interface{}{"state": "failed", "failure": err.Error()})
		services.RecordAudit(tenant.ID, action, release.ModuleName, "failed", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	database.DB.Model(&models.CustomizationRelease{}).
		Where("tenant_id = ? AND module_name = ? AND state = 'active'", tenant.ID, release.ModuleName).
		Update("state", "superseded")
	database.DB.Model(&release).Updates(map[string]interface{}{
		"state":        "active",
		"failure":      "",
		"activated_at": &now,
	})
	services.RecordAudit(tenant.ID, action, release.ModuleName, "success", fmt.Sprintf("release v%d", release.Version))
	c.JSON(http.StatusOK, release)
}

func recoverCustomizationActivation(tenant models.Tenant, activation services.ReleaseActivation) {
	if err := services.RestoreCustomizationActivation(tenant.Subdomain, activation); err != nil {
		return
	}
	_ = services.DockerRunModuleMaintenance(tenant, activation.Module)
	_ = restartTenantContainer(tenant)
}

func EnrollTenantMonitoring(c *gin.Context) {
	if !authorizeCustomization(c) {
		return
	}
	tenant, ok := tenantFromParam(c)
	if !ok {
		return
	}
	if tenant.AgentToken == "" {
		token, err := randomToken("agent_")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate agent credential"})
			return
		}
		tenant.AgentToken = token
		if err := database.DB.Model(&tenant).Update("agent_token", token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save agent credential"})
			return
		}
	}
	err := services.DockerStopTenant(tenant.ID)
	if err == nil {
		err = services.DockerRunModuleMaintenance(tenant, "tenant_management_agent")
	}
	if err == nil {
		err = restartTenantContainer(tenant)
	}
	if err != nil {
		services.RecordAudit(tenant.ID, "monitoring.enroll", "tenant_management_agent", "failed", err.Error())
		_ = restartTenantContainer(tenant)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	services.RecordAudit(tenant.ID, "monitoring.enroll", "tenant_management_agent", "success", "agent enrolled")
	c.JSON(http.StatusOK, gin.H{"message": "Tenant monitoring enrolled"})
}

func GetTenantAudit(c *gin.Context) {
	tenant, ok := tenantFromParam(c)
	if !ok {
		return
	}
	var events []models.PlatformAuditEvent
	database.DB.Where("tenant_id = ?", tenant.ID).Order("created_at desc").Limit(100).Find(&events)
	c.JSON(http.StatusOK, events)
}

func restartTenantContainer(tenant models.Tenant) error {
	hostProjectPath := services.GetEnv("HOST_PROJECT_PATH", "/home/nahom/Desktop/project/odoo_project")
	tenantAddons := filepath.Join(hostProjectPath, "tenants", tenant.Subdomain, "custom_addons")
	return services.DockerCreateAndStartDaemonContainer(
		tenant,
		services.GetEnv("POSTGRES_HOST", "db"),
		services.GetEnv("POSTGRES_PORT", "5432"),
		services.GetEnv("POSTGRES_USER", "odoo"),
		services.GetEnv("POSTGRES_PASSWORD", "odoo"),
		tenantAddons,
	)
}

func authorizeCustomization(c *gin.Context) bool {
	expected := os.Getenv("CUSTOMIZATION_ADMIN_TOKEN")
	provided := c.GetHeader("X-Customization-Admin-Token")
	if expected == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Customization deployment is not configured"})
		return false
	}
	if len(expected) != len(provided) || subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid customization administrator token"})
		return false
	}
	return true
}

func tenantFromParam(c *gin.Context) (models.Tenant, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return models.Tenant{}, false
	}
	tenant, err := services.GetTenantByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return models.Tenant{}, false
	}
	return tenant, true
}

func customizationUpload(c *gin.Context) ([]byte, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, services.MaxCustomizationSize+(1<<20))
	file, header, err := c.Request.FormFile("package")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Attach a ZIP package in the package field"})
		return nil, false
	}
	defer file.Close()
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customization package must be a ZIP file"})
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(file, services.MaxCustomizationSize+1))
	if err != nil || len(data) > services.MaxCustomizationSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Customization package exceeds the 25 MB limit"})
		return nil, false
	}
	return data, true
}
