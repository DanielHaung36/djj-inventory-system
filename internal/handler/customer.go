package handler

import (
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/service"
	"djj-inventory-system/internal/websocket"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	svc service.CustomerService
	hub *websocket.Hub
}

func NewCustomerHandler(rg *gin.RouterGroup, svc service.CustomerService, hub *websocket.Hub) {
	h := &CustomerHandler{svc, hub}
	grp := rg.Group("/customers")
	grp.GET("", h.List)
	grp.GET(":id", h.Get)
	grp.POST("", h.Create)
	grp.PUT(":id", h.Update)
	grp.DELETE(":id", h.Delete)
}

func (h *CustomerHandler) List(c *gin.Context) {
	cs, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cs)
}

func (h *CustomerHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	cust, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, cust)
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var input catalog.Customer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.svc.Create(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// broadcast to WebSocket subscribers on topic "customers"
	msg, _ := json.Marshal(gin.H{"event": "customerCreated", "payload": out})
	h.hub.Broadcast("customers", msg)
	c.JSON(http.StatusCreated, out)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	var input catalog.Customer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	out, err := h.svc.Update(c.Request.Context(), uint(id), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	msg, _ := json.Marshal(gin.H{"event": "customerUpdated", "payload": out})
	h.hub.Broadcast("customers", msg)
	c.JSON(http.StatusOK, out)
}

func (h *CustomerHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	msg, _ := json.Marshal(gin.H{"event": "customerDeleted", "payload": gin.H{"id": id}})
	h.hub.Broadcast("customers", msg)
	c.Status(http.StatusNoContent)
}
