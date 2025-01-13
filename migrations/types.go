package migrations

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	Version     string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error // For rollbacks if needed
}

// SchemaMigration tracks which migrations have been applied
type SchemaMigration struct {
	Version     string `gorm:"primaryKey"`
	Description string
	AppliedAt   time.Time
	Duration    time.Duration // Track how long the migration took
}

// MigrationError provides detailed migration failure information
type MigrationError struct {
	Version string
	Err     error
}

func (e *MigrationError) Error() string {
	return fmt.Sprintf("migration %s failed: %v", e.Version, e.Err)
}
