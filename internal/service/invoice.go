// internal/service/invoice_service.go
package service

import (
	"djj-inventory-system/internal/model"
	"djj-inventory-system/internal/model/sales"
	"djj-inventory-system/internal/repository"
	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	"bytes"
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"html/template"
	"time"
)

type InvoiceService struct {
	QuoteRepo *repository.QuoteRepository
}

func NewInvoiceService(qr *repository.QuoteRepository) *InvoiceService {
	return &InvoiceService{QuoteRepo: qr}
}

// GenerateQuotePDF 根据 quoteID 生成 PDF bytes
func (s *InvoiceService) GenerateQuotePDF(quoteID uint) ([]byte, error) {
	// 1) 从 DB 读取 quote + items
	q, err := s.QuoteRepo.GetByID(quoteID)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, ErrNotFound
	}

	// 2) 把 model.Quote 转成渲染用的 model.Invoice
	inv := sales.Invoice{
		InvoiceNumber:   q.QuoteNumber,
		InvoiceDate:     q.QuoteDate.Format("2006/01/02"),
		InvoiceType:     "SALES QUOTE",
		BillingAddress:  q.Customer.Address,
		DeliveryAddress: q.Customer.Address,
		CustomerCompany: q.Customer.Name,
		CustomerABN:     q.Customer.ABN,
		CustomerContact: q.Customer.Contact,
		CustomerPhone:   q.Customer.Phone,
		CustomerEmail:   q.Customer.Email,
		SalesRep:        q.SalesRep,
		SubtotalAmount:  q.SubTotal,
		GSTAmount:       q.GSTTotal,
		TotalAmount:     q.TotalAmount,
		Items:           toInvoiceItems(q.Items),
		// company info, logo, bank details 可从配置或另一表读入
	}

	// 3) 渲染模板 -> HTML
	tplPath := filepath.Join("templates", "invoice.tmpl")
	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		return nil, err
	}
	var htmlBuf bytes.Buffer
	if err := tpl.Execute(&htmlBuf, inv); err != nil {
		return nil, err
	}

	// 4) chromedp 打印成 PDF
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var pdf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate("data:text/html;base64,"+base64.StdEncoding.EncodeToString(htmlBuf.Bytes())),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithDisplayHeaderFooter(true).
				WithHeaderTemplate("<div></div>").
				WithFooterTemplate(
					`<div style="width:100%;text-align:center;font-size:10px;color:#666">
                        <span class="pageNumber"></span>/<span class="totalPages"></span>
                     </div>`).
				Do(ctx)
			pdf = buf
			return err
		}),
	); err != nil {
		return nil, err
	}
	return pdf, nil
}

// toInvoiceItems 把 repo QuoteItem 转到 Invoice Item
func toInvoiceItems(qis []sales.QuoteItem) []sales.Item {
	out := make([]sales.Item, len(qis))
	for i, qi := range qis {
		out[i] = sales.Item{
			DJJCode:           qi.Product.DJJCode,
			Description:       qi.Description,
			DetailDescription: qi.Product.Specs, // or qi.DetailDescription
			Quantity:          qi.Quantity,
			UnitPrice:         qi.UnitPrice,
			Discount:          qi.Discount,
			Subtotal:          qi.TotalPrice,
		}
	}
	return out
}
