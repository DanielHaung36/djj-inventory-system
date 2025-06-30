// internal/dto/customer.go
package dto

// CustomerCreateDTO 前端传入创建客户的数据
type CustomerCreateDTO struct {
	StoreID uint   `json:"store_id" binding:"required"`
	Type    string `json:"type"` // retail|wholesale|online
	Name    string `json:"name"  binding:"required"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

// CustomerUpdateDTO 更新客户
type CustomerUpdateDTO struct {
	Type    *string `json:"type,omitempty"`
	Name    *string `json:"name,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Email   *string `json:"email,omitempty"`
	Address *string `json:"address,omitempty"`
}
