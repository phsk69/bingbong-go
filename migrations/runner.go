// migrations/runner.go
package migrations

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Runner struct {
	db *gorm.DB
}

func NewRunner(db *gorm.DB) *Runner {
	return &Runner{db: db}
}

func (r *Runner) Run() error {
	// Create migrations table if it doesn't exist
	if err := r.db.AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	for _, migration := range Migrations {
		var sm SchemaMigration
		if r.db.Where("version = ?", migration.Version).First(&sm).RowsAffected == 0 {
			log.Printf("Running migration %s: %s", migration.Version, migration.Description)

			start := time.Now()
			if err := r.db.Transaction(func(tx *gorm.DB) error {
				if err := migration.Up(tx); err != nil {
					return &MigrationError{Version: migration.Version, Err: err}
				}
				return nil
			}); err != nil {
				return err
			}

			duration := time.Since(start)

			// Record successful migration
			r.db.Create(&SchemaMigration{
				Version:     migration.Version,
				Description: migration.Description,
				AppliedAt:   time.Now(),
				Duration:    duration,
			})

			log.Printf("Completed migration %s in %v", migration.Version, duration)
		}
	}
	return nil
}

func (r *Runner) Status() ([]SchemaMigration, error) {
	var migrations []SchemaMigration
	if err := r.db.Order("version").Find(&migrations).Error; err != nil {
		return nil, err
	}
	return migrations, nil
}

// For development/testing environments
func (r *Runner) Reset() error {
	log.Println("Resetting database - THIS SHOULD NOT BE USED IN PRODUCTION")

	// Run all down migrations in reverse order
	var migrations []SchemaMigration
	if err := r.db.Order("version desc").Find(&migrations).Error; err != nil {
		return err
	}

	for _, sm := range migrations {
		for _, m := range Migrations {
			if m.Version == sm.Version && m.Down != nil {
				if err := m.Down(r.db); err != nil {
					return err
				}
				if err := r.db.Delete(&sm).Error; err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}
