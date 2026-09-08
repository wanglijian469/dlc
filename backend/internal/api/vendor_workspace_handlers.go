package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var errWorkDraftConflict = errors.New("草稿版本已变化，请重新打开草稿；当前内容未被覆盖")
var workKeyPattern = regexp.MustCompile("^[a-zA-Z0-9_-]{8,80}$")
var workMediaPattern = regexp.MustCompile("/api/media/([0-9]+)")

type saveWorkDraftRequest struct {
	Kind       string          `json:"kind"`
	TargetType string          `json:"targetType"`
	TargetID   uint            `json:"targetId"`
	Version    uint            `json:"version"`
	Payload    json.RawMessage `json:"payload"`
}

func (h AdminHandler) ListWorkDrafts(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	rows := []model.VendorWorkDraft{}
	if err := h.DB.Where("vendor_id = ? AND user_id = ? AND committed_at IS NULL", vendorID, c.GetUint("userId")).Order("updated_at desc").Find(&rows).Error; err != nil {
		Fail(c, 500, 500, "草稿加载失败")
		return
	}
	c.Header("Cache-Control", "private, no-store")
	OK(c, rows)
}

func (h AdminHandler) SaveWorkDraft(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	var req saveWorkDraftRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	if c.ShouldBindJSON(&req) != nil || !workKeyPattern.MatchString(c.Param("key")) || (req.Kind != "product" && req.Kind != "profile") || len(req.Payload) == 0 || req.Payload[0] != '{' {
		Fail(c, 400, 400, "草稿格式不正确或内容过大")
		return
	}
	var row model.VendorWorkDraft
	err := h.DB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Transaction(func(tx *gorm.DB) error {
		// Serialize initial saves, too: retrying the same client key cannot create two drafts.
		var vendor model.Vendor
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&vendor, vendorID).Error; err != nil {
			return err
		}
		found := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("vendor_id = ? AND user_id = ? AND client_key = ?", vendorID, c.GetUint("userId"), c.Param("key")).First(&row).Error
		if found != nil && !errors.Is(found, gorm.ErrRecordNotFound) {
			return found
		}
		if found == nil {
			if row.CommittedAt != nil || row.Kind != req.Kind || row.TargetType != req.TargetType || row.TargetID != req.TargetID {
				return errWorkDraftConflict
			}
		} else {
			if req.Version != 0 {
				return errWorkDraftConflict
			}
			row = model.VendorWorkDraft{VendorID: vendorID, UserID: c.GetUint("userId"), ClientKey: c.Param("key"), Kind: req.Kind, TargetType: req.TargetType, TargetID: req.TargetID, BaseVersion: vendor.ContentVersion}
			if req.Kind == "product" {
				switch req.TargetType {
				case "":
					if req.TargetID != 0 {
						return gorm.ErrRecordNotFound
					}
					row.BaseVersion = 0
				case "supplier":
					var supplier model.ProductSupplier
					if err := tx.Where("id = ? AND vendor_id = ? AND status <> ?", req.TargetID, vendorID, "disabled").First(&supplier).Error; err != nil {
						return err
					}
					row.BaseVersion = supplier.ContentVersion
				case "submission":
					var sub model.ProductSubmission
					if err := tx.Where("id = ? AND vendor_id = ? AND supplier_id IS NULL AND status IN ?", req.TargetID, vendorID, []string{"pending", "rejected"}).First(&sub).Error; err != nil {
						return err
					}
				default:
					return gorm.ErrRecordNotFound
				}
			} else if req.TargetID != 0 || req.TargetType != "" {
				return gorm.ErrRecordNotFound
			}
		}
		previousPayload := string(row.Payload)
		if req.Kind == "product" {
			var input vendorProductDraft
			if json.Unmarshal(req.Payload, &input) != nil {
				return fmt.Errorf("产品草稿格式不正确")
			}
			if err := validateWorkProductAssets(tx, vendorID, c.GetString("username"), input); err != nil {
				return err
			}
			row.Payload, _ = json.Marshal(input)
		} else {
			var input model.Vendor
			if json.Unmarshal(req.Payload, &input) != nil {
				return fmt.Errorf("企业草稿格式不正确")
			}
			applyVendorEditableFields(&vendor, input)
			if input.Media != nil {
				var err error
				vendor.Media, err = sanitizeVendorMedia(input.Media)
				if err != nil {
					return err
				}
			}
			handler := h
			handler.DB = tx
			if err := handler.validateSubmissionAssets(c, vendorID, vendor); err != nil {
				return err
			}
			row.Payload, _ = json.Marshal(vendor)
		}
		if found == nil && row.Version != req.Version {
			if previousPayload == string(row.Payload) {
				return nil
			}
			return errWorkDraftConflict
		}
		row.Version++
		return tx.Save(&row).Error
	})
	if err != nil {
		workDraftError(c, err)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	OK(c, row)
}

