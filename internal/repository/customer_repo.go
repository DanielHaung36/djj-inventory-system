package repository

import (
	"djj-inventory-system/internal/model/catalog"
	"time"

	"gorm.io/gorm"
)

type CustomerRepo struct {
	db *gorm.DB
}

func NewCustomerRepo(db *gorm.DB) *CustomerRepo {
	return &CustomerRepo{db}
}

// GetAll 返回所有未删除的客户，预加载门店、地区和公司
func (r *CustomerRepo) GetAll() ([]catalog.Customer, error) {
	var cs []catalog.Customer
	if err := r.db.
		Preload("Store").
		Preload("Store.Region").
		Preload("Store.Company").
		Where("is_deleted = ?", false).
		Find(&cs).Error; err != nil {
		return nil, err
	}
	return cs, nil
}

// GetByID 查询单个客户
func (r *CustomerRepo) GetByID(id uint) (*catalog.Customer, error) {
	var c catalog.Customer
	if err := r.db.
		Preload("Store").
		Preload("Store.Region").
		Preload("Store.Company").
		First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// Create 新增客户
func (r *CustomerRepo) Create(cust *catalog.Customer) error {
	cust.CreatedAt = time.Now()
	cust.UpdatedAt = time.Now()
	cust.IsDeleted = false
	return r.db.Create(cust).Error
}

// Update 修改客户（软更新）
func (r *CustomerRepo) Update(cust *catalog.Customer) error {
	cust.UpdatedAt = time.Now()
	return r.db.Model(&catalog.Customer{}).
		Where("id = ?", cust.ID).
		Updates(cust).Error
}

// Delete 软删除客户
func (r *CustomerRepo) Delete(id uint) error {
	return r.db.Model(&catalog.Customer{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}
