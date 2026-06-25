package database

import (
	"fmt"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	serverDB, err := gorm.Open(mysql.Open(cfg.ServerDSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := serverDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName)).Error; err != nil {
		return nil, err
	}
	return gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Menu{},
		&model.Tag{},
		&model.Vendor{},
		&model.VendorTag{},
		&model.Category{},
		&model.Product{},
		&model.Banner{},
		&model.SiteConfig{},
		&model.AdminUser{},
	)
}
