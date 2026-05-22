package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"saas-superadmin-backend/database"
	"saas-superadmin-backend/models"
)

func HandleOdooSaleWebhook(c *gin.Context) {
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

	var tenant models.Tenant
	if err := database.DB.Where("db_name = ?", req.DbName).First(&tenant).Error; err != nil {
		log.Printf("Webhook error: Tenant with db_name %s not found", req.DbName)
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if req.AmountTotal >= tenant.MinNotificationAmount {
		var devices []models.MobileDevice
		database.DB.Where("tenant_id = ?", tenant.ID).Find(&devices)
		
		title := "Large Sale Alert!"
		body := fmt.Sprintf("Order %s for %.2f Birr by %s", req.OrderName, req.AmountTotal, req.CustomerName)
		
		for _, device := range devices {
			log.Printf("[PUSH NOTIFICATION DUMMY] Sending to token %s (Platform: %s): %s - %s", device.DeviceToken, device.Platform, title, body)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed"})
}
