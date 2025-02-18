// db/admin.go
package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"git.ssy.dk/noob/bingbong-go/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// InitializeAdminAccount creates or updates the admin account based on environment variables
func (db *Database) InitializeAdminAccount() error {
	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminUsername == "" || adminEmail == "" || adminPassword == "" {
		return fmt.Errorf("missing required admin environment variables")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %v", err)
	}

	// Use a transaction to ensure data consistency
	return db.Transaction(func(tx *gorm.DB) error {
		// Try to find existing admin user
		var user models.User
		result := tx.Where("email = ?", adminEmail).First(&user)

		if result.Error != nil {
			if result.Error != gorm.ErrRecordNotFound {
				return fmt.Errorf("error checking for existing admin: %v", result.Error)
			}

			// Create new admin user if not found
			user = models.User{
				Username:  adminUsername,
				Email:     adminEmail,
				Password:  string(hashedPassword),
				Active:    true,
				LastLogin: time.Now(),
			}

			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create admin user: %v", err)
			}

			log.Printf("Created new admin user: %s", adminUsername)
		} else {
			// Update existing admin user
			updates := map[string]interface{}{
				"username":   adminUsername,
				"password":   string(hashedPassword),
				"active":     true,
				"updated_at": time.Now(),
			}

			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update admin user: %v", err)
			}

			log.Printf("Updated existing admin user: %s", adminUsername)
		}

		// Ensure admin membership
		var adminMember models.AdminGroupMember
		result = tx.Where("user_id = ?", user.ID).First(&adminMember)

		if result.Error != nil {
			if result.Error != gorm.ErrRecordNotFound {
				return fmt.Errorf("error checking admin membership: %v", result.Error)
			}

			// Create admin membership if not exists
			adminMember = models.AdminGroupMember{
				UserID: user.ID,
				Active: true,
			}

			if err := tx.Create(&adminMember).Error; err != nil {
				return fmt.Errorf("failed to create admin membership: %v", err)
			}

			log.Printf("Created admin membership for user: %s", adminUsername)
		} else if !adminMember.Active {
			// Ensure admin membership is active
			if err := tx.Model(&adminMember).Update("active", true).Error; err != nil {
				return fmt.Errorf("failed to activate admin membership: %v", err)
			}

			log.Printf("Activated admin membership for user: %s", adminUsername)
		}

		return nil
	})
}
