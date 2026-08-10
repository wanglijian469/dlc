package database

import (
	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/gorm"
)

// PrioritizeVendorPublicNavigationV4 applies the vendor-first order once. It
// keeps labels, enabled states and custom menu rows intact.
func PrioritizeVendorPublicNavigationV4(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := reorderPublicMenus(tx, "top", []string{"/", "/vendors", "/products", "/service"}); err != nil {
			return err
		}
		if err := reorderPublicMenus(tx, "mobile_bottom", []string{"/", "/vendors", "/products", "/service", "/account/login"}); err != nil {
			return err
		}
		return tx.Model(&model.StaticPageBuild{}).
			Where("resource_type = ? AND status = ?", "vendor", model.StaticPageStatusReady).
			Updates(map[string]any{"status": model.StaticPageStatusStale, "error_message": "厂商优先导航已更新，请重新生成静态页面"}).Error
	})
}

func reorderPublicMenus(db *gorm.DB, menuType string, preferredPaths []string) error {
	var menus []model.Menu
	if err := db.Where("menu_type = ? AND parent_id = ?", menuType, 0).Order("sort_order asc, id asc").Find(&menus).Error; err != nil {
		return err
	}
	ordered := orderMenusByPath(menus, preferredPaths)
	for index, menu := range ordered {
		if err := db.Model(&model.Menu{}).Where("id = ?", menu.ID).UpdateColumn("sort_order", index+1).Error; err != nil {
			return err
		}
	}
	return nil
}

func orderMenusByPath(menus []model.Menu, preferredPaths []string) []model.Menu {
	consumed := make([]bool, len(menus))
	result := make([]model.Menu, 0, len(menus))
	for _, path := range preferredPaths {
		for index, menu := range menus {
			if !consumed[index] && menu.Path == path {
				result = append(result, menu)
				consumed[index] = true
				break
			}
		}
	}
	for index, menu := range menus {
		if !consumed[index] {
			result = append(result, menu)
		}
	}
	return result
}
