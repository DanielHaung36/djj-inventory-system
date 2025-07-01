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
		u := rbac.User{
			Username:     ru.Username,
			Email:        ru.Email,
			PasswordHash: string(hash),
			StoreID:      1, // 或者按规则分配
		}
		if err := db.Where("username = ?", u.Username).
			FirstOrCreate(&u).Error; err != nil {
			return fmt.Errorf("init user %s: %w", u.Username, err)
		}
	}
	logger.Infof("✔ users initialized")
	return nil
}

func assignRoles(db *gorm.DB) error {
	// 1) 先加载所有用户
	var users []rbac.User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("fetch users: %w", err)
	}

	for _, u := range users {
		var roleName string
		// 不再用 switch 精确匹配，而是前缀/包含匹配
		switch {
		case u.Username == "admin":
			roleName = "admin"
		case strings.HasPrefix(u.Username, "sales_leader"):
			roleName = "sales_leader"
		case strings.HasPrefix(u.Username, "sales_rep"):
			roleName = "sales_rep"
		case strings.HasPrefix(u.Username, "purchase_leader"):
			roleName = "purchase_leader"
		case strings.HasPrefix(u.Username, "purchase_rep"):
			roleName = "purchase_rep"
		case strings.HasPrefix(u.Username, "operations_leader"):
			roleName = "operations_leader"
		case strings.HasPrefix(u.Username, "operations_staff"):
			roleName = "operations_staff"
		case strings.HasPrefix(u.Username, "finance_leader"):
			roleName = "finance_leader"
		case strings.HasPrefix(u.Username, "finance_staff"):
			roleName = "finance_staff"
		default:
			// 如果不是我们关心的角色，就跳过
			continue
		}

		// 查出角色实体
		var role rbac.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			return fmt.Errorf("find role %s: %w", roleName, err)
		}

		// 绑到 user_roles
		ur := rbac.UserRole{UserID: u.ID, RoleID: role.ID}
		if err := db.FirstOrCreate(&ur, ur).Error; err != nil {
			return fmt.Errorf("assign role %s to user %s: %w", roleName, u.Username, err)
		}
	}

	logger.Infof("✔ roles assigned to users")
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
		if mod.Key == groupName {
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

// AutoGrantPermissions 遍历现有用户，根据绑定的角色授予对应的模块权限
func AutoGrantPermissions(db *gorm.DB) error {
	// 1) 取出所有用户
	var users []rbac.User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("fetch users: %w", err)
	}

	// 2) 遍历每个用户
	for _, u := range users {
		// 查出这个用户所有角色关联
		var urs []rbac.UserRole
		if err := db.Where("user_id = ?", u.ID).Find(&urs).Error; err != nil {
			return fmt.Errorf("fetch roles for user %d: %w", u.ID, err)
		}

		for _, ur := range urs {
			// 查出角色实体
			var role rbac.Role
			if err := db.First(&role, ur.RoleID).Error; err != nil {
				return fmt.Errorf("load role %d: %w", ur.RoleID, err)
			}

			// 根据角色名决定要给哪些权限组
			switch role.Name {
			case "sales_leader", "sales_rep":
				// 不论是 leader 还是 rep，都给 quote 和 sales 模块的权限
				if err := grantGroup(db, "quote", role.Name); err != nil {
					return err
				}
				if err := grantGroup(db, "sales", role.Name); err != nil {
					return err
				}

			case "purchase_leader", "purchase_rep":
				// leader/rep 一样，给 sales 模块权限
				if err := grantGroup(db, "sales", role.Name); err != nil {
					return err
				}

			case "operations_leader", "operations_staff":
				// leader/staff 一样，给 inventory 模块权限
				if err := grantGroup(db, "inventory", role.Name); err != nil {
					return err
				}

			case "finance_leader", "finance_staff":
				// leader/staff 一样，给 finance 模块权限
				if err := grantGroup(db, "finance", role.Name); err != nil {
					return err
				}
			}
		}
	}

	logger.Infof("✔ all existing users granted permissions according to their roles")
	return nil
}
