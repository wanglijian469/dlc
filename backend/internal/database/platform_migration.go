package database

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"gorm.io/gorm"
)

func BackfillPlatformData(db *gorm.DB) error {
	now := time.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		var vendors []model.Vendor
		if err := tx.Find(&vendors).Error; err != nil {
			return err
		}
		for i := range vendors {
			service.ApplyVendorSEO(&vendors[i])
			old := vendors[i].Slug
			if err := EnsureVendorSlug(tx, &vendors[i]); err != nil {
				return err
			}
			updates := map[string]any{
				"slug":                   vendors[i].Slug,
				"seo_title":              vendors[i].SEOTitle,
				"seo_description":        vendors[i].SEODescription,
				"seo_title_manual":       vendors[i].SEOTitleManual,
				"seo_description_manual": vendors[i].SEODescriptionManual,
			}
			if vendors[i].PublicationStatus == "published" && vendors[i].PublishedAt == nil {
				updates["published_at"] = firstTime(vendors[i].UpdatedAt, vendors[i].CreatedAt, now)
			}
			if err := tx.Model(&vendors[i]).Updates(updates).Error; err != nil {
				return err
			}
			if old == "" {
				_ = recordRedirect(tx, fmt.Sprintf("/vendors/%d", vendors[i].ID), "/v/"+vendors[i].Slug)
			}
			if err := ensureInitialRevision(tx, "vendor", vendors[i].ID, vendors[i].ContentVersion, vendors[i]); err != nil {
				return err
			}
		}

		var products []model.Product
		if err := tx.Find(&products).Error; err != nil {
			return err
		}
		for i := range products {
			old := products[i].Slug
			if err := EnsureProductSlug(tx, &products[i]); err != nil {
				return err
			}
			updates := map[string]any{"slug": products[i].Slug}
			if products[i].PublicationStatus == "published" && products[i].PublishedAt == nil {
				updates["published_at"] = firstTime(products[i].UpdatedAt, products[i].CreatedAt, now)
			}
			if err := tx.Model(&products[i]).Updates(updates).Error; err != nil {
				return err
			}
			if old == "" {
				_ = recordRedirect(tx, fmt.Sprintf("/products/%d", products[i].ID), "/products/"+products[i].Slug)
			}
			if err := ensureInitialRevision(tx, "product", products[i].ID, products[i].ContentVersion, products[i]); err != nil {
				return err
			}
		}

		var categories []model.Category
		if err := tx.Find(&categories).Error; err != nil {
			return err
		}
		for i := range categories {
			if categories[i].PublicationStatus == "" {
				categories[i].PublicationStatus = map[bool]string{true: "published", false: "archived"}[categories[i].IsEnabled]
			}
			if categories[i].ContentVersion == 0 {
				categories[i].ContentVersion = 1
			}
			if err := EnsureCategorySlug(tx, &categories[i]); err != nil {
				return err
			}
			updates := map[string]any{"slug": categories[i].Slug, "publication_status": categories[i].PublicationStatus, "content_version": categories[i].ContentVersion}
			if categories[i].PublicationStatus == "published" && categories[i].PublishedAt == nil {
				updates["published_at"] = firstTime(categories[i].UpdatedAt, categories[i].CreatedAt, now)
			}
			if err := tx.Model(&categories[i]).Updates(updates).Error; err != nil {
				return err
			}
			_ = recordRedirect(tx, "/products?categoryId="+strconv.FormatUint(uint64(categories[i].ID), 10), "/products/category/"+categories[i].Slug)
			if err := ensureInitialRevision(tx, "category", categories[i].ID, categories[i].ContentVersion, categories[i]); err != nil {
				return err
			}
		}

		var pages []model.ContentPage
		if err := tx.Find(&pages).Error; err != nil {
			return err
		}
		for i := range pages {
			status := pages[i].PublicationStatus
			if !pages[i].IsEnabled {
				status = "archived"
			} else if pages[i].PublishedAt != nil && pages[i].PublishedAt.After(now) {
				status = "scheduled"
			} else if status == "" {
				status = "published"
			}
			version := pages[i].ContentVersion
			if version == 0 {
				version = 1
			}
			if strings.TrimSpace(pages[i].Slug) == "" {
				pages[i].Slug = service.Slugify(pages[i].Title)
			}
			if err := tx.Model(&pages[i]).Updates(map[string]any{"slug": pages[i].Slug, "publication_status": status, "content_version": version}).Error; err != nil {
				return err
			}
			pages[i].PublicationStatus, pages[i].ContentVersion = status, version
			if err := ensureInitialRevision(tx, pageRevisionType(pages[i]), pages[i].ID, version, pages[i]); err != nil {
				return err
			}
		}
		if err := tx.Model(&model.Menu{}).Where("menu_type = ? AND path = ?", "mobile_bottom", "/admin/login").Update("path", "/account/login").Error; err != nil {
			return err
		}
		return nil
	})
}

