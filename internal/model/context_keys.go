package model

// 用来在 context 中存放／读取当前登陆用户 ID
type contextKey string

const (
	ContextUserIDKey contextKey = "userID"
)
