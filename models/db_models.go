package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey"`
	Username  string         `gorm:"type:varchar(255);unique;not null"`
	Email     string         `gorm:"type:varchar(255);unique;not null"`
	Password  string         `gorm:"type:text;not null"`
	PublicKey string         `gorm:"type:text"`
	Active    bool           `gorm:"default:true"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastLogin time.Time      `gorm:""`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relationships with cascade delete
	AdminAccess          []AdminGroupMember `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	GroupMemberships     []UserGroupMember  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	GroupInvitesSent     []UserGroupInvite  `gorm:"foreignKey:InviteInitiatorID;constraint:OnDelete:CASCADE;"`
	GroupInvitesReceived []UserGroupInvite  `gorm:"foreignKey:InviteeID;constraint:OnDelete:CASCADE;"`
	CreatedGroups        []UserGroup        `gorm:"foreignKey:CreatedByID;constraint:OnDelete:CASCADE;"`
	SentMessages         []GroupMessage     `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE;"`
	Notifications        []Notification     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

type UserGroup struct {
	ID          uint           `gorm:"primaryKey"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:varchar(1024)"`
	CreatedByID uint           `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships with cascade delete
	Creator  User              `gorm:"foreignKey:CreatedByID"`
	Members  []UserGroupMember `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE;"`
	Invites  []UserGroupInvite `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE;"`
	Messages []GroupMessage    `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE;"`
}

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

type AdminGroupMember struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"unique;not null"`
	Active    bool      `gorm:"default:true;not null"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationship
	User User `gorm:"foreignKey:UserID"`
}

type GroupMessage struct {
	ID        uint           `gorm:"primaryKey"`
	GroupID   uint           `gorm:"not null"`
	SenderID  uint           `gorm:"not null"`
	Content   string         `gorm:"type:text;not null"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relationships
	Group  UserGroup `gorm:"foreignKey:GroupID"`
	Sender User      `gorm:"foreignKey:SenderID"`
}

type Notification struct {
	ID        uint           `gorm:"primaryKey"`
	UserID    uint           `gorm:"not null"`
	Type      string         `gorm:"type:varchar(50);not null"`
	Content   string         `gorm:"type:jsonb;not null"`
	Read      bool           `gorm:"default:false;not null"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relationship
	User User `gorm:"foreignKey:UserID"`
}

// Methods remain unchanged
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

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
