package service

import (
	"context"

	"djj-inventory-system/internal/model/catalog"

	"gorm.io/gorm"
)

type RegionService struct {
	DB *gorm.DB
}

func NewRegionService(conn *gorm.DB) *RegionService {
	return &RegionService{DB: conn}
}

func (s *RegionService) List(ctx context.Context) ([]catalog.Region, error) {
	var list []catalog.Region
	err := s.DB.
		Preload("Company").
		Preload("Warehouses").
		Find(&list).Error
	return list, err
}

func (s *RegionService) GetByID(ctx context.Context, id uint) (*catalog.Region, error) {
	var r catalog.Region
	if err := s.DB.
		Preload("Company").
		Preload("Warehouses").
		First(&r, id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *RegionService) Create(ctx context.Context, in *catalog.Region) (*catalog.Region, error) {
	if err := s.DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

func (s *RegionService) Update(ctx context.Context, in *catalog.Region) (*catalog.Region, error) {
	if err := s.DB.Save(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

func (s *RegionService) Delete(ctx context.Context, id uint) error {
	return s.DB.Delete(&catalog.Region{}, id).Error
}
