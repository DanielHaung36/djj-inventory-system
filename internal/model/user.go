package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`                                                                       // 用户ID
	Version      int            `gorm:"type:int;not null;default:1;version" json:"version"`                                         // 乐观锁版本号
	Username     string         `gorm:"size:50;uniqueIndex;not null" json:"username"`                                               // 登录用户名
	Email        string         `gorm:"size:100;uniqueIndex;not null" json:"email"`                                                 // 邮箱（登录／联系用）
	PasswordHash string         `gorm:"size:256;not null" json:"-"`                                                                 // 密码哈希（不暴露）
	IsDeleted    bool           `gorm:"not null;default:false" json:"is_deleted"`                                                   // 软删除标记
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`                                                           // 创建时间
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`                                                           // 最后更新时间
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                                                                             // GORM 原生软删除字段（可选）
	Roles        []Role         `gorm:"many2many:user_roles;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"roles,omitempty"` // 用户拥有的角色
	// **NEW**: collect all perms through Role → RolePermission:
	Permissions []Permission `gorm:"-"`
}
