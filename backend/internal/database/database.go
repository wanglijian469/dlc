package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
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

func CleanupOrphanedMedia(db *gorm.DB, mediaDir string) error {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	var assets []model.MediaAsset
	if err := db.Where("status = ? AND created_at < ? AND NOT EXISTS (SELECT 1 FROM vendor_media WHERE vendor_media.asset_id = media_assets.id)", "staged", cutoff).Find(&assets).Error; err != nil {
		return err
	}
	for _, asset := range assets {
		url := fmt.Sprintf("/api/media/%d", asset.ID)
		var references int64
		if err := db.Model(&model.Vendor{}).
			Where("logo_asset_id = ? OR cover_asset_id = ? OR wechat_qr_code_asset_id = ?", asset.ID, asset.ID, asset.ID).
			Count(&references).Error; err != nil {
			return err
		}
		if references == 0 {
			if err := db.Model(&model.VendorSubmission{}).Where("status = ? AND INSTR(payload, ?) > 0", "pending", url).Count(&references).Error; err != nil {
				return err
			}
		}
		if references == 0 {
			if err := db.Model(&model.Product{}).Where("image = ? OR INSTR(gallery, ?) > 0 OR INSTR(specs, ?) > 0", url, url, url).Count(&references).Error; err != nil {
				return err
			}
		}
		if references == 0 {
			if err := db.Model(&model.ProductSupplier{}).Where("image = ? OR INSTR(gallery, ?) > 0 OR INSTR(specs, ?) > 0", url, url, url).Count(&references).Error; err != nil {
				return err
			}
		}
		if references == 0 {
			if err := db.Model(&model.ProductSubmission{}).Where("status = ? AND (INSTR(product_payload, ?) > 0 OR INSTR(supplier_payload, ?) > 0)", "pending", url, url).Count(&references).Error; err != nil {
				return err
			}
		}
		if references > 0 {
			continue
		}
		if err := db.Model(&asset).Update("status", "orphaned").Error; err != nil {
			return err
		}
		_ = os.Remove(filepath.Join(mediaDir, asset.StorageKey))
	}
	return nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.SchemaMigration{},
		&model.Menu{},
		&model.Tag{},
		&model.VendorCategory{},
		&model.VendorCategoryAssignment{},
		&model.Vendor{},
		&model.MediaAsset{},
		&model.VendorMedia{},
		&model.VendorTag{},
		&model.Category{},
		&model.Product{},
		&model.ProductSupplier{},
		&model.ProductSubmission{},
		&model.Banner{},
		&model.SiteConfig{},
		&model.ContentPage{},
		&model.FriendLink{},
		&model.OperationLog{},
		&model.AdminUser{},
		&model.VendorSubmission{},
		&model.AnalyticsEvent{},
		&model.SEORedirect{},
		&model.AuthSession{},
		&model.BuyerProfile{},
		&model.AppSession{},
		&model.MarketPost{},
		&model.MarketPostMedia{},
		&model.MarketContactAccessLog{},
		&model.ContentRevision{},
		&model.StaticPageBuild{},
		&model.StaticBuildJob{},
		&model.ContactAccessLog{},
		&model.ScrapeRiskEvent{},
		&model.ScrapeClientBlock{},
		&model.WatermarkBuildJob{},
	); err != nil {
		return err
	}
	if err := cleanupDuplicateVendorTags(db); err != nil {
		return err
	}
	if err := ensureVendorTagUniqueIndex(db); err != nil {
		return err
	}
	if err := ensureAnalyticsDedupeIndex(db); err != nil {
		return err
	}
	if err := cleanupLegacyDemoContent(db); err != nil {
		return err
	}
	if err := migrateVendorPublicationState(db); err != nil {
		return err
	}
	if err := dropRetiredVendorColumns(db); err != nil {
		return err
	}
	if err := backfillVendorImageAssetLinks(db); err != nil {
		return err
	}
	if err := migrateProductCatalog(db); err != nil {
		return err
	}
	if err := refreshLegacyDirectoryContent(db); err != nil {
		return err
	}
	if err := backfillCategoryMenuLinks(db); err != nil {
		return err
	}
	if err := migrateHomeDisplayConfiguration(db); err != nil {
		return err
	}
	if err := ensureAccessProtectionConfig(db); err != nil {
		return err
	}
	return ensureStructuredContent(db)
}

