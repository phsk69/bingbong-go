// Description: This file contains the database models for the application. These models are used to define the database schema and relationships between tables. The models are defined using GORM tags to specify the column types, constraints, and relationships. The models also contain hooks to handle timestamps and other operations before creating or updating records.
package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// User represents the user model
type User struct {
	ID        uint           `gorm:"primaryKey"`
	Username  string         `gorm:"type:varchar(255);unique;not null"`
	Email     string         `gorm:"type:varchar(255);unique;not null"`
	Password  string         `gorm:"type:text;not null"` // Changed from password_hash for Go naming convention
	PublicKey string         `gorm:"type:text"`
	Active    bool           `gorm:"default:true"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastLogin time.Time      `gorm:""`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete support

	// Relationships
	AdminAccess          []AdminGroupMember `gorm:"foreignKey:UserID"`
	GroupMemberships     []UserGroupMember  `gorm:"foreignKey:UserID"`
	GroupInvitesSent     []UserGroupInvite  `gorm:"foreignKey:InviteInitiatorID"`
	GroupInvitesReceived []UserGroupInvite  `gorm:"foreignKey:InviteeID"`
	CreatedGroups        []UserGroup        `gorm:"foreignKey:CreatedByID"`
	SentMessages         []GroupMessage     `gorm:"foreignKey:SenderID"`
	Notifications        []Notification     `gorm:"foreignKey:UserID"`
}

// UserGroup represents a group of users
type UserGroup struct {
	ID          uint           `gorm:"primaryKey"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:varchar(1024)"`
	CreatedByID uint           `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"` // Soft delete support

	// Relationships
	Creator  User              `gorm:"foreignKey:CreatedByID"`
	Members  []UserGroupMember `gorm:"foreignKey:GroupID"`
	Invites  []UserGroupInvite `gorm:"foreignKey:GroupID"`
	Messages []GroupMessage    `gorm:"foreignKey:GroupID"`
}

// UserGroupInvite represents an invitation to join a group
type UserGroupInvite struct {
	ID                uint      `gorm:"primaryKey"`
	GroupID           uint      `gorm:"not null"`
	InviteInitiatorID uint      `gorm:"not null"`
	InviteeID         uint      `gorm:"not null"`
	Accepted          bool      `gorm:"default:false;not null"`
	CreatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationships
	Group     UserGroup `gorm:"foreignKey:GroupID"`
	Initiator User      `gorm:"foreignKey:InviteInitiatorID"`
	Invitee   User      `gorm:"foreignKey:InviteeID"`
}

// UserGroupMember represents a user's membership in a group
type UserGroupMember struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	GroupID   uint      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationships
	User  User      `gorm:"foreignKey:UserID"`
	Group UserGroup `gorm:"foreignKey:GroupID"`
}

// AdminGroupMember represents an admin user
type AdminGroupMember struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"unique;not null"`
	Active    bool      `gorm:"default:true;not null"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationship
	User User `gorm:"foreignKey:UserID"`
}

// GroupMessage represents a message in a group
type GroupMessage struct {
	ID        uint           `gorm:"primaryKey"`
	GroupID   uint           `gorm:"not null"`
	SenderID  uint           `gorm:"not null"`
	Content   string         `gorm:"type:text;not null"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete support

	// Relationships
	Group  UserGroup `gorm:"foreignKey:GroupID"`
	Sender User      `gorm:"foreignKey:SenderID"`
}

// ToDict converts a GroupMessage to a map for JSON serialization
func (m *GroupMessage) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":              m.ID,
		"group_id":        m.GroupID,
		"sender_id":       m.SenderID,
		"sender_username": m.Sender.Username,
		"content":         m.Content,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
	}
}

// Notification represents a user notification
type Notification struct {
	ID        uint           `gorm:"primaryKey"`
	UserID    uint           `gorm:"not null"`
	Type      string         `gorm:"type:varchar(50);not null"`
	Content   string         `gorm:"type:jsonb;not null"` // Using JSONB for PostgreSQL
	Read      bool           `gorm:"default:false;not null"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete support

	// Relationship
	User User `gorm:"foreignKey:UserID"`
}

// ToDict converts a Notification to a map for JSON serialization
func (n *Notification) ToDict() map[string]interface{} {
	var content interface{}
	_ = json.Unmarshal([]byte(n.Content), &content)

	return map[string]interface{}{
		"id":         n.ID,
		"user_id":    n.UserID,
		"type":       n.Type,
		"content":    content,
		"read":       n.Read,
		"created_at": n.CreatedAt,
		"updated_at": n.UpdatedAt,
	}
}

// BeforeCreate hook to handle timestamps
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to handle timestamps
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
