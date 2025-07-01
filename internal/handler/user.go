// handler/user.go
package handler

import (
	"djj-inventory-system/internal/service"
	"net/http"
	"strconv"
	"time"

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

	// ---- 新增：直接赋予/回收 用户权限 ----
	grp.POST("/:id/permissions", h.GrantUserPermissions)
	grp.DELETE("/:id/permissions", h.RevokeUserPermissions)
	grp.GET("/:id/permissions", h.ListUserPermissions)
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
		Username  string   `json:"username" binding:"required"`
		Email     string   `json:"email" binding:"required,email"`
		Password  string   `json:"password" binding:"required"`
		RoleNames []string `json:"role_names" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.Create(c, in.Username, in.Email, in.Password, in.RoleNames)
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
	role := ""
	if len(u.Roles) > 0 {
		role = u.Roles[0].Name
	}
	dtoUser := UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		FullName:  u.Username,
		Email:     u.Email,
		Role:      role,
		IsActive:  !u.IsDeleted,
		LastLogin: u.UpdatedAt.Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, dtoUser)
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
	dtos := make([]UserDTO, 0, len(users))
	for _, u := range users {
		// 拿第一个角色名（如果你用户只有一个角色）
		roleName := ""
		if len(u.Roles) > 0 {
			roleName = u.Roles[0].Name
		}

		dtos = append(dtos, UserDTO{
			ID:        u.ID,
			Username:  u.Username,
			FullName:  u.Username,
			Email:     u.Email,
			Role:      roleName,
			IsActive:  !u.IsDeleted,
			LastLogin: u.UpdatedAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, dtos)
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

// GrantUserPermissions 批量给用户增加直接权限
func (h *UserHandler) GrantUserPermissions(c *gin.Context) {
	var body struct {
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.GrantUserPermissions(c.Request.Context(), uint(uid), body.PermissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// RevokeUserPermissions 批量从用户移除直接权限
func (h *UserHandler) RevokeUserPermissions(c *gin.Context) {
	var body struct {
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.RevokeUserPermissions(c.Request.Context(), uint(uid), body.PermissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

/*
	Preload("Roles.Permissions")

	首先通过 user_roles 把这个用户的所有角色都查出来

	然后再通过 role_permissions 把这些角色对应的所有权限一并拉进来

	Preload("DirectPermissions")

	把 user_permissions 里，这个用户直接关联（“单独赋予”）的所有权限查出来

	扁平化合并

	把第一步和第二步拿到的权限合并到一个 map 去重

	最终把 map 里的所有权限值填到 user.Permissions 里
*/

// ListUserPermissions 获取用户所有直接+继承权限
// ListUserPermissions godoc
// @Summary      获取用户权限及最新修改信息
// @Description  返回权限 ID 列表及最近一次修改的时间和操作者
// @Tags         users
// @Produce      json
// @Param        id   path     int  true  "用户 ID"
// @Success      200  {object} handler.UserPermissionDataDTO
// @Failure      500  {object} gin.H
// @Router       /users/{id}/permissions [get]

func (h *UserHandler) ListUserPermissions(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.GetUserPermissionData(c.Request.Context(), uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dto := UserPermissionDataDTO{
		UserID:      data.UserID,
		Permissions: data.PermissionIDs,
		ModifiedBy:  data.ModifiedBy,
	}
	if !data.LastModified.IsZero() {
		dto.LastModified = data.LastModified.Format(time.RFC3339)
	}
	// 返回扁平化后的 Permissions 字段
	c.JSON(http.StatusOK, dto)
}
