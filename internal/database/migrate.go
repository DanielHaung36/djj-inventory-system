package database

import (
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/model"
	"fmt"

	"gorm.io/gorm"
)

// 各模块的 CRUD 权限组
var permissionGroups = map[string][]string{
	"users":       {"users.create", "users.read", "users.update", "users.delete"},
	"roles":       {"roles.create", "roles.read", "roles.update", "roles.delete"},
	"permissions": {"permissions.create", "permissions.read", "permissions.update", "permissions.delete"},
	"products":    {"products.create", "products.read", "products.update", "products.delete"},
	"quotes":      {"quotes.create", "quotes.read", "quotes.update", "quotes.delete"},
	"orders":      {"orders.create", "orders.read", "orders.update", "orders.delete"},
	"inventory":   {"inventory.create", "inventory.read", "inventory.update", "inventory.delete"},
}

// initPermissions 写入所有 CRUD 权限，并把所有权限都授给 admin
func initPermissions(db *gorm.DB) error {
	// 写入权限
	for _, perms := range permissionGroups {
		for _, name := range perms {
			p := model.Permission{Name: name}
			if err := db.
				FirstOrCreate(&p, model.Permission{Name: name}).
				Error; err != nil {
				return fmt.Errorf("init perm %s: %w", name, err)
			}
		}
	}
	logger.Infof("✔ permissions initialized")

	// 授 admin 全权限
	var admin model.Role
	if err := db.Where("name = ?", "admin").First(&admin).Error; err != nil {
		return err
	}
	for _, perms := range permissionGroups {
		for _, name := range perms {
			var p model.Permission
			db.Where("name = ?", name).First(&p)
			rp := model.RolePermission{
				RoleID:       admin.ID,
				PermissionID: p.ID,
			}
			if err := db.FirstOrCreate(&rp, rp).Error; err != nil {
				return fmt.Errorf("grant %s to admin: %w", name, err)
			}
		}
	}
	logger.Infof("✔ admin granted all permissions")
	return nil
}

// initRoles 会确保有以下这几类角色（含普通和领导）
func initRoles(db *gorm.DB) error {
	roles := []model.Role{
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
		if err := db.FirstOrCreate(&r, model.Role{Name: r.Name}).Error; err != nil {
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
		// 1) 创建或查找用户
		u := model.User{
			Username: ru.Username,
			// TODO: 这里用实际的 bcrypt.Hash(ru.Password)
			Email: ru.Email,

			PasswordHash: ru.Password,
		}
		if err := db.
			Where("username = ?", u.Username).
			FirstOrCreate(&u).Error; err != nil {
			return fmt.Errorf("init user %s: %w", u.Username, err)
		}

		// 2) 绑定角色
		var role model.Role
		if err := db.Where("name = ?", ru.Role).First(&role).Error; err != nil {
			return fmt.Errorf("find role %s: %w", ru.Role, err)
		}
		ur := model.UserRole{UserID: u.ID, RoleID: role.ID}
		if err := db.FirstOrCreate(&ur, ur).Error; err != nil {
			return fmt.Errorf("assign role %s to user %s: %w", ru.Role, u.Username, err)
		}

		// 3) 给 leader 角色额外全部本模块权限
		if ru.Role == "sales_leader" {
			if err := grantGroup(db, "quotes", role.ID); err != nil {
				return err
			}
			if err := grantGroup(db, "orders", role.ID); err != nil {
				return err
			}
		}
		if ru.Role == "purchase_leader" {
			if err := grantGroup(db, "orders", role.ID); err != nil {
				return err
			}
		}
		if ru.Role == "operations_leader" {
			if err := grantGroup(db, "inventory", role.ID); err != nil {
				return err
			}
		}
		if ru.Role == "finance_leader" {
			// Finance leader = admin 下的所有模块，仅次于 admin
			for grp := range permissionGroups {
				if err := grantGroup(db, grp, role.ID); err != nil {
					return err
				}
			}
		}
	}

	logger.Infof("✔ users and user_roles initialized")
	return nil
}

// grantGroup 给某个角色(roleID)批量挂上某个模块(groupName)的所有权限
func grantGroup(db *gorm.DB, groupName string, roleID uint) error {
	perms, ok := permissionGroups[groupName]
	if !ok {
		return fmt.Errorf("unknown permission group %s", groupName)
	}
	for _, name := range perms {
		var p model.Permission
		if err := db.Where("name = ?", name).First(&p).Error; err != nil {
			return fmt.Errorf("find perm %s: %w", name, err)
		}
		rp := model.RolePermission{RoleID: roleID, PermissionID: p.ID}
		if err := db.FirstOrCreate(&rp, rp).Error; err != nil {
			return fmt.Errorf("grant %s to role %d: %w", name, roleID, err)
		}
	}
	return nil
}
