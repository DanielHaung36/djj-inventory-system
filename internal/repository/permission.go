// internal/repository/permission.go
package repository

import (
	"djj-inventory-system/internal/model/rbac"

	"gorm.io/gorm"
)

// PermRepo 定义了 Permission 的 CRUD
type PermRepo interface {
	Create(p *rbac.Permission) error
	FindByID(id uint) (*rbac.Permission, error)
	FindAll() ([]rbac.Permission, error)
	Update(p *rbac.Permission) error
	Delete(id uint) error
}

// permRepo 是 PermRepo 的 GORM 实现
type permRepo struct {
	db *gorm.DB
}

// NewPermRepo 返回一个新的 Permission 仓库
func NewPermRepo(db *gorm.DB) PermRepo {
	return &permRepo{db: db}
}

func (r *permRepo) Create(p *rbac.Permission) error {
	return r.db.Create(p).Error
}

func (r *permRepo) FindByID(id uint) (*rbac.Permission, error) {
	var p rbac.Permission
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *permRepo) FindAll() ([]rbac.Permission, error) {
	var list []rbac.Permission
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *permRepo) Update(p *rbac.Permission) error {
	return r.db.Save(p).Error
}

func (r *permRepo) Delete(id uint) error {
	return r.db.Delete(&rbac.Permission{}, id).Error
}
