package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

func CleanupOrphanedMedia(db *gorm.DB, mediaDir string) error {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	var assets []model.MediaAsset
	if err := db.Where("status = ? AND created_at < ? AND NOT EXISTS (SELECT 1 FROM vendor_media WHERE vendor_media.asset_id = media_assets.id)", "staged", cutoff).Find(&assets).Error; err != nil {
		return err
	}
	for _, asset := range assets {
		url := fmt.Sprintf("/api/media/%d", asset.ID)
		var references int64
		if err := db.Model(&model.Product{}).Where("image = ? OR INSTR(gallery, ?) > 0", url, url).Count(&references).Error; err != nil {
			return err
		}
		if references == 0 {
			if err := db.Model(&model.ProductSupplier{}).Where("image = ? OR INSTR(gallery, ?) > 0", url, url).Count(&references).Error; err != nil {
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
		&model.Menu{},
		&model.Tag{},
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
	); err != nil {
		return err
	}
	if err := cleanupDuplicateVendorTags(db); err != nil {
		return err
	}
	if err := ensureVendorTagUniqueIndex(db); err != nil {
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
	return ensureStructuredContent(db)
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
	for _, column := range []string{"quality_control", "supply_regions", "cooperation_terms", "source_url", "source_note"} {
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
			var count int64
			if err := tx.Model(&model.ProductSupplier{}).Where("product_id = ? AND vendor_id = ?", product.ID, product.VendorID).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			status := "approved"
			if product.Status != 1 {
				status = "pending"
			}
			supplier := model.ProductSupplier{ProductID: product.ID, VendorID: product.VendorID, VendorProductName: product.Name, Image: product.Image, GalleryRaw: product.GalleryRaw, CompatibleModels: product.CompatibleModels, Description: product.Description, PriceNote: product.PriceNote, InquiryText: product.InquiryText, InquiryPath: product.InquiryPath, Status: status, SourceType: "legacy", ContentVersion: 1}
			if err := tx.Create(&supplier).Error; err != nil {
				return err
			}
			if status == "pending" {
				productPayload, _ := json.Marshal(product)
				supplierPayload, _ := json.Marshal(supplier)
				productID, supplierID := product.ID, supplier.ID
				submission := model.ProductSubmission{VendorID: product.VendorID, ProductID: &productID, SupplierID: &supplierID, SubmissionType: "new_product", BaseVersion: supplier.ContentVersion, ProductPayload: string(productPayload), SupplierPayload: string(supplierPayload), Status: "pending", SubmittedBy: "migration"}
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
		"山东沃得农机配件有限公司", "河北金瑞农机制造有限公司", "江苏东成农机配件有限公司",
		"河南中联农机制造有限公司", "安徽豪华农机配件有限公司", "山东万鑫农机配件有限公司",
		"宁波动力机械有限公司", "浙江汉丰农机有限公司", "河北力捷机械有限公司",
		"辽宁佳丰农机配件有限公司", "四川川沃农机有限公司", "陕西恒农农机配件有限公司",
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
		{&model.SiteConfig{}, "config_key = ? AND (config_value LIKE ? OR config_value LIKE ?)", []interface{}{"home.stats", "%2000+%", "%10万+%"}, "config_value", "[]"},
	}
	for _, update := range updates {
		if err := db.Model(update.model).Where(update.where, update.args...).Update(update.column, update.value).Error; err != nil {
			return err
		}
	}
	return db.Where("LOWER(url) IN ?", []string{"https://example.com", "http://example.com", "https://www.example.com", "http://www.example.com"}).Delete(&model.FriendLink{}).Error
}

func ensureStructuredContent(db *gorm.DB) error {
	blocks := map[string][]model.ContentBlock{
		"join": {
			{Type: "hero", Title: "让更多采购商看见您的产品与实力", Text: "完成厂商账号绑定后，即可在 CMS 维护企业资料；新资料经平台审核后公开展示。", ButtonText: "已有账号，登录 CMS", ButtonPath: "/admin/login"},
			{Type: "text", Title: "入驻前请准备", Items: []string{"企业全称、所在地区与详细地址", "主营产品、适配机型和生产加工能力", "真实 Logo、厂房、设备、证书与产品图片", "联系人、联系电话、微信或企业官网"}},
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
