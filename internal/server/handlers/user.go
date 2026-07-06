package handlers

import (
	"net/http"
	"time"

	"bestsub/internal/model"
	"bestsub/internal/server/middleware"
	"bestsub/internal/server/resp"
	"bestsub/internal/server/router"
	"bestsub/internal/store"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/user").
		AddRoute(
			router.NewRoute("/login", http.MethodPost).
				Handle(login),
		)
	router.NewGroupRouter("/api/v1/user").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/check", http.MethodGet).
				Handle(check),
		).
		AddRoute(
			router.NewRoute("/logout", http.MethodPost).
				Handle(logout),
		).
		AddRoute(
			router.NewRoute("/change-password", http.MethodPost).
				Handle(changePassword),
		).
		AddRoute(
			router.NewRoute("/change-username", http.MethodPost).
				Handle(changeUsername),
		)
}

func login(c *gin.Context) {
	var user model.UserLogin
	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.UserVerify(user.Username, user.Password); err != nil {
		resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
		return
	}
	maxAge := int((15 * time.Minute).Seconds())
	if user.Trust {
		maxAge = int((30 * 24 * time.Hour).Seconds())
	}
	token := middleware.SignAuthSecret(store.UserAuthSecret(), maxAge)
	c.SetCookie("auth", token, maxAge, "/", "", false, true)
	resp.Success(c, "login successfully")
}

func check(c *gin.Context) {
	resp.Success(c, true)
}

func logout(c *gin.Context) {
	c.SetCookie("auth", "", -1, "/", "", false, true)
	resp.Success(c, "logout successfully")
}

func changePassword(c *gin.Context) {
	var user model.UserChangePassword
	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.UserChangePassword(user.OldPassword, user.NewPassword); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "password changed successfully")
}

func changeUsername(c *gin.Context) {
	var user model.UserChangeUsername
	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := store.UserChangeUsername(user.NewUsername); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "username changed successfully")
}
