package main

import (
	"djj-inventory-system/internal/database"
	"djj-inventory-system/internal/importdata"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/models"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

func main() {

	if err := logger.Init("./logs/app.log", zapcore.DebugLevel); err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Infof("djj-inventory-system 启动成功")

	// 1. 先打开 DB
	sqlDB := database.InitDB("djjinventory")
	gormDB := database.InitGormDB(sqlDB)
	file, err := excelize.OpenFile("E:\\learn\\js\\djj inventory system\\djj inventory system server\\cmd\\seed\\product.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	// 2. 依次调用导入
	if err := importdata.ImportUserData(gormDB); err != nil {
		logger.Fatalf("导入用户失败:", err)
	}
	if err := importdata.ImportCategories(gormDB, file); err != nil {
		logger.Fatalf("导入分类失败:", err)
	}

	//if err := importdata.ImportWarehouses(gormDB, file); err != nil {
	//	logger.Fatalf("导入仓库失败:", err)
	//}
	//if err := importdata.ImportProducts(db, "data/products.xlsx"); err != nil {
	//	logger.Fatalf("导入产品失败:", err)
	//}
	//if err := importdata.ImportInventory(db, "data/inventory.xlsx"); err != nil {
	//	logger.Fatalf("导入库存失败:", err)
	//}

	logger.Infof("初始数据导入完成")
}

// SeedUsers 将几个基础帐号写入 users 表
func SeedUsers(db *gorm.DB) error {
	users := []models.User{
		{Username: "Base Assistant", PasswordHash: "<hash>"},
		{Username: "Noah Zhang", PasswordHash: "<hash>"},
		{Username: "Daniel Huang", PasswordHash: "<hash>"},
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoNothing: true,
	}).Create(&users).Error
}
