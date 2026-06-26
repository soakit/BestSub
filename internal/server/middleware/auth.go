package middleware

import (
	"net/http"

	"bestsub/internal/server/resp"
	"bestsub/internal/store"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("authenticated") != true || session.Get("auth_version") != store.UserAuthVersion() {
			session.Clear()
			session.Options(sessions.Options{
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			_ = session.Save()
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}
