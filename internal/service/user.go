// internal/service/user.go
package service

import (
	"context"
	audit2 "djj-inventory-system/internal/model/audit"
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/rbac"
	_ "encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "gorm.io/gorm"

	"djj-inventory-system/internal/pkg/audit"
	"djj-inventory-system/internal/repository"
)

type UserService interface {
	Create(ctx context.Context, username, email, password string, roleNames []string) (*rbac.User, error)
	Get(ctx context.Context, id uint) (*rbac.User, error)
	List(ctx context.Context) ([]rbac.User, error)
	Update(ctx context.Context, id uint, email, password *string) (*rbac.User, error)
	Delete(ctx context.Context, id uint) error
	Authenticate(ctx context.Context, username, password string) (*rbac.User, *catalog.StoreDetails, error)
	AssignRole(ctx context.Context, userID, roleID uint) error
	RemoveRole(ctx context.Context, userID, roleID uint) error
	ListRoles(ctx context.Context, userID uint) ([]rbac.Role, error)

	// 用户直接权限管理
	GrantUserPermissions(ctx context.Context, userID uint, permIDs []uint) error
	RevokeUserPermissions(ctx context.Context, userID uint, permIDs []uint) error

	// 获取合并后的所有权限（角色继承 + 直接赋予）
	GetWithAllPermissions(ctx context.Context, userID uint) (*rbac.User, error)

	// 获取用户权限及最近的修改信息
	GetUserPermissionData(ctx context.Context, userID uint) (*UserPermissionData, error)
}

// UserPermissionData 包含权限ID及最近的修改信息
type UserPermissionData struct {
	UserID        uint
	PermissionIDs []uint
	LastModified  time.Time
	ModifiedBy    string
}

type userService struct {
	repo repository.UserRepo
	aud  audit.Recorder
}

func NewUserService(r repository.UserRepo, aud audit.Recorder) UserService {
	return &userService{repo: r, aud: aud}
}

// ---- 实现 Authenticate ----
func (s *userService) Authenticate(ctx context.Context, username, password string) (*rbac.User, *catalog.StoreDetails, error) {
	// 1) 根据用户名查用户
	u, err := s.repo.FindByEmail(username)
	if err != nil {
		return nil, nil, err
	}
	// 2) 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}
	// 3. 载入这个用户的角色列表
	//roles, err := s.repo.ListRoles(u.ID)
	//if err != nil {
	//	return nil, err
	//}
	//u.Roles = roles
	// 4. 载入所有这些角色对应的权限
	for _, r := range u.Roles {
		rps, err := s.repo.ListRolePermissions(ctx, r.ID)
		if err != nil {
			return nil, nil, err
		}
		for _, p := range rps {
			u.Permissions = append(u.Permissions, p)
		}
	}

	// 5) 载入用户的直接权限
	direct, err := s.repo.ListUserDirectPermissions(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}
	u.Permissions = append(u.Permissions, direct...)

	// 6) 去重
	seen := make(map[uint]struct{}, len(u.Permissions))
	dedup := make([]rbac.Permission, 0, len(u.Permissions))
	for _, p := range u.Permissions {
		if _, exists := seen[p.ID]; !exists {
			seen[p.ID] = struct{}{}
			dedup = append(dedup, p)
		}
	}
	u.Permissions = dedup
	// 6) 查门店＋区域＋公司
	storeDetails, err := s.repo.GetStoreFullDetails(ctx, u.StoreID)
	if err != nil {
		return nil, nil, fmt.Errorf("could not load store details: %w", err)
	}

	return u, storeDetails, nil
}

func (s *userService) Create(ctx context.Context, username, email, password string, roleNames []string) (*rbac.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &rbac.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Version:      1,
	}

	// 用新加的 CreateWithRoles 一步完成创建 + 关联
	if err := s.repo.CreateWithRoles(ctx, u, roleNames); err != nil {
		return nil, err
	}

	// 4. 根据刚关联的角色，批量拉取它们的权限并赋予给用户
	//    （这里用 repo.ListRolePermissions）
	rolePerms, err := s.repo.ListRolePermissions(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if len(rolePerms) > 0 {
		permIDs := make([]uint, len(rolePerms))
		for i, p := range rolePerms {
			permIDs[i] = p.ID
		}
		if err := s.repo.GrantUserPermissions(u.ID, permIDs); err != nil {
			return nil, err
		}
	}
	s.aud.Record(ctx, "users", u.ID, "create", *u)
	return s.repo.FindByID(ctx, u.ID)
}

