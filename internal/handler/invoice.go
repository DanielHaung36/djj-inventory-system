// internal/handler/invoice_handler.go
package handler

import (
	"djj-inventory-system/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	Svc *service.InvoiceService
}

func NewInvoiceHandler(svc *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{Svc: svc}
}

// GET /api/quotes/:id/pdf
func (h *InvoiceHandler) QuotePDF(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	pdf, err := h.Svc.GenerateQuotePDF(uint(id))
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "quote not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="quote_`+c.Param("id")+`.pdf"`)
	c.Writer.Write(pdf)
}
