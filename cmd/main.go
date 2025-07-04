package main

import (
	"djj-inventory-system/config"
	"djj-inventory-system/internal/database"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/pkg/setup"
	"fmt"

	"go.uber.org/zap/zapcore"
)

// @title           DJJ Inventory System API
// @version         1.0
// @description     用户注册 / 登录 / RBAC 管理
// @host            localhost:8080
// @BasePath        /api
func main() {
	// initialize logging
	config.Load()
	if err := logger.Init("./logs/app.log", zapcore.DebugLevel); err != nil {
		panic(err)
	}
	defer logger.Sync()
	// connect to DB
	db := database.Connect()
	router := setup.NewRouter(db)
	addr := fmt.Sprintf("%s:%s", config.Get("SERVER_IP"), config.Get("SERVER_PORT"))
	if err := router.Run(addr); err != nil {
		panic(err)
	}
}
