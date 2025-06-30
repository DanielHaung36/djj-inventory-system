package sales

import (
	"djj-inventory-system/internal/model/rbac"
	"time"
)

// PickingList 对应数据库表 picking_lists
type PickingList struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	OrderID         uint              `gorm:"not null;index" json:"orderId"`
	PickingNumber   string            `gorm:"size:50;unique;not null" json:"pickingNumber"`
	DeliveryAddress string            `gorm:"size:255;not null" json:"deliveryAddress"`
	Status          string            `gorm:"size:20;default:'draft'" json:"status"`
	Location        string            `gorm:"size:255;not null" json:"location"` // 门店地址
	CreatedBy       uint              `gorm:"not null" json:"createdBy"`         // 谁创建了订单
	UpdatedBy       *uint             `json:"updatedBy,omitempty"`               // 谁最后修改了
	SalesRepID      uint              `gorm:"not null" json:"salesRepId"`
	SalesRepUser    rbac.User         `gorm:"foreignKey:SalesRepID" json:"salesRepUser"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Items           []PickingListItem `gorm:"foreignKey:PickingListID" json:"items"`
}

func (PickingList) TableName() string { return "picking_lists" }