func (h AdminHandler) DeleteWorkDraft(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	result := h.DB.Where("id = ? AND vendor_id = ? AND user_id = ? AND committed_at IS NULL AND version = ?", c.Param("id"), vendorID, c.GetUint("userId"), queryUint(c, "version")).Delete(&model.VendorWorkDraft{})
	if result.Error != nil {
		Fail(c, 500, 500, "删除草稿失败")
		return
	}
	if result.RowsAffected == 0 {
		Fail(c, 409, 409, "草稿不存在或版本已变化")
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (h AdminHandler) CommitWorkDraft(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	var req struct {
		Version uint `json:"version"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, 400, 400, "请保存草稿后重试")
		return
	}
	var row model.VendorWorkDraft
	err := h.DB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Transaction(func(tx *gorm.DB) error {
		var vendor model.Vendor
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&vendor, vendorID).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND vendor_id = ? AND user_id = ?", c.Param("id"), vendorID, c.GetUint("userId")).First(&row).Error; err != nil {
			return err
		}
		if row.CommittedAt != nil {
			return nil
		}
		if row.Version != req.Version {
			return errWorkDraftConflict
		}
		if row.Kind == "profile" {
			if row.BaseVersion != vendor.ContentVersion {
				return errWorkDraftConflict
			}
			var input model.Vendor
			if err := json.Unmarshal(row.Payload, &input); err != nil {
				return err
			}
			applyVendorEditableFields(&vendor, input)
			vendor.Media = input.Media
			if err := validateVendorProfileContent(vendor); err != nil {
				return err
			}
			slug, err := database.ValidateVendorSiteSlug(tx, vendor.Slug, vendorID)
			if err != nil {
				return err
			}
			vendor.Slug = slug
			service.ApplyVendorSEO(&vendor)
			handler := h
			handler.DB = tx
			if err := handler.validateSubmissionAssets(c, vendorID, vendor); err != nil {
				return err
			}
			payload, _ := json.Marshal(vendor)
			if err := tx.Model(&model.VendorSubmission{}).Where("vendor_id = ? AND status = ?", vendorID, "pending").Update("status", "superseded").Error; err != nil {
				return err
			}
			sub := model.VendorSubmission{VendorID: vendorID, BaseVersion: vendor.ContentVersion, Status: "pending", Payload: string(payload), SubmittedBy: c.GetString("username")}
			if err := tx.Create(&sub).Error; err != nil {
				return err
			}
			row.SubmissionID = sub.ID
		} else {
			var input vendorProductDraft
			if err := json.Unmarshal(row.Payload, &input); err != nil {
				return err
			}
			if err := validateVendorProductDraft(tx, input); err != nil {
				return err
			}
			if err := validateWorkProductAssets(tx, vendorID, c.GetString("username"), input); err != nil {
				return err
			}
			handler := h
			handler.DB = tx
			exact, _, err := handler.checkVendorProductDuplicate(vendorID, input.VendorProductName, input.VendorModel, row.TargetType, row.TargetID)
			if err != nil {
				return err
			}
			if exact {
				return fmt.Errorf("本厂已存在名称和型号相同的产品，请编辑原记录")
			}
			if row.TargetType == "supplier" {
				var supplier model.ProductSupplier
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Product").Where("id = ? AND vendor_id = ? AND status <> ?", row.TargetID, vendorID, "disabled").First(&supplier).Error; err != nil {
					return err
				}
				if supplier.ContentVersion != row.BaseVersion {
					return errWorkDraftConflict
				}
				applyVendorProductDraft(&supplier, input)
				if err := tx.Model(&model.ProductSubmission{}).Where("supplier_id = ? AND status = ?", supplier.ID, "pending").Update("status", "superseded").Error; err != nil {
					return err
				}
				if err := createProductSubmission(tx, c.GetString("username"), "update_offer", supplier.Product, supplier); err != nil {
					return err
				}
				var sub model.ProductSubmission
				if err := tx.Where("supplier_id = ? AND status = ?", supplier.ID, "pending").Order("id desc").First(&sub).Error; err != nil {
					return err
				}
				row.SubmissionID = sub.ID
			} else {
				if row.TargetType == "submission" {
					result := tx.Model(&model.ProductSubmission{}).Where("id = ? AND vendor_id = ? AND supplier_id IS NULL AND status IN ?", row.TargetID, vendorID, []string{"pending", "rejected"}).Update("status", "superseded")
					if result.Error != nil {
						return result.Error
					}
					if result.RowsAffected != 1 {
						return errWorkDraftConflict
					}
				}
				sub, err := createStandaloneProductSubmission(tx, c.GetString("username"), vendorID, input)
				if err != nil {
					return err
				}
				row.SubmissionID = sub.ID
			}
		}
		now := time.Now()
		row.CommittedAt = &now
		return tx.Save(&row).Error
	})
	if err != nil {
		workDraftError(c, err)
		return
	}
	logOperation(h.DB, c.GetString("username"), "submit", "vendor-work-drafts", row.ID)
	OK(c, row)
}

func workDraftError(c *gin.Context, err error) {
	if errors.Is(err, errWorkDraftConflict) {
		Fail(c, 409, 409, errWorkDraftConflict.Error())
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		Fail(c, 404, 404, "草稿或关联资料不存在")
		return
	}
	// Never return database errors containing draft text or private contact details.
	if strings.Contains(err.Error(), "Error ") || strings.Contains(err.Error(), "sql") {
		Fail(c, 500, 500, "草稿保存失败，请稍后重试")
		return
	}
	Fail(c, 400, 400, err.Error())
}

func validateWorkProductAssets(db *gorm.DB, vendorID uint, username string, input vendorProductDraft) error {
	raw := input.Image + input.GalleryRaw + input.SpecsRaw
	if strings.Contains(raw, "/capture-") || strings.Contains(raw, "base64,") {
		return fmt.Errorf("请先选择并确认公开使用的产品图片")
	}
	for _, match := range workMediaPattern.FindAllStringSubmatch(raw, -1) {
		var asset model.MediaAsset
		if db.First(&asset, parseUint(match[1])).Error != nil || asset.Purpose == "capture_source" || asset.Purpose == "capture_crop" {
			return fmt.Errorf("产品图片无权使用或已失效")
		}
		if asset.VendorID != nil && *asset.VendorID == vendorID && (asset.Status == "published" || (asset.Status == "staged" && asset.OwnerUsername == username)) {
			continue
		}
		var references int64
		if asset.Status == "published" {
			db.Model(&model.ProductSupplier{}).Where("vendor_id = ? AND (image = ? OR INSTR(gallery, ?) > 0 OR INSTR(specs, ?) > 0)", vendorID, match[0], match[0], match[0]).Count(&references)
		}
		if references == 0 {
			return fmt.Errorf("产品图片无权使用或已失效")
		}
	}
	return nil
}

func (h AdminHandler) VendorWorkspace(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	var vendor model.Vendor
	if h.DB.First(&vendor, vendorID).Error != nil {
		Fail(c, 404, 404, "厂商不存在")
		return
	}
	records, err := h.vendorProductRecords(vendorID)
	if err != nil {
		Fail(c, 500, 500, "工作台加载失败")
		return
	}
	var drafts, profileDrafts, unread int64
	if err = h.DB.Model(&model.VendorWorkDraft{}).Where("vendor_id = ? AND user_id = ? AND committed_at IS NULL AND kind = ?", vendorID, c.GetUint("userId"), "product").Count(&drafts).Error; err != nil {
		Fail(c, 500, 500, "工作台加载失败")
		return
	}
	h.DB.Model(&model.VendorWorkDraft{}).Where("vendor_id = ? AND user_id = ? AND committed_at IS NULL AND kind = ?", vendorID, c.GetUint("userId"), "profile").Count(&profileDrafts)
	h.DB.Model(&model.UserNotification{}).Where("user_id = ? AND read_at IS NULL", c.GetUint("userId")).Count(&unread)
	counts := map[string]int{"pending": 0, "approved": 0, "rejected": 0, "expired": 0, "missingImage": 0}
	for _, r := range records {
		if r.Status == "pending_update" {
			counts["pending"]++
		} else if r.Status == "rejected_update" {
			counts["rejected"]++
		} else {
			counts[r.Status]++
		}
		if r.Image == "" {
			counts["missingImage"]++
		}
		if r.PriceValidUntil != nil && r.PriceValidUntil.Before(time.Now()) {
			counts["expired"]++
		}
	}
	var sub model.VendorSubmission
	h.DB.Where("vendor_id = ?", vendorID).Order("id desc").First(&sub)
	OK(c, gin.H{"vendor": vendor, "counts": counts, "draftCount": drafts, "profileDraftCount": profileDrafts, "unreadCount": unread, "profileStatus": sub.Status, "profileReviewNote": sub.ReviewNote})
}

func notifyVendorReview(tx *gorm.DB, vendorID, id uint, kind, status, note string) error {
	var users []model.AdminUser
	if err := tx.Where("vendor_id = ? AND role = ? AND is_enabled = ?", vendorID, "vendor", true).Find(&users).Error; err != nil {
		return err
	}
	title := "产品审核通过"
	if kind == "vendor_review" {
		title = "企业资料审核通过"
	}
	if status == "rejected" {
		title = strings.Replace(title, "通过", "需修改", 1)
	}
	for _, u := range users {
		if err := tx.Create(&model.UserNotification{UserID: u.ID, BusinessType: kind, BusinessID: id, Title: title, Content: firstNonEmpty(note, "审核结果已更新，请前往工作台查看。")}).Error; err != nil {
			return err
		}
	}
	return nil
}
