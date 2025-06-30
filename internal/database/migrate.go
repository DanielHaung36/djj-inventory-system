package database

import (
	"djj-inventory-system/internal/handler"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/model/rbac"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// initPermissions 用 handler.PermissionModules 初始化所有权限，并授予 admin 全权限
func initPermissions(db *gorm.DB) error {
	// 1. 写入所有权限
	var allPerms []rbac.Permission
	for _, mod := range handler.PermissionModules {
		allPerms = append(allPerms, mod.Permissions...)
	}
	for _, p := range allPerms {
		if err := db.FirstOrCreate(&p, rbac.Permission{ID: p.ID}).Error; err != nil {
			return fmt.Errorf("init perm %s: %w", p.Name, err)
		}
	}
	logger.Infof("✔ all business permissions initialized")

	// 2. 授 admin 全权限
	var admin rbac.Role
	if err := db.Where("name = ?", "admin").First(&admin).Error; err != nil {
		return err
	}
	for _, p := range allPerms {
		var perm rbac.Permission
		db.Where("id = ?", p.ID).First(&perm)
		rp := rbac.RolePermission{
			RoleID:       admin.ID,
			PermissionID: perm.ID,
		}
		if err := db.FirstOrCreate(&rp, rp).Error; err != nil {
			return fmt.Errorf("grant %s to admin: %w", p.Name, err)
		}
	}
	logger.Infof("✔ admin granted all permissions")
	return nil
}

// initRoles 会确保有以下这几类角色（含普通和领导）
func initRoles(db *gorm.DB) error {
	roles := []rbac.Role{
		{Name: "admin"},

		{Name: "sales_rep"},
		{Name: "sales_leader"},

		{Name: "purchase_rep"},
		{Name: "purchase_leader"},

		{Name: "operations_staff"},
		{Name: "operations_leader"},

		{Name: "finance_staff"},
		{Name: "finance_leader"},
	}
	for _, r := range roles {
		if err := db.FirstOrCreate(&r, rbac.Role{Name: r.Name}).Error; err != nil {
			return fmt.Errorf("init role %s: %w", r.Name, err)
		}
	}
	return nil
}

// initUsers 初始化账号，并分配角色、批量授予模块权限
func initUsers(db *gorm.DB) error {
	// 准备初始用户（明文密码示例：同用户名 + "123"）
	rawUsers := []struct {
		Username string
		Email    string
		Password string
		Role     string
	}{
		{"admin", "admin@example.com", "admin123", "admin"},

		{"sales_rep", "rep@example.com", "sales_rep123", "sales_rep"},
		{"sales_leader", "leader@example.com", "sales_leader123", "sales_leader"},

		{"purchase_rep", "prep@example.com", "purchase_rep123", "purchase_rep"},
		{"purchase_leader", "pleader@example.com", "purchase_leader123", "purchase_leader"},

		{"operations_staff", "opstaff@example.com", "operations_staff123", "operations_staff"},
		{"operations_leader", "opleader@example.com", "operations_leader123", "operations_leader"},

		{"finance_staff", "finstaff@example.com", "finance_staff123", "finance_staff"},
		{"finance_leader", "finleader@example.com", "finance_leader123", "finance_leader"},
	}

	for _, ru := range rawUsers {
		hash, err := bcrypt.GenerateFromPassword([]byte(ru.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", ru.Username, err)
		}
		// 1) 创建或查找用户
		u := rbac.User{
			Username:     ru.Username,
			Email:        ru.Email,
			PasswordHash: string(hash),
		}
		if err := db.
			Where("username = ?", u.Username).
			FirstOrCreate(&u).Error; err != nil {
			return fmt.Errorf("init user %s: %w", u.Username, err)
		}

		// 2) 绑定角色
		var role rbac.Role
		if err := db.Where("name = ?", ru.Role).First(&role).Error; err != nil {
			return fmt.Errorf("find role %s: %w", ru.Role, err)
		}
		ur := rbac.UserRole{UserID: u.ID, RoleID: role.ID}
		if err := db.FirstOrCreate(&ur, ur).Error; err != nil {
			return fmt.Errorf("assign role %s to user %s: %w", ru.Role, u.Username, err)
		}

		// 3) 给 leader 角色额外全部本模块权限
		if strings.Contains(ru.Role, "sale") {
			if err := grantGroup(db, "quotes", ru.Role); err != nil {
				return err
			}
			if err := grantGroup(db, "orders", ru.Role); err != nil {
				return err
			}
		}
		if strings.Contains(ru.Role, "purchase") {
			if err := grantGroup(db, "orders", ru.Role); err != nil {
				return err
			}
		}
		if strings.Contains(ru.Role, "operations") {
			if err := grantGroup(db, "inventory", ru.Role); err != nil {
				return err
			}
		}
		if strings.Contains(ru.Role, "finance") {
			// Finance leader = admin 下的所有模块，仅次于 admin
			for _, mod := range handler.PermissionModules {
				if err := grantGroup(db, mod.Module, ru.Role); err != nil {
					return err
				}
			}
		}
	}

	logger.Infof("✔ users and user_roles initialized")
	return nil
}

// grantGroup 给某个角色(roleID)批量挂上某个模块(groupName)的所有权限
func grantGroup(db *gorm.DB, groupName, roleName string) error {
	// 先把角色查出来
	var role rbac.Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return fmt.Errorf("find role %s: %w", roleName, err)
	}

	var perms []rbac.Permission
	found := false
	for _, mod := range handler.PermissionModules {
		if mod.Module == groupName {
			perms = mod.Permissions
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unknown permission group %s", groupName)
	}
	for _, perm := range perms {
		var p rbac.Permission
		if err := db.Where("id = ?", perm.ID).First(&p).Error; err != nil {
			return fmt.Errorf("find perm %s: %w", perm.Name, err)
		}
		rp := rbac.RolePermission{
			RoleID:       role.ID,
			PermissionID: p.ID,
		}
		if err := db.FirstOrCreate(&rp, rp).Error; err != nil {
			return fmt.Errorf("grant %s to %s: %w", perm.Name, roleName, err)
		}
	}
	return nil
}
