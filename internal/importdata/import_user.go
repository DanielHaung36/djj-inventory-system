package importdata

import (
	"djj-inventory-system/internal/logger"
	"fmt"
	"strings"
	"time"

	"djj-inventory-system/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ImportUserData(db *gorm.DB) error {
	// 1. 定义五个用户（用户名中最后一个单词代表角色）
	rawUsers := []string{
		"Nova Manager",
		"Terry Director",
		"Charles Staff",
		"customer1 User",
		"customer2 User",
	}

	// 2. 遍历 rawUsers，构造 User、Role、UserRole 列表
	var (
		users         []models.User
		rolesMap      = map[string]int{}      // 角色名称 -> Role ID
		rolesList     []models.Role           // 去重后的 Role 列表
		userRolesList []models.UserRole       // 用户和角色关联
		rolePermsList []models.RolePermission // 后续可补充权限映射
	)

	const defaultPassword = "qq123456"
	now := time.Now()

	// idxRole 为分配 Role ID 的计数器
	idxRole := 0

	for i, fullName := range rawUsers {
		// 3. 计算用户 ID（从 1 开始）
		userID := i + 1

		// 4. 解析用户名和角色名
		parts := strings.Fields(fullName)
		roleName := parts[len(parts)-1]      // “Manager”、“Director”等
		username := strings.Join(parts, " ") // 整个字符串作为 username

		// 5. 生成 bcrypt 密码哈希
		hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("生成密码哈希失败: %w", err)
		}

		// 6. 创建 User 结构体并追加到 users 列表
		users = append(users, models.User{
			ID:           userID,
			Username:     username,
			PasswordHash: string(hash),
			Version:      1,
			CreatedAt:    now,
			UpdatedAt:    now,
			IsDeleted:    false,
		})

		// 7. 如果 roleName 尚未在 rolesMap 中，先创建一个新的 Role
		if _, exists := rolesMap[roleName]; !exists {
			idxRole++
			rolesMap[roleName] = idxRole
			rolesList = append(rolesList, models.Role{
				ID:   idxRole,
				Name: roleName,
			})
		}

		// 8. 把 userID 与对应的 roleID 关联，追加到 userRolesList
		roleID := rolesMap[roleName]
		userRolesList = append(userRolesList, models.UserRole{
			UserID: userID,
			RoleID: roleID,
		})
	}

	// （可选）9. 如果要预先设置某些 RolePermission，可以在此处填充
	// 例如：
	// rolePermsList = append(rolePermsList, models.RolePermission{RoleID: 1, PermissionID: 101})
	// rolePermsList = append(rolePermsList, models.RolePermission{RoleID: 2, PermissionID: 202})
	// 目前示例不包含具体权限，留空即可

	// 10. 开启事务，将数据插入数据库
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 10.1 插入 Role 表（若存在同名角色，则忽略）
	if len(rolesList) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}}, // 以 id 为冲突判断
			DoNothing: true,
		}).Create(&rolesList).Error; err != nil {
			tx.Rollback()
			logger.Errorf("批量插入角色失败: %w", err)
			return err
		}
	}

	// 10.2 插入 User 表（若存在相同 id 或 username，则忽略）
	if len(users) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "username"}},
			DoNothing: true,
		}).Create(&users).Error; err != nil {
			tx.Rollback()
			logger.Errorf("批量插入用户失败: %w", err)
			return err
		}
	}

	// 10.3 插入 UserRole 中间表（若相同记录已存在，则忽略）
	if len(userRolesList) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
			DoNothing: true,
		}).Create(&userRolesList).Error; err != nil {
			tx.Rollback()
			logger.Errorf("批量插入用户角色关联失败: %w", err)
			return err
		}
	}

	// 10.4 插入 RolePermission（若你有初始化权限，可在此插入；示例为空，跳过）
	if len(rolePermsList) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "role_id"}, {Name: "permission_id"}},
			DoNothing: true,
		}).Create(&rolePermsList).Error; err != nil {
			tx.Rollback()
			logger.Errorf("批量插入角色权限关联失败: %w", err)
			return err
		}
	}

	return tx.Commit().Error
}
