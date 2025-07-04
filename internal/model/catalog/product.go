// internal/model/catalog/product.go
package catalog

import (
	"time"

	"gorm.io/datatypes"
)

// Product 对应数据库表 products
type Product struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	DJJCode          string         `gorm:"size:50;unique;not null" json:"djjCode"`
	NameCN           string         `gorm:"size:100;not null" json:"nameCn"`
	NameEN           string         `json:"nameEn"`
	Specs            string         `json:"specs"`
	TechnicalSpecs   datatypes.JSON `json:"technicalSpecs"`
	MarketingInfo    string         `json:"marketingInfo"`
	Remarks          string         `json:"remarks"`
	Manufacturer     string         `gorm:"size:100" json:"manufacturer"`
	ManufacturerCode string         `gorm:"size:100" json:"manufacturerCode"`
	Supplier         string         `gorm:"size:100" json:"supplier"`
	Model            string         `gorm:"size:100" json:"model"`
	Category         Category       `gorm:"size:50" json:"category"`
	Subcategory      string         `gorm:"size:50" json:"subcategory"`
	TertiaryCategory string         `gorm:"size:50" json:"tertiary_category"`
	Price            float64        `gorm:"type:numeric(12,2);not null" json:"price"`
	RRPPrice         float64        `gorm:"type:numeric(12,2)" json:"rrpPrice"`
	Currency         string         `gorm:"type:currency_code_enum;default:'AUD'" json:"currency"`

	// 审批、类型
	Status            ProductStatus `gorm:"type:product_status_enum;default:'draft'" json:"status"`
	ApplicationStatus string        `gorm:"type:application_status_enum;default:'open'" json:"applicationStatus"`
	ProductType       string        `gorm:"type:product_type_enum;default:'others'" json:"productType"`

	// 保修、文档、URL
	StandardWarranty string `gorm:"size:100" json:"standardWarranty"`
	// 如果要把它跟 DTO 里的 warranty 对应，你可以在映射里直接用 StandardWarranty
	TrainingDocs string `gorm:"type:text" json:"trainingDocs"`
	ProductURL   string `gorm:"size:255" json:"productUrl"`

	// 重量、规格
	WeightKG       float64        `gorm:"type:numeric(10,2)" json:"weightKg"`
	LiftCapacityKG float64        `gorm:"type:numeric(10,2)" json:"liftCapacityKg"`
	LiftHeightMM   float64        `gorm:"type:numeric(10,2)" json:"liftHeightMm"`
	PowerSource    string         `gorm:"size:100" json:"powerSource"`
	OtherSpecs     datatypes.JSON `json:"otherSpecs"`
	ExtraInfo      datatypes.JSON `json:"extraInfo"`
	Metadata       datatypes.JSON `json:"metadata"`

	Version   int64     `gorm:"default:1" json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsDeleted bool      `gorm:"default:false" json:"isDeleted"`
	VinEngine string    `json:"vinEngine"`
	// 库存 & 附件
	Images      []ProductImage `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"images"`
	Stocks      []ProductStock `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"stocks"`
	Attachments []Attachment   `gorm:"polymorphic:Ref;polymorphicValue:product;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"attachments,omitempty"`
	Standards   string         `json:"standards"`
	Unit        string         `json:"unit"`
	Warranty    string         `json:"warranty"`
}

func (Product) TableName() string { return "products" }
