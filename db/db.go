// db/db.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"git.ssy.dk/noob/bingbong-go/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	GormDB *gorm.DB // Renamed from DB to GormDB to avoid conflict
}

// InitDB initializes and returns a new Database instance
func InitDB() (*Database, error) {
	// Read environment variables
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s port=%s",
		host, user, dbname, sslmode, password, port)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	database := &Database{GormDB: gormDB}

	if err := database.setupConnection(); err != nil {
		return nil, err
	}

	if err := database.setupDatabase(); err != nil {
		return nil, err
	}

	// Initialize admin account
	if err := database.InitializeAdminAccount(); err != nil {
		log.Printf("Warning: Failed to initialize admin account: %v", err)
	}

	return database, nil
}

func (db *Database) setupConnection() error {
	sqlDB, err := db.GormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database: %v", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	return nil
}

func (db *Database) setupDatabase() error {
	// Ensure UUID extension exists
	if err := db.GormDB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("warning: failed to create extension: %v", err)
	}

	migrate := os.Getenv("DB_MIGRATE")
	if migrate == "false" {
		log.Println("Migrate is set to false, skipping database migration")
		return nil
	}

	runner := migrations.NewRunner(db.GormDB)
	if err := runner.Run(); err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}

	// Log migration status
	migrations, err := runner.Status()
	if err != nil {
		log.Printf("Warning: couldn't fetch migration status: %v", err)
	} else {
		log.Printf("Applied %d migrations", len(migrations))
	}

	return nil
}

// GetDB returns the underlying *gorm.DB instance
func (db *Database) GetDB() *gorm.DB {
	return db.GormDB
}

// GetSQLDB returns the underlying *sql.DB instance
func (db *Database) GetSQLDB() (*sql.DB, error) {
	return db.GormDB.DB()
}

// Transaction executes a transaction with the provided function
func (db *Database) Transaction(fc func(tx *gorm.DB) error) error {
	return db.GormDB.Transaction(fc)
}
