package database

import (
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/gorm"
)

// EnsurePublicMarketNavigationV5 makes the public supply-and-demand entry
// discoverable on upgraded databases and refreshes the two vendor-directory
// icons that are part of the canonical navigation.
func EnsurePublicMarketNavigationV5(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var rows []model.Menu
		if err := tx.Unscoped().Where("menu_type = ? AND parent_id = ? AND path = ?", "top", 0, "/purchase").Order("id asc").Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			if err := tx.Create(&model.Menu{Name: "供求信息", Icon: "clipboard", MenuType: "top", Path: "/purchase", SortOrder: 5, IsEnabled: true}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Unscoped().Model(&model.Menu{}).Where("id = ?", rows[0].ID).Updates(map[string]any{
				"name": "供求信息", "icon": "clipboard", "sort_order": 5, "is_enabled": true, "deleted_at": nil,
			}).Error; err != nil {
				return err
			}
			if len(rows) > 1 {
				ids := make([]uint, 0, len(rows)-1)
				for _, row := range rows[1:] {
					ids = append(ids, row.ID)
				}
				if err := tx.Unscoped().Model(&model.Menu{}).Where("id IN ?", ids).Update("is_enabled", false).Error; err != nil {
					return err
				}
			}
		}
		if err := reorderPublicMenus(tx, "top", []string{"/", "/vendors", "/products", "/service", "/purchase"}); err != nil {
			return err
		}
		if err := tx.Model(&model.VendorCategory{}).Where("name = ? AND parent_id = ?", "农机整机厂", 0).Update("icon", "tractor").Error; err != nil {
			return err
		}
		return tx.Model(&model.StaticPageBuild{}).
			Where("resource_type IN ? AND status = ?", []string{"vendor", "product"}, model.StaticPageStatusReady).
			Updates(map[string]any{"status": model.StaticPageStatusStale, "error_message": "供求公开导航已更新，请重新生成静态页面", "updated_at": time.Now()}).Error
	})
}
