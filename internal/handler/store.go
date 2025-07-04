// internal/handler/store.go
package handler

import (
	"djj-inventory-system/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StoreHandler struct {
	storeSvc *service.StoreService
}

func NewStoreHandler(rg *gin.RouterGroup, svc *service.StoreService) *StoreHandler {
	handler := &StoreHandler{storeSvc: svc}
	rg.GET("/stores", handler.ListStores)
	rg.GET("/stores/:id", handler.GetStoreByID)
	return handler
}

func (h *StoreHandler) ListStores(c *gin.Context) {
	stores, err := h.storeSvc.ListStores(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stores)
}

func (h *StoreHandler) GetStoreByID(c *gin.Context) {
	id := c.Param("id")
	store, err := h.storeSvc.GetStoreByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, store)
}
