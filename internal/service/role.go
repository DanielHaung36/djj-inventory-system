// internal/service/role.go
package service

import (
	"context"
	audit2 "djj-inventory-system/internal/model/audit"
	"djj-inventory-system/internal/model/rbac"
	"encoding/json"
	"fmt"

	"djj-inventory-system/internal/pkg/audit"
	"djj-inventory-system/internal/repository"
)

type RoleService interface {
	Create(ctx context.Context, name string) (*rbac.Role, error)
	Get(ctx context.Context, id uint) (*rbac.Role, error)
	List(ctx context.Context) ([]rbac.Role, error)
	Update(ctx context.Context, id uint, name string) (*rbac.Role, error)
	Delete(ctx context.Context, id uint) error
	// 新增：
	ListPermissions(ctx context.Context, roleID uint) ([]rbac.Permission, error)
}

type roleService struct {
	repo repository.RoleRepo
	aud  audit.Recorder
}

func NewRoleService(r repository.RoleRepo, aud audit.Recorder) RoleService {
	return &roleService{repo: r, aud: aud}
}

func (s *roleService) Create(ctx context.Context, name string) (*rbac.Role, error) {
	r := &rbac.Role{Name: name}
	if err := s.repo.Create(r); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	// 审计：写入创建前后的快照
	s.aud.Record(ctx, audit2.AuditedTableRoles, r.ID, "create", *r)
	return r, nil
}

func (s *roleService) Get(ctx context.Context, id uint) (*rbac.Role, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get role %d: %w", id, err)
	}
	return r, nil
}

func (s *roleService) List(ctx context.Context) ([]rbac.Role, error) {
	roles, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (s *roleService) Update(ctx context.Context, id uint, name string) (*rbac.Role, error) {
	// 先读取旧值以便审计
	old, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("find role %d: %w", id, err)
	}
	before, _ := json.Marshal(old)

	// 修改并保存
	old.Name = name
	if err := s.repo.Update(old); err != nil {
		return nil, fmt.Errorf("update role %d: %w", id, err)
	}

	// 审计：写入更新前的快照
	s.aud.Record(ctx, audit2.AuditedTableRoles, id, "update", string(before))
	return old, nil
}

func (s *roleService) Delete(ctx context.Context, id uint) error {
	// 先读出旧值审计
	old, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("find role %d: %w", id, err)
	}
	before, _ := json.Marshal(old)

	// 删除
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete role %d: %w", id, err)
	}

	// 审计：写入删除前的快照
	s.aud.Record(ctx, audit2.AuditedTableRoles, id, "delete", string(before))
	return nil
}

func (s *roleService) ListPermissions(ctx context.Context, roleID uint) ([]rbac.Permission, error) {
	perms, err := s.repo.ListPermissions(roleID)
	if err != nil {
		return nil, err
	}
	// 审计也可以放这里，但通常只是查询不审计
	return perms, nil
}
