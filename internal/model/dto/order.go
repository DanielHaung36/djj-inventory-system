package dto

// —— DTO for JSON input / template 渲染 ——

type OrderDTO struct {
	OrderNumber     string         `json:"orderNumber"`
	OrderDate       string         `json:"orderDate"`
	Currency        string         `json:"currency"`
	ShippingAddress string         `json:"shippingAddress"`
	Items           []OrderItemDTO `json:"items"`
	TotalAmount     float64        `json:"totalAmount"`
	Status          string         `json:"status"`
}

type OrderItemDTO struct {
	ProductID uint    `json:"productId"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}
