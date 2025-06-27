// internal/model/company.go
package company

import "time"

type Company struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Code       string    `gorm:"size:50;unique;not null" json:"code"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Email      string    `gorm:"size:100" json:"email"`
	Phone      string    `gorm:"size:50" json:"phone"`
	Website    string    `gorm:"size:255" json:"website"`
	ABN        string    `gorm:"size:20" json:"abn"`
	Address    string    `gorm:"size:255" json:"address"`
	LogoBase64 string    `gorm:"type:text" json:"logo_base64"`
	BankName   string    `gorm:"size:100" json:"bank_name"`
	BSB        string    `gorm:"size:20" json:"bsb"`
	AccountNo  string    `gorm:"size:50" json:"account_no"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
