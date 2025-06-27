package catalog

import (
	"time"

	"gorm.io/datatypes"
)

// Product 对应数据库表 products
type Product struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	DJJCode            string         `gorm:"size:50;unique;not null" json:"djjCode"`
	NameCN             string         `gorm:"size:100;not null" json:"nameCn"`
	NameEN             string         `gorm:"size:100" json:"nameEn"`
	Manufacturer       string         `gorm:"size:100" json:"manufacturer"`
	ManufacturerCode   string         `gorm:"size:100" json:"manufacturerCode"`
	Supplier           string         `gorm:"size:100" json:"supplier"`
	Model              string         `gorm:"size:100" json:"model"`
	CategoryID         *uint          `json:"categoryId,omitempty"`
	SubcategoryID      *uint          `json:"subcategoryId,omitempty"`
	TertiaryCategoryID *uint          `json:"tertiaryCategoryId,omitempty"`
	TechnicalSpecs     datatypes.JSON `json:"technicalSpecs"`
	Specs              string         `gorm:"type:text" json:"specs"`
	Price              float64        `gorm:"type:numeric(12,2);not null" json:"price"`
	RRPPrice           float64        `gorm:"type:numeric(12,2)" json:"rrpPrice"`
	Currency           string         `gorm:"type:currency_code_enum;default:'AUD'" json:"currency"`
	Status             string         `gorm:"type:product_status_enum;default:'draft'" json:"status"`
	ApplicationStatus  string         `gorm:"type:application_status_enum;default:'open'" json:"applicationStatus"`
	ProductType        string         `gorm:"type:product_type_enum;default:'others'" json:"productType"`
	StandardWarranty   string         `gorm:"size:100" json:"standardWarranty"`
	Remarks            string         `gorm:"type:text" json:"remarks"`
	MarketingInfo      string         `gorm:"type:text" json:"marketingInfo"`
	TrainingDocs       string         `gorm:"type:text" json:"trainingDocs"`
	WeightKG           float64        `gorm:"type:numeric(10,2)" json:"weightKg"`
	LiftCapacityKG     float64        `gorm:"type:numeric(10,2)" json:"liftCapacityKg"`
	LiftHeightMM       float64        `gorm:"type:numeric(10,2)" json:"liftHeightMm"`
	PowerSource        string         `gorm:"size:100" json:"powerSource"`
	OtherSpecs         datatypes.JSON `json:"otherSpecs"`
	ExtraInfo          datatypes.JSON `json:"extraInfo"`
	Metadata           datatypes.JSON `json:"metadata"`
	Version            int64          `gorm:"default:1" json:"version"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	IsDeleted          bool           `gorm:"default:false" json:"isDeleted"`
}

func (Product) TableName() string { return "products" }
