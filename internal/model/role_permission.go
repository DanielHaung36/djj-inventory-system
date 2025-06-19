package model

// 角色-权限 关联表 (many2many)
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"` // 角色ID
	PermissionID uint `gorm:"primaryKey"` // 权限ID
}
