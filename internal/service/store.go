// internal/service/store.go
package service

import (
	"context"
	"djj-inventory-system/internal/model/catalog"

	"gorm.io/gorm"
)

type StoreService struct {
	db *gorm.DB
}

func NewStoreService(db *gorm.DB) *StoreService {
	return &StoreService{db: db}
}

func (s *StoreService) ListStores(ctx context.Context) ([]catalog.Store, error) {
	var stores []catalog.Store
	err := s.db.Preload("Region.Company").Preload("Company").Preload("Manager").Find(&stores).Error
	return stores, err
}

func (s *StoreService) GetStoreByID(ctx context.Context, id string) (catalog.Store, error) {
	var store catalog.Store
	err := s.db.Preload("Region.Company").Preload("Company").Preload("Manager").
		First(&store, "id = ?", id).Error
	return store, err
}
