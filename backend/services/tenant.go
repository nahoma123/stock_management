package services

import (
	"encoding/json"
	"log"
	
	"saas-superadmin-backend/database"
	"saas-superadmin-backend/models"
	"saas-superadmin-backend/websocket"
)

func GetTenantByID(id int) (models.Tenant, error) {
	var tenant models.Tenant
	err := database.DB.First(&tenant, id).Error
	return tenant, err
}

func UpdateTenantState(id int, state string) {
	if err := database.DB.Model(&models.Tenant{}).Where("id = ?", id).Update("state", state).Error; err != nil {
		log.Printf("Error updating tenant %d state to %s: %v", id, state, err)
		return
	}
	if tenant, err := GetTenantByID(id); err == nil {
		BroadcastTenantUpdate(tenant)
	}
}

func UpdateLogAndBroadcast(id int, logMessage string) {
	err := database.DB.Exec("UPDATE saas_tenant SET creation_log = COALESCE(creation_log, '') || ? WHERE id = ?", logMessage, id).Error
	if err != nil {
		log.Printf("Error updating creation log for tenant %d: %v", id, err)
	}
	msg := websocket.WebSocketMessage{Type: "tenant_log", Payload: map[string]interface{}{"tenant_id": id, "log": logMessage}}
	if jsonMsg, err := json.Marshal(msg); err == nil {
		websocket.AppHub.Broadcast <- jsonMsg
	}
}

func BroadcastTenantUpdate(tenant models.Tenant) {
	msg := websocket.WebSocketMessage{Type: "tenant_updated", Payload: tenant}
	if jsonMsg, err := json.Marshal(msg); err == nil {
		websocket.AppHub.Broadcast <- jsonMsg
	}
}