func ensureAccessProtectionConfig(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.SiteConfig{}).Where("config_key = ?", "security.antiScrape").Count(&count).Error; err != nil || count > 0 {
		return err
	}
	raw, _ := json.Marshal(map[string]any{
		"enabled": true, "auditOnly": true, "windowMinutes": 10, "distinctResourceLimit": 120,
		"blockHours": 1, "escalationStrikes": 3, "escalatedBlockHours": 24,
		"blockedAiAgents": []string{"GPTBot", "Google-Extended", "ClaudeBot", "CCBot", "PerplexityBot", "OAI-SearchBot"},
		"allowCidrs":      []string{}, "watermarkEnabled": true, "watermarkOpacity": 25, "watermarkText": "大陆农机配件",
	})
	return db.Create(&model.SiteConfig{ConfigKey: "security.antiScrape", ConfigValue: string(raw), Description: "公开访问、AI 爬虫与图片水印保护配置"}).Error
}

const CurrentSchemaVersion uint = 16

// Migrate is invoked explicitly by cmd/initdb in production. Development may
// opt in through RUN_MIGRATIONS=true for the existing one-command workflow.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.SchemaMigration{}); err != nil {
		return err
	}
	var latest uint
	if err := db.Model(&model.SchemaMigration{}).Select("COALESCE(MAX(version), 0)").Scan(&latest).Error; err != nil {
		return err
	}
	if latest >= CurrentSchemaVersion {
		return nil
	}
	if latest < 15 {
		if err := removeVendorAfterSalesService(db); err != nil {
			return err
		}
	}
	if err := AutoMigrate(db); err != nil {
		return err
	}
	if latest < 5 {
		if err := InitializeVendorSEO(db); err != nil {
			return err
		}
	}
	if latest < 9 {
		if err := db.Model(&model.StaticPageBuild{}).Where("status = ?", "ready").Updates(map[string]any{
			"status": "stale", "error_message": "联系方式公开策略升级，请重新生成静态页面",
		}).Error; err != nil {
			return err
		}
	}
	if latest < 10 {
		if err := InitializeVendorCategories(db); err != nil {
			return err
		}
		if err := InitializeVendorCategoryTaxonomyV2(db); err != nil {
			return err
		}
	}
	if latest < 11 {
		if err := CleanupDisabledVendorCategoriesV3(db); err != nil {
			return err
		}
	}
	if latest < 12 {
		if err := PrioritizeVendorPublicNavigationV4(db); err != nil {
			return err
		}
	}
	if err := BackfillPlatformData(db); err != nil {
		return err
	}
	if err := PurgeOrdinaryAccounts(db); err != nil {
		return err
	}
	return db.Create(&model.SchemaMigration{Version: CurrentSchemaVersion, Name: "vendor-product-self-entry-v1", AppliedAt: time.Now()}).Error
}

func removeVendorAfterSalesService(db *gorm.DB) error {
	if db.Migrator().HasTable(&model.VendorSubmission{}) {
		if err := db.Exec(`UPDATE vendor_submissions
			SET payload = JSON_REMOVE(payload, '$.afterSalesService')
			WHERE JSON_VALID(payload) AND JSON_CONTAINS_PATH(payload, 'one', '$.afterSalesService')`).Error; err != nil {
			return fmt.Errorf("remove afterSalesService from vendor submissions: %w", err)
		}
	}
	if db.Migrator().HasTable(&model.ContentRevision{}) {
		if err := db.Exec(`UPDATE content_revisions
			SET snapshot = JSON_REMOVE(snapshot, '$.afterSalesService')
			WHERE resource_type = 'vendor' AND JSON_VALID(snapshot) AND JSON_CONTAINS_PATH(snapshot, 'one', '$.afterSalesService')`).Error; err != nil {
			return fmt.Errorf("remove afterSalesService from vendor revisions: %w", err)
		}
	}
	if db.Migrator().HasTable(&model.StaticPageBuild{}) {
		if err := db.Model(&model.StaticPageBuild{}).
			Where("resource_type = ?", "vendor").
			Updates(map[string]any{"status": model.StaticPageStatusStale, "error_message": ""}).Error; err != nil {
			return fmt.Errorf("mark vendor static pages stale: %w", err)
		}
	}
	if db.Migrator().HasColumn(&model.Vendor{}, "after_sales_service") {
		if err := db.Migrator().DropColumn(&model.Vendor{}, "after_sales_service"); err != nil {
			return fmt.Errorf("drop vendors.after_sales_service: %w", err)
		}
	}
	return nil
}

