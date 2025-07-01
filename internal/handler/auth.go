package handler

import (
	"djj-inventory-system/internal/pkg/auth"
	"djj-inventory-system/internal/service"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("DJJ_JWT_SECRET")

type AuthHandler struct {
	userSvc service.UserService
}

// NewAuthHandler 把 Session middleware 和 三个路由都挂到同一个 group
func NewAuthHandler(rg *gin.RouterGroup, us service.UserService) {
	h := &AuthHandler{userSvc: us}
	grp := rg.Group("auth") // 挂在/api下
	grp.POST("/register", h.Register)
	grp.POST("/login", h.Login)
	grp.POST("/logout", h.Logout)
}

// Register godoc
// @Summary      用户注册
// @Description  使用用户名、邮箱、密码和可选角色 ID 列表创建新用户，并下发登录 Cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body     RegisterRequest  true  "注册信息"
// @Success      201      {object} model.User
// @Failure      400      {object} ErrorResponse
// @Failure      500      {object} ErrorResponse
// @Router       /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var in RegisterRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	u, err := h.userSvc.Create(c, in.Username, in.Email, in.Password, in.RoleNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// 4. 发 JWT
	//token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	//	"sub":         u.ID,
	//	"role":        u.Roles[0].Name,
	//	"permissions": []string{}, // 默认无额外权限
	//	"exp":         time.Now().Add(24 * time.Hour).Unix(),
	//})
	//s, _ := token.SignedString(jwtSecret)

	//c.JSON(http.StatusOK, gin.H{
	//	"token": s,
	//	"user": gin.H{
	//		"id":         usr.ID,
	//		"name":       usr.FullName,
	//		"email":      usr.Email,
	//		"role":       role.Name,
	//		"avatar_url": usr.AvatarURL,
	//	},
	//})
	// 简单起见，注册后直接打登录 cookie
	c.SetCookie("uid", fmt.Sprint(u.ID), int((7 * 24 * time.Hour).Seconds()), "/", "", false, true)
	c.JSON(http.StatusCreated, u)
}

// Login godoc
// @Summary      用户登录
// @Description  使用用户名和密码登录，返回 Session Cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body     LoginRequest  true  "登录信息"
// @Success      200      {object} ResponseMessage
// @Failure      400      {object} ErrorResponse
// @Failure      401      {object} ErrorResponse
// @Router       /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var in LoginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Authenticate returns a User with Roles and Permissions preloaded
	u, sd, err := h.userSvc.Authenticate(c, in.Email, in.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户名或密码错误"})
		return
	}
	var rolePerms []string
	for i := 0; i < len(u.Permissions); i++ {
		rolePerms = append(rolePerms, u.Permissions[i].Name)
	}
	// extract role names
	permSet := make(map[string]struct{}, len(rolePerms)+len(u.DirectPermissions))
	for _, name := range rolePerms {
		permSet[name] = struct{}{}
	}
	for _, p := range u.DirectPermissions {
		permSet[p.Name] = struct{}{}
	}
	finalPerms := make([]string, 0, len(permSet))
	for name := range permSet {
		finalPerms = append(finalPerms, name)
	}
	role := strings.Join(func() []string {
		names := make([]string, len(u.Roles))
		for i, r := range u.Roles {
			names[i] = r.Name
		}
		return names
	}(), ",")

	// 5. 生成 JWT，包含 sub/role/permissions
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         u.ID,
		"name":        u.Username, // ← 新增这一行
		"role":        role,
		"permissions": finalPerms,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
		"avatar_url":  u.AvatarURL,
		"sd":          sd,
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}
	// 5. 写入 HttpOnly Cookie
	//    maxAge 单位是秒； path、domain、secure、httpOnly 根据你的需求调整
	c.SetCookie(
		"access_token", // name
		tokenString,    // value
		24*3600,        // maxAge: 24h
		"/",            // path
		"",             // domain: 改成你的域名，或留空字符串让浏览器自动匹配当前域
		true,           // secure: 生产环境请设为 true (HTTPS)
		true,           // httpOnly: JS 无法读取
	)
	// 6. 返回给前端
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":           u.ID,
			"name":         u.Username,
			"email":        u.Email,
			"role":         role,
			"permissions":  finalPerms,
			"storedetails": sd,
		}})

}

// Logout godoc
// @Summary      用户登出
// @Description  清除登录 Cookie
// @Tags         auth
// @Produce      json
// @Success      200      {object} ResponseMessage
// @Failure      500      {object} ErrorResponse
// @Router       /logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 调用我们 auth 包里的 ClearSession
	auth.ClearSession(c.Writer)

	c.JSON(http.StatusOK, ResponseMessage{Message: "logged out"})
}
