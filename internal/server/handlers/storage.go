package handlers

import (
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	storagepkg "github.com/bestruirui/bestsub/internal/storage"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/storage").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(storageList),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(storageCreate),
		).
		AddRoute(
			router.NewRoute("/test", http.MethodPost).
				Handle(storageTest),
		).
		AddRoute(
			router.NewRoute("/update/:id", http.MethodPut).
				Handle(storageUpdate),
		).
		AddRoute(
			router.NewRoute("/del/:id", http.MethodDelete).
				Handle(storageDelete),
		)
}

func storageList(c *gin.Context) {
	resp.Success(c, store.StorageList())
}

func storageCreate(c *gin.Context) {
	var config model.StorageTargetConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	storage := model.Storage{StorageTargetConfig: config}
	if err := store.StorageCreate(&storage); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, storage)
}

func storageTest(c *gin.Context) {
	var config model.StorageTargetConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := storagepkg.Test(c.Request.Context(), storagepkg.Type(config.Type), config.Params); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, "storage test successfully")
}

func storageUpdate(c *gin.Context) {
	var config model.StorageTargetConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.StorageUpdateConfig(c.Param("id"), config); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "storage updated successfully")
}

func storageDelete(c *gin.Context) {
	if err := store.StorageDelete(c.Param("id")); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "storage deleted successfully")
}
