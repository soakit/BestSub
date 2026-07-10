package handlers

import (
	"errors"
	"net/http"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/store"
	taskservice "github.com/bestruirui/bestsub/internal/task"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/task").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(taskList),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(taskCreate),
		).
		AddRoute(
			router.NewRoute("/get/:id", http.MethodGet).
				Handle(taskGet),
		).
		AddRoute(
			router.NewRoute("/update/:id", http.MethodPut).
				Handle(taskUpdate),
		).
		AddRoute(
			router.NewRoute("/del/:id", http.MethodDelete).
				Handle(taskDelete),
		).
		AddRoute(
			router.NewRoute("/run/:id", http.MethodPost).
				Handle(taskRun),
		).
		AddRoute(
			router.NewRoute("/stop/:id", http.MethodPost).
				Handle(taskStop),
		).
		AddRoute(
			router.NewRoute("/result/:id", http.MethodGet).
				Handle(taskResult),
		).
		AddRoute(
			router.NewRoute("/stream", http.MethodGet).
				Handle(taskStream),
		)
}

func taskList(c *gin.Context) {
	resp.Success(c, store.TaskList())
}

func taskGet(c *gin.Context) {
	task, ok := store.TaskGet(c.Param("id"))
	if !ok {
		resp.Error(c, http.StatusNotFound, resp.ErrResourceNotFound)
		return
	}
	resp.Success(c, task)
}

func taskCreate(c *gin.Context) {
	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.TaskCreate(&task); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, task)
}

func taskUpdate(c *gin.Context) {
	var config model.TaskConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.TaskUpdateConfig(c.Param("id"), config); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "task updated successfully")
}

func taskDelete(c *gin.Context) {
	if err := store.TaskDelete(c.Param("id")); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "task deleted successfully")
}

func taskRun(c *gin.Context) {
	if err := taskservice.Run(c.Param("id")); err != nil {
		if errors.Is(err, taskservice.ErrTaskNotFound) {
			resp.Error(c, http.StatusNotFound, resp.ErrResourceNotFound)
			return
		}
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, nil)
}

func taskStop(c *gin.Context) {
	if err := taskservice.StopTask(c.Param("id")); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, nil)
}

func taskResult(c *gin.Context) {
	if _, ok := store.TaskGet(c.Param("id")); !ok {
		resp.Error(c, http.StatusNotFound, resp.ErrResourceNotFound)
		return
	}
	resp.Success(c, taskservice.ResultCount(c.Param("id")))
}

func taskStream(c *gin.Context) {
	taskservice.ProgressEvents.Subscribe(c)
}
