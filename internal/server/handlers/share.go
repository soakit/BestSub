package handlers

import (
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	shareService "github.com/bestruirui/bestsub/internal/share"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/share").
		Use(middleware.Auth()).
		AddRoute(router.NewRoute("/list", http.MethodGet).Handle(shareList)).
		AddRoute(router.NewRoute("/get/:id", http.MethodGet).Handle(shareGet)).
		AddRoute(router.NewRoute("/create", http.MethodPost).Handle(shareCreate)).
		AddRoute(router.NewRoute("/update/:id", http.MethodPut).Handle(shareUpdate)).
		AddRoute(router.NewRoute("/del/:id", http.MethodDelete).Handle(shareDelete))
	router.NewGroupRouter("/share").
		AddRoute(router.NewRoute("/:token", http.MethodGet).Handle(shareContent))
}

func shareList(c *gin.Context) {
	shares := store.ShareList()
	for i := range shares {
		count, err := shareService.Count(shares[i])
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		shares[i].NodeCount = count
	}
	resp.Success(c, shares)
}

func shareGet(c *gin.Context) {
	share, ok := store.ShareGet(c.Param("id"))
	if !ok {
		resp.Error(c, http.StatusNotFound, resp.ErrResourceNotFound)
		return
	}
	count, err := shareService.Count(share)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.NodeCount = count
	resp.Success(c, share)
}

func shareCreate(c *gin.Context) {
	var share model.Share
	if err := c.ShouldBindJSON(&share); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := shareService.NormalizeConfig(&share.ShareConfig); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := store.ShareCreate(&share); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, share)
}

func shareUpdate(c *gin.Context) {
	if _, ok := store.ShareGet(c.Param("id")); !ok {
		resp.Error(c, http.StatusNotFound, resp.ErrResourceNotFound)
		return
	}
	var config model.ShareConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := shareService.NormalizeConfig(&config); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := store.ShareUpdateConfig(c.Param("id"), config); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "share updated successfully")
}

func shareDelete(c *gin.Context) {
	if _, ok := store.ShareGet(c.Param("id")); !ok {
		resp.Error(c, http.StatusNotFound, resp.ErrResourceNotFound)
		return
	}
	if err := store.ShareDelete(c.Param("id")); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "share deleted successfully")
}

func shareContent(c *gin.Context) {
	share, ok := store.ShareGetByToken(c.Param("token"))
	if !ok {
		c.String(http.StatusNotFound, "share not found")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Type", "text/plain; charset=utf-8")
	if err := shareService.Write(share, c.Writer); err != nil {
		log.Errorf("render share %s: %v", share.ID, err)
		if !c.Writer.Written() {
			c.Header("Content-Type", "text/plain; charset=utf-8")
			c.String(http.StatusInternalServerError, "share render failed")
		}
		return
	}
}
