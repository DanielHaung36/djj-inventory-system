// internal/model/sales/order.go
package sales

import (
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/rbac"
	"time"
)

// Order 对应数据库表 orders
type Order struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	QuoteID         *uint            `json:"quoteId,omitempty"`
	StoreID         uint             `gorm:"not null"  json:"storeId"` // 一定要 non-null
	Store           catalog.Store    `gorm:"foreignKey:StoreID" json:"store"`
	CustomerID      uint             `gorm:"not null" json:"customerId"`
	Customer        catalog.Customer `gorm:"foreignKey:CustomerID" json:"customer"`
	OrderNumber     string           `gorm:"size:50;unique;not null" json:"orderNumber"`
	OrderDate       time.Time        `gorm:"type:date;not null" json:"orderDate"`
	Currency        string           `gorm:"type:currency_code_enum;default:'AUD'" json:"currency"`
	ShippingAddress string           `gorm:"size:255;not null" json:"shippingAddress"`
	TotalAmount     float64          `gorm:"type:numeric(14,2)" json:"totalAmount"`
	Status          string           `gorm:"type:order_status_enum;default:'draft'" json:"status"`
	// 直接存门店地址，无需前端传
	Location  string      `gorm:"size:255;not null" json:"location"` // 门店地址
	CreatedBy uint        `gorm:"not null" json:"createdBy"`         // 谁创建了订单
	UpdatedBy *uint       `json:"updatedBy,omitempty"`               // 谁最后修改了
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
	Items     []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	// ← 新增这一行，把销售人员也关联成一个用户
	SalesRepID   uint      `gorm:"not null" json:"salesRepId"`
	SalesRepUser rbac.User `gorm:"foreignKey:SalesRepID" json:"salesRepUser"`
}

func (Order) TableName() string { return "orders" }
