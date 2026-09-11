package database

import (
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"gorm.io/gorm"
)

// Drop child tables first. Each step is repeatable after an interrupted MySQL DDL migration.
func removeSupplyDemand(db *gorm.DB) error {
	for _, table := range []string{"market_post_media", "market_contact_access_logs", "market_posts"} {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				return err
			}
		}
	}
	if db.Migrator().HasTable(&model.Menu{}) {
		var menus []model.Menu
		if err := db.Find(&menus).Error; err != nil {
			return err
		}
		removed := map[uint]bool{}
		for _, menu := range menus {
			if service.IsRetiredPublicPath(menu.Path) {
				removed[menu.ID] = true
			}
		}
		for changed := true; changed; {
			changed = false
			for _, menu := range menus {
				if removed[menu.ParentID] && !removed[menu.ID] {
					removed[menu.ID] = true
					changed = true
				}
			}
		}
		for id := range removed {
			if err := db.Unscoped().Delete(&model.Menu{}, id).Error; err != nil {
				return err
			}
		}
	}
	if db.Migrator().HasTable(&model.ContentPage{}) {
		if err := db.Unscoped().Where("slug = ?", "purchase").Delete(&model.ContentPage{}).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable(&model.StaticPageBuild{}) {
		if err := db.Model(&model.StaticPageBuild{}).Where("status = ?", "ready").Updates(map[string]any{"status": "stale", "error_message": "导航功能调整，请重新生成静态页面"}).Error; err != nil {
			return err
		}
	}
	return nil
}
