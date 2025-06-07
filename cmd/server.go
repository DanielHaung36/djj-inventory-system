// main.go
package main

import (
	"go.uber.org/zap/zapcore"

	"djj-inventory-system/internal/logger"
)

func main() {
	if err := logger.Init("./logs/app.log", zapcore.DebugLevel); err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Infof("djj-inventory-system 启动成功")
}
