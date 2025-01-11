package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.ssy.dk/noob/snakey-go/v2/models"
	"git.ssy.dk/noob/snakey-go/v2/routes"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDB() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables.")
	}

	// Read environment variables
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s port=%s",
		host, user, dbname, sslmode, password, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %v", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	// Setup extensions and migrations
	if err := setupDatabase(db); err != nil {
		return nil, err
	}

	return db, nil
}

func setupDatabase(db *gorm.DB) error {
	// Ensure UUID extension exists
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("warning: failed to create extension: %v", err)
	}

	migrate := os.Getenv("DB_MIGRATE")
	if migrate == "false" {
		log.Println("Migrate is set to false, skipping database migration")
		return nil
	}

	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.UserGroup{},
		&models.UserGroupInvite{},
		&models.UserGroupMember{},
		&models.AdminGroupMember{},
		&models.GroupMessage{},
		&models.Notification{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}

func main() {
	// Initialize DB
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize router with routes
	router := routes.NewRouter(db)
	router.SetupRoutes()

	server_port := os.Getenv("SERVER_PORT")

	// Create server
	srv := &http.Server{
		Addr:    ":" + server_port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", server_port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a timeout of 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Close DB connection
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	log.Println("Server exiting")
}
