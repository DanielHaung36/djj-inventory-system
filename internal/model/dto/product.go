package dto

import "gorm.io/datatypes"

// —— DTO for前端 Create/Update JSON ——

type ProductDTO struct {
	DJJCode            string         `json:"djjCode"`
	NameCN             string         `json:"nameCn"`
	NameEN             string         `json:"nameEn,omitempty"`
	Manufacturer       string         `json:"manufacturer,omitempty"`
	ManufacturerCode   string         `json:"manufacturerCode,omitempty"`
	Supplier           string         `json:"supplier,omitempty"`
	Model              string         `json:"model,omitempty"`
	CategoryID         *uint          `json:"categoryId,omitempty"`
	SubcategoryID      *uint          `json:"subcategoryId,omitempty"`
	TertiaryCategoryID *uint          `json:"tertiaryCategoryId,omitempty"`
	TechnicalSpecs     datatypes.JSON `json:"technicalSpecs,omitempty"`
	Specs              string         `json:"specs,omitempty"`
	Price              float64        `json:"price"`
	RRPPrice           float64        `json:"rrpPrice,omitempty"`
	Currency           string         `json:"currency,omitempty"`
	Status             string         `json:"status,omitempty"`
	ApplicationStatus  string         `json:"applicationStatus,omitempty"`
	ProductType        string         `json:"productType,omitempty"`
	StandardWarranty   string         `json:"standardWarranty,omitempty"`
	Remarks            string         `json:"remarks,omitempty"`
	MarketingInfo      string         `json:"marketingInfo,omitempty"`
	TrainingDocs       string         `json:"trainingDocs,omitempty"`
	WeightKG           float64        `json:"weightKg,omitempty"`
	LiftCapacityKG     float64        `json:"liftCapacityKg,omitempty"`
	LiftHeightMM       float64        `json:"liftHeightMm,omitempty"`
	PowerSource        string         `json:"powerSource,omitempty"`
	OtherSpecs         datatypes.JSON `json:"otherSpecs,omitempty"`
	ExtraInfo          datatypes.JSON `json:"extraInfo,omitempty"`
	Metadata           datatypes.JSON `json:"metadata,omitempty"`
}
