package store

import (
	"github.com/bestruirui/bestsub/internal/conf"
	"github.com/bestruirui/bestsub/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

var db *gorm.DB

func InitDB() error {
	var err error
	gormConfig := gorm.Config{Logger: logger.Discard}
	if conf.IsDebug() {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}
	db, err = gorm.Open(sqlite.Open(conf.AppConfig.Database.Path), &gormConfig)
	if err != nil {
		return err
	}
	return db.AutoMigrate(
		new(model.User),
		new(model.Subscription),
		new(model.Node),
		new(model.Tag),
		new(model.Storage),
		new(model.Task),
		new(model.Setting),
	)
}

func InitStore() error {
	if err := initSetting(); err != nil {
		return err
	}
	if err := UserInit(); err != nil {
		return err
	}
	if err := initSubscription(); err != nil {
		return err
	}
	if err := initNode(); err != nil {
		return err
	}
	if err := initTag(); err != nil {
		return err
	}
	if err := initStorage(); err != nil {
		return err
	}
	if err := initTask(); err != nil {
		return err
	}
	return nil
}

func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
