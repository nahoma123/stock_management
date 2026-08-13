package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"saas-superadmin-backend/database"
	"saas-superadmin-backend/handlers"
	"saas-superadmin-backend/websocket"
)

func main() {
	database.InitDB()

	websocket.AppHub = websocket.NewHub()
	go websocket.AppHub.Run()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	r.GET("/ws", func(c *gin.Context) { websocket.ServeWs(websocket.AppHub, c.Writer, c.Request) })

	api := r.Group("/api")
	{
		api.GET("/tenants", handlers.GetTenants)
		api.POST("/tenants", handlers.CreateTenant)
		api.POST("/tenants/:id/disable", handlers.DisableTenant)
		api.POST("/tenants/:id/enable", handlers.EnableTenant)
		api.PUT("/tenants/:id/expiry", handlers.SetTenantExpiry)
		api.DELETE("/tenants/:id", handlers.DeleteTenant)
		api.GET("/tenants/:id/monitoring", handlers.GetTenantMonitoring)
		api.POST("/tenants/:id/customizations/validate", handlers.ValidateTenantCustomization)
		api.POST("/tenants/:id/customizations/deploy", handlers.DeployTenantCustomization)

		api.GET("/mobile/tenants", handlers.GetTenantsList)
		api.POST("/mobile/login", handlers.MobileLogin)
		api.GET("/mobile/stats", handlers.GetMobileStats)
		api.POST("/mobile/devices", handlers.RegisterDevice)
		api.DELETE("/mobile/devices/:token", handlers.UnregisterDevice)
		api.PUT("/mobile/settings", handlers.UpdateMobileSettings)

		api.POST("/webhooks/odoo/sale", handlers.HandleOdooSaleWebhook)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
