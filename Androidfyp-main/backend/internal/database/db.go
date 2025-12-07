package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Global GORM DB connection
var DB *gorm.DB

// Connect initializes the GORM connection to Supabase PostgreSQL
func Connect() error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=require",
		"aws-0-eu-north-1.pooler.supabase.com",
		"postgres.npeusanizvcyjwsgbhfn",
		"Hamoudi123?",
		"postgres",
		6543,
	)

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

	fmt.Println("✅ Successfully connected to Supabase PostgreSQL using GORM")
	return nil
}
