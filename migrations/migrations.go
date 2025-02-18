// migrations/migrations.go
package migrations

import (
	"git.ssy.dk/noob/bingbong-go/models"
	"gorm.io/gorm"
)

type Migration struct {
	Version     string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

type Runner struct {
	db *gorm.DB
}

func NewRunner(db *gorm.DB) *Runner {
	return &Runner{db: db}
}

func (r *Runner) Run() error {
	// Implement migration logic here
	return nil
}

func (r *Runner) Status() ([]string, error) {
	// Implement status fetching logic here
	return []string{}, nil
}

var Migrations = []Migration{
	{
		Version:     "2025.01.13.01",
		Description: "Create base user and group tables",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(
				&models.User{},
				&models.UserGroup{},
			)
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(
				&models.UserGroup{},
				&models.User{},
			)
		},
	},
	{
		Version:     "2025.01.13.02",
		Description: "Create relationship tables",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(
				&models.UserGroupMember{},
				&models.UserGroupInvite{},
				&models.AdminGroupMember{},
			)
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(
				&models.AdminGroupMember{},
				&models.UserGroupInvite{},
				&models.UserGroupMember{},
			)
		},
	},
	{
		Version:     "2025.01.13.03",
		Description: "Create message and notification tables",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(
				&models.GroupMessage{},
				&models.Notification{},
			)
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(
				&models.Notification{},
				&models.GroupMessage{},
			)
		},
	},
	{
		Version:     "2025.01.13.04",
		Description: "Add indexes for performance",
		Up: func(db *gorm.DB) error {
			// Add specific indexes for better query performance
			if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_group_messages_created_at ON group_messages(created_at)`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, read)`).Error; err != nil {
				return err
			}
			return nil
		},
		Down: func(db *gorm.DB) error {
			if err := db.Exec(`DROP INDEX IF EXISTS idx_group_messages_created_at`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP INDEX IF EXISTS idx_notifications_user_read`).Error; err != nil {
				return err
			}
			return nil
		},
	},
}
