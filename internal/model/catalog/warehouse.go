// internal/model/catalog/warehouse.go
package catalog

import "time"

// Warehouse 对应数据库表 warehouses
type Warehouse struct {
	ID        uint      `gorm:"primaryKey" json:"id"` // 仓库ID
	Name      string    `gorm:"size:100;unique;not null" json:"name"`
	Location  string    `gorm:"size:255" json:"location"` // 地址（可为空）
	Version   int64     `gorm:"not null;default:1" json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsDeleted bool      `gorm:"not null;default:false" json:"isDeleted"`

	// 关联到 Regions，通过 region_warehouses
	Regions []Region `gorm:"many2many:region_warehouses" json:"regions,omitempty"`
}

// TableName 指定表名
func (Warehouse) TableName() string { return "warehouses" }
