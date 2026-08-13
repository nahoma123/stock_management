package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"saas-superadmin-backend/database"
	"saas-superadmin-backend/models"
)

func GetMobileStats(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
		return
	}

	var tenant models.Tenant
	if err := database.DB.Where("api_key = ?", apiKey).First(&tenant).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
		return
	}

	tenantDB, err := database.GetTenantDBConnection(tenant.DbName)
	if err != nil {
		log.Printf("Error connecting to tenant DB %s: %v", tenant.DbName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not connect to tenant database"})
		return
	}

	sqlDB, err := tenantDB.DB()
	if err == nil {
		defer sqlDB.Close()
	}

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
		TenantName          string  `json:"tenant_name"`
		Subdomain           string  `json:"subdomain"`
		State               string  `json:"state"`
		LicenseExpiryDate   string  `json:"license_expiry_date"`
		TotalSalesAllTime   float64 `json:"total_sales_all_time"`
		OrderCountAllTime   int     `json:"order_count_all_time"`
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
func authenticateMobileAPI(c *gin.Context) (models.Tenant, bool) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
		return models.Tenant{}, false
	}

	var tenant models.Tenant
	if err := database.DB.Where("api_key = ?", apiKey).First(&tenant).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
		return models.Tenant{}, false
	}
	return tenant, true
}

func RegisterDevice(c *gin.Context) {
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

	device := models.MobileDevice{
		TenantID:    tenant.ID,
		DeviceToken: req.DeviceToken,
		Platform:    req.Platform,
		CreatedAt:   time.Now(),
	}

	if err := database.DB.Where("device_token = ?", req.DeviceToken).Assign(device).FirstOrCreate(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device registered successfully"})
}

func UnregisterDevice(c *gin.Context) {
	tenant, ok := authenticateMobileAPI(c)
	if !ok {
		return
	}

	token := c.Param("token")
	if err := database.DB.Where("tenant_id = ? AND device_token = ?", tenant.ID, token).Delete(&models.MobileDevice{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unregister device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device unregistered successfully"})
}

func UpdateMobileSettings(c *gin.Context) {
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

	if err := database.DB.Model(&models.Tenant{}).Where("id = ?", tenant.ID).Update("min_notification_amount", req.MinNotificationAmount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully", "min_notification_amount": req.MinNotificationAmount})
}

// GetTenantsList returns a list of active tenants for the mobile app dropdown
func GetTenantsList(c *gin.Context) {
	var tenants []models.Tenant
	// Only return basic info for active tenants
	if err := database.DB.Select("id", "name", "subdomain").Where("state = ?", "active").Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
		return
	}

	var response []map[string]interface{}
	for _, t := range tenants {
		response = append(response, map[string]interface{}{
			"id":        t.ID,
			"name":      t.Name,
			"subdomain": t.Subdomain,
		})
	}

	c.JSON(http.StatusOK, response)
}

// MobileLogin authenticates an Odoo user and returns the Tenant's API Key
func MobileLogin(c *gin.Context) {
	var req struct {
		TenantID int    `json:"tenant_id" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tenant models.Tenant
	if err := database.DB.Where("id = ?", req.TenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if tenant.State != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tenant is not active"})
		return
	}

	// Prepare JSON-RPC request to Odoo container
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"params": map[string]interface{}{
			"db":       tenant.DbName,
			"login":    req.Email,
			"password": req.Password,
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	// Since we are running in docker, the odoo container is accessible at http://odoo:8069
	resp, err := http.Post("http://odoo:8069/web/session/authenticate", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to authentication server"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var odooResp struct {
		Result *struct {
			UID int `json:"uid"`
		} `json:"result"`
		Error *interface{} `json:"error"`
	}

	if err := json.Unmarshal(body, &odooResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from authentication server"})
		return
	}

	// If Result is nil or Error is not nil, auth failed
	if odooResp.Error != nil || odooResp.Result == nil || odooResp.Result.UID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Auth successful, return the API Key so the mobile app can continue using existing endpoints
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"api_key": tenant.ApiKey,
	})
}
