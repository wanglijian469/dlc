package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var errRevisionConflict = errors.New("content version conflict")
var errProductRequiresVendor = errors.New("published product requires a visible published vendor")

type revisionDraftRequest struct {
	ResourceType string          `json:"resourceType"`
	ResourceID   uint            `json:"resourceId"`
	BaseVersion  uint            `json:"baseVersion"`
	Snapshot     json.RawMessage `json:"snapshot"`
}

type revisionReviewRequest struct {
	Note        string     `json:"note"`
	ScheduledAt *time.Time `json:"scheduledAt"`
}

func (h AdminHandler) ListRevisions(c *gin.Context) {
	var rows []model.ContentRevision
	query := h.DB.Model(&model.ContentRevision{})
	if value := strings.TrimSpace(c.Query("resourceType")); value != "" {
		query = query.Where("resource_type = ?", value)
	}
	if value := strings.TrimSpace(c.Query("status")); value != "" {
		query = query.Where("status = ?", value)
	}
	if value := strings.TrimSpace(c.Query("resourceId")); value != "" {
		query = query.Where("resource_id = ?", value)
	}
	page, pageSize := pageParams(c, 30)
	result, err := paginate(query.Order("updated_at desc, id desc"), &rows, page, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "修订列表加载失败")
		return
	}
	OK(c, result)
}

func (h AdminHandler) PreviewRevision(c *gin.Context) {
	var revision model.ContentRevision
	if h.DB.First(&revision, idParam(c)).Error != nil {
		Fail(c, http.StatusNotFound, 404, "修订不存在")
		return
	}
	c.Header("X-Robots-Tag", "noindex, nofollow")
	c.Header("Cache-Control", "private, no-store")
	var snapshot any
	if json.Unmarshal([]byte(revision.Snapshot), &snapshot) != nil {
		Fail(c, 500, 500, "修订快照无法读取")
		return
	}
	var baseSnapshot any
	if payload, _, err := loadRevisionSource(h.DB, revision.ResourceType, revision.ResourceID); err == nil {
		_ = json.Unmarshal(payload, &baseSnapshot)
	}
	OK(c, gin.H{"revision": revision, "snapshot": snapshot, "baseSnapshot": baseSnapshot})
}

func (h AdminHandler) SaveRevision(c *gin.Context) {
	var req revisionDraftRequest
	if c.ShouldBindJSON(&req) != nil || !validRevisionType(req.ResourceType) || req.ResourceID == 0 {
		Fail(c, http.StatusBadRequest, 400, "资源类型、资源 ID 或修订内容无效")
		return
	}
	if len(req.Snapshot) == 0 {
		payload, version, err := loadRevisionSource(h.DB, req.ResourceType, req.ResourceID)
		if err != nil {
			Fail(c, http.StatusNotFound, 404, "待修订内容不存在")
			return
		}
		req.Snapshot, req.BaseVersion = payload, version
	}
	if !json.Valid(req.Snapshot) {
		Fail(c, http.StatusBadRequest, 400, "修订快照必须是有效 JSON")
		return
	}
	currentVersion, err := contentVersion(h.DB, req.ResourceType, req.ResourceID)
	if err != nil {
		Fail(c, http.StatusNotFound, 404, "待修订内容不存在")
		return
	}
	if req.BaseVersion == 0 {
		req.BaseVersion = currentVersion
	}
	if req.BaseVersion != currentVersion {
		Fail(c, http.StatusConflict, 409, "内容已被他人更新，请刷新后再编辑")
		return
	}
	var latest uint
	h.DB.Model(&model.ContentRevision{}).Where("resource_type = ? AND resource_id = ?", req.ResourceType, req.ResourceID).Select("COALESCE(MAX(version), 0)").Scan(&latest)
	revision := model.ContentRevision{
		ResourceType: req.ResourceType, ResourceID: req.ResourceID, Version: latest + 1,
		BaseVersion: req.BaseVersion, Status: "draft", Snapshot: string(req.Snapshot),
		AuthorUsername: c.GetString("username"),
	}
	if err := h.DB.Create(&revision).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "修订保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "save-revision", req.ResourceType, req.ResourceID)
	OK(c, revision)
}