func CheckMigrations(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.SchemaMigration{}) {
		return fmt.Errorf("database migrations are pending; run cmd/initdb before starting production")
	}
	var latest uint
	if err := db.Model(&model.SchemaMigration{}).Select("COALESCE(MAX(version), 0)").Scan(&latest).Error; err != nil {
		return err
	}
	if latest < CurrentSchemaVersion {
		return fmt.Errorf("database schema is version %d, require %d; run cmd/initdb", latest, CurrentSchemaVersion)
	}
	return nil
}

func ensureAnalyticsDedupeIndex(db *gorm.DB) error {
	if db.Migrator().HasIndex(&model.AnalyticsEvent{}, "idx_analytics_dedupe") {
		if err := db.Migrator().DropIndex(&model.AnalyticsEvent{}, "idx_analytics_dedupe"); err != nil {
			return err
		}
	}
	return db.Migrator().CreateIndex(&model.AnalyticsEvent{}, "idx_analytics_dedupe")
}

// CleanupAnalytics removes pseudonymous visit events after their stated
// retention period. No visitor-level analytics data is retained long-term.
func CleanupAnalytics(db *gorm.DB, retention time.Duration) error {
	if db == nil {
		return nil
	}
	return db.Where("created_at < ?", time.Now().Add(-retention)).Delete(&model.AnalyticsEvent{}).Error
}

// migrateHomeDisplayConfiguration removes settings for homepage blocks that
// are no longer rendered and keeps the three displayed modules as one source
// of truth. It also normalizes the public enrollment wording in persisted data.
func migrateHomeDisplayConfiguration(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("config_key IN ?", []string{"home.sections", "home.stats", "home.safeguards", "home.join"}).Delete(&model.SiteConfig{}).Error; err != nil {
			return err
		}

		var moduleConfig model.SiteConfig
		modules := service.DefaultHomeModules()
		if err := tx.Where("config_key = ?", "home.modules").First(&moduleConfig).Error; err == nil {
			var stored []service.HomeModule
			if json.Unmarshal([]byte(moduleConfig.ConfigValue), &stored) == nil {
				allowed := map[string]bool{"recommendedVendors": true, "moreVendors": true, "processingServices": true}
				filtered := make([]service.HomeModule, 0, len(stored))
				seen := map[string]bool{}
				for _, item := range stored {
					if allowed[item.Type] && !seen[item.Type] {
						filtered = append(filtered, item)
						seen[item.Type] = true
					}
				}
				defaults := service.DefaultHomeModules()
				byType := make(map[string]service.HomeModule, len(filtered))
				for _, item := range filtered {
					byType[item.Type] = item
				}
				modules = make([]service.HomeModule, 0, len(defaults))
				for _, fallback := range defaults {
					if item, ok := byType[fallback.Type]; ok {
						modules = append(modules, item)
					} else {
						modules = append(modules, fallback)
					}
				}
			}
		}
		payload, _ := json.Marshal(modules)
		if moduleConfig.ID == 0 {
			if err := tx.Create(&model.SiteConfig{ConfigKey: "home.modules", ConfigValue: string(payload), Description: "首页实际展示模块配置"}).Error; err != nil {
				return err
			}
		} else if err := tx.Model(&moduleConfig).Updates(map[string]interface{}{"config_value": string(payload), "description": "首页实际展示模块配置"}).Error; err != nil {
			return err
		}

		var metaConfig model.SiteConfig
		if err := tx.Where("config_key = ?", "site.meta").First(&metaConfig).Error; err == nil {
			var meta map[string]interface{}
			if json.Unmarshal([]byte(metaConfig.ConfigValue), &meta) == nil {
				meta["submitVendorText"] = "厂商入驻"
				if value, err := json.Marshal(meta); err == nil {
					if err := tx.Model(&metaConfig).Update("config_value", string(value)).Error; err != nil {
						return err
					}
				}
			}
		}
		if err := tx.Model(&model.Menu{}).Where("path = ? AND name = ?", "/join", "提交厂商").Update("name", "厂商入驻").Error; err != nil {
			return err
		}
		return tx.Model(&model.ContentPage{}).Where("slug = ?", "join").Updates(map[string]interface{}{"title": "厂商入驻", "summary": "提交入驻资料后平台运营人员会尽快联系。"}).Error
	})
}