func (s *userService) Get(ctx context.Context, id uint) (*rbac.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *userService) List(ctx context.Context) ([]rbac.User, error) {
	return s.repo.FindAll()
}

func (s *userService) Update(ctx context.Context, id uint, email, password *string) (*rbac.User, error) {
	u, err := s.repo.FindByID(ctx, id)
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
	s.aud.Record(ctx, audit2.AuditedTableUsers, u.ID, "update", before)
	return u, nil
}

func (s *userService) Delete(ctx context.Context, id uint) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	before := *u
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.aud.Record(ctx, audit2.AuditedTableUsers, id, "delete", before)
	return nil
}

func (s *userService) AssignRole(ctx context.Context, userID, roleID uint) error {
	if err := s.repo.AddRole(userID, roleID); err != nil {
		return err
	}
	// 审计：记录 user_role 表的变更
	s.aud.Record(ctx, audit2.AuditedTableUserRoles, userID, "assign_role", map[string]uint{"role_id": roleID})
	return nil
}

func (s *userService) RemoveRole(ctx context.Context, userID, roleID uint) error {
	if err := s.repo.RemoveRole(userID, roleID); err != nil {
		return err
	}
	s.aud.Record(ctx, audit2.AuditedTableUserRoles, userID, "remove_role", map[string]uint{"role_id": roleID})
	return nil
}

func (s *userService) ListRoles(ctx context.Context, userID uint) ([]rbac.Role, error) {
	return s.repo.ListRoles(userID)
}

// GrantUserPermissions 批量授予用户直接权限
func (s *userService) GrantUserPermissions(ctx context.Context, userID uint, permIDs []uint) error {
	if err := s.repo.GrantUserPermissions(userID, permIDs); err != nil {
		return err
	}
	s.aud.Record(ctx, audit2.AuditedTableUserRoles, userID,
		"grant_user_permissions", map[string]interface{}{"perm_ids": permIDs})
	return nil
}

// RevokeUserPermissions 批量撤销用户直接权限
func (s *userService) RevokeUserPermissions(ctx context.Context, userID uint, permIDs []uint) error {
	if err := s.repo.RevokeUserPermissions(userID, permIDs); err != nil {
		return err
	}
	s.aud.Record(ctx, audit2.AuditedTableUserRoles, userID,
		"revoke_user_permissions", map[string]interface{}{"perm_ids": permIDs})
	return nil
}

// GetWithAllPermissions 获取用户角色继承 + 直接赋予后的全部权限
func (s *userService) GetWithAllPermissions(ctx context.Context, userID uint) (*rbac.User, error) {
	// repo 层预加载了 Roles.Permissions 和 DirectPermissions，并做了扁平去重
	u, err := s.repo.FindWithAllPerms(userID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserPermissionData 返回用户权限及最后的修改记录
func (s *userService) GetUserPermissionData(ctx context.Context, userID uint) (*UserPermissionData, error) {
	u, err := s.repo.FindWithAllPerms(userID)
	if err != nil {
		return nil, err
	}
	permIDs := make([]uint, 0, len(u.Permissions))
	for _, p := range u.Permissions {
		permIDs = append(permIDs, p.ID)
	}

	var last time.Time
	var by string
	if ah, err := s.repo.LastPermissionChange(userID); err == nil && ah != nil {
		last = ah.ChangedAt
		if modUser, err := s.repo.FindByID(ctx, uint(ah.ChangedBy)); err == nil && modUser != nil {
			by = modUser.Username
		}
	}

	return &UserPermissionData{
		UserID:        userID,
		PermissionIDs: permIDs,
		LastModified:  last,
		ModifiedBy:    by,
	}, nil
}
