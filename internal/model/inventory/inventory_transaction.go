package inventory

import (
	"djj-inventory-system/internal/model/catalog"
	"time"
)

const TableNameInventoryTransaction = "inventory_transaction"

// InventoryTransaction 库存流水
type InventoryTransaction struct {
	ID          uint                 `gorm:"primaryKey" json:"id"`
	InventoryID uint                 `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"inventoryID"`
	Inventory   catalog.ProductStock `gorm:"foreignKey:InventoryID" json:"inventory"`
	TxType      TransactionType      `gorm:"column:tx_type;type:transaction_type;not null;default:'IN'" json:"txType"`
	Quantity    int                  `gorm:"not null" json:"quantity"`
	Operator    string               `gorm:"size:100;not null" json:"operator"`
	Note        string               `gorm:"size:500" json:"note"`
	CreatedAt   time.Time            `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName User's table name
func (*InventoryTransaction) TableName() string {
	return TableNameInventoryTransaction
}
