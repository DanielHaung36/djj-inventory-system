package setup

import (
	"djj-inventory-system/internal/database"
	"djj-inventory-system/internal/handler"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/pkg/audit"
	"djj-inventory-system/internal/repository"
	"djj-inventory-system/internal/service"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTest 初始化测试环境，包括内存数据库和 Gin 引擎
func SetupTest(t *testing.T) (*gorm.DB, *gin.Engine, error) {
	// 1. 初始化日志
	logger.Init("test.log", zapcore.DebugLevel)

	// 2. 初始化内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to in-memory database: %v", err)
		return nil, nil, err
	}
	t.Log("In-memory database connected.")

	// 3. 运行数据库迁移
	database.Migrate(db)
	t.Log("Database migration successful.")

	// 4. 初始化 Gin 引擎
	router := gin.Default()
	api := router.Group("/api/v1")

	// 5. 初始化依赖注入 (repository, service, handler)
	// 使用 mock recorder
	mockRecorder := &audit.MockRecorder{}
	permRepo := repository.NewPermRepo(db)
	permSvc := service.NewPermService(permRepo, mockRecorder)
	handler.NewPermHandler(api, permSvc) // 注册 /permissions 路由

	// 你可以在这里继续注册其他 handler
	// ...

	return db, router, nil
}
