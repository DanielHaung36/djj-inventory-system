// internal/repository/stock_repo.go
package repository

import (
	"context"
	"djj-inventory-system/internal/model/catalog"

	"gorm.io/gorm"
)

type StockRepository struct {
	DB *gorm.DB
}

func NewStockRepository(db *gorm.DB) *StockRepository {
	return &StockRepository{DB: db}
}

func (r *StockRepository) DeleteByProduct(ctx context.Context, productID uint) error {
	return r.DB.WithContext(ctx).
		Where("product_id = ?", productID).
		Delete(&catalog.ProductStock{}).Error
}

func (r *StockRepository) Create(ctx context.Context, ps *catalog.ProductStock) error {
	return r.DB.WithContext(ctx).Create(ps).Error
}
