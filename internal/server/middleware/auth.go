package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bestruirui/bestsub/internal/conf"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/gin-gonic/gin"
)

// SignAuthSecret 生成签名 cookie 值，格式：expiry.signature
func SignAuthSecret(secret string, maxAge int) string {
	expiry := time.Now().Add(time.Duration(maxAge) * time.Second).Unix()
	payload := secret + ":" + strconv.FormatInt(expiry, 10)
	mac := hmac.New(sha256.New, []byte(conf.APP_NAME))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return strconv.FormatInt(expiry, 10) + "." + sig
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("auth")
		if err != nil || token == "" {
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(token, ".", 2)
		if len(parts) != 2 {
			c.SetCookie("auth", "", -1, "/", "", false, true)
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}

		expiry, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || time.Now().Unix() > expiry {
			c.SetCookie("auth", "", -1, "/", "", false, true)
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}

		payload := store.UserAuthSecret() + ":" + parts[0]
		mac := hmac.New(sha256.New, []byte(conf.APP_NAME))
		mac.Write([]byte(payload))
		expected := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(parts[1]), []byte(expected)) {
			c.SetCookie("auth", "", -1, "/", "", false, true)
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Next()
	}
}
