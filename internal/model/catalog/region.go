// internal/model/catalog/region.go
package catalog

import (
	"time"

	"djj-inventory-system/internal/model/company"
)

// Region 对应数据库表 regions
type Region struct {
	ID        uint            `gorm:"primaryKey" json:"id"`                 // 地区ID
	Name      string          `gorm:"size:100;unique;not null" json:"name"` // 地区名称
	CompanyID uint            `gorm:"not null" json:"companyId"`            // 所属公司
	Company   company.Company `gorm:"foreignKey:CompanyID" json:"company"`  // 关联公司

	// 多对多关联到 Warehouses，通过 region_warehouses 联表
	Warehouses []Warehouse `gorm:"many2many:region_warehouses" json:"warehouses,omitempty"`

	CreatedAt time.Time `json:"createdAt"` // 创建时间
	UpdatedAt time.Time `json:"updatedAt"` // 更新时间
}

// TableName 指定这张表在数据库里的名字
func (Region) TableName() string { return "regions" }
