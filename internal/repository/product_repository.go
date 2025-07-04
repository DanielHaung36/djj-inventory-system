package repository

import (
	"context"
	"djj-inventory-system/internal/model/catalog"

	"gorm.io/gorm"
)

type ProductRepository struct {
	DB *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *catalog.Product) error {
	return r.DB.
		Session(&gorm.Session{FullSaveAssociations: true}).
		WithContext(ctx).
		Create(p).
		Error
}

func (r *ProductRepository) Update(ctx context.Context, p *catalog.Product) error {
	return r.DB.WithContext(ctx).Save(p).Error
}

func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(&catalog.Product{}, id).Error
}

func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*catalog.Product, error) {
	var p catalog.Product
	err := r.DB.WithContext(ctx).
		Preload("Stocks.Warehouse").
		Preload("Images").
		Preload("Attachments").
		First(&p, id).Error
	return &p, err
}

func (r *ProductRepository) List(ctx context.Context, offset, limit int) ([]catalog.Product, int64, error) {
	var (
		products []catalog.Product
		total    int64
	)
	q := r.DB.WithContext(ctx).Model(&catalog.Product{})
	q.Count(&total)
	err := q.
		Preload("Stocks.Warehouse").
		Preload("Images").
		Preload("Attachments").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&products).Error

	return products, total, err
}
