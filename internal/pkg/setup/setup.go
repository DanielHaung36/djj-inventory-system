package setup

import (
	_ "djj-inventory-system/docs" // <-- 一定要导入，才能注册 docs.SwaggerInfo
	"djj-inventory-system/internal/handler"
	"djj-inventory-system/internal/pkg/audit"
	"djj-inventory-system/internal/repository"
	"djj-inventory-system/internal/service"

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

	// router
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Use(handler.SessionAuthMiddleware())
	public := r.Group("/api")
	// 挂载 Swagger UI
	handler.NewAuthHandler(public, userSvc)
	r.POST("/api/generate-pdf", handler.GeneratePDF)
	handler.NewRoleHandler(public, roleService)
	protected := r.Group("/api")
	protected.Use(handler.RequireLogin())
	handler.NewUserHandler(protected, userSvc)
	handler.NewPermHandler(protected, permSvc)

	return r
}
