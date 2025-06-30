// internal/handler/permission.go
package handler

import (
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermHandler struct {
	svc service.PermService
}

// NewPermHandler 在 /permissions 下挂载 CRUD 路由
func NewPermHandler(rg *gin.RouterGroup, svc service.PermService) {
	h := &PermHandler{svc: svc}
	grp := rg.Group("/permissions")
	grp.POST("", h.Create)             // 创建权限
	grp.GET("", h.List)                // 列表
	grp.GET("/:id", h.Get)             // 取单条
	grp.PUT("/:id", h.Update)          // 更新
	grp.DELETE("/:id", h.Delete)       // 删除
	grp.GET("/modules", h.ListModules) // 新增：权限模块分组
}

// Create godoc
// @Summary      创建权限
// @Description 使用 action 和 object 字段创建权限
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        payload  body     handler.PermCreateRequest  true  "权限信息"
// @Success      201      {object} model.Permission
// @Failure      400      {object} gin.H
// @Failure      500      {object} gin.H
// @Router       /permissions [post]
func (h *PermHandler) Create(c *gin.Context) {
	var in struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.Create(c, in.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// Get godoc
// @Summary      查询单个权限
// @Description 根据 ID 获取权限详细
// @Tags         permissions
// @Produce      json
// @Param        id       path     int  true  "权限 ID"
// @Success      200      {object} model.Permission
// @Failure      404      {object} gin.H
// @Router       /permissions/{id} [get]
func (h *PermHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := h.svc.Get(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// List godoc
// @Summary      列表权限
// @Description 获取所有权限
// @Tags         permissions
// @Produce      json
// @Success      200      {array}  model.Permission
// @Failure      500      {object} gin.H
// @Router       /permissions [get]
func (h *PermHandler) List(c *gin.Context) {
	perms, err := h.svc.List(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

// Update godoc
// @Summary      更新权限
// @Description 根据 ID 修改 action 和 object
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        id       path     int                      true  "权限 ID"
// @Param        payload  body     handler.PermUpdateRequest true  "更新信息"
// @Success      200      {object} model.Permission
// @Failure      400      {object} gin.H
// @Failure      500      {object} gin.H
// @Router       /permissions/{id} [put]
func (h *PermHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var in struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.Update(c, uint(id), in.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// Delete godoc
// @Summary      删除权限
// @Description 根据 ID 删除权限
// @Tags         permissions
// @Param        id       path     int  true  "权限 ID"
// @Success      204      "No Content"
// @Failure      500      {object} gin.H
// @Router       /permissions/{id} [delete]
func (h *PermHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.Delete(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /permissions/modules
func (h *PermHandler) ListModules(c *gin.Context) {
	// 1. 从数据库获取所有权限
	perms, err := h.svc.List(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. 按模块分组返回
	modules := make([]PermissionModuleDTO, len(PermissionModules))
	for i, mod := range PermissionModules {
		modules[i] = PermissionModuleDTO{
			Module:      mod.Module,
			Icon:        mod.Icon,
			Description: mod.Description,
			Permissions: make([]PermissionDTO, 0, len(mod.Permissions)),
		}

		// 遍历该模块下的权限ID，从数据库权限列表中找到对应权限
		for _, expectedPerm := range mod.Permissions {
			for _, dbPerm := range perms {
				if dbPerm.ID == expectedPerm.ID {
					modules[i].Permissions = append(modules[i].Permissions, PermissionDTO{
						ID:          dbPerm.ID,
						Name:        dbPerm.Name,
						Label:       dbPerm.Label,
						Description: dbPerm.Description,
					})
					break
				}
			}
		}
	}

	logger.Infof("✔ returning %d permission modules with %d total permissions",
		len(modules),
		func() int {
			total := 0
			for _, m := range modules {
				total += len(m.Permissions)
			}
			return total
		}(),
	)
	c.JSON(http.StatusOK, modules)
}
