package company

import "time"

// Company 对应 companies 表
type Company struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Code          string    `gorm:"size:50;uniqueIndex;not null" json:"code"` // ← 新增：业务可用的唯一标识
	Name          string    `gorm:"size:100;not null"               json:"name"`
	Email         string    `gorm:"size:100"                        json:"email"`
	Phone         string    `gorm:"size:50"                         json:"phone"`
	Website       string    `gorm:"size:255"                        json:"website"`
	ABN           string    `gorm:"size:50"                         json:"abn"`
	Address       string    `gorm:"size:255"                        json:"address"`
	BankName      string    `gorm:"size:100"                        json:"bank_name"`
	BSB           string    `gorm:"size:20"                         json:"bsb"`
	AccountNumber string    `gorm:"size:50"                         json:"account_number"`
	IsDefault     bool      `gorm:"not null;default:false"         json:"is_default"`
	CreatedAt     time.Time `gorm:"autoCreateTime"                 json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"                 json:"updated_at"`
}

func (Company) TableName() string { return "companies" }
