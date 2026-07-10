package handlers

import (
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/store"
	"strconv"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/tag").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(tagList),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(tagCreate),
		).
		AddRoute(
			router.NewRoute("/del/:id", http.MethodDelete).
				Handle(tagDelete),
		)
}

func tagList(c *gin.Context) {
	resp.Success(c, store.TagList())
}

func tagCreate(c *gin.Context) {
	var tag model.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.TagCreate(&tag); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, tag)
}

func tagDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.TagDelete(uint(id)); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "tag deleted successfully")
}
