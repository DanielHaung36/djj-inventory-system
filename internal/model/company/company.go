package company

import "time"

type Company struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Contact       string    `gorm:"size:100" json:"contact"`
	Phone         string    `gorm:"size:50" json:"phone"`
	Email         string    `gorm:"size:100" json:"email"`
	Website       string    `gorm:"size:255" json:"website"`
	ABN           string    `gorm:"size:20" json:"abn"`
	Address       string    `gorm:"size:255" json:"address"`
	BankName      string    `gorm:"size:100" json:"bankName"`
	BSB           string    `gorm:"size:20" json:"bsb"`
	AccountNumber string    `gorm:"size:50" json:"accountNumber"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (Company) TableName() string { return "companies" }
