// handler/user.go
package handler

import (
	"djj-inventory-system/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(rg *gin.RouterGroup, svc service.UserService) {
	h := &UserHandler{svc}
	grp := rg.Group("/users")
	grp.POST("", h.Create)
	grp.GET("", h.List)
	grp.GET("/:id", h.Get)
	grp.PUT("/:id", h.Update)
	grp.DELETE("/:id", h.Delete)
	// 角色分配
	grp.POST("/:id/roles/:rid", h.AssignRole)
	grp.DELETE("/:id/roles/:rid", h.RemoveRole)
	grp.GET("/:id/roles", h.ListRoles)
}

// Create godoc
// @Summary      创建用户
// @Description 使用用户名、邮箱、密码和可选角色列表创建用户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        payload  body     handler.UserCreateRequest  true  "用户信息"
// @Success      201      {object} model.User
// @Failure      400      {object} gin.H
// @Failure      500      {object} gin.H
// @Router       /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var in struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		RoleIDs  []uint `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.Create(c, in.Username, in.Email, in.Password, in.RoleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, u)
}

// Get godoc
// @Summary      查询单个用户
// @Description 根据 ID 获取用户详情
// @Tags         users
// @Produce      json
// @Param        id       path     int  true  "用户 ID"
// @Success      200      {object} model.User
// @Failure      404      {object} gin.H
// @Router       /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	u, err := h.svc.Get(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

// List godoc
// @Summary      列表用户
// @Description 获取所有用户
// @Tags         users
// @Produce      json
// @Success      200      {array}  model.User
// @Failure      500      {object} gin.H
// @Router       /users [get]
func (h *UserHandler) List(c *gin.Context) {
	users, _ := h.svc.List(c)
	c.JSON(http.StatusOK, users)
}

// Update godoc
// @Summary      更新用户
// @Description 根据 ID 修改邮箱或密码
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id       path     int                      true  "用户 ID"
// @Param        payload  body     handler.UserUpdateRequest true  "更新信息"
// @Success      200      {object} model.User
// @Failure      400      {object} gin.H
// @Failure      500      {object} gin.H
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var in struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.Update(c, uint(id), in.Email, in.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

// Delete godoc
// @Summary      删除用户
// @Description 根据 ID 删除用户
// @Tags         users
// @Param        id       path     int  true  "用户 ID"
// @Success      204      "No Content"
// @Failure      500      {object} gin.H
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.Delete(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// AssignRole godoc
// @Summary      给用户分配角色
// @Description 为指定用户添加角色
// @Tags         users
// @Param        id       path  int  true  "用户 ID"
// @Param        rid      path  int  true  "角色 ID"
// @Success      204      "No Content"
// @Failure      500      {object} gin.H
// @Router       /users/{id}/roles/{rid} [post]
func (h *UserHandler) AssignRole(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Param("id"))
	rid, _ := strconv.Atoi(c.Param("rid"))
	if err := h.svc.AssignRole(c, uint(uid), uint(rid)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveRole godoc
// @Summary      从用户移除角色
// @Description 删除指定用户的某个角色
// @Tags         users
// @Param        id       path  int  true  "用户 ID"
// @Param        rid      path  int  true  "角色 ID"
// @Success      204      "No Content"
// @Failure      500      {object} gin.H
// @Router       /users/{id}/roles/{rid} [delete]
func (h *UserHandler) RemoveRole(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Param("id"))
	rid, _ := strconv.Atoi(c.Param("rid"))
	if err := h.svc.RemoveRole(c, uint(uid), uint(rid)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListRoles godoc
// @Summary      列出用户角色
// @Description 获取指定用户拥有的所有角色
// @Tags         users
// @Param        id       path  int  true  "用户 ID"
// @Produce      json
// @Success      200      {array}  model.Role
// @Failure      500      {object} gin.H
// @Router       /users/{id}/roles [get]
func (h *UserHandler) ListRoles(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Param("id"))
	roles, err := h.svc.ListRoles(c, uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}
