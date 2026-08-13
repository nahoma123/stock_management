package models

import "time"

type Tenant struct {
	ID                    int        `json:"id" gorm:"primaryKey"`
	Name                  string     `json:"name"`
	Subdomain             string     `json:"subdomain" gorm:"unique"`
	DbName                string     `json:"db_name" gorm:"column:db_name;unique"`
	State                 string     `json:"state"`
	ApiKey                string     `json:"api_key" gorm:"column:api_key"`
	AgentToken            string     `json:"-" gorm:"column:agent_token;not null"`
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

type CustomizationRelease struct {
	ID          int        `json:"id" gorm:"primaryKey"`
	TenantID    int        `json:"tenant_id" gorm:"index;not null"`
	ModuleName  string     `json:"module_name" gorm:"not null"`
	Version     int        `json:"version" gorm:"not null"`
	State       string     `json:"state" gorm:"not null"`
	FileCount   int        `json:"file_count"`
	SizeBytes   int64      `json:"size_bytes"`
	ArchiveHash string     `json:"archive_hash"`
	Failure     string     `json:"failure,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
}

type PlatformAuditEvent struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	TenantID  int       `json:"tenant_id" gorm:"index;not null"`
	Action    string    `json:"action" gorm:"not null"`
	Resource  string    `json:"resource"`
	Status    string    `json:"status" gorm:"not null"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}
