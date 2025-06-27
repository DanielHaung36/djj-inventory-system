package main

import (
	"context"
	"encoding/base64"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Item represents a line item in the invoice
type Item struct {
	DjjCode           string
	Description       string
	DetailDescription string
	VinEngine         string
	Quantity          int
	Location          string
	UnitPrice         float64
	Discount          float64
	Subtotal          float64
}

// Invoice holds all data for rendering
type Invoice struct {
	LogoBase64     string
	CompanyName    string
	CompanyEmail   string
	CompanyPhone   string
	CompanyWebsite string
	CompanyABN     string
	CompanyAddress string

	InvoiceNumber   string
	InvoiceDate     string
	InvoiceType     string
	IsQuote         bool
	BillingAddress  string
	DeliveryAddress string
	CustomerCompany string
	CustomerABN     string
	CustomerContact string
	CustomerPhone   string
	CustomerEmail   string
	SalesRep        string

	Items          []Item
	TotalQuantity  int
	SubtotalAmount float64
	GSTAmount      float64
	TotalAmount    float64

	BankName      string
	BSB           string
	AccountNumber string
}

func main() {
	// 1. 解析模板
	tmpl, err := template.ParseFiles("../invoice.tmpl")
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}

	// 2. 读 logo.png 并 base64
	logoBytes, err := ioutil.ReadFile("./logo.png")
	if err != nil {
		log.Fatalf("读取 logo.png 失败: %v", err)
	}
	logoBase64 := base64.StdEncoding.EncodeToString(logoBytes)

	// 3. 构造两个测试用的 Invoice
	tests := []struct {
		name string // 用来命名输出文件
		data Invoice
	}{
		{
			name: "picking",
			data: Invoice{
				LogoBase64:     logoBase64,
				CompanyName:    "DJJ PERTH PTY LTD",
				CompanyEmail:   "sales@djjequipment.com.au",
				CompanyPhone:   "1800 355 388",
				CompanyWebsite: "https://djjequipment.com.au",
				CompanyABN:     "95 663 874 664",
				CompanyAddress: "56 Clavering Road Bayswater, WA Australia 6053",

				InvoiceNumber:   "INV-25093P",
				InvoiceDate:     "2025/06/25",
				InvoiceType:     "PICKING LIST",
				IsQuote:         false,
				BillingAddress:  "PO Box 227, Belmont WA 6984",
				DeliveryAddress: "49 McKenna Drive Cardup WA 6122",
				CustomerCompany: "PJ'S HIRE N HAULAGE PTY LTD",
				CustomerABN:     "68 619 555 387",
				CustomerContact: "Patrick Fitzgibbon",
				CustomerPhone:   "0400797622",
				CustomerEmail:   "patfitz74@gmail.com",
				SalesRep:        "Mark WANG",

				Items: []Item{
					{
						DjjCode:     "DJJ01859",
						Description: "Hangcha 2.5T Diesel Rough Terrain Forklift",
						DetailDescription: `CPCD25-XW33E-RT
											Brand New Hangcha 2.5T Diesel Rough Terrain 2WD Forklift
											Integrated side sideshifter
											1220mm fork/tyne
											4.5m triplex mast`,
						VinEngine: "19BC01040\nE2968",
						Quantity:  1,
						Location:  "Perth",
					},
					{
						DjjCode:           "",
						Description:       "Delivery Service",
						DetailDescription: "Deliver to…",
						VinEngine:         "LG93024010003",
						Quantity:          1,
						Location:          "Perth",
					},
				},
				TotalQuantity: 2,
				BankName:      "DJJ PERTH PTY LTD",
				BSB:           "082-309",
				AccountNumber: "70 774 6500",
			},
		},
		{
			name: "quote",
			data: Invoice{
				LogoBase64:     logoBase64,
				CompanyName:    "DJJ PERTH PTY LTD",
				CompanyEmail:   "sales@djjequipment.com.au",
				CompanyPhone:   "1800 355 388",
				CompanyWebsite: "https://djjequipment.com.au",
				CompanyABN:     "95 663 874 664",
				CompanyAddress: "56 Clavering Road Bayswater, WA Australia 6053",

				InvoiceNumber:   "QTE-12345",
				InvoiceDate:     "25/06/2025",
				InvoiceType:     "SALES QUOTE",
				IsQuote:         true,
				BillingAddress:  "8 Kimberley St, Wyndham WA 6740",
				DeliveryAddress: "8 Kimberley St, Wyndham WA 6740",
				CustomerCompany: "ACME Pty Ltd",
				CustomerABN:     "12 345 678 901",
				CustomerContact: "Alice Lee",
				CustomerPhone:   "0412345678",
				CustomerEmail:   "alice@example.com",
				SalesRep:        "Mark WANG",

				Items: []Item{
					{
						DjjCode:           "DJJ00001",
						Description:       "LGMA Wheel Loader – LM930",
						DetailDescription: "Brand new QH, A/C Cab…",
						Quantity:          1,
						UnitPrice:         24990,
						Discount:          0,
						Subtotal:          24990,
					},
					{
						DjjCode:           "DJJ00032",
						Description:       "GP Bucket – LM930",
						DetailDescription: "GP Bucket LM930",
						Quantity:          1,
						UnitPrice:         2800,
						Discount:          2800,
						Subtotal:          0,
					},
				},
				SubtotalAmount: 24990,
				GSTAmount:      2499,
				TotalAmount:    27489,

				BankName:      "DJJ PERTH PTY LTD",
				BSB:           "082-309",
				AccountNumber: "70 774 6500",
			},
		},
	}

	// 4. 对每个测试项：渲染 HTML + 用 chromedp 打印 PDF
	for _, tt := range tests {
		// 4.1 渲染 HTML
		htmlFile := tt.name + ".html"
		f, err := os.Create(htmlFile)
		if err != nil {
			log.Fatalf("[%s] 无法创建 %s: %v", tt.name, htmlFile, err)
		}
		if err := tmpl.Execute(f, tt.data); err != nil {
			log.Fatalf("[%s] 渲染模板失败: %v", tt.name, err)
		}
		f.Close()
		log.Printf("[%s] 已生成 %s", tt.name, htmlFile)

		// 4.2 用 chromedp 打印成 PDF
		ctx, cancel := chromedp.NewContext(context.Background())
		ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		abs, _ := filepath.Abs(htmlFile)
		url := "file://" + abs
		outPDF := tt.name + ".pdf"

		var buf []byte
		if err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body"),
			chromedp.ActionFunc(func(ctx context.Context) error {
				b, _, err := page.PrintToPDF().
					WithPrintBackground(true).
					WithDisplayHeaderFooter(true).
					WithHeaderTemplate("<div></div>").
					WithFooterTemplate(`<div style="width:100%;text-align:center;font-size:12px;color:#666"><span class="pageNumber"></span>/<span class="totalPages"></span></div>`).
					WithMarginTop(0.4).
					WithMarginBottom(0.5).
					WithMarginLeft(0.3).
					WithMarginRight(0.3).
					Do(ctx)
				if err != nil {
					return err
				}
				buf = b
				return nil
			}),
		); err != nil {
			log.Fatalf("[%s] 生成 PDF 失败: %v", tt.name, err)
		}

		if err := ioutil.WriteFile(outPDF, buf, 0644); err != nil {
			log.Fatalf("[%s] 写入 %s 失败: %v", tt.name, outPDF, err)
		}
		log.Printf("[%s] 已生成 %s\n", tt.name, outPDF)
	}
}
