package sales

import (
	"djj-inventory-system/internal/model/rbac"
	"time"

	"djj-inventory-system/internal/model/catalog"
	"djj-inventory-system/internal/model/company"
)

// Quote 对应数据库表 quotes
type Quote struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	StoreID       uint             `gorm:"not null" json:"storeId"`
	Store         catalog.Store    `gorm:"foreignKey:StoreID" json:"store"`
	CompanyID     uint             `json:"company_id"` // ← 新增
	Company       company.Company  `gorm:"foreignKey:CompanyID" json:"company"`
	CustomerID    uint             `gorm:"not null" json:"customerId"`
	Customer      catalog.Customer `gorm:"foreignKey:CustomerID"` // ← 新增
	QuoteNumber   string           `gorm:"size:50;unique;not null" json:"quoteNumber"`
	SalesRepID    uint             `gorm:"not null" json:"salesRepId"`
	SalesRepUser  rbac.User        `gorm:"foreignKey:SalesRepID" json:"salesRepUser"`
	QuoteDate     time.Time        `gorm:"type:date;not null"     json:"quoteDate"`
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