func (h AdminHandler) SubmitRevision(c *gin.Context) {
	revision, ok := h.editableRevision(c)
	if !ok {
		return
	}
	if revision.Status != "draft" && revision.Status != "rejected" {
		Fail(c, http.StatusConflict, 409, "只有草稿或已退回修订可以送审")
		return
	}
	if err := h.DB.Model(&revision).Updates(map[string]any{"status": "in_review", "reviewer": "", "review_note": "", "scheduled_at": nil}).Error; err != nil {
		Fail(c, 500, 500, "送审失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "submit-review", revision.ResourceType, revision.ResourceID)
	h.DB.First(&revision, revision.ID)
	OK(c, revision)
}

func (h AdminHandler) RejectRevision(c *gin.Context) {
	var req revisionReviewRequest
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Note) == "" {
		Fail(c, http.StatusBadRequest, 400, "退回原因不能为空")
		return
	}
	revision, ok := h.reviewableRevision(c)
	if !ok {
		return
	}
	if err := h.DB.Model(&revision).Updates(map[string]any{"status": "rejected", "reviewer": c.GetString("username"), "review_note": strings.TrimSpace(req.Note), "scheduled_at": nil}).Error; err != nil {
		Fail(c, 500, 500, "退回失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "reject-revision", revision.ResourceType, revision.ResourceID)
	h.DB.First(&revision, revision.ID)
	OK(c, revision)
}

func (h AdminHandler) ApproveRevision(c *gin.Context) {
	var req revisionReviewRequest
	_ = c.ShouldBindJSON(&req)
	revision, ok := h.reviewableRevision(c)
	if !ok {
		return
	}
	username, role := c.GetString("username"), c.GetString("role")
	if revision.AuthorUsername == username && (role != "admin" || strings.TrimSpace(req.Note) == "") {
		Fail(c, http.StatusForbidden, 403, "不能审核自己的修改；管理员紧急发布必须填写原因")
		return
	}
	if req.ScheduledAt != nil && req.ScheduledAt.After(time.Now()) {
		updates := map[string]any{"status": "scheduled", "reviewer": username, "review_note": strings.TrimSpace(req.Note), "scheduled_at": req.ScheduledAt}
		if err := h.DB.Model(&revision).Updates(updates).Error; err != nil {
			Fail(c, 500, 500, "定时发布保存失败")
			return
		}
		logOperation(h.DB, username, "schedule-revision", revision.ResourceType, revision.ResourceID)
		h.DB.First(&revision, revision.ID)
		OK(c, revision)
		return
	}
	if err := publishRevision(h.DB, &revision, username, strings.TrimSpace(req.Note)); err != nil {
		if errors.Is(err, errRevisionConflict) {
			Fail(c, http.StatusConflict, 409, "基础版本已变化，请重新比较并送审")
		} else if errors.Is(err, errProductRequiresVendor) {
			Fail(c, http.StatusBadRequest, 400, "产品发布前必须关联至少一家已发布且前台可见的厂商")
		} else {
			Fail(c, 500, 500, "发布事务失败")
		}
		return
	}
	logOperation(h.DB, username, "publish-revision", revision.ResourceType, revision.ResourceID)
	OK(c, revision)
}

func (h AdminHandler) ArchiveRevisionResource(c *gin.Context) {
	revision, ok := h.reviewableRevision(c)
	if !ok {
		return
	}
	if err := archiveRevisionResource(h.DB, revision.ResourceType, revision.ResourceID, revision.BaseVersion); err != nil {
		if errors.Is(err, errRevisionConflict) {
			Fail(c, http.StatusConflict, 409, "内容版本已变化")
		} else {
			Fail(c, 500, 500, "撤下失败")
		}
		return
	}
	h.DB.Model(&revision).Updates(map[string]any{"status": "archived", "reviewer": c.GetString("username")})
	logOperation(h.DB, c.GetString("username"), "archive", revision.ResourceType, revision.ResourceID)
	OK(c, gin.H{"archived": true})
}

func (h AdminHandler) editableRevision(c *gin.Context) (model.ContentRevision, bool) {
	var revision model.ContentRevision
	if h.DB.First(&revision, idParam(c)).Error != nil {
		Fail(c, http.StatusNotFound, 404, "修订不存在")
		return revision, false
	}
	if revision.AuthorUsername != c.GetString("username") && c.GetString("role") != "admin" {
		Fail(c, http.StatusForbidden, 403, "只能操作自己的修订")
		return revision, false
	}
	return revision, true
}

func (h AdminHandler) reviewableRevision(c *gin.Context) (model.ContentRevision, bool) {
	var revision model.ContentRevision
	if h.DB.First(&revision, idParam(c)).Error != nil {
		Fail(c, http.StatusNotFound, 404, "修订不存在")
		return revision, false
	}
	if revision.Status != "in_review" {
		Fail(c, http.StatusConflict, 409, "修订当前不在审核中")
		return revision, false
	}
	if revision.AuthorUsername == c.GetString("username") && c.GetString("role") != "admin" {
		Fail(c, http.StatusForbidden, 403, "不能审核、退回或撤下自己的修改")
		return revision, false
	}
	return revision, true
}

func validRevisionType(value string) bool {
	switch value {
	case "vendor", "product", "category", "page", "article":
		return true
	default:
		return false
	}
}

func loadRevisionSource(db *gorm.DB, resourceType string, id uint) (json.RawMessage, uint, error) {
	var value any
	switch resourceType {
	case "vendor":
		value = &model.Vendor{ID: id}
	case "product":
		value = &model.Product{ID: id}
	case "category":
		value = &model.Category{ID: id}
	case "page", "article":
		value = &model.ContentPage{ID: id}
	default:
		return nil, 0, gorm.ErrRecordNotFound
	}
	if err := db.First(value, id).Error; err != nil {
		return nil, 0, err
	}
	payload, err := json.Marshal(value)
	version, _ := contentVersion(db, resourceType, id)
	return payload, version, err
}

func contentVersion(db *gorm.DB, resourceType string, id uint) (uint, error) {
	table := map[string]string{"vendor": "vendors", "product": "products", "category": "categories", "page": "content_pages", "article": "content_pages"}[resourceType]
	if table == "" {
		return 0, gorm.ErrRecordNotFound
	}
	var row struct{ ContentVersion uint }
	if err := db.Table(table).Select("content_version").Where("id = ?", id).Take(&row).Error; err != nil {
		return 0, err
	}
	return row.ContentVersion, nil
}

func publishRevision(db *gorm.DB, revision *model.ContentRevision, reviewer, note string) error {
	now := time.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		current, err := contentVersion(tx, revision.ResourceType, revision.ResourceID)
		if err != nil {
			return err
		}
		if current != revision.BaseVersion {
			return errRevisionConflict
		}
		if err := applyRevisionSnapshot(tx, revision, current+1, now); err != nil {
			return err
		}
		result := tx.Model(&model.ContentRevision{}).Where("id = ? AND status IN ?", revision.ID, []string{"in_review", "scheduled"}).Updates(map[string]any{
			"status": "published", "reviewer": reviewer, "review_note": note, "published_at": &now, "scheduled_at": nil,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errRevisionConflict
		}
		revision.Status, revision.Reviewer, revision.ReviewNote, revision.PublishedAt = "published", reviewer, note, &now
		return nil
	})
}

func applyRevisionSnapshot(tx *gorm.DB, revision *model.ContentRevision, version uint, publishedAt time.Time) error {
	switch revision.ResourceType {
	case "vendor":
		var item model.Vendor
		if err := json.Unmarshal([]byte(revision.Snapshot), &item); err != nil {
			return err
		}
		item.ID, item.ContentVersion, item.PublicationStatus, item.IsVisible, item.PublishedAt = revision.ResourceID, version, "published", true, &publishedAt
		service.ApplyVendorSEO(&item)
		if err := bindVendorImageAssets(tx, &item); err != nil {
			return err
		}
		if err := database.EnsureVendorSlug(tx, &item); err != nil {
			return err
		}
		tagIDs, media := uniqueUintIDs(item.TagIDs), item.Media
		item.Tags, item.Media = nil, nil
		if err := optimisticSave(tx, &item, revision.ResourceID, revision.BaseVersion, "Tags", "Media"); err != nil {
			return err
		}
		if media != nil {
			if err := replaceVendorMedia(tx, item.ID, media); err != nil {
				return err
			}
		}
		if tagIDs != nil {
			var tags []model.Tag
			if len(tagIDs) > 0 {
				if err := tx.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
					return err
				}
				if len(tags) != len(tagIDs) {
					return gorm.ErrRecordNotFound
				}
			}
			if err := tx.Model(&item).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}
		assetIDs := make([]uint, 0, len(media)+2)
		if item.LogoAssetID != nil {
			assetIDs = append(assetIDs, *item.LogoAssetID)
		}
		if item.CoverAssetID != nil {
			assetIDs = append(assetIDs, *item.CoverAssetID)
		}
		for _, entry := range media {
			if entry.AssetID != nil {
				assetIDs = append(assetIDs, *entry.AssetID)
			}
		}
		if len(assetIDs) > 0 {
			if err := tx.Model(&model.MediaAsset{}).Where("id IN ?", uniqueUintIDs(assetIDs)).Updates(map[string]any{"status": "published", "published_at": &publishedAt}).Error; err != nil {
				return err
			}
		}
		return nil
	case "product":
		var item model.Product
		if err := json.Unmarshal([]byte(revision.Snapshot), &item); err != nil {
			return err
		}
		item.ID, item.ContentVersion, item.PublicationStatus, item.Status, item.PublishedAt = revision.ResourceID, version, "published", 1, &publishedAt
		if !hasPublishableSupplier(tx, item.ID, nil) {
			return errProductRequiresVendor
		}
		if err := database.EnsureProductSlug(tx, &item); err != nil {
			return err
		}
		if err := optimisticSave(tx, &item, revision.ResourceID, revision.BaseVersion, "Category", "Vendor"); err != nil {
			return err
		}
		publishProductMedia(tx, item)
		return nil
	case "category":
		var item model.Category
		if err := json.Unmarshal([]byte(revision.Snapshot), &item); err != nil {
			return err
		}
		item.ID, item.ContentVersion, item.PublicationStatus, item.IsEnabled, item.PublishedAt = revision.ResourceID, version, "published", true, &publishedAt
		if err := database.EnsureCategorySlug(tx, &item); err != nil {
			return err
		}
		return optimisticSave(tx, &item, revision.ResourceID, revision.BaseVersion)
	case "page", "article":
		var item model.ContentPage
		if err := json.Unmarshal([]byte(revision.Snapshot), &item); err != nil {
			return err
		}
		item.ID, item.ContentVersion, item.PublicationStatus, item.IsEnabled, item.PublishedAt = revision.ResourceID, version, "published", true, &publishedAt
		if revision.ResourceType == "article" {
			item.PageType = "article"
		}
		if err := database.EnsurePageSlug(tx, &item); err != nil {
			return err
		}
		return optimisticSave(tx, &item, revision.ResourceID, revision.BaseVersion)
	default:
		return gorm.ErrInvalidData
	}
}

func optimisticSave(tx *gorm.DB, item any, id, baseVersion uint, omit ...string) error {
	query := tx.Model(item).Where("id = ? AND content_version = ?", id, baseVersion).Select("*").Omit("id", "created_at", "deleted_at")
	if len(omit) > 0 {
		query = query.Omit(omit...)
	}
	result := query.Updates(item)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errRevisionConflict
	}
	return nil
}

func archiveRevisionResource(db *gorm.DB, resourceType string, id, baseVersion uint) error {
	table := map[string]string{"vendor": "vendors", "product": "products", "category": "categories", "page": "content_pages", "article": "content_pages"}[resourceType]
	updates := map[string]any{"publication_status": "archived", "content_version": gorm.Expr("content_version + 1")}
	switch resourceType {
	case "vendor":
		updates["is_visible"] = false
	case "product":
		updates["status"] = 2
	case "category":
		updates["is_enabled"] = false
	case "page", "article":
		updates["is_enabled"] = false
	}
	result := db.Table(table).Where("id = ? AND content_version = ?", id, baseVersion).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errRevisionConflict
	}
	return nil
}

// PublishScheduledRevisions is safe to call from more than one instance: the
// status/version predicates make racing workers converge on one publication.
func PublishScheduledRevisions(db *gorm.DB) error {
	var rows []model.ContentRevision
	if err := db.Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", "scheduled", time.Now()).Order("scheduled_at asc").Limit(100).Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		if err := publishRevision(db, &rows[i], rows[i].Reviewer, rows[i].ReviewNote); err != nil && !errors.Is(err, errRevisionConflict) {
			return err
		}
	}
	return nil
}
