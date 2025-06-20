package handler

import (
	"context"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/model"
	"djj-inventory-system/internal/pkg/auth"
	"djj-inventory-system/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware injects ContextUserIDKey
func AuthMiddleware(svc service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if uid, err := auth.GetUserID(r); err == nil && uid > 0 {
				ctx := context.WithValue(r.Context(), model.ContextUserIDKey, uid)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := auth.GetUserID(c.Request)
		logger.Debugf("uid: %v", uid)
		logger.Debugf(c.FullPath())
		if c.FullPath() == "/api/login" || c.FullPath() == "/api/logout" || c.FullPath() == "/api/register" || c.FullPath() == "/api/roles" {
			c.Next()
			return // ← 加上这句 跑你的登录 handler 然后直接 return，不会再继续执行后面的登录检查中间件。
		}
		if err != nil || uid == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		// 3. 注入到 Gin 自己的上下文里
		c.Set(string(model.ContextUserIDKey), uid)
		c.Next()
	}
}

// RequireLogin 只允许已登录的用户继续，未登录的返回 401
func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get(string(model.ContextUserIDKey)); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要先登录"})
			c.Abort()
			return
		}
		c.Next()
	}
}

//// PermissionMiddleware guards by permission name
//func PermissionMiddleware(svc service.UserService, perm string) func(http.Handler) http.Handler {
//	return func(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			uid := r.Context().Value(model.ContextUserIDKey).(uint)
//			ok, err := svc.CheckPermission(r.Context(), uid, perm)
//			if err != nil || !ok {
//				http.Error(w, "Forbidden", http.StatusForbidden)
//				return
//			}
//			next.ServeHTTP(w, r)
//		})
//	}
//}
