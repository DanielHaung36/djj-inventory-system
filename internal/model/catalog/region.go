// internal/model/catalog/region.go
package catalog

import "time"

// Region 对应数据库表 regions
type Region struct {
	ID        uint      `gorm:"primaryKey" json:"id"` // 地区ID
	Name      string    `gorm:"size:100;unique;not null" json:"name"`
	CreatedAt time.Time `json:"createdAt"` // 创建时间
	UpdatedAt time.Time `json:"updatedAt"` // 更新时间

	// 关联到 Warehouses，通过 region_warehouses
	Warehouses []Warehouse `gorm:"many2many:region_warehouses" json:"warehouses,omitempty"`
}

// TableName 指定表名
func (Region) TableName() string { return "regions" }
