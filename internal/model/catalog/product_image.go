package catalog

import "time"

// ProductImage 对应数据库表 product_images
type ProductImage struct {
	ID        uint      `gorm:"primaryKey"           json:"id"`
	ProductID uint      `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	URL       string    `gorm:"type:text;not null"   json:"url"`
	Alt       string    `gorm:"size:255"             json:"alt"`
	IsPrimary bool      `gorm:"not null;default:false" json:"isPrimary"`
	CreatedAt time.Time `gorm:"autoCreateTime"       json:"createdAt"`
}

func (ProductImage) TableName() string { return "product_images" }
