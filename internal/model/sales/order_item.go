package sales

import (
	"djj-inventory-system/internal/model/catalog"
	"time"
)

// OrderItem 对应数据库表 order_items
type OrderItem struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	OrderID   uint `gorm:"not null;index" json:"orderId"`
	ProductID uint `gorm:"not null" json:"productId"`
	// 下面这一行：
	Product   catalog.Product `gorm:"foreignKey:ProductID;references:ID" json:"product"`
	Quantity  int             `gorm:"not null" json:"quantity"`
	UnitPrice float64         `gorm:"type:numeric(12,2);not null" json:"unitPrice"`
	CreatedAt time.Time       `json:"createdAt"`
}

func (OrderItem) TableName() string { return "order_items" }
