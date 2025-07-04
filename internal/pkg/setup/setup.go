package setup

import (
	_ "djj-inventory-system/docs" // <-- 一定要导入，才能注册 docs.SwaggerInfo
	"djj-inventory-system/internal/handler"
	"djj-inventory-system/internal/pkg/audit"
	"djj-inventory-system/internal/repository"
	"djj-inventory-system/internal/service"
	"djj-inventory-system/internal/websocket"
	"log"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func ServerStart(db *gorm.DB) {

	// set up repos + services *once*
	userRepo := repository.NewUserRepo(db)
	auditor := audit.NewGormAuditor(db)
	userSvc := service.NewUserService(userRepo, auditor)

	roleRepo := repository.NewRoleRepo(db)
	roleSvc := service.NewRoleService(roleRepo, auditor)

	permRepo := repository.NewPermRepo(db)
	permSvc := service.NewPermService(permRepo, auditor)

	// new Gin router
	r := gin.Default()

	// 1) global — always try to decode session cookie into context
	//r.Use(handler.SessionAuthMiddleware())

	// 2) public endpoints: register & login (no RequireLogin here)
	public := r.Group("/api")
	handler.NewAuthHandler(public, userSvc)

	// 3) protected endpoints: everything under here needs a valid session
	protected := r.Group("/api")
	protected.Use(handler.RequireLogin())
	{
		handler.NewUserHandler(protected, userSvc)
		handler.NewRoleHandler(protected, roleSvc)
		handler.NewPermHandler(protected, permSvc)

		// in future: handler.NewQuoteHandler(protected, quoteSvc) …
	}

	// 4) start server
	if err := r.Run("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}

// NewRouter 返回组装好所有路由但不启动的 gin.Engine
func NewRouter(db *gorm.DB) *gin.Engine {
	// repos + svc
	userRepo := repository.NewUserRepo(db)
	auditor := audit.NewGormAuditor(db)
	userSvc := service.NewUserService(userRepo, auditor)

	roleRepo := repository.NewRoleRepo(db)
	roleService := service.NewRoleService(roleRepo, auditor)

	permRepo := repository.NewPermRepo(db)
	permSvc := service.NewPermService(permRepo, auditor)
	hub := websocket.NewHub()
	customerRepo := repository.NewCustomerRepo(db)
	customerService := service.NewCustomerService(customerRepo)
	storeService := service.NewStoreService(db)
	regionService := service.NewRegionService(db)

	stockRepository := repository.NewStockRepository(db)
	productRepository := repository.NewProductRepository(db)
	prodSvc := service.NewProductService(productRepository, stockRepository)

	// router
	// 假设配置里 STORAGE_PATH="./"（项目根目录）
	baseDir, _ := filepath.Abs("./")
	uploadDir := filepath.Join(baseDir, "uploads")
	log.Println("→ Serving static files from:", uploadDir)

	r := gin.Default()
	r.Static("/uploads", uploadDir)
	r.Use(handler.SessionAuthMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://192.168.1.244:5173"}, // 或者 ["*"] 开发时
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//websocket
	r.GET("/ws/:topic", websocket.ServeWS(hub))
	public := r.Group("/api")
	// 挂载 Swagger UI
	handler.NewAuthHandler(public, userSvc)
	r.POST("/api/generate-pdf", handler.GeneratePDF)
	handler.NewRoleHandler(public, roleService)
	protected := r.Group("/api")
	protected.Use(handler.RequireLogin())
	handler.NewUserHandler(protected, userSvc)
	handler.NewPermHandler(protected, permSvc)
	handler.NewCustomerHandler(protected, customerService, hub)
	handler.NewStoreHandler(protected, storeService)
	handler.NewRegionHandler(protected, regionService)
	handler.NewProductHandler(protected, prodSvc, hub)
	handler.NewUploadHandler(protected, "uploads", "")
	return r
}
