package handler

// RegisterRequest 用户注册请求体
// swagger:model
// @Description 使用用户名、邮箱、密码和可选角色 ID 列表创建新用户
// @Param username body string true "用户名"
// @Param email body string true "邮箱"
// @Param password body string true "密码"
// @Param role_ids body []uint false "角色 ID 列表"
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	RoleIDs  []uint `json:"role_ids"`
}

// LoginRequest 用户登录请求体
// swagger:model
// @Description 使用用户名和密码登录
// @Param username body string true "用户名"
// @Param password body string true "密码"
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RoleCreateRequest 创建角色请求体
// swagger:model
// @Description 使用 name 字段创建新角色
// @Param name body string true "角色名称"
type RoleCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

// RoleUpdateRequest 更新角色请求体
// swagger:model
// @Description 更新角色名称
// @Param name body string true "角色名称"
type RoleUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

// PermCreateRequest 创建权限请求体
// swagger:model
// @Description 使用 action 和 object 字段创建权限
// @Param action body string true "权限动作"
// @Param object body string true "权限对象"
type PermCreateRequest struct {
	Action string `json:"action" binding:"required"`
	Object string `json:"object" binding:"required"`
}

// PermUpdateRequest 更新权限请求体
// swagger:model
// @Description 更新权限的 action 和 object
// @Param action body string true "权限动作"
// @Param object body string true "权限对象"
type PermUpdateRequest struct {
	Name string `json:"name" binding:"required"`
	Id   int    `json:"id" `
}

// UserCreateRequest 创建用户请求体
// swagger:model
// @Description 使用用户名、邮箱、密码和可选角色列表创建用户
// @Param username body string true "用户名"
// @Param email body string true "邮箱"
// @Param password body string true "密码"
// @Param role_ids body []uint false "角色 ID 列表"
type UserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	RoleIDs  []uint `json:"role_ids"`
}

// UserUpdateRequest 更新用户请求体
// swagger:model
// @Description 可选更新用户名、邮箱或密码
// @Param username body string false "用户名"
// @Param email body string false "邮箱"
// @Param password body string false "密码"
type UserUpdateRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

// ResponseMessage 通用消息响应体
// swagger:model
// @Description 通用返回消息结构
// @Param message body string true "消息内容"
type ResponseMessage struct {
	Message string `json:"message"`
}

// ErrorResponse 错误响应体
// swagger:model
// @Description 错误返回结构
// @Param error body string true "错误信息"
type ErrorResponse struct {
	Error string `json:"error"`
}
