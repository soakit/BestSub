package handlers

import (
	"net/http"

	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/storage").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(storageList),
		)
}

func storageList(c *gin.Context) {
	resp.Success(c, store.StorageList())
}
