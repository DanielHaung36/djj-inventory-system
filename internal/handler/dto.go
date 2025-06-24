// internal/handler/dto.go
package handler

// 用户返回给前端的格式
type UserDTO struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
	LastLogin string `json:"lastLogin"`
}

// 权限模块返回给前端的格式
type PermissionModuleDTO struct {
	Module      string          `json:"module"`
	Icon        string          `json:"icon"`
	Description string          `json:"description"`
	Permissions []PermissionDTO `json:"permissions"`
}

// 单个权限项
type PermissionDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// 更新／查询用户权限时的 Response
type UserPermissionDataDTO struct {
	UserID       uint   `json:"userId"`
	Permissions  []uint `json:"permissions"`
	LastModified string `json:"lastModified"`
	ModifiedBy   string `json:"modifiedBy"`
}
