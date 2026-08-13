package services

import (
	"log"

	"saas-superadmin-backend/database"
	"saas-superadmin-backend/models"
)

func RecordAudit(tenantID int, action, resource, status, detail string) {
	event := models.PlatformAuditEvent{
		TenantID: tenantID,
		Action:   action,
		Resource: resource,
		Status:   status,
		Detail:   detail,
	}
	if err := database.DB.Create(&event).Error; err != nil {
		log.Printf("failed to record audit event for tenant %d: %v", tenantID, err)
	}
}
