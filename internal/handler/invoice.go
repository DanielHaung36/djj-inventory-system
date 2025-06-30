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
	// 1. 解析并校验 ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 2. 调用 Service，把 Gin 的 Context 里的 context.Context 传下去
	pdfBytes, err := h.Svc.GenerateQuotePDF(c.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "quote not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 3. 返回 PDF 文件
	filename := "quote_" + c.Param("id") + ".pdf"
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GET /api/orders/:id/picking.pdf
func (h *InvoiceHandler) PickingPDF(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	pdf, err := h.Svc.GeneratePickingPDF(c.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(404, gin.H{"error": "order not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="picking_`+c.Param("id")+`.pdf"`)
	c.Writer.Write(pdf)
}
