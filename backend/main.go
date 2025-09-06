package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Tenant struct {
	ID                int        `json:"id" gorm:"primaryKey"`
	Name              string     `json:"name"`
	Subdomain         string     `json:"subdomain" gorm:"unique"`
	DbName            string     `json:"db_name" gorm:"column:db_name;unique"`
	State             string     `json:"state"`
	CreationLog       *string    `json:"creation_log"`
	LicenseExpiryDate *time.Time `json:"license_expiry_date"`
	CreateDate        *time.Time `json:"create_date" gorm:"column:create_date"`
	WriteDate         *time.Time `json:"write_date" gorm:"column:write_date"`
}

func (Tenant) TableName() string {
	return "saas_tenant"
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

var db *gorm.DB
var hub *Hub

func main() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_USER", "odoo"),
		getEnv("POSTGRES_PASSWORD", "odoo"),
		getEnv("POSTGRES_DB", "odoo"),
		getEnv("POSTGRES_PORT", "5432"),
	)
	log.Println("Connecting to database with DSN:", dsn)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	for i := 0; i < 10; i++ {
		err = sqlDB.Ping()
		if err == nil {
			break
		}
		log.Println("Database connection test failed, retrying...", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Database connection test failed:", err)
	}
	log.Println("Successfully connected to the database.")

	hub = newHub()
	go hub.run()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	r.GET("/ws", func(c *gin.Context) { serveWs(hub, c.Writer, c.Request) })

	api := r.Group("/api")
	{
		api.GET("/tenants", getTenants)
		api.POST("/tenants", createTenant)
		api.POST("/tenants/:id/disable", disableTenant)
		api.POST("/tenants/:id/enable", enableTenant)
		api.PUT("/tenants/:id/expiry", setTenantExpiry)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

// --- API Handlers ---

func getTenants(c *gin.Context) {
	var tenants []Tenant
	if err := db.Order("id desc").Find(&tenants).Error; err != nil {
		log.Printf("Error getting tenants: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tenants"})
		return
	}
	c.JSON(http.StatusOK, tenants)
}

func disableTenant(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	updateTenantState(id, "disabled")
	c.Status(http.StatusOK)
}

func enableTenant(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	updateTenantState(id, "active")
	c.Status(http.StatusOK)
}

type SetExpiryRequest struct {
	ExpiryDate string `json:"expiry_date" binding:"required"` // Expecting "YYYY-MM-DD"
}

func setTenantExpiry(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req SetExpiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := db.Model(&Tenant{}).Where("id = ?", id).Update("license_expiry_date", req.ExpiryDate).Error; err != nil {
		log.Printf("Error updating expiry for tenant %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expiry date"})
		return
	}

	if tenant, err := getTenantByID(id); err == nil {
		broadcastTenantUpdate(tenant)
	}
	c.Status(http.StatusOK)
}

type CreateTenantRequest struct {
	Name      string `json:"name" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
}

func createTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbName := strings.ToLower(strings.ReplaceAll(req.Subdomain, ".", "_")) + "_db"
	now := time.Now()
	newTenant := Tenant{
		Name:       req.Name,
		Subdomain:  req.Subdomain,
		DbName:     dbName,
		State:      "creating",
		CreateDate: &now,
		WriteDate:  &now,
	}

	if err := db.Create(&newTenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant record"})
		return
	}

	broadcastTenantUpdate(newTenant)
	c.JSON(http.StatusCreated, newTenant)

	go createTenantInBackground(newTenant)
}

func createTenantInBackground(tenant Tenant) {
	updateLogAndBroadcast(tenant.ID, "Starting tenant creation process...\n")
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "odoo")
	dbPassword := getEnv("POSTGRES_PASSWORD", "odoo")
	updateLogAndBroadcast(tenant.ID, fmt.Sprintf("Attempting to create database '%s'...\n", tenant.DbName))
	cmdCreate := exec.Command("createdb", "--host", dbHost, "--port", dbPort, "--username", dbUser, "--owner", dbUser, tenant.DbName)
	cmdCreate.Env = append(os.Environ(), "PGPASSWORD="+dbPassword)
	if err := runCommandAndLog(cmdCreate, tenant.ID); err != nil {
		updateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to create database: %v\n", err))
		updateTenantState(tenant.ID, "error")
		return
	}
	updateLogAndBroadcast(tenant.ID, "Database created successfully.\n")
	updateLogAndBroadcast(tenant.ID, "Initializing Odoo modules...\n")
	modules := "base,web,shopping_portal"
	addonsPath := "/mnt/extra-addons"
	cmdInit := exec.Command("odoo", "--config=/dev/null", "--database", tenant.DbName, "--db_host", dbHost, "--db_port", dbPort, "--db_user", dbUser, "--db_password", dbPassword, "--addons-path", addonsPath, "--init", modules, "--stop-after-init", "--log-level=info")
	if err := runCommandAndLog(cmdInit, tenant.ID); err != nil {
		updateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to initialize Odoo: %v\n", err))
		updateTenantState(tenant.ID, "error")
		return
	}
	updateLogAndBroadcast(tenant.ID, "Odoo initialized successfully.\n")
	updateLogAndBroadcast(tenant.ID, "Tenant creation process finished.\n")
	updateTenantState(tenant.ID, "active")
}

// --- Helper Functions ---

func getTenantByID(id int) (Tenant, error) {
	var tenant Tenant
	err := db.First(&tenant, id).Error
	return tenant, err
}

func updateTenantState(id int, state string) {
	if err := db.Model(&Tenant{}).Where("id = ?", id).Update("state", state).Error; err != nil {
		log.Printf("Error updating tenant %d state to %s: %v", id, state, err)
		return
	}
	if tenant, err := getTenantByID(id); err == nil {
		broadcastTenantUpdate(tenant)
	}
}

func updateLogAndBroadcast(id int, logMessage string) {
	// Using Raw SQL for COALESCE function
	err := db.Exec("UPDATE saas_tenant SET creation_log = COALESCE(creation_log, '') || ? WHERE id = ?", logMessage, id).Error
	if err != nil {
		log.Printf("Error updating creation log for tenant %d: %v", id, err)
	}
	msg := WebSocketMessage{Type: "tenant_log", Payload: gin.H{"tenant_id": id, "log": logMessage}}
	if jsonMsg, err := json.Marshal(msg); err == nil {
		hub.broadcast <- jsonMsg
	}
}

func broadcastTenantUpdate(tenant Tenant) {
	msg := WebSocketMessage{Type: "tenant_updated", Payload: tenant}
	if jsonMsg, err := json.Marshal(msg); err == nil {
		hub.broadcast <- jsonMsg
	}
}

func runCommandAndLog(cmd *exec.Cmd, tenantID int) error {
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()
    cmd.Start()
    go func() {
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() { updateLogAndBroadcast(tenantID, scanner.Text()+"\n") }
    }()
    go func() {
        scanner := bufio.NewScanner(stderr)
        for scanner.Scan() { updateLogAndBroadcast(tenantID, "[STDERR] "+scanner.Text()+"\n") }
    }()
    return cmd.Wait()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key);
	ok {
		return value
	}
	return fallback
}