// InitializeVendorSEO is the one-time compatibility step for schema version 5.
// Existing non-empty metadata is treated as editorial content and must never be
// overwritten by the new automatic generator.
func InitializeVendorSEO(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var vendors []model.Vendor
		if err := tx.Find(&vendors).Error; err != nil {
			return err
		}
		for i := range vendors {
			vendors[i].SEOTitleManual = strings.TrimSpace(vendors[i].SEOTitle) != ""
			vendors[i].SEODescriptionManual = strings.TrimSpace(vendors[i].SEODescription) != ""
			service.ApplyVendorSEO(&vendors[i])
			if err := tx.Model(&vendors[i]).Updates(map[string]any{
				"seo_title":              vendors[i].SEOTitle,
				"seo_description":        vendors[i].SEODescription,
				"seo_title_manual":       vendors[i].SEOTitleManual,
				"seo_description_manual": vendors[i].SEODescriptionManual,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PurgeOrdinaryAccounts permanently removes the retired consumer-account
// surface, including browser sessions and every log tied to those accounts.
// This is intentionally a one-way migration and must run only after backup.
func PurgeOrdinaryAccounts(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var users []struct {
			ID       uint
			Username string
		}
		if err := tx.Unscoped().Model(&model.AdminUser{}).Select("id, username").Where("role = ?", "user").Find(&users).Error; err != nil {
			return err
		}

		if len(users) > 0 {
			ids := make([]uint, 0, len(users))
			usernames := make([]string, 0, len(users))
			for _, user := range users {
				ids = append(ids, user.ID)
				usernames = append(usernames, user.Username)
			}
			if err := tx.Where("user_id IN ?", ids).Delete(&model.AuthSession{}).Error; err != nil {
				return err
			}
			if err := tx.Where("username IN ? OR (resource = ? AND record_id IN ?)", usernames, "users", ids).Delete(&model.OperationLog{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("id IN ?", ids).Delete(&model.AdminUser{}).Error; err != nil {
				return err
			}
		}

		return tx.Model(&model.Menu{}).
			Where("menu_type = ? AND path = ?", "mobile_bottom", "/account/login").
			Update("name", "厂商").Error
	})
}

func firstTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value
		}
	}
	return time.Now()
}

func ensureInitialRevision(db *gorm.DB, resourceType string, resourceID, version uint, snapshot any) error {
	var count int64
	if err := db.Model(&model.ContentRevision{}).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	now := time.Now()
	return db.Create(&model.ContentRevision{ResourceType: resourceType, ResourceID: resourceID, Version: version, BaseVersion: version, Status: "published", Snapshot: string(payload), AuthorUsername: "migration", Reviewer: "migration", PublishedAt: &now}).Error
}

func pageRevisionType(page model.ContentPage) string {
	if page.PageType == "article" {
		return "article"
	}
	return "page"
}

func EnsureVendorSlug(db *gorm.DB, item *model.Vendor) error {
	old, err := existingSlug(db, "vendors", item.ID)
	if err != nil {
		return err
	}
	item.Slug = uniqueSlug(db, "vendors", item.ID, firstNonEmptySlug(item.Slug, item.Name), "vendor")
	if old != "" && old != item.Slug {
		if err := db.Model(&model.SEORedirect{}).Where("destination_path = ?", "/v/"+old).Update("destination_path", "/v/"+item.Slug).Error; err != nil {
			return err
		}
		if err := recordRedirect(db, "/v/"+old, "/v/"+item.Slug); err != nil {
			return err
		}
		if err := recordRedirect(db, "/vendors/"+old, "/v/"+item.Slug); err != nil {
			return err
		}
	}
	return recordRedirect(db, "/vendors/"+item.Slug, "/v/"+item.Slug)
}

var vendorSiteSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidateVendorSiteSlug keeps vendor-controlled site addresses predictable and
// prevents a pending submission from silently being assigned a different URL.
func ValidateVendorSiteSlug(db *gorm.DB, value string, vendorID uint) (string, error) {
	slug, err := normalizeVendorSiteSlug(value)
	if err != nil {
		return "", err
	}
	var count int64
	query := db.Model(&model.Vendor{}).Where("slug = ?", slug)
	if vendorID > 0 {
		query = query.Where("id <> ?", vendorID)
	}
	if err := query.Count(&count).Error; err != nil {
		return "", err
	}
	if count > 0 {
		return "", fmt.Errorf("该厂商网站地址标识已被使用")
	}
	return slug, nil
}

func normalizeVendorSiteSlug(value string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(value))
	if len(slug) == 0 || len(slug) > 72 || !vendorSiteSlugPattern.MatchString(slug) {
		return "", fmt.Errorf("厂商网站地址标识仅支持小写英文、数字和连字符，且不能以连字符开头或结尾")
	}
	return slug, nil
}

