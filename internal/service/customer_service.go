package service

import (
	"context"
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/repository"
	"fmt"
)

type CustomerService interface {
	List(ctx context.Context) ([]catalog.Customer, error)
	Get(ctx context.Context, id uint) (*catalog.Customer, error)
	Create(ctx context.Context, input *catalog.Customer) (*catalog.Customer, error)
	Update(ctx context.Context, id uint, input *catalog.Customer) (*catalog.Customer, error)
	Delete(ctx context.Context, id uint) error
}

type customerService struct {
	repo *repository.CustomerRepo
}

func NewCustomerService(repo *repository.CustomerRepo) CustomerService {
	return &customerService{repo}
}

func (s *customerService) List(ctx context.Context) ([]catalog.Customer, error) {
	return s.repo.GetAll()
}

func (s *customerService) Get(ctx context.Context, id uint) (*catalog.Customer, error) {
	c, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("service: get customer %d: %w", id, err)
	}
	return c, nil
}

func (s *customerService) Create(ctx context.Context, input *catalog.Customer) (*catalog.Customer, error) {
	if err := s.repo.Create(input); err != nil {
		return nil, fmt.Errorf("service: create customer: %w", err)
	}
	return input, nil
}

func (s *customerService) Update(ctx context.Context, id uint, input *catalog.Customer) (*catalog.Customer, error) {
	input.ID = id
	if err := s.repo.Update(input); err != nil {
		return nil, fmt.Errorf("service: update customer %d: %w", id, err)
	}
	return input, nil
}

func (s *customerService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("service: delete customer %d: %w", id, err)
	}
	return nil
}
