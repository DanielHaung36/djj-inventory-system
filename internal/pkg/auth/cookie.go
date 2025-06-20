package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

type SessionData struct {
	UserID      uint     `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

var (
	// ❗️生产环境请用 securecookie.GenerateRandomKey(32) 生成，并注入 ENV 32位用于加密解密
	hashKey  = []byte("0123456789abdjjf0123456789abcdef") // 32 bytes
	blockKey = []byte("fedcba9djj543210fedcba9876s43210") // 32 bytes
	s        = securecookie.New(hashKey, blockKey)
)

// SetSession 写入 “session” Cookie，7 天后过期
func SetSession(sd *SessionData, w http.ResponseWriter) error {
	encoded, err := s.Encode("session", sd)
	if err != nil {
		return err
	}
	maxAge := 7 * 24 * 60 * 60 // 7 天 (秒)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    encoded,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// GetSessionData 解析 “session” Cookie，返回完整的 SessionData
func GetSessionData(r *http.Request) (*SessionData, error) {
	c, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}
	var sd SessionData
	if err := s.Decode("session", c.Value, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}

// ClearSession 删除 “session” Cookie
func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetUserID 只是从 SessionData 里取 UserID（保留旧接口兼容）
func GetUserID(r *http.Request) (uint, error) {
	sd, err := GetSessionData(r)
	if err != nil {
		return 0, err
	}
	if sd.UserID == 0 {
		return 0, errors.New("no user id in session")
	}
	return sd.UserID, nil
}
