package sales

import (
	"time"

	"djj-inventory-system/internal/model/catalog"
)

// QuoteItem 对应数据库表 quote_items
type QuoteItem struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	QuoteID           uint             `gorm:"not null;index"         json:"quoteId"`
	ProductID         *uint            `json:"productId,omitempty"`
	Product           *catalog.Product `gorm:"foreignKey:ProductID"   json:"product,omitempty"`
	Description       string           `gorm:"type:text;not null"     json:"description"`
	DetailDescription string           `gorm:"column:detail_description;type:text" json:"detailDescription"`
	Quantity          int              `gorm:"not null"              json:"quantity"`
	Unit              string           `gorm:"size:20;not null"      json:"unit"`
	UnitPrice         float64          `gorm:"type:numeric(12,2);not null" json:"unitPrice"`
	Discount          float64          `gorm:"type:numeric(12,2);default:0"  json:"discount"`
	TotalPrice        float64          `gorm:"type:numeric(14,2);not null"   json:"totalPrice"`
	GoodsNature       string           `gorm:"type:goods_nature_enum;default:'contract'" json:"goodsNature"`
	CreatedAt         time.Time        `json:"createdAt"`
}

// TableName 指定这张表在数据库中的名字
func (QuoteItem) TableName() string {
	return "quote_items"
}
