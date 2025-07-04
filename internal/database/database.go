package database

import (
	"database/sql"
	"djj-inventory-system/config"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/model/audit"
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/company"
	"djj-inventory-system/internal/model/inventory"
	"djj-inventory-system/internal/model/rbac"
	"djj-inventory-system/internal/model/sales"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	_ "github.com/lib/pq" // <------------ here
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	logv1 "gorm.io/gorm/logger"
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
	config.Load()
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
	// 自定义一个 logger，打印出 INFO 级别以上的所有 SQL，并显示执行时间
	newLogger := logv1.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io.Writer，这里打印到 stdout
		logv1.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
			LogLevel:                  logv1.Info,             // 这里设为 Info 或者 Debug
			IgnoreRecordNotFoundError: true,                   // 忽略 ErrRecordNotFound
			Colorful:                  false,                  // 关闭彩色输出
		},
	)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		logger.Fatalf("使用 GORM 连接数据库失败: ", err)

	}
	logger.Infof("成功使用 GORM 连接到数据库")
	return gormDB
}

func Migrate(db *gorm.DB) {

	// 在 AutoMigrate 之前
	stmt := `
			DO $$
			BEGIN
			  -- 产品状态枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'product_status_enum') THEN
				CREATE TYPE product_status_enum AS ENUM (
				  'draft',
				  'pending_tech',
				  'pending_purchase',
				  'pending_finance',
				  'ready_published',
				  'published',
				  'rejected',
				  'closed'
				);
			  END IF;
		   -- 审批流程状态枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'application_status_enum') THEN
				CREATE TYPE application_status_enum AS ENUM (
				  'open',    -- 审批中
				  'closed'   -- 已结束
				);
			  END IF;
			  -- 货物性质枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'goods_nature_enum') THEN
				CREATE TYPE goods_nature_enum AS ENUM (
				  'contract',
				  'multi_contract',
				  'partial_contract',
				  'warranty',
				  'gift',
				  'self_purchased',
				  'consignment'
				);
			  END IF;
			
			  -- 审批结果枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'approval_status_enum') THEN
				CREATE TYPE approval_status_enum AS ENUM (
				  'pending',
				  'approved',
				  'rejected'
				);
			  END IF;
			
			  -- 库存状态枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'stock_status_enum') THEN
				CREATE TYPE stock_status_enum AS ENUM (
				  'pending',
				  'in_stock',
				  'not_applicable'
				);
			  END IF;
			
			  -- 订单状态枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status_enum') THEN
				CREATE TYPE order_status_enum AS ENUM (
				  'draft',
				  'ordered',
				  'deposit_received',
				  'final_payment_received',
				  'pre_delivery_inspection',
				  'shipped',
				  'delivered',
				  'order_closed',
				  'cancelled'
				);
			  END IF;
			
			  -- 币种代码枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'currency_code_enum') THEN
				CREATE TYPE currency_code_enum AS ENUM (
				  'AUD',
				  'USD',
				  'CNY',
				  'EUR',
				  'GBP'
				);
			  END IF;
			
			  -- 客户类型枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'customer_type_enum') THEN
				CREATE TYPE customer_type_enum AS ENUM (
				  'retail',
				  'wholesale',
				  'online'
				);
			  END IF;
			
			  -- 订单类型枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_type_enum') THEN
				CREATE TYPE order_type_enum AS ENUM (
				  'purchase',
				  'sales'
				);
			  END IF;
              -- 客户类型枚举
			  IF NOT EXISTS (
				SELECT 1 FROM pg_type WHERE typname = 'customer_type_enum'
			  ) THEN
				CREATE TYPE customer_type_enum AS ENUM (
				  'retail',    -- 零售
				  'wholesale', -- 批发
				  'online'     -- 电商
				);
			  END IF;
			  -- 产品主分类枚举
			  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'product_type_enum') THEN
				CREATE TYPE product_type_enum AS ENUM (
				  'machine',
				  'parts',
				  'attachment',
				  'tools',
				  'others'
				);
			  END IF;
			END$$;
			`
	if err := db.Exec(stmt).Error; err != nil {
		log.Fatalf("failed to ensure currency_code_enum exists: %v", err)
	}

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// v1: 初始的 RBAC 表
		{
			ID: "20250627_add_customers",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&catalog.Customer{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("customers")
			},
		},
		{
			ID: "20250701_add_companies",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&company.Company{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("customers")
			},
		},
		{
			ID: "20250701_add_catalog",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&catalog.Store{},
					&catalog.Region{},
					&catalog.Warehouse{},
					&catalog.RegionWarehouse{}, // ← 确保在这里
				)
			}},
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
					&sales.PickingList{}, &sales.PickingListItem{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					"picking_list_items", "picking_lists",
				)
			},
		},
		{
			ID: "20250704_add_attatchment,stock",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&catalog.Attachment{}, &catalog.ProductStock{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					"attachments", "product_stocks",
				)
			},
		},
		{
			ID: "20250704_add_images",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&catalog.ProductImage{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					"product_images",
				)
			},
		},
		{
			ID: "20250705_add_transaction_type_and_inventory_transaction",
			Migrate: func(tx *gorm.DB) error {
				// 1. 创建 PostgreSQL enum 类型
				if err := tx.Exec(`
            DO $$
            BEGIN
              IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_type') THEN
                CREATE TYPE transaction_type AS ENUM ('IN', 'OUT', 'SALE');
              END IF;
            END$$;
        `).Error; err != nil {
					return err
				}

				// 2. 先保证 product_stocks 表已存在（如果还没加 on_hand/reserved/generated 列也可以一并写在这里）
				if err := tx.AutoMigrate(&catalog.ProductStock{}); err != nil {
					return err
				}

				// 3. 创建 inventory_transaction 表
				return tx.AutoMigrate(&inventory.InventoryTransaction{})
			},
			Rollback: func(tx *gorm.DB) error {
				// 先删表
				if err := tx.Migrator().DropTable("inventory_transaction"); err != nil {
					return err
				}
				// 如果你之前也想同时撤销 product_stocks，那这儿也可以 DropTable
				// 然后删 enum 类型
				return tx.Exec(`DROP TYPE IF EXISTS transaction_type;`).Error
			},
		},
		// 在你 migrate.go 的 migrations 列表里，追加一段：
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
	////初始化RBAC
	if err := initRoles(db); err != nil {
		logger.Fatalf("❌ initRoles: %v", err)
	}
	if err := initPermissions(db); err != nil {
		logger.Fatalf("❌ initPermissions: %v", err)
	}
	if err := NewSeeder(db).Run(); err != nil {
		logger.Fatalf("❌ 测试数据种子失败: %v", err)
	}
	InitRBACSeed(db)

}

// InitRBACSeed 全量种子：角色、权限、用户、角色-权限关联
func InitRBACSeed(db *gorm.DB) {
	if err := initUsers(db); err != nil {
		logger.Fatalf("❌ initUsers: %v", err)
	}
	if err := assignRoles(db); err != nil {
		logger.Fatalf("❌ assignRoles: %v", err)
	}
	if err := AutoGrantPermissions(db); err != nil {
		logger.Fatalf("❌ AutoGrantPermissions: %v", err)
	}
	logger.Infof("🎉 database seeding completed")
}
