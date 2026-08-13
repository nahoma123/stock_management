package handlers

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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
	report, err := services.DeployCustomizationArchive(tenant.Subdomain, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	services.UpdateLogAndBroadcast(tenant.ID, fmt.Sprintf("Customization module %s deployed to tenant addon layer.\n", report.Module))
	c.JSON(http.StatusOK, report)
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
