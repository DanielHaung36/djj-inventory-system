// internal/handler/auth_helpers.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// writeAuthResponse 负责：
//  1. 把 token 写到 HttpOnly Cookie
//  2. 把同名 JSON body 写回客户端
func writeAuthResponse(c *gin.Context, token string, payload any) {
	// 1) 写 Cookie
	c.SetCookie(
		"access_token",
		token,
		24*3600, // 1 天
		"/",
		"",    // domain
		false, // secure
		true,  // httpOnly
	)

	// 2) 写 JSON
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  payload,
	})
}
