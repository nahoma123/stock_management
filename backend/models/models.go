package models

import "time"

type Tenant struct {
	ID                    int        `json:"id" gorm:"primaryKey"`
	Name                  string     `json:"name"`
	Subdomain             string     `json:"subdomain" gorm:"unique"`
	DbName                string     `json:"db_name" gorm:"column:db_name;unique"`
	State                 string     `json:"state"`
	ApiKey                string     `json:"api_key" gorm:"column:api_key"`
	CreationLog           *string    `json:"creation_log"`
	LicenseExpiryDate     *time.Time `json:"license_expiry_date"`
	MinNotificationAmount float64    `json:"min_notification_amount"`
	CreateDate            *time.Time `json:"create_date" gorm:"column:create_date"`
	WriteDate             *time.Time `json:"write_date" gorm:"column:write_date"`
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
