// internal/model/catalog/region_warehouse.go
package catalog

// RegionWarehouse 对应数据库表 region_warehouses
// 这是一个纯关联表，如果你不需要额外字段，可以不用单独建 struct，
// GORM 会自动管理 many2many。但若要在代码里直接操作，也可以这样写：
type RegionWarehouse struct {
	RegionID    uint `gorm:"primaryKey" json:"regionId"`
	WarehouseID uint `gorm:"primaryKey" json:"warehouseId"`
}

// TableName 指定表名
func (RegionWarehouse) TableName() string { return "region_warehouses" }
