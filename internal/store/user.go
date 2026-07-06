package store

import (
	"errors"
	"fmt"

	"bestsub/internal/model"

	"github.com/charmbracelet/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var userCache model.User

func UserInit() error {
	if err := db.First(&userCache).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to query user: %w", err)
		}
		userCache.Username = "admin"
		userCache.Password = hashPassword("admin")
		if err := db.Create(&userCache).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		log.Infof("initial user: admin, password: admin")
	}
	return nil
}

func UserChangePassword(oldPassword, newPassword string) error {
	if err := comparePassword(userCache.Password, oldPassword); err != nil {
		return fmt.Errorf("incorrect old password: %w", err)
	}

	hashed := hashPassword(newPassword)
	if err := db.Model(&userCache).Updates(model.UserUpdate{Password: hashed}).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	userCache.Password = hashed
	return nil
}

func UserChangeUsername(newUsername string) error {
	if userCache.Username == newUsername {
		return fmt.Errorf("new username is the same as the old username")
	}
	userCache.Username = newUsername
	if err := db.Model(&userCache).Updates(model.UserUpdate{Username: userCache.Username}).Error; err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}
	return nil
}

func UserSet(username, password string) error {
	hashed := hashPassword(password)

	var user model.User
	if err := db.First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get user: %w", err)
		}
		userCache.Username = username
		userCache.Password = hashed
		if err := db.Create(&userCache).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		return nil
	}

	userCache.ID = user.ID
	userCache.Username = username
	userCache.Password = hashed
	if err := db.Model(&user).Updates(userCache).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func UserVerify(username, password string) error {
	if username != userCache.Username {
		return fmt.Errorf("incorrect username")
	}
	if err := comparePassword(userCache.Password, password); err != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil
}

func UserAuthSecret() string {
	return userCache.Username + ":" + userCache.Password
}

func hashPassword(plain string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(hashed)
}

func comparePassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}
