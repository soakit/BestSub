package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	config := cors.DefaultConfig()
	// 携带 cookie 的跨域响应不能使用 "*"，这里回显 Origin 以保留原本允许任意来源的行为。
	config.AllowOriginFunc = func(origin string) bool {
		return origin != ""
	}
	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	return cors.New(config)
}
