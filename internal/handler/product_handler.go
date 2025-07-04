// internal/handler/product_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"djj-inventory-system/internal/model/dto"
	"djj-inventory-system/internal/service"
	"djj-inventory-system/internal/websocket"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	Svc *service.ProductService
	Hub *websocket.Hub
}

func NewProductHandler(
	rg *gin.RouterGroup,
	svc *service.ProductService,
	hub *websocket.Hub,
) {
	h := &ProductHandler{Svc: svc, Hub: hub}
	grp := rg.Group("/products")
	grp.GET("", h.List)
	grp.GET("/:id", h.Get)
	grp.POST("", h.Create)
	grp.PUT("/:id", h.Update)
	grp.DELETE("/:id", h.Delete)
}

// List 返回分页列表： /api/products?offset=0&limit=20
func (h *ProductHandler) List(c *gin.Context) {
	off, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	lim, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	products, total, err := h.Svc.List(c.Request.Context(), off, lim)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"products": products,
	})
}

// Get 单条查询： /api/products/:id
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	pr, err := h.Svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, pr)
}

// Create 新建： POST /api/products
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr, err := h.Svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 广播给所有订阅 "products" 频道的客户端
	msg, _ := json.Marshal(gin.H{"event": "productCreated", "payload": pr})
	h.Hub.Broadcast("products", msg)

	c.JSON(http.StatusCreated, pr)
}

// Update 修改： PUT /api/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr, err := h.Svc.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	msg, _ := json.Marshal(gin.H{"event": "productUpdated", "payload": pr})
	h.Hub.Broadcast("products", msg)

	c.JSON(http.StatusOK, pr)
}

// Delete 删除： DELETE /api/products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.Svc.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 仅广播删除的 ID
	msg, _ := json.Marshal(gin.H{"event": "productDeleted", "payload": gin.H{"id": id}})
	h.Hub.Broadcast("products", msg)

	c.Status(http.StatusNoContent)
}
