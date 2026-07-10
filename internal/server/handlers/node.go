package handlers

import (
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/node").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(nodeList),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(nodeCreate),
		).
		AddRoute(
			router.NewRoute("/update/:id", http.MethodPut).
				Handle(nodeUpdate),
		).
		AddRoute(
			router.NewRoute("/del/:id", http.MethodDelete).
				Handle(nodeDelete),
		)
}

func nodeList(c *gin.Context) {
	resp.Success(c, store.NodeList())
}

func nodeCreate(c *gin.Context) {
	var node model.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.NodeCreate(&node); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, node)
}

func nodeUpdate(c *gin.Context) {
	id := c.Param("id")
	var node model.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.NodeUpdate(id, &node); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "node updated successfully")
}

func nodeDelete(c *gin.Context) {
	id := c.Param("id")
	if err := store.NodeDelete(id); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "node deleted successfully")
}
