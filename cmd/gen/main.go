package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
	"log"
)

func init() {

}

func main() {
	// 1. 打开底层 *sql.DB（假设你已经有了 db *sql.DB）
	//    这里我们直接用 gorm.Open 演示
	dsn := "host=127.0.0.1 user=djj password=qq123456 dbname=djjinventory port=5432 sslmode=disable"
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	// 2. 从 pg_catalog 拿到所有表名
	var tables []string
	if err := gormDB.Raw(`
        SELECT tablename
          FROM pg_catalog.pg_tables
         WHERE schemaname = 'public'
    `).Scan(&tables).Error; err != nil {
		log.Fatalf("query tables failed: %v", err)
	}
	fmt.Printf("found tables: %v\n", tables)
	//gen.NewGenerator(gen.Config{})
	// 3. 初始化 gen
	g := gen.NewGenerator(gen.Config{
		OutPath:      "./internal/models",
		ModelPkgPath: "github.com/youruser/djj-inventory-system/internal/models",
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})
	//g := gen.NewGenerator(gen.Config{
	//	OutPath: "./internal/models",
	//})

	g.UseDB(gormDB)

	// generate all table from database
	g.ApplyBasic(g.GenerateAllTable()...)

	g.Execute()

}
