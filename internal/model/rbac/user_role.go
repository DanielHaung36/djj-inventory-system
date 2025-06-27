package rbac

// 用户-角色 关联表 (many2many)
type UserRole struct {
	UserID uint `gorm:"primaryKey"` // 用户ID
	RoleID uint `gorm:"primaryKey"` // 角色ID
}
