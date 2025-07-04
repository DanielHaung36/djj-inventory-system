package handler

import (
	"context"
	"djj-inventory-system/internal/model/common"
	"djj-inventory-system/internal/pkg/auth"
	"djj-inventory-system/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware injects ContextUserIDKey
func AuthMiddleware(svc service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if uid, err := auth.GetUserID(r); err == nil && uid > 0 {
				ctx := context.WithValue(r.Context(), common.ContextUserIDKey, uid)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/api/auth/login" || c.FullPath() == "/api/auth//logout" || c.FullPath() == "/api/auth//register" || c.FullPath() == "/api/auth/roles" {
			c.Next()
			return // ← 加上这句 跑你的登录 handler 然后直接 return，不会再继续执行后面的登录检查中间件。
		}

		// 1. 从 Cookie 里读 token
		tokenString, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}

		// 2. 解析并校验 JWT
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的 token"})
			return
		}

		// 3. 取出 claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无法读取 token 声明"})
			return
		}
		// 4. 拿出 sub（userID）和 name（fullName）
		if sub, ok := claims["sub"].(float64); ok {
			c.Set("currentUserId", int32(sub))
		}
		if name, ok := claims["name"].(string); ok {
			c.Set("currentUser", name)
		}
		if role, ok := claims["role"].(string); ok {
			c.Set("currentUserRole", role)
		}
		if perms, ok := claims["permissions"].([]interface{}); ok {
			// 转成 []string
			pp := make([]string, len(perms))
			for i, v := range perms {
				if s, ok := v.(string); ok {
					pp[i] = s
				}
			}
			c.Set("currentUserPermissions", pp)
		}
		// 3. 注入到 Gin 自己的上下文里
		c.Set(string(common.ContextUserIDKey), claims["sub"])
		c.Next()
	}
}

// RequireLogin 只允许已登录的用户继续，未登录的返回 401
func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get(string(common.ContextUserIDKey)); !exists {
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
