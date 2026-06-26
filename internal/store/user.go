package store

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"bestsub/internal/model"

	"github.com/charmbracelet/log"
	"gorm.io/gorm"
)

var (
	userCache model.User
	// 用户资料变更时推进版本号，用于让旧 cookie session 失效。
	userAuthVersion atomic.Int64
)

func UserInit() error {
	if err := db.First(&userCache).Error; err == nil {
		userAuthVersion.Store(time.Now().UnixNano())
		return nil
	}
	userCache.Username = "admin"
	userCache.Password = "admin"
	if err := userCache.HashPassword(); err != nil {
		return err
	}
	if err := db.Create(&userCache).Error; err != nil {
		return err
	}
	userAuthVersion.Store(time.Now().UnixNano())
	log.Infof("initial user: admin,password: admin")
	return nil
}

func UserChangePassword(oldPassword, newPassword string) error {
	if err := userCache.ComparePassword(oldPassword); err != nil {
		return fmt.Errorf("incorrect old password: %w", err)
	}

	userCache.Password = newPassword
	if err := userCache.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := db.Model(&userCache).Update("password", userCache.Password).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	userAuthVersion.Store(time.Now().UnixNano())
	return nil
}

func UserChangeUsername(newUsername string) error {
	if userCache.Username == newUsername {
		return fmt.Errorf("new username is the same as the old username")
	}
	userCache.Username = newUsername
	if err := db.Model(&userCache).Update("username", userCache.Username).Error; err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}
	userAuthVersion.Store(time.Now().UnixNano())
	return nil
}

func UserSet(username, password string) error {
	userCache.Username = username
	userCache.Password = password
	if err := userCache.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	var user model.User
	if err := db.First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get user: %w", err)
		}
		if err := db.Create(&userCache).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		userAuthVersion.Store(time.Now().UnixNano())
		return nil
	}

	userCache.ID = user.ID
	if err := db.Model(&user).Updates(map[string]any{
		"username": userCache.Username,
		"password": userCache.Password,
	}).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	userAuthVersion.Store(time.Now().UnixNano())
	return nil
}

func UserVerify(username, password string) error {
	if username != userCache.Username {
		return fmt.Errorf("incorrect username")
	}
	if err := userCache.ComparePassword(password); err != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil
}

func UserAuthVersion() int64 {
	return userAuthVersion.Load()
}
