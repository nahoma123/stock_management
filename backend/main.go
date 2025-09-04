package main

import (
	"bufio"
	"database/sql"
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
	_ "github.com/lib/pq"
)

// ... (Tenant and WebSocketMessage structs remain the same)
type Tenant struct {
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Subdomain         string     `json:"subdomain"`
	DbName            string     `json:"db_name"`
	State             string     `json:"state"`
	CreationLog       *string    `json:"creation_log"`
	LicenseExpiryDate *time.Time `json:"license_expiry_date"`
	CreateDate        time.Time  `json:"create_date"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

var db *sql.DB
var hub *Hub

func main() {
	// ... (database connection and hub setup remains the same)
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_USER", "odoo"),
		getEnv("POSTGRES_PASSWORD", "odoo"),
		getEnv("POSTGRES_DB", "odoo"),
	)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil { log.Fatal("Failed to connect to database:", err) }
	defer db.Close()
	err = db.Ping()
	if err != nil { log.Fatal("Database connection test failed:", err) }
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
	rows, err := db.Query("SELECT id, name, subdomain, db_name, state, creation_log, license_expiry_date, create_date FROM saas_tenant ORDER BY id DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tenants"})
		return
	}
	defer rows.Close()

	tenants := []Tenant{}
	for rows.Next() {
		t, err := scanTenant(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process tenant data"})
			return
		}
		tenants = append(tenants, t)
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

	_, err := db.Exec("UPDATE saas_tenant SET license_expiry_date = $1, write_date = NOW() WHERE id = $2", req.ExpiryDate, id)
	if err != nil {
		log.Printf("Error updating expiry for tenant %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expiry date"})
		return
	}

	if tenant, err := getTenantByID(id); err == nil {
		broadcastTenantUpdate(tenant)
	}
	c.Status(http.StatusOK)
}

// ... (createTenant and background worker remain the same)
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
	var newTenantID int
	err := db.QueryRow(
		"INSERT INTO saas_tenant (name, subdomain, db_name, state, create_date, write_date) VALUES ($1, $2, $3, 'creating', NOW(), NOW()) RETURNING id",
		req.Name, req.Subdomain, dbName,
	).Scan(&newTenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant record"})
		return
	}
	newTenant, err := getTenantByID(newTenantID)
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{"message": "Tenant creation started"})
	} else {
		broadcastTenantUpdate(newTenant)
		c.JSON(http.StatusCreated, newTenant)
	}
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

type scannable interface {
    Scan(dest ...interface{}) error
}

func scanTenant(row scannable) (Tenant, error) {
    var t Tenant
    var creationLog sql.NullString
    var licenseExpiryDate sql.NullTime
    if err := row.Scan(&t.ID, &t.Name, &t.Subdomain, &t.DbName, &t.State, &creationLog, &licenseExpiryDate, &t.CreateDate); err != nil {
        return t, err
    }
    if creationLog.Valid {
        t.CreationLog = &creationLog.String
    }
    if licenseExpiryDate.Valid {
        t.LicenseExpiryDate = &licenseExpiryDate.Time
		// Dynamic 'expired' state
		if t.State == "active" && time.Now().After(*t.LicenseExpiryDate) {
			t.State = "expired"
		}
    }
    return t, nil
}

func getTenantByID(id int) (Tenant, error) {
	row := db.QueryRow("SELECT id, name, subdomain, db_name, state, creation_log, license_expiry_date, create_date FROM saas_tenant WHERE id = $1", id)
	return scanTenant(row)
}

// ... (other helpers remain the same)
func updateTenantState(id int, state string) {
	_, err := db.Exec("UPDATE saas_tenant SET state = $1, write_date = NOW() WHERE id = $2", state, id)
	if err != nil {
		log.Printf("Error updating tenant %d state to %s: %v", id, state, err)
		return
	}
	if tenant, err := getTenantByID(id); err == nil {
		broadcastTenantUpdate(tenant)
	}
}
func updateLogAndBroadcast(id int, logMessage string) {
	_, err := db.Exec("UPDATE saas_tenant SET creation_log = COALESCE(creation_log, '') || $1 WHERE id = $2", logMessage, id)
	if err != nil { log.Printf("Error updating creation log for tenant %d: %v", id, err) }
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
	if value, ok := os.LookupEnv(key); ok { return value }
	return fallback
}
