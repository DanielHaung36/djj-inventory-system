// internal/model/catalog/product_stock.go
package catalog

import "time"

type ProductStock struct {
	ID          uint `gorm:"primaryKey"`
	ProductID   uint `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	WarehouseID uint `gorm:"not null;index"`

	// —— 新增字段 ——
	OnHand   int `gorm:"not null;default:0" json:"on_hand"`
	Reserved int `gorm:"not null;default:0" json:"reserved"`
	// 如果 DB 有生成列，就可以直接 Preload 出 available
	Available int `gorm:"->;type:integer" json:"available"`

	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// GORM 预加载关联
	Product   Product   `gorm:"foreignKey:ProductID"   json:"-"`
	Warehouse Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse"`
}

func (ProductStock) TableName() string { return "product_stocks" }