func EnsureProductSlug(db *gorm.DB, item *model.Product) error {
	old, err := existingSlug(db, "products", item.ID)
	if err != nil {
		return err
	}
	item.Slug = uniqueSlug(db, "products", item.ID, firstNonEmptySlug(item.Slug, item.Name), "product")
	if old != "" && old != item.Slug {
		return recordRedirect(db, "/products/"+old, "/products/"+item.Slug)
	}
	return nil
}

func EnsureCategorySlug(db *gorm.DB, item *model.Category) error {
	old, err := existingSlug(db, "categories", item.ID)
	if err != nil {
		return err
	}
	item.Slug = uniqueSlug(db, "categories", item.ID, firstNonEmptySlug(item.Slug, item.Name), "category")
	if old != "" && old != item.Slug {
		return recordRedirect(db, "/products/category/"+old, "/products/category/"+item.Slug)
	}
	return nil
}

func EnsurePageSlug(db *gorm.DB, item *model.ContentPage) error {
	var old struct{ Slug, PageType string }
	if item.ID > 0 {
		if err := db.Table("content_pages").Select("COALESCE(slug, '') AS slug, COALESCE(page_type, '') AS page_type").Where("id = ?", item.ID).Scan(&old).Error; err != nil {
			return err
		}
	}
	item.Slug = uniqueSlug(db, "content_pages", item.ID, firstNonEmptySlug(item.Slug, item.Title), "page")
	if old.Slug != "" && old.Slug != item.Slug {
		oldPath, newPath := "/"+old.Slug, "/"+item.Slug
		if old.PageType == "article" {
			oldPath = "/guides/" + old.Slug
		}
		if item.PageType == "article" {
			newPath = "/guides/" + item.Slug
		}
		return recordRedirect(db, oldPath, newPath)
	}
	return nil
}

func existingSlug(db *gorm.DB, table string, id uint) (string, error) {
	if id == 0 {
		return "", nil
	}
	var slug string
	err := db.Table(table).Select("COALESCE(slug, '')").Where("id = ?", id).Scan(&slug).Error
	return slug, err
}

func firstNonEmptySlug(values ...string) string {
	for _, value := range values {
		if slug := service.Slugify(value); slug != "" {
			return slug
		}
	}
	return "content"
}

func uniqueSlug(db *gorm.DB, table string, id uint, preferred, fallback string) string {
	base := firstNonEmptySlug(preferred, fallback)
	candidate := base
	for suffix := 0; suffix < 10000; suffix++ {
		var count int64
		query := db.Table(table).Where("slug = ?", candidate)
		if id > 0 {
			query = query.Where("id <> ?", id)
		}
		if query.Count(&count).Error == nil && count == 0 {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", base, suffix+2)
	}
	return fmt.Sprintf("%s-%d", base, time.Now().Unix())
}

func recordRedirect(db *gorm.DB, source, destination string) error {
	if source == destination || source == "" || destination == "" {
		return nil
	}
	item := model.SEORedirect{SourcePath: source, DestinationPath: destination, StatusCode: 301}
	return db.Where("source_path = ?", source).Assign(map[string]any{"destination_path": destination, "status_code": 301}).FirstOrCreate(&item).Error
}
