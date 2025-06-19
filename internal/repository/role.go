// internal/repository/role.go
package repository

import (
	"djj-inventory-system/internal/model"

	"gorm.io/gorm"
)

type RoleRepo interface {
	Create(r *model.Role) error
	FindByID(id uint) (*model.Role, error)
	FindAll() ([]model.Role, error)
	Update(r *model.Role) error
	Delete(id uint) error
	ListPermissions(roleID uint) ([]model.Permission, error)
}

type roleGormRepo struct{ db *gorm.DB }

func NewRoleRepo(db *gorm.DB) RoleRepo             { return &roleGormRepo{db} }
func (r *roleGormRepo) Create(x *model.Role) error { return r.db.Create(x).Error }
func (r *roleGormRepo) FindByID(id uint) (*model.Role, error) {
	var x model.Role
	return &x, r.db.First(&x, id).Error
}
func (r *roleGormRepo) FindAll() ([]model.Role, error) {
	var xs []model.Role
	return xs, r.db.Find(&xs).Error
}
func (r *roleGormRepo) Update(x *model.Role) error { return r.db.Save(x).Error }
func (r *roleGormRepo) Delete(id uint) error       { return r.db.Delete(&model.Role{}, id).Error }

func (r *roleGormRepo) ListPermissions(roleID uint) ([]model.Permission, error) {
	var perms []model.Permission
	// 假设 role_permissions 是你 join 表，字段是 role_id / permission_id
	err := r.db.
		Table("permissions").
		Select("permissions.*").
		Joins("join role_permissions on role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&perms).Error
	return perms, err
}
