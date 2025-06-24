package model

import "time"

// 权限表
type Permission struct {
	ID          uint      `gorm:"primaryKey"`               // 权限ID
	Name        string    `gorm:"size:100;unique;not null"` // 权限名，如 "users.create"
	CreatedAt   time.Time `gorm:"autoCreateTime"`           // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`           // 更新时间
	Label       string    `gorm:"size:100" json:"label"`
	Description string    `gorm:"size:255" json:"description"`
}
