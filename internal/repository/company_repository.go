package repository

import (
	"context"
	"djj-inventory-system/internal/model/company"
	"errors"

	"gorm.io/gorm"
)

// CompanyRepository 封装对 companies 表的访问
type CompanyRepository struct {
	DB *gorm.DB
}

// NewCompanyRepository 返回一个新的 CompanyRepository
func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{DB: db}
}

// FindDefault 从 companies 表中查出标记为默认的那家公司
func (r *CompanyRepository) FindDefault(ctx context.Context) (*company.Company, error) {
	var co company.Company
	err := r.DB.WithContext(ctx).
		Where("is_default = ?", true).
		First(&co).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &co, nil
}
