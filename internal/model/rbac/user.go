package rbac

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Version           int            `gorm:"type:int;not null;default:1;version" json:"version"`
	Username          string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	StoreID           uint           `json:"store_id"`
	Email             string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash      string         `gorm:"size:256;not null" json:"-"`
	IsDeleted         bool           `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	Roles             []Role         `gorm:"many2many:user_roles;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"roles,omitempty"`
	DirectPermissions []Permission   `gorm:"many2many:user_permissions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"direct_permissions,omitempty"`
	Permissions       []Permission   `gorm:"-" json:"permissions,omitempty"`
	AvatarURL         string         `gorm:"size:255;not null" json:"avatar_url"`
}
