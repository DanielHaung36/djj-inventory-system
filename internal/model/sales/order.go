package sales

import "time"

// Order 对应数据库表 orders
type Order struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	QuoteID         *uint       `json:"quoteId,omitempty"`
	StoreID         *uint       `json:"storeId,omitempty"`
	CustomerID      *uint       `json:"customerId,omitempty"`
	OrderNumber     string      `gorm:"size:50;unique;not null" json:"orderNumber"`
	OrderDate       time.Time   `gorm:"type:date;not null" json:"orderDate"`
	Currency        string      `gorm:"type:currency_code_enum;default:'AUD'" json:"currency"`
	ShippingAddress string      `gorm:"size:255;not null" json:"shippingAddress"`
	TotalAmount     float64     `gorm:"type:numeric(14,2)" json:"totalAmount"`
	Status          string      `gorm:"type:order_status_enum;default:'draft'" json:"status"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
	Items           []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

func (Order) TableName() string { return "orders" }

// OrderItem 对应数据库表 order_items
type OrderItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"not null;index" json:"orderId"`
	ProductID uint      `gorm:"not null" json:"productId"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	UnitPrice float64   `gorm:"type:numeric(12,2);not null" json:"unitPrice"`
	CreatedAt time.Time `json:"createdAt"`
}

func (OrderItem) TableName() string { return "order_items" }
