package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"git.ssy.dk/noob/snakey-go/v2/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Try to load .env file, fallback to environment variables
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

	// DSN for PostgreSQL connection
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s port=%s", host, user, dbname, sslmode, password, port)

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Configure the database connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get database: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	// Ensure UUID extension exists and migrate the schema
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("warning: failed to create extension: %v", err)
	}

	if err := db.AutoMigrate(&models.JournalEntry{}); err != nil {
		log.Printf("warning: failed to migrate database: %v", err)
	}

	router := gin.Default()

	// GET - Retrieve all journal entries
	router.GET("/api/journal", func(c *gin.Context) {
		var journals []models.JournalEntry
		if result := db.Find(&journals); result.Error != nil {
			c.JSON(500, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(200, journals)
	})

	// GET - Retrieve a journal entry by ID
	router.GET("/api/journal/:id", func(c *gin.Context) {
		id := c.Param("id")
		var journal models.JournalEntry
		if result := db.First(&journal, "id = ?", id); result.Error != nil {
			c.JSON(404, gin.H{"error": "journal entry not found"})
			return
		}
		c.JSON(200, journal)
	})

	// POST - Create a new journal entry
	router.POST("/api/journal", func(c *gin.Context) {
		var input models.JournalEntry
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		if result := db.Create(&input); result.Error != nil {
			c.JSON(500, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(201, input)
	})

	// PATCH - Update an existing journal entry by ID
	router.PATCH("/api/journal/:id", func(c *gin.Context) {
		id := c.Param("id")
		var journal models.JournalEntry
		if result := db.First(&journal, "id = ?", id); result.Error != nil {
			c.JSON(404, gin.H{"error": "journal entry not found"})
			return
		}

		var input models.JournalEntry
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		if result := db.Model(&journal).Updates(input); result.Error != nil {
			c.JSON(500, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(200, journal)
	})

	// PUT - Replace an existing journal entry by ID
	router.PUT("/api/journal/:id", func(c *gin.Context) {
		id := c.Param("id")
		var journal models.JournalEntry
		if result := db.First(&journal, "id = ?", id); result.Error != nil {
			c.JSON(404, gin.H{"error": "journal entry not found"})
			return
		}

		var input models.JournalEntry
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		journal.Title = input.Title
		journal.Body = input.Body
		journal.Thumbnail = input.Thumbnail
		if result := db.Save(&journal); result.Error != nil {
			c.JSON(500, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(200, journal)
	})

	// DELETE - Delete a journal entry by ID
	router.DELETE("/api/journal/:id", func(c *gin.Context) {
		id := c.Param("id")
		var journal models.JournalEntry
		if result := db.First(&journal, "id = ?", id); result.Error != nil {
			c.JSON(404, gin.H{"error": "journal entry not found"})
			return
		}

		if result := db.Delete(&journal); result.Error != nil {
			c.JSON(500, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "journal entry deleted successfully"})
	})

	// DELETE - Batch of journal entries by ID
	// Example cURL: curl -X DELETE -H "Content-Type: application/json" -d '["id1", "id2"]' http://localhost:8080/api/journal
	router.DELETE("/api/journal", func(c *gin.Context) {
		var input []string
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		// Start a transaction
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(500, gin.H{"error": "failed to start transaction"})
			return
		}

		// Perform the delete operation
		if result := tx.Where("id IN ?", input).Delete(&models.JournalEntry{}); result.Error != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": result.Error.Error()})
			return
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "failed to commit transaction"})
			return
		}

		c.JSON(200, gin.H{"message": "journal entries deleted successfully"})
	})

	log.Println("Server starting on port 8080")
	router.Run(":8080")
}
