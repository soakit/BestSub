package handlers

import (
	"net/http"

	"github.com/bestruirui/bestsub/internal/conf"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1").
		AddRoute(router.NewRoute("/version", http.MethodGet).Handle(version))
}

func version(c *gin.Context) {
	resp.Success(c, conf.Version)
}
