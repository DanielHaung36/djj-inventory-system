// internal/dto/product.go
package dto

import (
	"encoding/json"
	"time"
)

type ProductImageDTO struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	IsPrimary bool   `json:"is_primary"`
}

type SalesDataDTO struct {
	Month   string  `json:"month"`
	Sales   int     `json:"sales"`
	Revenue float64 `json:"revenue"`
	Profit  float64 `json:"profit"`
}

// 前端提交通用的库存条目
type StockEntry struct {
	WarehouseID   uint   `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name,omitempty"`

	// —— 新增字段 ——
	OnHand   int `gorm:"not null;default:0" json:"on_hand"`
	Reserved int `gorm:"not null;default:0" json:"reserved"`
	// 如果 DB 有生成列，就可以直接 Preload 出 available
	Available int `gorm:"->;type:integer" json:"available"`
}

// ------ Create / Update 请求 ------
type CreateProductRequest struct {
	DJJCode          string `json:"djj_code" binding:"required"`
	Status           string `json:"status"` // Active/Inactive/Discontinued
	Supplier         string `json:"supplier"`
	ManufacturerCode string `json:"manufacturer_code"`
	Category         string `json:"category"`
	Subcategory      string `json:"subcategory"`
	TertiaryCategory string `json:"tertiary_category"`

	NameCN           string  `json:"name_cn" binding:"required"`
	NameEN           string  `json:"name_en"`
	Specs            string  `json:"specs"`
	Standards        string  `json:"standards"`
	Unit             string  `json:"unit"`
	Price            float64 `json:"price"`
	Currency         string  `json:"currency"`
	RRPPrice         float64 `json:"rrp_price"`
	StandardWarranty string  `json:"standard_warranty"`
	Remarks          string  `json:"remarks"`

	WeightKG       float64 `json:"weight_kg"`
	LiftCapacityKG float64 `json:"lift_capacity_kg,omitempty"`
	LiftHeightMM   float64 `json:"lift_height_mm,omitempty"`
	PowerSource    string  `json:"power_source,omitempty"`

	Warranty      string `json:"warranty"`
	MarketingInfo string `json:"marketing_info"`
	TrainingDocs  string `json:"training_docs"`
	ProductURL    string `json:"product_url"`

	// —— 之前用 syd_stock/per_stock/bne_stock —— 改为：
	Stocks []StockEntry `json:"stocks"`

	LastUpdate     time.Time `json:"last_update"`
	LastModifiedBy string    `json:"last_modified_by"`

	MonthlySales int     `json:"monthly_sales"`
	TotalSales   int     `json:"total_sales"`
	ProfitMargin float64 `json:"profit_margin"`

	OtherInfo      json.RawMessage `json:"other_info,omitempty"`
	OtherSpecs     json.RawMessage `json:"other_specs,omitempty"`
	TechnicalSpecs json.RawMessage `json:"technical_specs,omitempty"`
	ExtraInfo      json.RawMessage `json:"extra_info,omitempty"`

	Images    []ProductImageDTO `json:"images"`
	SalesData []SalesDataDTO    `json:"sales_data"`
}

type UpdateProductRequest = CreateProductRequest

// ------ Response 返回给前端用 ------
type ProductResponse struct {
	ID               uint   `json:"id"`
	DJJCode          string `json:"djj_code"`
	Status           string `json:"status"`
	Supplier         string `json:"supplier"`
	ManufacturerCode string `json:"manufacturer_code"`
	Category         string `json:"category"`
	Subcategory      string `json:"subcategory"`
	TertiaryCategory string `json:"tertiary_category"`

	NameCN    string `json:"name_cn"`
	NameEN    string `json:"name_en"`
	Specs     string `json:"specs"`
	Standards string `json:"standards"`
	Unit      string `json:"unit"`

	Currency         string  `json:"currency"`
	RRPPrice         float64 `json:"rrp_price"`
	StandardWarranty string  `json:"standard_warranty"`
	Remarks          string  `json:"remarks"`

	WeightKG       float64         `json:"weight_kg"`
	LiftCapacityKG float64         `json:"lift_capacity_kg,omitempty"`
	LiftHeightMM   float64         `json:"lift_height_mm,omitempty"`
	PowerSource    string          `json:"power_source,omitempty"`
	OtherSpecs     json.RawMessage `json:"other_specs,omitempty"`

	Warranty      string `json:"warranty"`
	MarketingInfo string `json:"marketing_info"`
	TrainingDocs  string `json:"training_docs"`
	ProductURL    string `json:"product_url"`

	// —— 统一列表 ——
	Stocks []StockEntry `json:"stocks"`

	LastUpdate     time.Time `json:"last_update"`
	LastModifiedBy string    `json:"last_modified_by"`

	MonthlySales int     `json:"monthly_sales"`
	TotalSales   int     `json:"total_sales"`
	ProfitMargin float64 `json:"profit_margin"`

	TechnicalSpecs json.RawMessage `json:"technical_specs"`
	OtherInfo      json.RawMessage `json:"other_info,omitempty"`

	Images    []ProductImageDTO `json:"images"`
	SalesData []SalesDataDTO    `json:"sales_data"`

	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
