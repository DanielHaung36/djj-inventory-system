package handler

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

type InvoiceItem struct {
	ID          string  `json:"id"`
	DJJCode     string  `json:"djjCode"`
	Description string  `json:"description"`
	VinEngine   string  `json:"vinEngine"`
	Quantity    int     `json:"quantity"`
	Location    string  `json:"location"`
	UnitPrice   float64 `json:"unitPrice"`
	TotalPrice  float64 `json:"totalPrice"`
}

type InvoiceData struct {
	// Company Info
	CompanyName    string `json:"companyName"`
	CompanyEmail   string `json:"companyEmail"`
	CompanyPhone   string `json:"companyPhone"`
	CompanyWebsite string `json:"companyWebsite"`
	CompanyABN     string `json:"companyABN"`
	CompanyAddress string `json:"companyAddress"`

	// Invoice Details
	InvoiceNumber string `json:"invoiceNumber"`
	InvoiceDate   string `json:"invoiceDate"`
	InvoiceType   string `json:"invoiceType"`

	// Customer Info
	BillingAddress  string `json:"billingAddress"`
	DeliveryAddress string `json:"deliveryAddress"`
	CustomerCompany string `json:"customerCompany"`
	CustomerABN     string `json:"customerABN"`
	CustomerContact string `json:"customerContact"`
	CustomerPhone   string `json:"customerPhone"`
	CustomerEmail   string `json:"customerEmail"`
	SalesRep        string `json:"salesRep"`

	// Items
	Items []InvoiceItem `json:"items"`

	// Payment Info
	BankName      string `json:"bankName"`
	BSB           string `json:"bsb"`
	AccountNumber string `json:"accountNumber"`

	// Terms
	TermsAndConditions string `json:"termsAndConditions"`
	ShowPrices         bool   `json:"showPrices"`
}

func (i *InvoiceData) CalculateTotal() float64 {
	total := 0.0
	for _, item := range i.Items {
		total += item.TotalPrice
	}
	return total
}

func (i *InvoiceData) TotalQuantity() int {
	total := 0
	for _, item := range i.Items {
		total += item.Quantity
	}
	return total
}

func formatPrice(price float64) string {
	return fmt.Sprintf("%.2f", price)
}

func formatPriceWithCommas(price float64) string {
	priceStr := fmt.Sprintf("%.2f", price)
	parts := strings.Split(priceStr, ".")
	intPart := parts[0]
	decPart := parts[1]

	// Add commas to integer part
	if len(intPart) > 3 {
		var result []string
		for i, digit := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result = append(result, ",")
			}
			result = append(result, string(digit))
		}
		intPart = strings.Join(result, "")
	}

	return intPart + "." + decPart
}

func splitDescription(desc string) (string, string) {
	lines := strings.Split(desc, "\n")
	if len(lines) == 1 {
		return lines[0], ""
	}
	return lines[0], strings.Join(lines[1:], "\n")
}

const invoiceTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>{{.InvoiceNumber}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: Arial, sans-serif; 
            font-size: 12px; 
            line-height: 1.4;
            color: #000;
            background: white;
            padding: 20px;
        }
        .header { margin-bottom: 20px; }
        .company-name { font-size: 18px; font-weight: bold; margin-bottom: 10px; }
        .company-info { display: flex; justify-content: space-between; margin-bottom: 20px; }
        .company-left, .company-right { width: 48%; }
        .invoice-type { 
            text-align: center; 
            font-size: 16px; 
            font-weight: bold; 
            margin: 20px 0; 
            border-top: 1px solid #000;
            border-bottom: 1px solid #000;
            padding: 10px 0;
        }
        .customer-info { 
            display: flex; 
            justify-content: space-between; 
            margin-bottom: 20px; 
        }
        .customer-left, .customer-right { width: 48%; }
        .info-line { margin-bottom: 3px; }
        .items-table { 
            width: 100%; 
            border-collapse: collapse; 
            margin-bottom: 20px;
            border: 1px solid #000;
        }
        .items-table th, .items-table td { 
            border: 1px solid #000; 
            padding: 8px; 
            text-align: left; 
            vertical-align: top;
        }
        .items-table th { 
            background-color: #f5f5f5; 
            font-weight: bold; 
        }
        .djj-code { width: 80px; }
        .description { width: 250px; }
        .vin { width: 100px; }
        .qty { width: 50px; text-align: center; }
        .location { width: 80px; }
        .unit-price { width: 80px; text-align: right; }
        .total-price { width: 90px; text-align: right; }
        .total-row { 
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-weight: bold; 
            margin-bottom: 20px; 
        }
        .terms { margin-bottom: 20px; }
        .terms h3 { font-weight: bold; margin-bottom: 10px; }
        .terms-content { font-size: 10px; line-height: 1.3; white-space: pre-line; }
        .payment-details { 
            border-top: 1px solid #000; 
            padding-top: 15px; 
        }
        .payment-details h3 { font-weight: bold; margin-bottom: 10px; }
        .description-text { white-space: pre-line; }
        .total-amount { font-size: 14px; }
        @media print {
            body { padding: 0; }
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="company-name">{{.CompanyName}}</div>
        <div class="company-info">
            <div class="company-left">
                <div class="info-line"><strong>Email:</strong> {{.CompanyEmail}}</div>
                <div class="info-line"><strong>Phone:</strong> {{.CompanyPhone}}</div>
                <div class="info-line"><strong>Website:</strong>{{.CompanyWebsite}}</div>
            </div>
            <div class="company-right">
                <div class="info-line"><strong>ABN:</strong> {{.CompanyABN}}</div>
                <div class="info-line"><strong>Address:</strong> {{.CompanyAddress}}</div>
            </div>
        </div>
    </div>

    <div class="invoice-type">{{.InvoiceType}}</div>

    <div class="customer-info">
        <div class="customer-left">
            <div class="info-line"><strong>Billing Address:</strong> {{.BillingAddress}}</div>
            <div class="info-line"><strong>Company:</strong> {{.CustomerCompany}}</div>
            <div class="info-line"><strong>Contact:</strong> {{.CustomerContact}}</div>
            <div class="info-line"><strong>Email:</strong> {{.CustomerEmail}}</div>
        </div>
        <div class="customer-right">
            <div class="info-line"><strong>Delivery Address:</strong> {{.DeliveryAddress}}</div>
            <div class="info-line"><strong>ABN:</strong> {{.CustomerABN}}</div>
            <div class="info-line"><strong>Invoice Number:</strong> {{.InvoiceNumber}}</div>
            <div class="info-line"><strong>Invoice Date:</strong> {{.InvoiceDate}}</div>
            <div class="info-line"><strong>Phone:</strong> {{.CustomerPhone}}</div>
            <div class="info-line"><strong>Sales Rep:</strong> {{.SalesRep}}</div>
        </div>
    </div>

    <table class="items-table">
        <thead>
            <tr>
                <th class="djj-code">DJJ Code</th>
                <th class="description">Product / Description</th>
                <th class="vin">VIN / Engine No.</th>
                <th class="qty">Qty.</th>
                <th class="location">Location</th>
                {{if .ShowPrices}}
                <th class="unit-price">Unit Price</th>
                <th class="total-price">Total</th>
                {{end}}
            </tr>
        </thead>
        <tbody>
            {{range .Items}}
            <tr>
                <td class="djj-code">{{.DJJCode}}</td>
                <td class="description">
                    {{$first, $rest := splitDescription .Description}}
                    {{if .DJJCode}}
                        <strong>{{$first}}</strong>
                    {{else}}
                        <strong>- {{$first}}</strong>
                    {{end}}
                    {{if $rest}}<br>{{$rest}}{{end}}
                </td>
                <td class="vin">{{.VinEngine}}</td>
                <td class="qty">{{.Quantity}}</td>
                <td class="location">{{.Location}}</td>
                {{if $.ShowPrices}}
                <td class="unit-price">${{formatPriceWithCommas .UnitPrice}}</td>
                <td class="total-price"><strong>${{formatPriceWithCommas .TotalPrice}}</strong></td>
                {{end}}
            </tr>
            {{end}}
        </tbody>
    </table>

    <div class="total-row">
        <div><strong>Total Qty: {{.TotalQuantity}}</strong></div>
        {{if .ShowPrices}}
        <div class="total-amount"><strong>Total Amount: ${{formatPriceWithCommas .CalculateTotal}}</strong></div>
        {{end}}
    </div>

    <div class="terms">
        <h3>Terms & Conditions:</h3>
        <div class="terms-content">{{.TermsAndConditions}}</div>
    </div>

    <div class="payment-details">
        <h3>EFT Payment Details:</h3>
        <div class="info-line"><strong>Name:</strong> {{.BankName}}</div>
        <div class="info-line"><strong>BSB:</strong> {{.BSB}}</div>
        <div class="info-line"><strong>Acct No.:</strong> {{.AccountNumber}}</div>
        <div class="info-line"><strong>Ref.:</strong> {{.InvoiceNumber}}</div>
    </div>
</body>
</html>
`

func GeneratePDF(c *gin.Context) {
	var invoiceData InvoiceData
	if err := c.ShouldBindJSON(&invoiceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create template with custom functions
	tmpl := template.New("invoice").Funcs(template.FuncMap{
		"formatPrice":           formatPrice,
		"formatPriceWithCommas": formatPriceWithCommas,
		"splitDescription":      splitDescription,
	})

	tmpl, err := tmpl.Parse(invoiceTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Template parsing failed"})
		return
	}

	// Generate HTML
	var htmlBuffer strings.Builder
	if err := tmpl.Execute(&htmlBuffer, invoiceData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Template execution failed"})
		return
	}

	// Create context for chromedp
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pdfBuffer []byte

	// Generate PDF using chromedp
	err = chromedp.Run(ctx,
		chromedp.Navigate("data:text/html,"+htmlBuffer.String()),
		chromedp.WaitVisible("body"),
		//chromedp.ActionFunc(func(ctx context.Context) error {
		//	var err error
		//	pdfBuffer, _, err = chromedp.PrintToPDF().
		//		WithPrintBackground(true).
		//		WithPaperWidth(8.27).  // A4 width in inches
		//		WithPaperHeight(11.7). // A4 height in inches
		//		WithMarginTop(0.4).
		//		WithMarginBottom(0.4).
		//		WithMarginLeft(0.4).
		//		WithMarginRight(0.4).
		//		Do(ctx)
		//	return err
		//}),
	)

	if err != nil {
		log.Printf("PDF generation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
		return
	}

	// Return PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", invoiceData.InvoiceNumber))
	c.Data(http.StatusOK, "application/pdf", pdfBuffer)
}
