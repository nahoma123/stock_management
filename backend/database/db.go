package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"saas-superadmin-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_USER", "odoo"),
		getEnv("POSTGRES_PASSWORD", "odoo"),
		getEnv("POSTGRES_DB", "odoo"),
		getEnv("POSTGRES_PORT", "5432"),
	)
	log.Println("Connecting to database with DSN:", dsn)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := DB.DB()
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

	log.Println("Running AutoMigration...")
	if err := DB.AutoMigrate(&models.Tenant{}, &models.MobileDevice{}); err != nil {
		log.Fatal("Failed to auto-migrate models:", err)
	}
}

func GetTenantDBConnection(dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_USER", "odoo"),
		getEnv("POSTGRES_PASSWORD", "odoo"),
		dbName,
		getEnv("POSTGRES_PORT", "5432"),
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
