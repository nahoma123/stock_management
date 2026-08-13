package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"saas-superadmin-backend/database"
	"saas-superadmin-backend/models"
	"saas-superadmin-backend/services"
	"saas-superadmin-backend/websocket"
)

var tenantSubdomainPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)

func GetTenants(c *gin.Context) {
	var tenants []models.Tenant
	if err := database.DB.Order("id desc").Find(&tenants).Error; err != nil {
		log.Printf("Error getting tenants: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tenants"})
		return
	}
	c.JSON(http.StatusOK, tenants)
}

func DisableTenant(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	containerName := fmt.Sprintf("odoo_tenant_%d", id)

	services.UpdateLogAndBroadcast(id, "Disabling tenant container...\n")
	_, status, err := services.CallDockerAPI("POST", fmt.Sprintf("/containers/%s/stop?t=10", containerName), nil)
	if err != nil || status >= 400 {
		log.Printf("Error stopping container %s: %v (status %d)", containerName, err, status)
		services.UpdateLogAndBroadcast(id, "WARNING: Failed to stop tenant container.\n")
	} else {
		services.UpdateLogAndBroadcast(id, "Tenant container stopped.\n")
	}

	services.UpdateTenantState(id, "disabled")
	c.Status(http.StatusOK)
}

func EnableTenant(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := services.GetTenantByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	containerName := fmt.Sprintf("odoo_tenant_%d", id)

	services.UpdateLogAndBroadcast(id, "Enabling tenant container...\n")

	_, status, err := services.CallDockerAPI("POST", fmt.Sprintf("/containers/%s/start", containerName), nil)
	if err != nil || status >= 400 {
		log.Printf("Error starting container %s: %v (status %d)", containerName, err, status)
		services.UpdateLogAndBroadcast(id, "WARNING: Failed to start tenant container.\n")
	} else {
		services.UpdateLogAndBroadcast(id, "Tenant container started.\n")
	}

	services.UpdateTenantState(id, "active")
	c.Status(http.StatusOK)
}

func DeleteTenant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := services.GetTenantByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	containerName := fmt.Sprintf("odoo_tenant_%d", id)
	log.Printf("Removing Docker container %s for tenant %d", containerName, id)
	_, status, err := services.CallDockerAPI("DELETE", fmt.Sprintf("/containers/%s?force=true", containerName), nil)
	if err != nil || status >= 400 {
		log.Printf("Error removing container %s: %v (status %d)", containerName, err, status)
	}

	tenantDir := filepath.Join("/app/tenants", tenant.Subdomain)
	log.Printf("Deleting directory %s for tenant %d", tenantDir, id)
	if err := os.RemoveAll(tenantDir); err != nil {
		log.Printf("Error deleting tenant directory %s: %v", tenantDir, err)
	}

	dropDbSQL := fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", tenant.DbName)
	if err := database.DB.Exec(dropDbSQL).Error; err != nil {
		log.Printf("Error dropping database for tenant %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to drop tenant database"})
		return
	}

	if err := database.DB.Delete(&tenant).Error; err != nil {
		log.Printf("Error deleting tenant %d record: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant record"})
		return
	}

	msg := websocket.WebSocketMessage{Type: "tenant_deleted", Payload: gin.H{"id": id}}
	if jsonMsg, err := json.Marshal(msg); err == nil {
		websocket.AppHub.Broadcast <- jsonMsg
	}

	c.Status(http.StatusOK)
}

type SetExpiryRequest struct {
	ExpiryDate string `json:"expiry_date" binding:"required"`
}

func SetTenantExpiry(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req SetExpiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := database.DB.Model(&models.Tenant{}).Where("id = ?", id).Update("license_expiry_date", req.ExpiryDate).Error; err != nil {
		log.Printf("Error updating expiry for tenant %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expiry date"})
		return
	}

	if tenant, err := services.GetTenantByID(id); err == nil {
		services.BroadcastTenantUpdate(tenant)
	}
	c.Status(http.StatusOK)
}

