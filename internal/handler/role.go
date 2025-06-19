// internal/handler/role.go
package handler

import (
	"djj-inventory-system/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct{ svc service.RoleService }

func NewRoleHandler(rg *gin.RouterGroup, svc service.RoleService) {
	h := &RoleHandler{svc}
	grp := rg.Group("/roles")
	grp.POST("", h.Create)
	grp.GET("", h.List)
	grp.GET("/:id", h.Get)
	grp.PUT("/:id", h.Update)
	grp.DELETE("/:id", h.Delete)
}

// Create godoc
// @Summary      创建角色
// @Description 使用 name 字段创建一个新角色
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        payload  body     handler.RoleCreateRequest  true  "角色名称"
// @Success      201      {object} model.Role
// @Failure      400      {object} gin.H
// @Failure      500      {object} gin.H
// @Router       /roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var in struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.svc.Create(c, in.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

// List godoc
// @Summary      列表角色
// @Description 获取所有角色列表
// @Tags         roles
// @Produce      json
// @Success      200      {array}  model.Role
// @Failure      500      {object} gin.H
// @Router       /roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	xs, _ := h.svc.List(c)
	c.JSON(http.StatusOK, xs)
}

// Get godoc
// @Summary      查询单个角色
// @Description 根据 ID 返回角色信息
// @Tags         roles
// @Produce      json
// @Param        id       path     int  true  "角色 ID"
// @Success      200      {object} model.Role
// @Failure      404      {object} gin.H
// @Router       /roles/{id} [get]
func (h *RoleHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	x, err := h.svc.Get(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, x)
}

// Update godoc
// @Summary      更新角色
// @Description 根据 ID 修改角色名称
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id       path     int                      true  "角色 ID"
// @Param        payload  body     handler.RoleUpdateRequest true  "新角色名称"
// @Success      200      {object} model.Role
// @Failure      400      {object} gin.H
// @Failure      500      {object} gin.H
// @Router       /roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var in struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	x, err := h.svc.Update(c, uint(id), in.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, x)
}

// Delete godoc
// @Summary      删除角色
// @Description 根据 ID 删除角色
// @Tags         roles
// @Param        id       path     int  true  "角色 ID"
// @Success      204      "No Content"
// @Failure      500      {object} gin.H
// @Router       /roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.Delete(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
