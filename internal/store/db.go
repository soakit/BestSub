package store

import (
	"net/url"

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
	params := url.Values{}
	params.Add("_pragma", "journal_mode(WAL)")
	params.Add("_pragma", "synchronous(NORMAL)")
	params.Add("_pragma", "cache_size(10000)")
	params.Add("_pragma", "busy_timeout(5000)")
	params.Add("_pragma", "foreign_keys(ON)")
	params.Add("_pragma", "auto_vacuum(INCREMENTAL)")
	params.Add("_pragma", "mmap_size(268435456)")
	params.Add("_pragma", "locking_mode(NORMAL)")
	db, err = gorm.Open(sqlite.Open(conf.AppConfig.Database.Path+"?"+params.Encode()), &gormConfig)
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
		new(model.Share),
		new(model.Setting),
		new(model.RenameTemplate),
	)
}

func InitStore() error {
	if err := initSetting(); err != nil {
		return err
	}
	if err := initRenameTemplate(); err != nil {
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
	if err := initShare(); err != nil {
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
