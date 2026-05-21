package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	ApiKey                string     `json:"api_key" gorm:"column:api_key"`
	CreationLog           *string    `json:"creation_log"`
	LicenseExpiryDate     *time.Time `json:"license_expiry_date"`
	MinNotificationAmount float64    `json:"min_notification_amount"`
	CreateDate            *time.Time `json:"create_date" gorm:"column:create_date"`
	WriteDate         *time.Time `json:"write_date" gorm:"column:write_date"`
}

func (Tenant) TableName() string {
	return "saas_tenant"
}

type MobileDevice struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	TenantID    int       `json:"tenant_id" gorm:"index"`
	DeviceToken string    `json:"device_token" gorm:"uniqueIndex"`
	Platform    string    `json:"platform"`
	CreatedAt   time.Time `json:"created_at"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Docker API JSON payload structs
type RestartPolicy struct {
	Name string `json:"Name"`
}

type HostConfig struct {
	NetworkMode   string        `json:"NetworkMode"`
	Binds         []string      `json:"Binds"`
	RestartPolicy RestartPolicy `json:"RestartPolicy,omitempty"`
}

type ContainerConfig struct {
	Image      string            `json:"Image"`
	Cmd        []string          `json:"Cmd"`
	Env        []string          `json:"Env"`
	HostConfig HostConfig        `json:"HostConfig"`
	Labels     map[string]string `json:"Labels"`
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

	// Auto-migrate models
	log.Println("Running AutoMigration...")
	if err := db.AutoMigrate(&Tenant{}, &MobileDevice{}); err != nil {
		log.Fatal("Failed to auto-migrate models:", err)
	}

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
		api.DELETE("/tenants/:id", deleteTenant)
		
		api.GET("/mobile/stats", getMobileStats)
		api.POST("/mobile/devices", registerDevice)
		api.DELETE("/mobile/devices/:token", unregisterDevice)
		api.PUT("/mobile/settings", updateMobileSettings)
		
		api.POST("/webhooks/odoo/sale", handleOdooSaleWebhook)
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
	containerName := fmt.Sprintf("odoo_tenant_%d", id)

	updateLogAndBroadcast(id, "Disabling tenant container...\n")
	_, status, err := callDockerAPI("POST", fmt.Sprintf("/containers/%s/stop?t=10", containerName), nil)
	if err != nil || status >= 400 {
		log.Printf("Error stopping container %s: %v (status %d)", containerName, err, status)
		updateLogAndBroadcast(id, "WARNING: Failed to stop tenant container.\n")
	} else {
		updateLogAndBroadcast(id, "Tenant container stopped.\n")
	}

	updateTenantState(id, "disabled")
	c.Status(http.StatusOK)
}

func enableTenant(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tenant, err := getTenantByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	containerName := fmt.Sprintf("odoo_tenant_%d", id)

	updateLogAndBroadcast(id, "Enabling tenant container...\n")
	
	// Ensure the dummy module exists so the addons-path validation doesn't crash Odoo
	if err := ensureDummyModuleExists(tenant.Subdomain); err != nil {
		log.Printf("Error ensuring dummy module exists: %v", err)
	}

	_, status, err := callDockerAPI("POST", fmt.Sprintf("/containers/%s/start", containerName), nil)
	if err != nil || status >= 400 {
		log.Printf("Error starting container %s: %v (status %d)", containerName, err, status)
		updateLogAndBroadcast(id, "WARNING: Failed to start tenant container.\n")
	} else {
		updateLogAndBroadcast(id, "Tenant container started.\n")
	}

	updateTenantState(id, "active")
	c.Status(http.StatusOK)
}

func deleteTenant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := getTenantByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Stop and remove the Docker container
	containerName := fmt.Sprintf("odoo_tenant_%d", id)
	log.Printf("Removing Docker container %s for tenant %d", containerName, id)
	_, status, err := callDockerAPI("DELETE", fmt.Sprintf("/containers/%s?force=true", containerName), nil)
	if err != nil || status >= 400 {
		log.Printf("Error removing container %s: %v (status %d)", containerName, err, status)
	}

	// Delete tenant directories
	tenantDir := filepath.Join("/app/tenants", tenant.Subdomain)
	log.Printf("Deleting directory %s for tenant %d", tenantDir, id)
	if err := os.RemoveAll(tenantDir); err != nil {
		log.Printf("Error deleting tenant directory %s: %v", tenantDir, err)
	}

	// Drop the database
	dropDbSQL := fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", tenant.DbName)
	if err := db.Exec(dropDbSQL).Error; err != nil {
		log.Printf("Error dropping database for tenant %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to drop tenant database"})
		return
	}

	// Delete the tenant record
	if err := db.Delete(&tenant).Error; err != nil {
		log.Printf("Error deleting tenant %d record: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant record"})
		return
	}

	msg := WebSocketMessage{Type: "tenant_deleted", Payload: gin.H{"id": id}}
	if jsonMsg, err := json.Marshal(msg); err == nil {
		hub.broadcast <- jsonMsg
	}

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
	
	// Generate unique API Key for the mobile stats app
	apiKey := "tenant_key_" + generateRandomString(24)

	newTenant := Tenant{
		Name:       req.Name,
		Subdomain:  req.Subdomain,
		DbName:     dbName,
		State:      "creating",
		ApiKey:     apiKey,
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
	
	// 1. Create the tenant database
	updateLogAndBroadcast(tenant.ID, fmt.Sprintf("Attempting to create database '%s'...\n", tenant.DbName))
	createDbSQL := fmt.Sprintf("CREATE DATABASE \"%s\" OWNER \"%s\"", tenant.DbName, dbUser)
	if err := db.Exec(createDbSQL).Error; err != nil {
		updateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to create database: %v\n", err))
		updateTenantState(tenant.ID, "error")
		return
	}
	updateLogAndBroadcast(tenant.ID, "Database created successfully.\n")

	// 2. Prepare isolated tenant code directory
	updateLogAndBroadcast(tenant.ID, "Preparing isolated custom addons directory...\n")
	if err := ensureDummyModuleExists(tenant.Subdomain); err != nil {
		updateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to create tenant directories or dummy module: %v\n", err))
		updateTenantState(tenant.ID, "error")
		return
	}

	// 3. Setup host path and dynamic network configuration
	hostProjectPath := getEnv("HOST_PROJECT_PATH", "/home/nahom/Desktop/project/odoo_project")
	hostVolumePath := filepath.Join(hostProjectPath, "tenants", tenant.Subdomain, "custom_addons")

	// Run transient init container to populate database tables
	err := dockerCreateAndStartInitContainer(tenant, dbHost, dbPort, dbUser, dbPassword, hostVolumePath)
	if err != nil {
		updateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to initialize database: %v\n", err))
		updateTenantState(tenant.ID, "error")
		return
	}
	updateLogAndBroadcast(tenant.ID, "Odoo database initialized successfully.\n")

	// 4. Start the tenant container daemon
	err = dockerCreateAndStartDaemonContainer(tenant, dbHost, dbPort, dbUser, dbPassword, hostVolumePath)
	if err != nil {
		updateLogAndBroadcast(tenant.ID, fmt.Sprintf("ERROR: Failed to start tenant daemon container: %v\n", err))
		updateTenantState(tenant.ID, "error")
		return
	}

	updateLogAndBroadcast(tenant.ID, "Tenant creation process finished.\n")
	updateTenantState(tenant.ID, "active")
}

func getMobileStats(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
		return
	}

	var tenant Tenant
	if err := db.Where("api_key = ?", apiKey).First(&tenant).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
		return
	}

	tenantDB, err := getTenantDBConnection(tenant.DbName)
	if err != nil {
		log.Printf("Error connecting to tenant DB %s: %v", tenant.DbName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not connect to tenant database"})
		return
	}

	sqlDB, err := tenantDB.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// Graceful handling if Odoo tables are not initialized yet
	if !tenantDB.Migrator().HasTable("sale_order") {
		c.JSON(http.StatusOK, gin.H{
			"status":      "initializing",
			"message":     "Tenant database is being initialized. Sales stats are not available yet.",
			"tenant_name": tenant.Name,
			"subdomain":   tenant.Subdomain,
		})
		return
	}

	var stats struct {
		TenantName        string  `json:"tenant_name"`
		Subdomain         string  `json:"subdomain"`
		State             string  `json:"state"`
		LicenseExpiryDate string  `json:"license_expiry_date"`
		TotalSalesAllTime float64 `json:"total_sales_all_time"`
		OrderCountAllTime int     `json:"order_count_all_time"`
		TotalSalesToday     float64 `json:"total_sales_today"`
		OrderCountToday     int     `json:"order_count_today"`
		TotalSalesYesterday float64 `json:"total_sales_yesterday"`
		OrderCountYesterday int     `json:"order_count_yesterday"`
		TotalSalesMonth     float64 `json:"total_sales_month"`
		OrderCountMonth     int     `json:"order_count_month"`
	}

	stats.TenantName = tenant.Name
	stats.Subdomain = tenant.Subdomain
	stats.State = tenant.State
	if tenant.LicenseExpiryDate != nil {
		stats.LicenseExpiryDate = tenant.LicenseExpiryDate.Format("2006-01-02")
	} else {
		stats.LicenseExpiryDate = "N/A"
	}

	// Query metrics
	tenantDB.Raw(`
		SELECT COALESCE(SUM(amount_total), 0) AS total, COUNT(id) AS cnt 
		FROM sale_order 
		WHERE state IN ('sale', 'done')
	`).Row().Scan(&stats.TotalSalesAllTime, &stats.OrderCountAllTime)

	tenantDB.Raw(`
		SELECT COALESCE(SUM(amount_total), 0) AS total, COUNT(id) AS cnt 
		FROM sale_order 
		WHERE state IN ('sale', 'done') AND date_order >= CURRENT_DATE
	`).Row().Scan(&stats.TotalSalesToday, &stats.OrderCountToday)

	tenantDB.Raw(`
		SELECT COALESCE(SUM(amount_total), 0) AS total, COUNT(id) AS cnt 
		FROM sale_order 
		WHERE state IN ('sale', 'done') AND date_order >= CURRENT_DATE - INTERVAL '1 day' AND date_order < CURRENT_DATE
	`).Row().Scan(&stats.TotalSalesYesterday, &stats.OrderCountYesterday)

	tenantDB.Raw(`
		SELECT COALESCE(SUM(amount_total), 0) AS total, COUNT(id) AS cnt 
		FROM sale_order 
		WHERE state IN ('sale', 'done') AND date_order >= DATE_TRUNC('month', CURRENT_DATE)
	`).Row().Scan(&stats.TotalSalesMonth, &stats.OrderCountMonth)

	type DailyTrend struct {
		Date  string  `json:"date"`
		Total float64 `json:"total_revenue"`
		Count int     `json:"order_count"`
	}
	var weeklyTrend []DailyTrend
	tenantDB.Raw(`
		SELECT TO_CHAR(DATE_TRUNC('day', date_order), 'YYYY-MM-DD') AS date,
		       COALESCE(SUM(amount_total), 0) AS total,
		       COUNT(id) AS count
		FROM sale_order
		WHERE state IN ('sale', 'done') AND date_order >= CURRENT_DATE - INTERVAL '6 days'
		GROUP BY DATE_TRUNC('day', date_order)
		ORDER BY date ASC
	`).Scan(&weeklyTrend)

	type RecentOrder struct {
		Name        string    `json:"name"`
		DateOrder   time.Time `json:"date_order"`
		AmountTotal float64   `json:"amount_total"`
		State       string    `json:"state"`
	}
	var recentOrders []RecentOrder
	tenantDB.Raw(`
		SELECT name, date_order, amount_total, state 
		FROM sale_order 
		ORDER BY date_order DESC 
		LIMIT 5
	`).Scan(&recentOrders)

	type TopProduct struct {
		Name  string  `json:"name"`
		Qty   float64 `json:"quantity"`
		Total float64 `json:"total_revenue"`
	}
	var topProducts []TopProduct
	// Safely joins to template and parses name field (which is JSONB translation dictionary in Odoo 17/18)
	tenantDB.Raw(`
		SELECT COALESCE(pt.name->>'en_US', COALESCE(pt.name::text, 'Unknown Product')) AS name, 
		       SUM(sol.product_uom_qty) AS qty, 
		       SUM(sol.price_total) AS total
		FROM sale_order_line sol
		JOIN sale_order so ON sol.order_id = so.id
		JOIN product_product pp ON sol.product_id = pp.id
		JOIN product_template pt ON pp.product_tmpl_id = pt.id
		WHERE so.state IN ('sale', 'done')
		GROUP BY pt.name, sol.product_id
		ORDER BY qty DESC
		LIMIT 5
	`).Scan(&topProducts)

	c.JSON(http.StatusOK, gin.H{
		"stats":         stats,
		"weekly_trend":  weeklyTrend,
		"recent_orders": recentOrders,
		"top_products":  topProducts,
	})
}

// --- Mobile App & Push Notification Handlers ---

func authenticateMobileAPI(c *gin.Context) (Tenant, bool) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
		return Tenant{}, false
	}

	var tenant Tenant
	if err := db.Where("api_key = ?", apiKey).First(&tenant).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
		return Tenant{}, false
	}
	return tenant, true
}

func registerDevice(c *gin.Context) {
	tenant, ok := authenticateMobileAPI(c)
	if !ok {
		return
	}

	var req struct {
		DeviceToken string `json:"device_token" binding:"required"`
		Platform    string `json:"platform" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := MobileDevice{
		TenantID:    tenant.ID,
		DeviceToken: req.DeviceToken,
		Platform:    req.Platform,
		CreatedAt:   time.Now(),
	}

	// Upsert to handle re-registrations
	if err := db.Where("device_token = ?", req.DeviceToken).Assign(device).FirstOrCreate(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device registered successfully"})
}