// backfillCategoryMenuLinks turns existing sidebar/mobile category rows into
// stable anchors. They keep optional shortcut children but no longer own the
// displayed category name, icon, sort order, or path.
func backfillCategoryMenuLinks(db *gorm.DB) error {
	var categories []model.Category
	if err := db.Where("parent_id = ?", 0).Find(&categories).Error; err != nil {
		return err
	}
	if len(categories) == 0 {
		return nil
	}
	var menus []model.Menu
	if err := db.Where("menu_type IN ? AND parent_id = ? AND category_id = ?", []string{"sidebar", "mobile"}, 0, 0).Find(&menus).Error; err != nil {
		return err
	}
	for _, menu := range menus {
		if categoryID := categoryIDForLegacyMenu(menu, categories); categoryID > 0 {
			if err := db.Model(&menu).Update("category_id", categoryID).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func categoryIDForLegacyMenu(menu model.Menu, categories []model.Category) uint {
	if parsed, err := url.Parse(menu.Path); err == nil {
		if raw := parsed.Query().Get("categoryId"); raw != "" {
			if id, err := strconv.ParseUint(raw, 10, 64); err == nil {
				for _, category := range categories {
					if category.ID == uint(id) {
						return category.ID
					}
				}
			}
		}
	}
	for _, category := range categories {
		if strings.EqualFold(strings.TrimSpace(menu.Name), strings.TrimSpace(category.Name)) {
			return category.ID
		}
	}
	return 0
}

func refreshLegacyDirectoryContent(db *gorm.DB) error {
	const legacyBannerTitle = "查农机配件，找公开厂商资料"
	const updatedBannerTitle = "农机供应链，查农机配件，厂商信息"
	const updatedHotKeywords = "收割机链条,齿轮,皮带,液压油泵,刀片,滤芯"
	retiredVendorNames := []string{
		"江苏东成农机配件有限公司", "河南中联农机制造有限公司", "安徽豪华农机配件有限公司", "山东万鑫农机配件有限公司", "宁波动力机械有限公司",
		"浙江汉丰农机有限公司", "河北力捷机械有限公司", "辽宁佳丰农机配件有限公司", "四川川沃农机有限公司", "陕西恒农农机配件有限公司",
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Banner{}).Where("title = ?", legacyBannerTitle).Updates(map[string]interface{}{"title": updatedBannerTitle, "hot_keywords": updatedHotKeywords}).Error; err != nil {
			return err
		}

		var vendorIDs []uint
		if err := tx.Unscoped().Model(&model.Vendor{}).Where("name IN ?", retiredVendorNames).Pluck("id", &vendorIDs).Error; err != nil || len(vendorIDs) == 0 {
			return err
		}

		var productIDs []uint
		if err := tx.Unscoped().Model(&model.Product{}).Where("vendor_id IN ?", vendorIDs).Pluck("id", &productIDs).Error; err != nil {
			return err
		}
		supplierQuery := tx.Unscoped().Model(&model.ProductSupplier{}).Where("vendor_id IN ?", vendorIDs)
		if len(productIDs) > 0 {
			supplierQuery = supplierQuery.Or("product_id IN ?", productIDs)
		}
		var supplierIDs []uint
		if err := supplierQuery.Pluck("id", &supplierIDs).Error; err != nil {
			return err
		}

		submissionQuery := tx.Unscoped().Where("vendor_id IN ?", vendorIDs)
		if len(productIDs) > 0 {
			submissionQuery = submissionQuery.Or("product_id IN ?", productIDs)
		}
		if len(supplierIDs) > 0 {
			submissionQuery = submissionQuery.Or("supplier_id IN ?", supplierIDs)
		}
		if err := submissionQuery.Delete(&model.ProductSubmission{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("vendor_id IN ?", vendorIDs).Delete(&model.VendorSubmission{}).Error; err != nil {
			return err
		}
		if err := supplierQuery.Delete(&model.ProductSupplier{}).Error; err != nil {
			return err
		}
		if len(productIDs) > 0 {
			if err := tx.Unscoped().Where("id IN ?", productIDs).Delete(&model.Product{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Where("vendor_id IN ?", vendorIDs).Delete(&model.VendorMedia{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("vendor_id IN ?", vendorIDs).Delete(&model.VendorTag{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("vendor_id IN ?", vendorIDs).Delete(&model.MediaAsset{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("vendor_id IN ?", vendorIDs).Delete(&model.AdminUser{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Where("id IN ?", vendorIDs).Delete(&model.Vendor{}).Error
	})
}

func backfillVendorImageAssetLinks(db *gorm.DB) error {
	var vendors []model.Vendor
	if err := db.Where("(logo LIKE ? AND logo_asset_id IS NULL) OR (cover_image LIKE ? AND cover_asset_id IS NULL)", "/api/media/%", "/api/media/%").Find(&vendors).Error; err != nil {
		return err
	}
	for _, vendor := range vendors {
		updates := map[string]interface{}{}
		if vendor.LogoAssetID == nil {
			if id := assetIDFromMediaURL(vendor.Logo); id > 0 && mediaAssetExists(db, id) {
				updates["logo_asset_id"] = id
			}
		}
		if vendor.CoverAssetID == nil {
			if id := assetIDFromMediaURL(vendor.CoverImage); id > 0 && mediaAssetExists(db, id) {
				updates["cover_asset_id"] = id
			}
		}
		if len(updates) > 0 {
			if err := db.Model(&vendor).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func assetIDFromMediaURL(value string) uint {
	if !strings.HasPrefix(value, "/api/media/") {
		return 0
	}
	id, err := strconv.ParseUint(strings.TrimPrefix(value, "/api/media/"), 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}

func mediaAssetExists(db *gorm.DB, id uint) bool {
	var count int64
	db.Model(&model.MediaAsset{}).Where("id = ?", id).Count(&count)
	return count > 0
}

func dropRetiredVendorColumns(db *gorm.DB) error {
	for _, column := range []string{"quality_control", "supply_regions", "cooperation_terms", "source_url", "source_note", "service_models"} {
		if db.Migrator().HasColumn(&model.Vendor{}, column) {
			if err := db.Migrator().DropColumn(&model.Vendor{}, column); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateProductCatalog(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Product{}).Where("content_version = 0").Update("content_version", 1).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Product{}).Where("status = ?", 1).Updates(map[string]interface{}{"publication_status": "published"}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Product{}).Where("status = ?", 2).Updates(map[string]interface{}{"publication_status": "draft"}).Error; err != nil {
			return err
		}
		var products []model.Product
		if err := tx.Unscoped().Where("vendor_id > 0 AND deleted_at IS NULL").Find(&products).Error; err != nil {
			return err
		}
		for _, product := range products {
			vendorID := product.VendorIDValue()
			if vendorID == 0 {
				continue
			}
			status := "approved"
			if product.Status != 1 {
				status = "pending"
			}
			supplier := model.ProductSupplier{ProductID: product.ID, VendorID: vendorID, VendorProductName: product.Name, Image: product.Image, GalleryRaw: product.GalleryRaw, CompatibleModels: product.CompatibleModels, Description: product.Description, PriceNote: product.PriceNote, InquiryText: product.InquiryText, InquiryPath: product.InquiryPath, Status: status, SourceType: "legacy", ContentVersion: 1}

			// The unique index does not include deleted_at. A normal scoped lookup misses a
			// soft-deleted relationship and a subsequent INSERT then fails with duplicate
			// product_id/vendor_id. Include deleted rows and revive the existing record so
			// this migration remains safe to run on every startup.
			var existing model.ProductSupplier
			result := tx.Unscoped().Where("product_id = ? AND vendor_id = ?", product.ID, vendorID).First(&existing)
			switch {
			case result.Error == nil:
				if !existing.DeletedAt.Valid {
					continue
				}
				supplier.ID = existing.ID
				supplier.CreatedAt = existing.CreatedAt
				if err := tx.Unscoped().Save(&supplier).Error; err != nil {
					return err
				}
			case errors.Is(result.Error, gorm.ErrRecordNotFound):
				if err := tx.Create(&supplier).Error; err != nil {
					return err
				}
			default:
				return result.Error
			}
			if status == "pending" {
				productPayload, _ := json.Marshal(product)
				supplierPayload, _ := json.Marshal(supplier)
				productID, supplierID := product.ID, supplier.ID
				submission := model.ProductSubmission{VendorID: vendorID, ProductID: &productID, SupplierID: &supplierID, SubmissionType: "new_product", BaseVersion: supplier.ContentVersion, ProductPayload: string(productPayload), SupplierPayload: string(supplierPayload), Status: "pending", SubmittedBy: "migration"}
				if err := tx.Create(&submission).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func migrateVendorPublicationState(db *gorm.DB) error {
	if err := db.Model(&model.Vendor{}).Where("content_version = 0").Update("content_version", 1).Error; err != nil {
		return err
	}
	if err := db.Model(&model.Vendor{}).Where("data_origin = '' OR data_origin IS NULL").Update("data_origin", "admin").Error; err != nil {
		return err
	}
	if err := db.Model(&model.Vendor{}).Where("publication_status = '' OR publication_status IS NULL").
		Update("publication_status", gorm.Expr("CASE WHEN is_visible = 1 THEN 'published' ELSE 'hidden' END")).Error; err != nil {
		return err
	}
	templateNames := []string{
		"山东沃得农机配件有限公司", "河北金瑞农机制造有限公司",
	}
	if err := db.Model(&model.Vendor{}).
		Where("name IN ? AND established_year = ? AND factory_area = ? AND employee_count = ? AND annual_capacity <> ''", templateNames, "2012 年", "12000 平方米", "80 人").
		Updates(map[string]interface{}{"data_origin": "demo", "publication_status": "hidden", "is_visible": false, "is_verified": false, "is_recommended": false}).Error; err != nil {
		return err
	}
	return db.Exec(`DELETE vt FROM vendor_tags vt JOIN vendors v ON v.id = vt.vendor_id JOIN tags t ON t.id = vt.tag_id WHERE v.data_origin = 'demo' AND t.name IN ('源头厂商','支持定制','现货充足','实地认证')`).Error
}

func cleanupLegacyDemoContent(db *gorm.DB) error {
	updates := []struct {
		model  interface{}
		where  string
		args   []interface{}
		column string
		value  interface{}
	}{
		{&model.Vendor{}, "logo LIKE ?", []interface{}{"%dummyimage.com%"}, "logo", ""},
		{&model.Vendor{}, "cover_image LIKE ?", []interface{}{"%dummyimage.com%"}, "cover_image", ""},
		{&model.Vendor{}, "LOWER(website_url) IN ?", []interface{}{[]string{"https://example.com", "http://example.com", "https://www.example.com", "http://www.example.com"}}, "website_url", ""},
		{&model.Product{}, "image LIKE ?", []interface{}{"%dummyimage.com%"}, "image", ""},
		{&model.Banner{}, "background_image LIKE ?", []interface{}{"%dummyimage.com%"}, "background_image", ""},
	}
	for _, update := range updates {
		if err := db.Model(update.model).Where(update.where, update.args...).Update(update.column, update.value).Error; err != nil {
			return err
		}
	}
	return db.Where("LOWER(url) IN ?", []string{"https://example.com", "http://example.com", "https://www.example.com", "http://www.example.com"}).Delete(&model.FriendLink{}).Error
}

func ensureStructuredContent(db *gorm.DB) error {
	privacy := model.ContentPage{Slug: "privacy", Title: "隐私说明", Summary: "本平台以最小化方式统计匿名访问数据，用于改善公开资源展示与服务。", Content: "平台默认启用匿名访问统计，用于汇总页面浏览量、匿名访客数、省级访问分布和公开内容访问排行。统计使用第一方 Cookie 生成匿名标识，原始 IP 仅用于本地省份解析后立即丢弃；不会保存搜索关键词、联系电话、微信、账号或可识别身份信息。匿名事件明细和 Cookie 最多保留 90 天。", SEOKeywords: "隐私说明,访问统计", IsEnabled: true, SortOrder: 90}
	if err := db.Where("slug = ?", privacy.Slug).FirstOrCreate(&privacy).Error; err != nil {
		return err
	}
	blocks := map[string][]model.ContentBlock{
		"join": {
			{Type: "hero", Title: "让更多采购商看见您的产品与实力", Text: "完成厂商账号绑定后，即可在 CMS 维护企业资料；新资料经平台审核后公开展示。", ButtonText: "已有账号，登录 CMS", ButtonPath: "/admin/login"},
			{Type: "text", Title: "入驻前请准备", Items: []string{"企业全称、所在地区与详细地址", "主营产品和生产加工能力", "真实 Logo、厂房、设备、证书与产品图片", "联系人、联系电话、微信或企业官网"}},
			{Type: "steps", Title: "入驻流程", Items: []string{"联系平台运营人员核验企业信息并创建厂商档案", "获取厂商 CMS 账号，登录后完善企业展示资料", "提交资料等待平台审核，驳回后可按意见重新修改", "审核通过后自动更新前台厂商目录和详情页"}},
			{Type: "cta", Title: "已经获得厂商账号？", Text: "登录 CMS 更新企业资料，审核期间原有公开资料不会受到影响。", ButtonText: "登录厂商 CMS", ButtonPath: "/admin/login"},
			{Type: "faq", Title: "常见问题", Items: []string{"提交后会立即展示吗？|不会。资料需要管理员审核通过后才会更新到前台。", "审核期间旧资料是否下线？|不会，平台继续展示上一版已审核资料。", "可以上传哪些资料？|支持 Logo、封面、厂房、设备和证书等企业图片。"}},
		},
		"about":    {{Type: "hero", Title: "连接农机采购需求与源头厂商", Text: "大陆农机配件聚合厂商、配件产品和加工能力信息，为维修门店、经销商和采购人员提供清晰可信的行业目录。"}, {Type: "text", Title: "平台价值", Items: []string{"按地区、品类和服务能力快速筛选厂商", "集中查看产品适配信息、企业能力与联系方式", "厂商自主维护资料，平台审核后发布"}}, {Type: "steps", Title: "信息保障", Items: []string{"厂商资料变更留存审核记录", "隐藏厂商和下架产品不对外展示", "公开来源信息与厂商自有资料分开标识"}}},
		"purchase": {{Type: "hero", Title: "按品类查产品，按能力找厂商", Text: "当前平台以公开行业目录为核心，您可以通过搜索、产品分类和厂商筛选快速定位供应资源。", ButtonText: "浏览配件产品", ButtonPath: "/products"}, {Type: "cta", Title: "需要定制加工？", Text: "查看支持来图来样、数控加工和批量代工的厂商。", ButtonText: "查看加工服务", ButtonPath: "/service"}},
		"links":    {{Type: "text", Title: "行业合作入口", Text: "友情链接仅展示经平台维护的农机行业与服务合作伙伴。"}},
	}
	for slug, pageBlocks := range blocks {
		payload, _ := json.Marshal(pageBlocks)
		if err := db.Model(&model.ContentPage{}).Where("slug = ? AND (blocks IS NULL OR blocks = '')", slug).Update("blocks", string(payload)).Error; err != nil {
			return err
		}
	}
	return nil
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
