package database

import (
	"database/sql"
	"djj-inventory-system/internal/logger"
	"fmt"
	_ "github.com/lib/pq" // <------------ here
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

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

func Migrate() {

}
