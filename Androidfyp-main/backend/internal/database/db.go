package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Global GORM DB connection
var DB *gorm.DB

// Connect initializes the GORM connection to Supabase PostgreSQL
func Connect() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is not set (see backend .env.example)")
	}

	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Disables prepared statements - fixes the cache issue
	}), &gorm.Config{
		PrepareStmt: false, // Disables prepared statement cache completely
	})

	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}

	fmt.Println("✅ Successfully connected to PostgreSQL using GORM")
	return nil
}