type CreateTenantRequest struct {
	Name      string `json:"name" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
}

func CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Subdomain = strings.ToLower(strings.TrimSpace(req.Subdomain))
	if !tenantSubdomainPattern.MatchString(req.Subdomain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subdomain must start with a letter and contain only lowercase letters, numbers, hyphens, or underscores"})
		return
	}

	dbName := strings.ToLower(strings.ReplaceAll(req.Subdomain, ".", "_")) + "_db"
	now := time.Now()

	// Generate separate credentials for the tenant-facing API and private management agent.
	apiKey, err := randomToken("tenant_key_")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tenant credentials"})
		return
	}
	agentToken, err := randomToken("agent_")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tenant credentials"})
		return
	}

	newTenant := models.Tenant{
		Name:       req.Name,
		Subdomain:  req.Subdomain,
		DbName:     dbName,
		State:      "creating",
		ApiKey:     apiKey,
		AgentToken: agentToken,
		CreateDate: &now,
		WriteDate:  &now,
	}

	if err := database.DB.Create(&newTenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant record"})
		return
	}

	services.BroadcastTenantUpdate(newTenant)
	c.JSON(http.StatusCreated, newTenant)

	go createTenantInBackground(newTenant)
}

func randomToken(prefix string) (string, error) {
	value := make([]byte, 24)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(value), nil
}

func createTenantInBackground(tenant models.Tenant) {
	services.UpdateLogAndBroadcast(tenant.ID, "Starting tenant creation process...\n")
	dbHost := services.GetEnv("POSTGRES_HOST", "localhost")
	dbPort := services.GetEnv("POSTGRES_PORT", "5432")
	dbUser := services.GetEnv("POSTGRES_USER", "odoo")
	dbPassword := services.GetEnv("POSTGRES_PASSWORD", "odoo")

	services.UpdateLogAndBroadcast(tenant.ID, fmt.Sprintf("Attempting to create database '%s'...\n", tenant.DbName))
	createDbSQL := fmt.Sprintf("CREATE DATABASE \"%s\" OWNER \"%s\"", tenant.DbName, dbUser)
	if err := database.DB.Exec(createDbSQL).Error; err != nil {
		services.UpdateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to create database: %v\n", err))
		services.UpdateTenantState(tenant.ID, "error")
		return
	}
	services.UpdateLogAndBroadcast(tenant.ID, "Database created successfully.\n")

	services.UpdateLogAndBroadcast(tenant.ID, "Preparing isolated custom addons directory...\n")
	customAddonsPath := filepath.Join("/app/tenants", tenant.Subdomain, "custom_addons")
	if err := os.MkdirAll(customAddonsPath, 0755); err != nil {
		services.UpdateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to create tenant addon directory: %v\n", err))
		services.UpdateTenantState(tenant.ID, "error")
		return
	}

	hostProjectPath := services.GetEnv("HOST_PROJECT_PATH", "/home/nahom/Desktop/project/odoo_project")
	hostVolumePath := filepath.Join(hostProjectPath, "tenants", tenant.Subdomain, "custom_addons")

	err := services.DockerCreateAndStartInitContainer(tenant, dbHost, dbPort, dbUser, dbPassword, hostVolumePath)
	if err != nil {
		services.UpdateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to initialize database: %v\n", err))
		services.UpdateTenantState(tenant.ID, "error")
		return
	}
	services.UpdateLogAndBroadcast(tenant.ID, "Odoo database initialized successfully.\n")

	err = services.DockerCreateAndStartDaemonContainer(tenant, dbHost, dbPort, dbUser, dbPassword, hostVolumePath)
	if err != nil {
		services.UpdateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to start tenant daemon container: %v\n", err))
		services.UpdateTenantState(tenant.ID, "error")
		return
	}

	services.UpdateLogAndBroadcast(tenant.ID, "Tenant creation process finished.\n")
	services.UpdateTenantState(tenant.ID, "active")
}
