package handler

import (
	"net/http"
	"strconv"

	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/service"

	"github.com/gin-gonic/gin"
)

type RegionHandler struct {
	regionSvc *service.RegionService
}

func NewRegionHandler(rg *gin.RouterGroup, svc *service.RegionService) *RegionHandler {
	h := &RegionHandler{regionSvc: svc}
	reg := rg.Group("/regions")
	reg.GET("", h.ListRegions)
	reg.GET("/:id", h.GetByID)
	reg.POST("", h.CreateRegion)
	reg.PUT("/:id", h.UpdateRegion)
	reg.DELETE("/:id", h.DeleteRegion)
	return h
}

func (h *RegionHandler) ListRegions(c *gin.Context) {
	list, err := h.regionSvc.List(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *RegionHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.regionSvc.GetByID(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *RegionHandler) CreateRegion(c *gin.Context) {
	var in catalog.Region
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.regionSvc.Create(c, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *RegionHandler) UpdateRegion(c *gin.Context) {
	var in catalog.Region
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	in.ID = uint(id)
	out, err := h.regionSvc.Update(c, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *RegionHandler) DeleteRegion(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.regionSvc.Delete(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
