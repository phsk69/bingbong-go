// models/journal.go
package models

import (
	"gorm.io/gorm"
)

type JournalEntry struct {
	gorm.Model
	ID        string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Title     string `gorm:"type:varchar(255);"`
	Body      string `gorm:"type:text"`
	Thumbnail []byte `gorm:"type:bytea;default:null"`
}
