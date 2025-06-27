package sales

// InvoiceData 是整份数据的顶层：以编号做 key
type InvoiceData map[string]Invoice

// Invoice 对应每一单发票／报价单
type Invoice struct {
	CompanyName        string `json:"companyName"`
	CompanyEmail       string `json:"companyEmail"`
	CompanyPhone       string `json:"companyPhone"`
	CompanyWebsite     string `json:"companyWebsite"`
	CompanyABN         string `json:"companyABN"`
	CompanyAddress     string `json:"companyAddress"`
	InvoiceNumber      string `json:"invoiceNumber"`
	InvoiceDate        string `json:"invoiceDate"`
	InvoiceType        string `json:"invoiceType"`
	BillingAddress     string `json:"billingAddress"`
	DeliveryAddress    string `json:"deliveryAddress"`
	CustomerCompany    string `json:"customerCompany"`
	CustomerABN        string `json:"customerABN"`
	CustomerContact    string `json:"customerContact"`
	CustomerPhone      string `json:"customerPhone"`
	CustomerEmail      string `json:"customerEmail"`
	SalesRep           string `json:"salesRep"`
	Items              []Item `json:"items"`
	BankName           string `json:"bankName"`
	BSB                string `json:"bsb"`
	AccountNumber      string `json:"accountNumber"`
	TermsAndConditions string `json:"termsAndConditions"`

	// 以下字段只有报价（SALES QUOTE）会用到
	SubtotalAmount float64 `json:"subtotalAmount,omitempty"`
	GSTAmount      float64 `json:"gstAmount,omitempty"`
	TotalAmount    float64 `json:"totalAmount,omitempty"`
}

// Item 对应 items 数组里的每一行
type Item struct {
	ID                string `json:"id"`
	DJJCode           string `json:"djjCode"`
	Description       string `json:"description"`
	DetailDescription string `json:"detailDescription,omitempty"`
	VinEngine         string `json:"vinEngine,omitempty"`
	Quantity          int    `json:"quantity"`
	Location          string `json:"location,omitempty"`

	// 以下报价专用
	UnitPrice float64 `json:"unitPrice,omitempty"`
	Discount  float64 `json:"discount,omitempty"`
	Subtotal  float64 `json:"subtotal,omitempty"`
}
