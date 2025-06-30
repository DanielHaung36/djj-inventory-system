// internal/repository/quote_repository.go
package repository

import (
	"context"
	"djj-inventory-system/internal/model/sales"
	"errors"

	"gorm.io/gorm"
)

// QuoteRepository 封装对 quotes 表的访问
type QuoteRepository struct {
	DB *gorm.DB
}

func NewQuoteRepository(db *gorm.DB) *QuoteRepository {
	return &QuoteRepository{DB: db}
}

// FindByID 根据主键读取报价单，并 preload 所有关联
func (r *QuoteRepository) FindByID(ctx context.Context, id uint) (*sales.Quote, error) {
	var q sales.Quote
	err := r.DB.
		WithContext(ctx).
		// 公司信息
		Preload("Company").
		// 门店信息：门店自身、门店负责人、门店所属区域、以及该区域下的所有仓库
		Preload("Store").
		Preload("Store.Manager").
		Preload("Store.Region").
		Preload("Store.Region.Warehouses").
		// 客户信息：客户本身；以及客户所在门店的负责人/区域/仓库
		Preload("Customer").
		// 报价明细及明细关联的产品
		Preload("Items").
		Preload("Items.Product").
		First(&q, "id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &q, nil
}
