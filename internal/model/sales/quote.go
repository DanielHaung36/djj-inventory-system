package sales

import (
	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/company"
	"time"
)

// Quote 对应数据库表 quotes
type Quote struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	StoreID       uint             `gorm:"not null" json:"storeId"`
	CompanyID     uint             `json:"company_id"` // ← 新增
	Company       company.Company  `gorm:"foreignKey:CompanyID" json:"company"`
	CustomerID    uint             `gorm:"not null" json:"customerId"`
	Customer      catalog.Customer `gorm:"foreignKey:CustomerID"` // ← 新增
	QuoteNumber   string           `gorm:"size:50;unique;not null" json:"quoteNumber"`
	SalesRep      string           `gorm:"size:100" json:"salesRep"`
	QuoteDate     time.Time        `gorm:"type:date;not null" json:"quoteDate"`
	Currency      string           `gorm:"type:currency_code_enum;default:'AUD'" json:"currency"`
	SubTotal      float64          `gorm:"type:numeric(14,2);not null" json:"subTotal"`
	GSTTotal      float64          `gorm:"type:numeric(14,2);not null" json:"gstTotal"`
	TotalAmount   float64          `gorm:"type:numeric(14,2);not null" json:"totalAmount"`
	Remarks       string           `gorm:"type:text" json:"remarks"`
	WarrantyNotes string           `gorm:"type:text" json:"warrantyNotes"`
	Status        string           `gorm:"type:approval_status_enum;default:'pending'" json:"status"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
	Items         []QuoteItem      `gorm:"foreignKey:QuoteID" json:"items"`
}

func (Quote) TableName() string { return "quotes" }

// QuoteItem 对应数据库表 quote_items
type QuoteItem struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	QuoteID           uint             `gorm:"not null;index" json:"quote_id"`
	ProductID         *uint            `json:"product_id,omitempty"`
	Product           *catalog.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Description       string           `gorm:"type:text;not null" json:"description"`
	DetailDescription string           `gorm:"column:detail_description;type:text" json:"detail_description"`
	Quantity          int              `gorm:"not null" json:"quantity"`
	Unit              string           `gorm:"size:20;not null" json:"unit"`
	UnitPrice         float64          `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Discount          float64          `gorm:"type:numeric(12,2);default:0" json:"discount"`
	TotalPrice        float64          `gorm:"type:numeric(14,2);not null" json:"total_price"`
	GoodsNature       string           `gorm:"type:goods_nature_enum;default:'contract'" json:"goods_nature"`
	CreatedAt         time.Time        `json:"created_at"`
}

func (QuoteItem) TableName() string { return "quote_items" }
