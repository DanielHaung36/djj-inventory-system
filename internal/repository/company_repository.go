package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"djj-inventory-system/internal/model/company"
)

type CompanyRepository struct {
	DB *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{DB: db}
}

// FindDefault 加载 is_default = true 的那家
func (r *CompanyRepository) FindDefault(ctx context.Context) (*company.Company, error) {
	var co company.Company
	err := r.DB.WithContext(ctx).
		Where("is_default = ?", true).
		First(&co).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &co, err
}

// FindByID 根据主键加载
func (r *CompanyRepository) FindByID(ctx context.Context, id uint) (*company.Company, error) {
	var co company.Company
	err := r.DB.WithContext(ctx).First(&co, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &co, err
}

// FindByCode 根据唯一 code 加载
func (r *CompanyRepository) FindByCode(ctx context.Context, code string) (*company.Company, error) {
	var co company.Company
	err := r.DB.WithContext(ctx).
		Where("code = ?", code).
		First(&co).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &co, err
}

// ListAll 列出所有公司（分页／筛选可自行扩展）
func (r *CompanyRepository) ListAll(ctx context.Context) ([]company.Company, error) {
	var list []company.Company
	if err := r.DB.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
