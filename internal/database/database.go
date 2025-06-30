package database

import (
	"database/sql"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/model"
	"djj-inventory-system/internal/model/audit"
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/rbac"
	"djj-inventory-system/internal/model/sales"
	"fmt"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	_ "github.com/lib/pq" // <------------ here
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	// 1. 先打开 DB
	sqlDB := InitDB("djjinventory")
	gormDB := InitGormDB(sqlDB)
	_, err := excelize.OpenFile("/mnt/a/code/go/djj-inventory-system/cmd/seed/product.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	Migrate(gormDB)
	return gormDB
}
func InitDB(dbName string) *sql.DB {
	// 连接到目标数据库
	connStrTarget := fmt.Sprintf("host=localhost user=djj password=qq123456 dbname=%s sslmode=disable", dbName)
	dbTarget, err := sql.Open("postgres", connStrTarget)
	if err != nil {
		logger.Fatalf("fail to connect to the %s", dbName, err.Error())
	}
	err = dbTarget.Ping()
	if err != nil {
		log.Fatal(err)
	}
	logger.Infof("Connecting to database %s", dbName)
	return dbTarget
}

func InitGormDB(db *sql.DB) *gorm.DB {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		logger.Fatalf("使用 GORM 连接数据库失败: ", err)

	}
	logger.Infof("成功使用 GORM 连接到数据库")
	return gormDB
}

func Migrate(db *gorm.DB) {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// v1: 初始的 RBAC 表
		{
			ID: "20250611_init_rbac",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&rbac.User{}, &rbac.Role{}, &rbac.Permission{}, &rbac.UserRole{}, &rbac.RolePermission{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("user_roles", "role_permissions", "permissions", "roles", "users")
			},
		},
		{
			ID: "20250611_add_audit_history",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&audit.AuditedHistory{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("audit_histories")
			},
		},
		// v3: 产品表
		{
			ID: "20250625_add_products",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&catalog.Product{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("products")
			},
		},
		// v4: 报价单 + 报价明细
		{
			ID: "20250625_add_quotes",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&sales.Quote{}, &sales.QuoteItem{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					"quote_items", "quotes",
				)
			},
		},
		// v5: 订单 + 订单明细
		{
			ID: "20250625_add_orders",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&sales.Order{}, &sales.OrderItem{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					"order_items", "orders",
				)
			},
		},
		// v6: 拣货单 + 拣货单明细
		{
			ID: "20250625_add_picking_lists",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&sales.PickingList{}, &model.PickingListItem{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					"picking_list_items", "picking_lists",
				)
			},
		},
		{
			ID: "20250627_add_customers",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&catalog.Customer{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("customers")
			},
		},
		//{
		//	ID: "20250611_add_deleted_at_to_users",
		//	Migrate: func(tx *gorm.DB) error {
		//		// 会执行：ALTER TABLE users ADD COLUMN deleted_at timestamptz NULL;
		//		return tx.Migrator().AddColumn(&model.User{}, "DeletedAt")
		//	},
		//	Rollback: func(tx *gorm.DB) error {
		//		// 回滚时删掉这一列
		//		return tx.Migrator().DropColumn(&model.User{}, "DeletedAt")
		//	},
		//},
		//// v2: 产品、报价
		//{
		//	ID: "20250201_add_product_quote",
		//	Migrate: func(tx *gorm.DB) error {
		//		return tx.AutoMigrate(&Product{}, &Quote{}, &QuoteItem{})
		//	},
		//	Rollback: func(tx *gorm.DB) error {
		//		return tx.Migrator().DropTable("quote_items", "quotes", "products")
		//	},
		//},
		//
		//// ★ 如果你下一步要加“订单”表，就在这里插一段：
		//{
		//	ID: "20250315_add_order_orderitem",
		//	Migrate: func(tx *gorm.DB) error {
		//		// 顺序不重要，gorm 会按 ID 升序执行
		//		return tx.AutoMigrate(&Order{}, &OrderItem{})
		//	},
		//	Rollback: func(tx *gorm.DB) error {
		//		return tx.Migrator().DropTable("order_items", "orders")
		//	},
		//},

		// 再下次要加“库存”表，就继续追加一个新的 Migration{…}
	})

	if err := m.Migrate(); err != nil {
		log.Fatalf("could not migrate: %v", err)
	}
	//初始化RBAC
	InitRBACSeed(db)
	if err := SeedTestData(db); err != nil {
		logger.Fatalf("❌ 测试数据种子失败: %v", err)
	}

}

// InitRBACSeed 全量种子：角色、权限、用户、角色-权限关联
func InitRBACSeed(db *gorm.DB) {
	if err := initRoles(db); err != nil {
		logger.Fatalf("❌ initRoles: %v", err)
	}
	if err := initPermissions(db); err != nil {
		logger.Fatalf("❌ initPermissions: %v", err)
	}
	if err := initUsers(db); err != nil {
		logger.Fatalf("❌ initUsers: %v", err)
	}
	logger.Infof("🎉 database seeding completed")
}
