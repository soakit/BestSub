package handlers

import (
	"bytes"
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/store"
	"github.com/bestruirui/bestsub/pkg/countries"

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
			router.NewRoute("/convert", http.MethodPost).
				Handle(nodeConvert),
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
	nodes := store.NodeList()
	// 国家名称只在接口响应时生成，避免展示字段进入节点缓存和检测写回链路。
	for i := range nodes {
		nodes[i].CountryName = countries.Get(nodes[i].CountryCode).NameZh
	}
	resp.Success(c, nodes)
}

func nodeConvert(c *gin.Context) {
	var current model.NodeConfig
	if err := c.ShouldBindJSON(&current); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	converted, err := node.Convert(c.Request.Context(), []byte(current.Content), node.ConvertTargetMihomo)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	current.Content = string(bytes.TrimPrefix(bytes.Split(converted, []byte("\n"))[1], []byte(" - ")))
	resp.Success(c, current.Content)
}

func nodeCreate(c *gin.Context) {
	var current model.Node
	if err := c.ShouldBindJSON(&current); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	current.TrafficMultiplier = node.ParseTrafficMultiplier(current.Content)
	if err := store.NodeCreate(&current); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, current)
}

func nodeUpdate(c *gin.Context) {
	id := c.Param("id")
	var current model.Node
	if err := c.ShouldBindJSON(&current); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	current.TrafficMultiplier = node.ParseTrafficMultiplier(current.Content)
	if err := store.NodeUpdate(id, &current); err != nil {
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
