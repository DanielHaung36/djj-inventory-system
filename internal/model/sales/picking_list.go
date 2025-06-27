package sales

import "time"

// PickingList 对应数据库表 picking_lists
type PickingList struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	OrderID         uint              `gorm:"not null;index" json:"orderId"`
	PickingNumber   string            `gorm:"size:50;unique;not null" json:"pickingNumber"`
	DeliveryAddress string            `gorm:"size:255;not null" json:"deliveryAddress"`
	Status          string            `gorm:"size:20;default:'draft'" json:"status"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Items           []PickingListItem `gorm:"foreignKey:PickingListID" json:"items"`
}

func (PickingList) TableName() string { return "picking_lists" }
