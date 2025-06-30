package sales

import (
	"djj-inventory-system/internal/model/catalog"
	"time"
)

// PickingListItem 对应数据库表 picking_list_items
type PickingListItem struct {
	ID            uint `gorm:"primaryKey" json:"id"`
	PickingListID uint `gorm:"not null;index" json:"pickingListId"`
	ProductID     uint `gorm:"not null" json:"productId"`
	// 下面这一行：
	Product   catalog.Product `gorm:"foreignKey:ProductID;references:ID" json:"product"`
	Quantity  int             `gorm:"not null" json:"quantity"`
	Location  string          `gorm:"size:100" json:"location"`
	CreatedAt time.Time       `json:"createdAt"`
}

func (PickingListItem) TableName() string { return "picking_list_items" }
