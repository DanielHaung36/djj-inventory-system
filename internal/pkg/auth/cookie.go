package auth

import (
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
	// generate your own keys once and store them in env/config
	hashKey  = []byte("djj-very-secret-hash-key")
	blockKey = []byte("a-lot-secret-block-key-djj")
	s        = securecookie.New(hashKey, blockKey)
)

// SetSession writes a signed “user_id” cookie
// SetSession 写入 “session” Cookie，7 天后过期
func SetSession(sd *SessionData, w http.ResponseWriter) error {
	encoded, err := s.Encode("session", sd)
	if err != nil {
		return err
	}

	// 7 天后过期
	maxAge := 7 * 24 * 60 * 60 // seconds

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    encoded,
		Path:     "/",
		Domain:   "",     // 或者你自己的域名
		MaxAge:   maxAge, // 浏览器会自动删除
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		HttpOnly: true,
		Secure:   true, // 如在 HTTPS 下
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// ClearSession removes the cookie
func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

// GetUserID parses and returns the user_id from cookie (or zero)
func GetUserID(r *http.Request) (uint, error) {
	if c, err := r.Cookie("session"); err == nil {
		var value map[string]interface{}
		if err = s.Decode("session", c.Value, &value); err == nil {
			if idf, ok := value["user_id"].(float64); ok {
				return uint(idf), nil
			}
		}
	}
	return 0, http.ErrNoCookie
}
