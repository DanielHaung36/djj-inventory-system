package dto

// —— DTO for JSON input / template 渲染 ——

type QuoteDTO struct {
	QuoteNumber   string         `json:"quoteNumber"`
	QuoteDate     string         `json:"quoteDate"`
	SalesRep      string         `json:"salesRep"`
	Currency      string         `json:"currency"`
	Items         []QuoteItemDTO `json:"items"`
	SubTotal      float64        `json:"subTotal"`
	GSTTotal      float64        `json:"gstTotal"`
	TotalAmount   float64        `json:"totalAmount"`
	Remarks       string         `json:"remarks"`
	WarrantyNotes string         `json:"warrantyNotes"`
	Status        string         `json:"status"`
}

type QuoteItemDTO struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Discount    float64 `json:"discount"`
	TotalPrice  float64 `json:"totalPrice"`
}