func unregisterDevice(c *gin.Context) {
	tenant, ok := authenticateMobileAPI(c)
	if !ok {
		return
	}

	token := c.Param("token")
	if err := db.Where("tenant_id = ? AND device_token = ?", tenant.ID, token).Delete(&MobileDevice{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unregister device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device unregistered successfully"})
}

func updateMobileSettings(c *gin.Context) {
	tenant, ok := authenticateMobileAPI(c)
	if !ok {
		return
	}

	var req struct {
		MinNotificationAmount float64 `json:"min_notification_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Tenant{}).Where("id = ?", tenant.ID).Update("min_notification_amount", req.MinNotificationAmount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully", "min_notification_amount": req.MinNotificationAmount})
}

func handleOdooSaleWebhook(c *gin.Context) {
	var req struct {
		DbName       string  `json:"db_name" binding:"required"`
		OrderID      int     `json:"order_id" binding:"required"`
		OrderName    string  `json:"order_name" binding:"required"`
		AmountTotal  float64 `json:"amount_total" binding:"required"`
		CustomerName string  `json:"customer_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tenant Tenant
	if err := db.Where("db_name = ?", req.DbName).First(&tenant).Error; err != nil {
		log.Printf("Webhook error: Tenant with db_name %s not found", req.DbName)
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Check if notification threshold is met
	if req.AmountTotal >= tenant.MinNotificationAmount {
		var devices []MobileDevice
		db.Where("tenant_id = ?", tenant.ID).Find(&devices)
		
		title := "Large Sale Alert!"
		body := fmt.Sprintf("Order %s for %.2f Birr by %s", req.OrderName, req.AmountTotal, req.CustomerName)
		
		for _, device := range devices {
			// Placeholder for sending actual push notification via FCM/APNs
			log.Printf("[PUSH NOTIFICATION DUMMY] Sending to token %s (Platform: %s): %s - %s", device.DeviceToken, device.Platform, title, body)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed"})
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

func getTenantDBConnection(dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_USER", "odoo"),
		getEnv("POSTGRES_PASSWORD", "odoo"),
		dbName,
		getEnv("POSTGRES_PORT", "5432"),
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func generateRandomString(n int) string {
	b := make([]byte, n/2)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Docker Engine API helpers over unix socket
func callDockerAPI(method, path string, payload interface{}) ([]byte, int, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	var reqBody io.Reader
	if payload != nil {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, "http://localhost"+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func dockerCreateAndStartInitContainer(tenant Tenant, dbHost, dbPort, dbUser, dbPassword string, hostVolumePath string) error {
	dockerNetwork := getEnv("DOCKER_NETWORK", "saas_net")
	containerName := fmt.Sprintf("odoo_init_%d", tenant.ID)

	// Clean up previous init container if it exists
	callDockerAPI("DELETE", fmt.Sprintf("/containers/%s?force=true", containerName), nil)

	config := ContainerConfig{
		Image: "odoo_project-odoo:latest",
		Cmd: []string{
			"odoo", "--config=/dev/null", "--database", tenant.DbName,
			"--db_host", dbHost, "--db_port", dbPort, "--db_user", dbUser, "--db_password", dbPassword,
			"--addons-path=/mnt/extra-addons,/usr/lib/python3/dist-packages/odoo/addons",
			"--init", "base,web,sale_management,daily_sales_report",
			"--stop-after-init",
		},
		HostConfig: HostConfig{
			NetworkMode: dockerNetwork,
			Binds:       []string{fmt.Sprintf("%s:/mnt/extra-addons", hostVolumePath)},
		},
	}

	updateLogAndBroadcast(tenant.ID, "Creating transient initialization container...\n")
	resp, status, err := callDockerAPI("POST", fmt.Sprintf("/containers/create?name=%s", containerName), config)
	if err != nil {
		return fmt.Errorf("failed to create init container: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("failed to create init container (status %d): %s", status, string(resp))
	}

	updateLogAndBroadcast(tenant.ID, "Starting initialization container...\n")
	resp, status, err = callDockerAPI("POST", fmt.Sprintf("/containers/%s/start", containerName), nil)
	if err != nil {
		return fmt.Errorf("failed to start init container: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("failed to start init container (status %d): %s", status, string(resp))
	}

	// Wait for container to exit
	updateLogAndBroadcast(tenant.ID, "Waiting for database initialization to complete (this may take 1-2 minutes)...\n")
	resp, status, err = callDockerAPI("POST", fmt.Sprintf("/containers/%s/wait", containerName), nil)
	if err != nil {
		return fmt.Errorf("failed to wait for init container: %w", err)
	}

	// Fetch logs and write to tenant logs
	logResp, _, _ := callDockerAPI("GET", fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true", containerName), nil)
	cleanLogs := parseDockerLogs(logResp)
	updateLogAndBroadcast(tenant.ID, cleanLogs)

	// Clean up transient container
	callDockerAPI("DELETE", fmt.Sprintf("/containers/%s?force=true", containerName), nil)

	return nil
}

func dockerCreateAndStartDaemonContainer(tenant Tenant, dbHost, dbPort, dbUser, dbPassword string, hostVolumePath string) error {
	dockerNetwork := getEnv("DOCKER_NETWORK", "saas_net")
	containerName := fmt.Sprintf("odoo_tenant_%d", tenant.ID)

	// Clean up previous daemon if exists
	callDockerAPI("DELETE", fmt.Sprintf("/containers/%s?force=true", containerName), nil)

	config := ContainerConfig{
		Image: "odoo_project-odoo:latest",
		Cmd: []string{
			"odoo", "-r", dbUser, "-w", dbPassword, "--db_host", dbHost, "--db_port", dbPort, "--database", tenant.DbName,
			"--addons-path=/mnt/extra-addons,/usr/lib/python3/dist-packages/odoo/addons",
		},
		Env: []string{
			fmt.Sprintf("DB_HOST=%s", dbHost),
			fmt.Sprintf("DB_PORT=%s", dbPort),
			fmt.Sprintf("DB_USER=%s", dbUser),
			fmt.Sprintf("DB_PASSWORD=%s", dbPassword),
		},
		HostConfig: HostConfig{
			NetworkMode: dockerNetwork,
			Binds:       []string{fmt.Sprintf("%s:/mnt/extra-addons", hostVolumePath)},
			RestartPolicy: RestartPolicy{
				Name: "unless-stopped",
			},
		},
		Labels: map[string]string{
			"traefik.enable":                                                  "true",
			fmt.Sprintf("traefik.http.routers.%s.rule", containerName):        fmt.Sprintf("Host(`%s.localhost`)", tenant.Subdomain),
			fmt.Sprintf("traefik.http.routers.%s.entrypoints", containerName): "web",
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", containerName): "8069",
		},
	}

	updateLogAndBroadcast(tenant.ID, "Creating tenant container daemon...\n")
	resp, status, err := callDockerAPI("POST", fmt.Sprintf("/containers/create?name=%s", containerName), config)
	if err != nil {
		return fmt.Errorf("failed to create daemon container: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("failed to create daemon container (status %d): %s", status, string(resp))
	}

	updateLogAndBroadcast(tenant.ID, "Starting tenant container daemon...\n")
	resp, status, err = callDockerAPI("POST", fmt.Sprintf("/containers/%s/start", containerName), nil)
	if err != nil {
		return fmt.Errorf("failed to start daemon container: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("failed to start daemon container (status %d): %s", status, string(resp))
	}

	return nil
}

func parseDockerLogs(raw []byte) string {
	var builder strings.Builder
	i := 0
	for i < len(raw) {
		if i+8 > len(raw) {
			break
		}
		// Docker stream frame: 1 byte stream type, 3 bytes padding, 4 bytes big-endian payload size
		size := int(raw[i+4])<<24 | int(raw[i+5])<<16 | int(raw[i+6])<<8 | int(raw[i+7])
		i += 8
		if i+size > len(raw) {
			builder.Write(raw[i:])
			break
		}
		builder.Write(raw[i : i+size])
		i += size
	}
	return builder.String()
}

func ensureDummyModuleExists(subdomain string) error {
	tenantPath := filepath.Join("/app/tenants", subdomain, "custom_addons")
	if err := os.MkdirAll(tenantPath, 0755); err != nil {
		return err
	}
	
	dummyModulePath := filepath.Join(tenantPath, "dummy_module")
	if err := os.MkdirAll(dummyModulePath, 0755); err != nil {
		return err
	}

	initPyPath := filepath.Join(dummyModulePath, "__init__.py")
	if _, err := os.Stat(initPyPath); os.IsNotExist(err) {
		if err := os.WriteFile(initPyPath, []byte(""), 0644); err != nil {
			return err
		}
	}

	manifestPyPath := filepath.Join(dummyModulePath, "__manifest__.py")
	if _, err := os.Stat(manifestPyPath); os.IsNotExist(err) {
		manifestContent := `{
    'name': 'Dummy Module',
    'version': '1.0',
    'category': 'Hidden',
    'summary': 'Dummy module for path validation',
    'depends': ['base'],
    'installable': True,
}`
		if err := os.WriteFile(manifestPyPath, []byte(manifestContent), 0644); err != nil {
			return err
		}
	}
	return nil
}
