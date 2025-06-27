// internal/repository/quote_repository.go
package repository

import (
	"djj-inventory-system/internal/model"
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

// GetByID 根据主键读取报价单（连同明细）
func (r *QuoteRepository) GetByID(id uint) (*model.Quote, error) {
	var q model.Quote
	// 预加载 Items
	if err := r.DB.Preload("Items").
		First(&q, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &q, nil
}
