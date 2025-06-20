package handler

import (
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/pkg/auth"
	"djj-inventory-system/internal/service"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userSvc service.UserService
}

// NewAuthHandler 把 Session middleware 和 三个路由都挂到同一个 group
func NewAuthHandler(rg *gin.RouterGroup, us service.UserService) {
	h := &AuthHandler{userSvc: us}
	grp := rg.Group("") // 挂在/api下
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
	u, err := h.userSvc.Create(c, in.Username, in.Email, in.Password, in.RoleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
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
	u, err := h.userSvc.Authenticate(c, in.Username, in.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户名或密码错误"})
		return
	}

	// extract role names
	roleNames := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roleNames[i] = r.Name
	}

	// extract permission names
	permNames := make([]string, len(u.Permissions))
	for i, p := range u.Permissions {
		permNames[i] = p.Name
	}

	// prepare session data
	sd := &auth.SessionData{
		UserID:      u.ID,
		Roles:       roleNames,
		Permissions: permNames,
	}

	// write the securecookie
	if err := auth.SetSession(sd, c.Writer); err != nil {
		logger.Errorf("SetSession err: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "无法写入会话"})
		return
	}

	c.JSON(http.StatusOK, ResponseMessage{Message: "登录成功"})
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
