package service

import (
	"context"
	audit2 "djj-inventory-system/internal/model/audit"
	"djj-inventory-system/internal/model/rbac"
	"djj-inventory-system/internal/pkg/audit"
	"djj-inventory-system/internal/repository"
)

type PermService interface {
	Create(ctx context.Context, name string) (*rbac.Permission, error)
	Get(ctx context.Context, id uint) (*rbac.Permission, error)
	List(ctx context.Context) ([]rbac.Permission, error)
	Update(ctx context.Context, id uint, name string) (*rbac.Permission, error)
	Delete(ctx context.Context, id uint) error
}

type permService struct {
	repo repository.PermRepo
	aud  audit.Recorder
}

func NewPermService(r repository.PermRepo, aud audit.Recorder) PermService {
	return &permService{repo: r, aud: aud}
}

func (s *permService) Create(ctx context.Context, name string) (*rbac.Permission, error) {
	p := &rbac.Permission{Name: name}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	// 注意：表名要用复数，对应 AuditedTableEnum 常量
	s.aud.Record(ctx, audit2.AuditedTablePermissions, p.ID, "create", *p)
	return p, nil
}

func (s *permService) Get(ctx context.Context, id uint) (*rbac.Permission, error) {
	return s.repo.FindByID(id)
}

func (s *permService) List(ctx context.Context) ([]rbac.Permission, error) {
	return s.repo.FindAll()
}

func (s *permService) Update(ctx context.Context, id uint, name string) (*rbac.Permission, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	before := *p // 先拷贝旧值
	p.Name = name
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	s.aud.Record(ctx, audit2.AuditedTablePermissions, id, "update", before)
	return p, nil
}

func (s *permService) Delete(ctx context.Context, id uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	before := *p
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.aud.Record(ctx, audit2.AuditedTablePermissions, id, "delete", before)
	return nil
}
