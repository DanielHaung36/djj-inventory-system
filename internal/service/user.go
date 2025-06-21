// internal/service/user.go
package service

import (
	"context"
	_ "encoding/json"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	_ "gorm.io/gorm"

	"djj-inventory-system/internal/model"
	"djj-inventory-system/internal/pkg/audit"
	"djj-inventory-system/internal/repository"
)

type UserService interface {
	Create(ctx context.Context, username, email, password string, roleIDs []uint) (*model.User, error)
	Get(ctx context.Context, id uint) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, id uint, email, password *string) (*model.User, error)
	Delete(ctx context.Context, id uint) error
	Authenticate(ctx context.Context, username, password string) (*model.User, error)
	AssignRole(ctx context.Context, userID, roleID uint) error
	RemoveRole(ctx context.Context, userID, roleID uint) error
	ListRoles(ctx context.Context, userID uint) ([]model.Role, error)

	// 用户直接权限管理
	GrantUserPermissions(ctx context.Context, userID uint, permIDs []uint) error
	RevokeUserPermissions(ctx context.Context, userID uint, permIDs []uint) error

	// 获取合并后的所有权限（角色继承 + 直接赋予）
	GetWithAllPermissions(ctx context.Context, userID uint) (*model.User, error)
}

type userService struct {
	repo repository.UserRepo
	aud  audit.Recorder
}

func NewUserService(r repository.UserRepo, aud audit.Recorder) UserService {
	return &userService{repo: r, aud: aud}
}

// ---- 实现 Authenticate ----
func (s *userService) Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	// 1) 根据用户名查用户
	u, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	// 2) 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	// 3. 载入这个用户的角色列表
	roles, err := s.repo.ListRoles(u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	// 4. 载入所有这些角色对应的权限
	for _, r := range roles {
		rps, err := s.repo.ListRolePermissions(r.ID)
		if err != nil {
			return nil, err
		}
		for _, p := range rps {
			u.Permissions = append(u.Permissions, p)
		}
	}

	return u, nil
}

func (s *userService) Create(ctx context.Context, username, email, password string, roleIDs []uint) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Version:      1,
	}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	// 分配角色
	for _, rid := range roleIDs {
		if err := s.repo.AddRole(u.ID, rid); err != nil {
			return nil, err
		}
	}
	//分配权限

	// 审计
	s.aud.Record(ctx, model.AuditedTableUsers, u.ID, "create", *u)
	return u, nil
}

func (s *userService) Get(ctx context.Context, id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *userService) List(ctx context.Context) ([]model.User, error) {
	return s.repo.FindAll()
}

func (s *userService) Update(ctx context.Context, id uint, email, password *string) (*model.User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 拷贝旧值用于审计
	before := *u

	if email != nil {
		u.Email = *email
	}
	if password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		u.PasswordHash = string(hash)
	}
	if err := s.repo.Update(u); err != nil {
		return nil, err
	}
	s.aud.Record(ctx, model.AuditedTableUsers, u.ID, "update", before)
	return u, nil
}

func (s *userService) Delete(ctx context.Context, id uint) error {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	before := *u
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.aud.Record(ctx, model.AuditedTableUsers, id, "delete", before)
	return nil
}

func (s *userService) AssignRole(ctx context.Context, userID, roleID uint) error {
	if err := s.repo.AddRole(userID, roleID); err != nil {
		return err
	}
	// 审计：记录 user_role 表的变更
	s.aud.Record(ctx, model.AuditedTableUserRoles, userID, "assign_role", map[string]uint{"role_id": roleID})
	return nil
}

func (s *userService) RemoveRole(ctx context.Context, userID, roleID uint) error {
	if err := s.repo.RemoveRole(userID, roleID); err != nil {
		return err
	}
	s.aud.Record(ctx, model.AuditedTableUserRoles, userID, "remove_role", map[string]uint{"role_id": roleID})
	return nil
}

func (s *userService) ListRoles(ctx context.Context, userID uint) ([]model.Role, error) {
	return s.repo.ListRoles(userID)
}

// GrantUserPermissions 批量授予用户直接权限
func (s *userService) GrantUserPermissions(ctx context.Context, userID uint, permIDs []uint) error {
	if err := s.repo.GrantUserPermissions(userID, permIDs); err != nil {
		return err
	}
	s.aud.Record(ctx, model.AuditedTableUserRoles, userID,
		"grant_user_permissions", map[string]interface{}{"perm_ids": permIDs})
	return nil
}

// RevokeUserPermissions 批量撤销用户直接权限
func (s *userService) RevokeUserPermissions(ctx context.Context, userID uint, permIDs []uint) error {
	if err := s.repo.RevokeUserPermissions(userID, permIDs); err != nil {
		return err
	}
	s.aud.Record(ctx, model.AuditedTableUserRoles, userID,
		"revoke_user_permissions", map[string]interface{}{"perm_ids": permIDs})
	return nil
}

// GetWithAllPermissions 获取用户角色继承 + 直接赋予后的全部权限
func (s *userService) GetWithAllPermissions(ctx context.Context, userID uint) (*model.User, error) {
	// repo 层预加载了 Roles.Permissions 和 DirectPermissions，并做了扁平去重
	u, err := s.repo.FindWithAllPerms(userID)
	if err != nil {
		return nil, err
	}
	return u, nil
}
