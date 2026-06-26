package middleware

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func StaticEmbed(urlPrefix string, embeddedFS fs.FS) gin.HandlerFunc {
	return static(urlPrefix, http.FS(embeddedFS))
}

func StaticLocal(urlPrefix string, localPath string) gin.HandlerFunc {
	return static(urlPrefix, http.Dir(localPath))
}

func static(urlPrefix string, fileSystem http.FileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fileSystem)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}
		if _, err := fileSystem.Open(c.Request.URL.Path); err == nil {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}
