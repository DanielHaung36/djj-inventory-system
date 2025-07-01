// internal/model/catalog/store_details.go
package catalog

import "djj-inventory-system/internal/model/company"

// StoreDetails 用于承载门店 + 区域 + 公司 + 仓库信息
type StoreDetails struct {
	Store      Store           `json:"store"`
	Region     Region          `json:"region"`
	Company    company.Company `json:"company"`
	Warehouses []Warehouse     `json:"warehouses"`
}
