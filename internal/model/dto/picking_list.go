package dto

// —— DTO for JSON input / template 渲染 ——

type PickingListDTO struct {
	PickingNumber   string               `json:"pickingNumber"`
	DeliveryAddress string               `json:"deliveryAddress"`
	Items           []PickingListItemDTO `json:"items"`
}

type PickingListItemDTO struct {
	ProductID uint   `json:"productId"`
	Quantity  int    `json:"quantity"`
	Location  string `json:"location"`
}
