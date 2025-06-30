// repository/user.go
package repository

import (
	"djj-inventory-system/internal/model/audit"
	"djj-inventory-system/internal/model/rbac"

	"gorm.io/gorm"
)

type UserRepo interface {
	Create(*rbac.User) error
	FindByID(uint) (*rbac.User, error)
	FindAll() ([]rbac.User, error)
	Update(*rbac.User) error
	Delete(uint) error

	// 角色关联
	AddRole(userID, roleID uint) error
	RemoveRole(userID, roleID uint) error
	ListRoles(userID uint) ([]rbac.Role, error)
	ListRolePermissions(userID uint) ([]rbac.Permission, error)
	// 新增这一行
	FindByUsername(username string) (*rbac.User, error)

	// 直接给用户批量增删权限
	GrantUserPermissions(userID uint, permIDs []uint) error
	RevokeUserPermissions(userID uint, permIDs []uint) error

	// 查询用户：角色继承+直接权限扁平去重后的全部权限
	FindWithAllPerms(userID uint) (*rbac.User, error)

	// 获取该用户权限最后一次变更的审计记录
	LastPermissionChange(userID uint) (*audit.AuditedHistory, error)
}

type userRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db}
}

func (r *userRepo) Create(u *rbac.User) error {
	return r.db.Create(u).Error
}

func (r *userRepo) FindByID(id uint) (*rbac.User, error) {
	var u rbac.User
	if err := r.db.Preload("Roles").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) FindAll() ([]rbac.User, error) {
	var list []rbac.User
	if err := r.db.Preload("Roles").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *userRepo) Update(u *rbac.User) error {
	return r.db.Save(u).Error
}

func (r *userRepo) Delete(id uint) error {
	return r.db.Delete(&rbac.User{}, id).Error
}

func (r *userRepo) AddRole(userID, roleID uint) error {
	// 1. 删除该用户现有的所有角色（如果没有记纪录，DELETE 不会出错）
	if err := r.db.
		Where("user_id = ?", userID).
		Delete(&rbac.UserRole{}).Error; err != nil {
		return err
	}

	// 2. 插入新的角色
	ur := rbac.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	if err := r.db.Create(&ur).Error; err != nil {
		return err
	}

	return nil
}

func (r *userRepo) RemoveRole(userID, roleID uint) error {
	return r.db.
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&rbac.UserRole{}).Error
}

func (r *userRepo) ListRoles(userID uint) ([]rbac.Role, error) {
	var roles []rbac.Role
	err := r.db.
		Table("roles").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *userRepo) FindByUsername(username string) (*rbac.User, error) {
	var u rbac.User
	if err := r.db.
		Preload("Roles").
		Where("username = ? AND is_deleted = FALSE", username).
		First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// userGormRepo 实现

/*
	Table("permissions")
	告诉 GORM 从 permissions 表开始查询，相当于 SQL 里的 FROM permissions。
	Select("permissions.*")
	表示要选取这个表里所有列（permissions.*）。
	Joins("join role_permissions on role_permissions.permission_id = permissions.id")
	用 SQL 的 JOIN 把 permissions 和中间表 role_permissions 关联起来：
	role_permissions.permission_id = permissions.id 这一句指定了连接条件，意思是“role_permissions 里指向某条权限的 permission_id，要和 permissions 表的主键 id 对上号”。
	Where("role_permissions.role_id = ?", roleID)
	加一个过滤条件，只拿出那些在 role_permissions 表里，其 role_id 等于我们传进来的 roleID 的行。
	Find(&perms)
	把最终筛出来的权限记录读到 perms 这个切片里。

	SELECT permissions.*

		 FROM permissions
		 JOIN role_permissions
		   ON role_permissions.permission_id = permissions.id
		WHERE role_permissions.role_id = ?;
*/

func (r *userRepo) ListRolePermissions(userID uint) ([]rbac.Permission, error) {
	var perms []rbac.Permission
	// 假设 user_roles、role_permissions、permissions 三表关联：
	// user_roles(user_id, role_id)
	// role_permissions(role_id, permission_id)
	// permissions(id,...)
	err := r.db.
		Table("permissions").
		Select("permissions.*").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", userID).
		Find(&perms).
		Error
	return perms, err
}

func (r *userRepo) GrantUserPermissions(userID uint, permIDs []uint) error {
	// 1. 把 permIDs 对应的 Permission 记录 load 出来
	var perms []rbac.Permission
	if err := r.db.
		Where("id IN ?", permIDs).
		Find(&perms).
		Error; err != nil {
		return err
	}
	// 2. 用 GORM 的 Association.Append 批量插入 user_permissions
	return r.db.
		Model(&rbac.User{ID: userID}).
		Association("DirectPermissions").
		Append(perms)
}

func (r *userRepo) RevokeUserPermissions(userID uint, permIDs []uint) error {
	// 1. load 出要删的那些权限
	var perms []rbac.Permission
	if err := r.db.
		Where("id IN ?", permIDs).
		Find(&perms).
		Error; err != nil {
		return err
	}
	// 2. 批量从 user_permissions 删除
	return r.db.
		Model(&rbac.User{ID: userID}).
		Association("DirectPermissions").
		Delete(perms)
}

func (r *userRepo) FindWithAllPerms(userID uint) (*rbac.User, error) {
	// 1. 预加载 Roles → Permissions，以及 DirectPermissions
	var u rbac.User
	if err := r.db.
		Preload("Roles.Permissions").
		Preload("DirectPermissions").
		First(&u, userID).
		Error; err != nil {
		return nil, err
	}

	// 2. 扁平化去重：把角色的权限 和 直接权限 合并到 u.Permissions
	permMap := make(map[uint]rbac.Permission)
	for _, role := range u.Roles {
		for _, p := range role.Permissions {
			permMap[p.ID] = p
		}
	}
	for _, p := range u.DirectPermissions {
		permMap[p.ID] = p
	}
	for _, p := range permMap {
		u.Permissions = append(u.Permissions, p)
	}

	return &u, nil
}

// LastPermissionChange 查询用户权限的最近一次审计记录
func (r *userRepo) LastPermissionChange(userID uint) (*audit.AuditedHistory, error) {
	var ah audit.AuditedHistory
	err := r.db.
		Where("table_name = ? AND record_id = ?", audit.AuditedTableUserRoles, userID).
		Order("changed_at DESC").
		Limit(1).
		Take(&ah).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ah, nil
}
