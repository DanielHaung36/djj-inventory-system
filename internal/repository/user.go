// repository/user.go
package repository

import (
	"djj-inventory-system/internal/model"

	"gorm.io/gorm"
)

type UserRepo interface {
	Create(*model.User) error
	FindByID(uint) (*model.User, error)
	FindAll() ([]model.User, error)
	Update(*model.User) error
	Delete(uint) error

	// 角色关联
	AddRole(userID, roleID uint) error
	RemoveRole(userID, roleID uint) error
	ListRoles(userID uint) ([]model.Role, error)
	ListRolePermissions(userID uint) ([]model.Permission, error)
	// 新增这一行
	FindByUsername(username string) (*model.User, error)
}

type userRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db}
}

func (r *userRepo) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *userRepo) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.Preload("Roles").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) FindAll() ([]model.User, error) {
	var list []model.User
	if err := r.db.Preload("Roles").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *userRepo) Update(u *model.User) error {
	return r.db.Save(u).Error
}

func (r *userRepo) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *userRepo) AddRole(userID, roleID uint) error {
	ur := model.UserRole{UserID: userID, RoleID: roleID}
	return r.db.FirstOrCreate(&ur, ur).Error
}

func (r *userRepo) RemoveRole(userID, roleID uint) error {
	return r.db.
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.UserRole{}).Error
}

func (r *userRepo) ListRoles(userID uint) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.
		Table("roles").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *userRepo) FindByUsername(username string) (*model.User, error) {
	var u model.User
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

func (r *userRepo) ListRolePermissions(userID uint) ([]model.Permission, error) {
	var perms []model.Permission
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
