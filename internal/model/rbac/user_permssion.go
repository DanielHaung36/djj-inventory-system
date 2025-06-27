package rbac

// 用户-权限 关联表 (many2many)
type UserPermission struct {
	UserID       uint `gorm:"primaryKey"` // 用户ID
	PermissionID uint `gorm:"primaryKey"` // 权限ID
}
