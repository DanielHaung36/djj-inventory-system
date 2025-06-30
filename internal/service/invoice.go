// internal/service/invoice_service.go
package service

import (
	"bytes"
	"context"
	"djj-inventory-system/assets"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/repository"
	"encoding/base64"
	"fmt"
	"html/template"
	"time"

	"djj-inventory-system/internal/model/sales"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type InvoiceService struct {
	QuoteRepo   *repository.QuoteRepository
	OrderRepo   *repository.OrderRepository
	CompanyRepo *repository.CompanyRepository
	tmpl        *template.Template
	logoBase64  string
}

func NewInvoiceService(
	qr *repository.QuoteRepository,
	or *repository.OrderRepository,
	cr *repository.CompanyRepository,
	tplPath string,
) *InvoiceService {
	// 1) base64 logo（同理 embed logo.png）
	// 2) 解析模板

	tmpl, err := template.New("invoice").Parse(assets.InvoiceTmplSrc)
	if err != nil {
		logger.Fatalf("解析 invoice.tmpl 失败: %w", err)
	}
	return &InvoiceService{
		QuoteRepo:   qr,
		OrderRepo:   or,
		CompanyRepo: cr,
		tmpl:        tmpl,
	}
}

func (s *InvoiceService) GenerateQuotePDF(ctx context.Context, quoteID uint) ([]byte, error) {
	// 1) 读报价
	q, err := s.QuoteRepo.FindByID(ctx, quoteID)
	if err != nil {
		return nil, err
	}

	// 2) 读公司
	co, err := s.CompanyRepo.FindDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("加载公司信息失败: %w", err)
	}

	// 3) 组装渲染数据
	inv := sales.Invoice{
		LogoBase64:     LogoBase64,
		CompanyName:    co.Name,
		CompanyEmail:   co.Email,
		CompanyPhone:   co.Phone,
		CompanyWebsite: co.Website,
		CompanyABN:     co.ABN,
		CompanyAddress: co.Address,
		BankName:       co.BankName,
		BSB:            co.BSB,
		AccountNumber:  co.AccountNumber,

		InvoiceNumber:   q.QuoteNumber,
		InvoiceDate:     q.QuoteDate.Format("2006/01/02"),
		InvoiceType:     "SALES QUOTE",
		IsQuote:         true,
		BillingAddress:  q.Customer.Address,
		DeliveryAddress: q.Customer.Address,
		CustomerCompany: q.Customer.Name,
		CustomerABN:     q.Customer.ABN,
		CustomerContact: q.Customer.Contact,
		CustomerPhone:   q.Customer.Phone,
		CustomerEmail:   q.Customer.Email,
		SalesRep:        q.SalesRepUser.Username,
		Items:           toInvoiceItemsFromQuote(q.Items),
		SubtotalAmount:  q.SubTotal,
		GSTAmount:       q.GSTTotal,
		TotalAmount:     q.TotalAmount,
		// company info, logo, bank details 可从配置或另一表读入
	}

	return s.renderAndPrintPDF(ctx, inv)
}

// toInvoiceItemsFromQuote 用来把 []QuoteItem 转成 []Item
func toInvoiceItemsFromQuote(qis []sales.QuoteItem) []sales.Item {
	out := make([]sales.Item, len(qis))
	for i, qi := range qis {
		out[i] = sales.Item{
			DJJCode:           qi.Product.DJJCode,
			Description:       qi.Description,
			DetailDescription: qi.DetailDescription,
			VinEngine:         qi.Product.VinEngine,
			Quantity:          qi.Quantity,
			// Quote 不用 Location，所以留空
			UnitPrice: qi.UnitPrice,
			Discount:  qi.Discount,
			Subtotal:  qi.TotalPrice,
		}
	}
	return out
}

