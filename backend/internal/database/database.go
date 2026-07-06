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
	if err := db.AutoMigrate(
		&model.Menu{},
		&model.Tag{},
		&model.Vendor{},
		&model.VendorTag{},
		&model.Category{},
		&model.Product{},
		&model.Banner{},
		&model.SiteConfig{},
		&model.ContentPage{},
		&model.FriendLink{},
		&model.OperationLog{},
		&model.AdminUser{},
	); err != nil {
		return err
	}
	if err := cleanupDuplicateVendorTags(db); err != nil {
		return err
	}
	return ensureVendorTagUniqueIndex(db)
}

func cleanupDuplicateVendorTags(db *gorm.DB) error {
	return db.Exec(`
		DELETE vt1 FROM vendor_tags vt1
		INNER JOIN vendor_tags vt2
			ON vt1.vendor_id = vt2.vendor_id
			AND vt1.tag_id = vt2.tag_id
			AND vt1.id > vt2.id
	`).Error
}

func ensureVendorTagUniqueIndex(db *gorm.DB) error {
	var count int64
	if err := db.Raw(`
		SELECT COUNT(1)
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
			AND table_name = 'vendor_tags'
			AND index_name = 'idx_vendor_tags_vendor_id_tag_id'
	`).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Exec("CREATE UNIQUE INDEX idx_vendor_tags_vendor_id_tag_id ON vendor_tags (vendor_id, tag_id)").Error
}
