package handlers

import (
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/service"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/sub").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(list),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(create),
		).
		AddRoute(
			router.NewRoute("/:id", http.MethodPut).
				Handle(update),
		).
		AddRoute(
			router.NewRoute("/:id", http.MethodDelete).
				Handle(delete),
		).
		AddRoute(
			router.NewRoute("/refresh/:id", http.MethodPost).
				Handle(refresh),
		).
		AddRoute(
			router.NewRoute("/refresh", http.MethodGet).
				Handle(refreshStream),
		).
		AddRoute(
			router.NewRoute("/:id", http.MethodGet).
				Handle(get),
		)
}

func list(c *gin.Context) {
	resp.Success(c, store.SubscriptionList())
}

func get(c *gin.Context) {
	sub, ok := store.SubscriptionGet(c.Param("id"))
	if !ok {
		resp.Error(c, http.StatusNotFound, resp.ErrResourceNotFound)
		return
	}
	resp.Success(c, sub)
}

func create(c *gin.Context) {
	var sub model.Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.SubscriptionCreate(&sub); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, sub)
}

func update(c *gin.Context) {
	var config model.SubscriptionConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.SubscriptionUpdateConfig(c.Param("id"), config); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "subscription updated successfully")
}

func delete(c *gin.Context) {
	if err := store.SubscriptionDelete(c.Param("id")); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "subscription deleted successfully")
}

func refresh(c *gin.Context) {
	if err := service.RefreshSubscription(c.Param("id")); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, nil)
}

func refreshStream(c *gin.Context) {
	service.RefreshEvents.Subscribe(c)
}
