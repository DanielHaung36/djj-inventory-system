package model

import "time"

// 角色表
type Role struct {
	ID          uint         `gorm:"primaryKey"`                 // 角色ID
	Name        string       `gorm:"size:50;unique;not null"`    // 角色名
	CreatedAt   time.Time    `gorm:"autoCreateTime"`             // 创建时间
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`             // 更新时间
	Permissions []Permission `gorm:"many2many:role_permissions"` // 多对多：角色拥有的权限
}
