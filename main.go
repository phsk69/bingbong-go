package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.ssy.dk/noob/bingbong-go/db"
	"git.ssy.dk/noob/bingbong-go/redis"
	"git.ssy.dk/noob/bingbong-go/routes"
	"github.com/joho/godotenv"
)

func setupServer(router http.Handler, port string) *http.Server {
	return &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
}

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}
	return port
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables.")
	}

	// Initialize DB
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis and WebSocket hub
	hub, err := redis.InitRedis()
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Initialize router with routes
	router := routes.NewRouter(database.GetDB()) // Use GetDB() to get *gorm.DB
	router.SetHub(hub)
	router.SetupRoutes()

	// Setup HTTP server
	server_port := getServerPort()
	srv := setupServer(router, server_port)

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

	// Attempt to shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Close DB connection
	if sqlDB, err := database.GetSQLDB(); err == nil {
		log.Println("Closing database connection...")
		sqlDB.Close()
	}

	// Stop Redis hub
	if hub != nil {
		hub.Stop()
		log.Println("Redis hub stopped")
	}

	log.Println("Server exiting")
}