func (s *InvoiceService) GeneratePickingPDF(ctx context.Context, orderID uint) ([]byte, error) {
	// 1) 读拣货单
	o, err := s.OrderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 2) 读公司
	co, err := s.CompanyRepo.FindDefault(ctx)
	if err != nil {
		return nil, err
	}

	// 3) 组装渲染数据
	inv := sales.Invoice{
		LogoBase64:         LogoBase64,
		CompanyName:        co.Name,
		CompanyEmail:       co.Email,
		CompanyPhone:       co.Phone,
		CompanyWebsite:     co.Website,
		CompanyABN:         co.ABN,
		CompanyAddress:     co.Address,
		InvoiceNumber:      o.OrderNumber,
		InvoiceDate:        o.OrderDate.Format("2006/01/02"),
		InvoiceType:        "PICKING LIST",
		IsQuote:            false,
		BillingAddress:     o.ShippingAddress,
		DeliveryAddress:    o.ShippingAddress,
		CustomerCompany:    o.Customer.Name,
		CustomerABN:        o.Customer.ABN,
		CustomerContact:    o.Customer.Contact,
		CustomerPhone:      o.Customer.Phone,
		CustomerEmail:      o.Customer.Email,
		SalesRep:           o.SalesRepUser.Username,
		Items:              toInvoiceItemsFromOrder(o.Items, o.Location),
		BankName:           co.BankName,
		BSB:                co.BSB,
		AccountNumber:      co.AccountNumber,
		TermsAndConditions: "",
		SubtotalAmount:     0,
		GSTAmount:          0,
		TotalAmount:        0,
	}

	return s.renderAndPrintPDF(ctx, inv)
}

// renderAndPrintPDF 负责：
//  1. 渲染模板为 HTML
//  2. 用 chromedp 打印成 PDF
func (s *InvoiceService) renderAndPrintPDF(ctx context.Context, inv sales.Invoice) ([]byte, error) {
	// 渲染
	var htmlBuf bytes.Buffer
	if err := s.tmpl.Execute(&htmlBuf, inv); err != nil {
		return nil, fmt.Errorf("填充模板失败: %w", err)
	}

	// 打印
	pdfCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	pdfCtx, cancel = context.WithTimeout(pdfCtx, 20*time.Second)
	defer cancel()

	var pdf []byte
	if err := chromedp.Run(pdfCtx,
		chromedp.Navigate("data:text/html;base64,"+base64.StdEncoding.EncodeToString(htmlBuf.Bytes())),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithDisplayHeaderFooter(true).
				WithHeaderTemplate(`<div></div>`).
				WithFooterTemplate(`
					<div style="width:100%;text-align:center;font-size:10px;color:#666;">
					  <span class="pageNumber"></span>/<span class="totalPages"></span>
					</div>
				`).
				WithMarginTop(0.4).
				WithMarginBottom(0.5).
				WithMarginLeft(0.3).
				WithMarginRight(0.3).
				Do(ctx)
			if err != nil {
				return err
			}
			pdf = buf
			return nil
		}),
	); err != nil {
		return nil, fmt.Errorf("生成 PDF 失败: %w", err)
	}

	return pdf, nil
}

// toInvoiceItems 把 repo QuoteItem 转到 Invoice Item
func toInvoiceItems(qis []model.QuoteItem) []model.Item {
	out := make([]model.Item, len(qis))
	for i, qi := range qis {
		out[i] = sales.Item{
			DJJCode:           qi.Product.DJJCode,
			Description:       qi.Description,
			DetailDescription: qi.DetailDescription,
			VinEngine:         qi.Product.VinEngine,
			Quantity:          qi.Quantity,
			// Quote 不用 Location，所以留空
			UnitPrice: qi.UnitPrice,
			Discount:  qi.Discount,
			Subtotal:  qi.TotalPrice,
		}
	}
	return out
}

// toInvoiceItemsFromOrder 用来把 []OrderItem 转成 []Item，
// 并且把主单传过来的 location 塞进去
func toInvoiceItemsFromOrder(items []sales.OrderItem, location string) []sales.Item {
	out := make([]sales.Item, len(items))
	for i, it := range items {
		out[i] = sales.Item{
			DJJCode:     it.Product.DJJCode, // 你要确保在 repo 里 Preload("Items.Product")
			Description: it.Product.NameCN,  // 或者随你想显示哪个字段
			Quantity:    it.Quantity,
			Location:    location,
		}
	}
	return out
}
