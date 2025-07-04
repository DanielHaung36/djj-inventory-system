// internal/model/customer.go
package catalog

import "time"

// Customer 对应数据库 customers 表
type Customer struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	StoreID   uint      `gorm:"column:store_id"      json:"store_id"`
	Store     Store     `gorm:"foreignKey:StoreID" json:"store"`
	Type      string    `gorm:"type:customer_type_enum;default:'retail';column:type" json:"type"`
	Company   string    `gorm:"type:varchar(255);column:company" json:"company"`
	Name      string    `gorm:"size:100;not null;column:name"   json:"name"`
	Phone     string    `gorm:"size:20;column:phone"            json:"phone"`
	Email     string    `gorm:"size:100;column:email"           json:"email"`
	ABN       string    `gorm:"size:50"          json:"abn"`     // Customer.ABN
	Address   string    `gorm:"size:255"         json:"address"` // Customer.Address
	Version   int64     `gorm:"default:1;column:version"        json:"version"`
	CreatedAt time.Time `gorm:"column:created_at"               json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"               json:"updated_at"`
	IsDeleted bool      `gorm:"default:false;column:is_deleted" json:"is_deleted"`
	Contact   string    `gorm:"size:100" json:"contact"`
}

// TableName 显式指定表名
func (Customer) TableName() string {
	return "customers"
}
