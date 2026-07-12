package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type vendorSubmissionView struct {
	model.VendorSubmission
	Draft model.Vendor `json:"draft"`
}

func (h AdminHandler) GetVendorProfile(c *gin.Context) {
	vendorID := c.GetUint("vendorId")
	if c.GetString("role") != "vendor" || vendorID == 0 {
		Fail(c, http.StatusForbidden, 403, "当前账号未绑定厂商")
		return
	}
	var vendor model.Vendor
	if err := h.DB.Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).First(&vendor, vendorID).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "绑定的厂商不存在")
		return
	}
	var submission model.VendorSubmission
	response := gin.H{"vendor": vendor, "draft": vendor, "submission": nil}
	if err := h.DB.Where("vendor_id = ?", vendorID).Order("id desc").First(&submission).Error; err == nil {
		var draft model.Vendor
		if json.Unmarshal([]byte(submission.Payload), &draft) == nil {
			response["draft"] = draft
		}
		response["submission"] = vendorSubmissionView{VendorSubmission: submission, Draft: draft}
	}
	OK(c, response)
}

func (h AdminHandler) SubmitVendorProfile(c *gin.Context) {
	vendorID := c.GetUint("vendorId")
	if c.GetString("role") != "vendor" || vendorID == 0 {
		Fail(c, http.StatusForbidden, 403, "当前账号未绑定厂商")
		return
	}
	var current model.Vendor
	if err := h.DB.Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).First(&current, vendorID).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "绑定的厂商不存在")
		return
	}
	var proposed model.Vendor
	if err := c.ShouldBindJSON(&proposed); err != nil {
		Fail(c, http.StatusBadRequest, 400, "提交内容格式不正确")
		return
	}
	applyVendorEditableFields(&current, proposed)
	if proposed.Media != nil {
		media, mediaErr := sanitizeVendorMedia(proposed.Media)
		if mediaErr != nil {
			Fail(c, http.StatusBadRequest, 400, mediaErr.Error())
			return
		}
		current.Media = media
	}
	if err := h.validateSubmissionAssets(c, vendorID, current); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if strings.TrimSpace(current.Name) == "" {
		Fail(c, http.StatusBadRequest, 400, "厂商名称不能为空")
		return
	}
	if current.WebsiteURL != "" && !validURL(current.WebsiteURL) {
		Fail(c, http.StatusBadRequest, 400, "官网地址格式不正确")
		return
	}
	payload, err := json.Marshal(current)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "提交内容保存失败")
		return
	}
	// Each submit is an immutable revision. Older pending rows remain available
	// for audit, but cannot later be approved over the newer revision.
	submission := model.VendorSubmission{VendorID: vendorID, BaseVersion: current.ContentVersion, Status: "pending", Payload: string(payload), SubmittedBy: c.GetString("username")}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.VendorSubmission{}).Where("vendor_id = ? AND status = ?", vendorID, "pending").Update("status", "superseded").Error; err != nil {
			return err
		}
		return tx.Create(&submission).Error
	}); err != nil {
		Fail(c, http.StatusInternalServerError, 500, "提交内容保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "submit", "vendor-submissions", submission.ID)
	OK(c, vendorSubmissionView{VendorSubmission: submission, Draft: current})
}

func (h AdminHandler) ListVendorSubmissions(c *gin.Context) {
	var rows []model.VendorSubmission
	query := h.DB.Model(&model.VendorSubmission{}).Preload("Vendor").Preload("Vendor.Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Order("created_at desc, id desc")
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	page, pageSize := pageParams(c, 20)
	var total int64
	if adminPaginationRequested(c) {
		query.Count(&total)
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	if err := query.Find(&rows).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "审核列表加载失败")
		return
	}
	views := make([]vendorSubmissionView, 0, len(rows))
	for _, row := range rows {
		var draft model.Vendor
		_ = json.Unmarshal([]byte(row.Payload), &draft)
		views = append(views, vendorSubmissionView{VendorSubmission: row, Draft: draft})
	}
	if adminPaginationRequested(c) {
		OK(c, PageResult{Items: views, Page: page, PageSize: pageSize, Total: total})
		return
	}
	OK(c, views)
}

func (h AdminHandler) ReviewVendorSubmission(c *gin.Context) {
	var req struct {
		Status     string `json:"status"`
		ReviewNote string `json:"reviewNote"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != "approved" && req.Status != "rejected") {
		Fail(c, http.StatusBadRequest, 400, "审核状态只能是 approved 或 rejected")
		return
	}
	if req.Status == "rejected" && len([]rune(strings.TrimSpace(req.ReviewNote))) < 2 {
		Fail(c, http.StatusBadRequest, 400, "驳回时必须填写有效原因")
		return
	}
	var result model.VendorSubmission
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&result, c.Param("id")).Error; err != nil {
			return err
		}
		if result.Status != "pending" {
			return errSubmissionReviewed
		}
		if req.Status == "approved" {
			var draft model.Vendor
			if err := json.Unmarshal([]byte(result.Payload), &draft); err != nil {
				return err
			}
			var vendor model.Vendor
			if err := tx.First(&vendor, result.VendorID).Error; err != nil {
				return err
			}
			if vendor.ContentVersion != result.BaseVersion {
				return errSubmissionConflict
			}
			applyVendorEditableFields(&vendor, draft)
			vendor.ReviewStatus = "verified"
			if vendor.PublicationStatus == "" || vendor.PublicationStatus == "draft" {
				vendor.PublicationStatus = "published"
				vendor.IsVisible = true
			}
			vendor.ContentVersion++
			if err := tx.Save(&vendor).Error; err != nil {
				return err
			}
			if err := replaceVendorMedia(tx, vendor.ID, draft.Media); err != nil {
				return err
			}
			assetIDs := make([]uint, 0, len(draft.Media))
			if draft.LogoAssetID != nil {
				assetIDs = append(assetIDs, *draft.LogoAssetID)
			}
			if draft.CoverAssetID != nil {
				assetIDs = append(assetIDs, *draft.CoverAssetID)
			}
			for _, media := range draft.Media {
				if media.AssetID != nil {
					assetIDs = append(assetIDs, *media.AssetID)
				}
			}
			if len(assetIDs) > 0 {
				now := time.Now()
				if err := tx.Model(&model.MediaAsset{}).Where("id IN ?", assetIDs).Updates(map[string]interface{}{"status": "published", "published_at": &now}).Error; err != nil {
					return err
				}
			}
		}
		now := time.Now()
		result.Status = req.Status
		result.ReviewNote = strings.TrimSpace(req.ReviewNote)
		result.ReviewedBy = c.GetString("username")
		result.ReviewedAt = &now
		return tx.Save(&result).Error
	})
	if err != nil {
		if err == errSubmissionReviewed {
			Fail(c, http.StatusConflict, 409, "该提交已审核")
		} else if err == errSubmissionConflict {
			Fail(c, http.StatusConflict, 409, "正式资料已在提交后发生变化，请重新比较并提交")
		} else if err == gorm.ErrRecordNotFound {
			Fail(c, http.StatusNotFound, 404, "提交记录不存在")
		} else {
			Fail(c, http.StatusInternalServerError, 500, "审核保存失败")
		}
		return
	}
	logOperation(h.DB, c.GetString("username"), req.Status, "vendor-submissions", result.ID)
	OK(c, result)
}

var errSubmissionReviewed = &workflowError{"submission already reviewed"}
var errSubmissionConflict = &workflowError{"vendor content version conflict"}

type workflowError struct{ message string }

func (e *workflowError) Error() string { return e.message }

func applyVendorEditableFields(dst *model.Vendor, src model.Vendor) {
	dst.Name = strings.TrimSpace(src.Name)
	dst.ShortName = src.ShortName
	dst.Logo = src.Logo
	dst.LogoAssetID = src.LogoAssetID
	dst.CoverImage = src.CoverImage
	dst.CoverAssetID = src.CoverAssetID
	dst.Province = src.Province
	dst.City = src.City
	dst.County = src.County
	dst.Address = src.Address
	dst.MainProducts = src.MainProducts
	dst.ServiceModels = src.ServiceModels
	dst.ServiceAdvantages = src.ServiceAdvantages
	dst.Description = src.Description
	dst.EstablishedYear = src.EstablishedYear
	dst.FactoryArea = src.FactoryArea
	dst.EmployeeCount = src.EmployeeCount
	dst.AnnualCapacity = src.AnnualCapacity
	dst.Equipment = src.Equipment
	dst.Certifications = src.Certifications
	dst.AfterSalesService = src.AfterSalesService
	dst.ProvidesProcessing = src.ProvidesProcessing
	dst.ProcessingServices = src.ProcessingServices
	dst.ProcessingMaterials = src.ProcessingMaterials
	dst.ProcessingEquipment = src.ProcessingEquipment
	dst.ProcessingCapacity = src.ProcessingCapacity
	dst.ProcessingRegions = src.ProcessingRegions
	dst.ProcessingNotes = src.ProcessingNotes
	dst.WebsiteURL = strings.TrimSpace(src.WebsiteURL)
	dst.Phone = src.Phone
	dst.Wechat = src.Wechat
	dst.ContactName = src.ContactName
}

func sanitizeVendorMedia(rows []model.VendorMedia) ([]model.VendorMedia, error) {
	allowed := map[string]bool{"factory": true, "equipment": true, "certificate": true}
	clean := make([]model.VendorMedia, 0, len(rows))
	for _, row := range rows {
		row.Kind = strings.TrimSpace(row.Kind)
		row.URL = strings.TrimSpace(row.URL)
		row.Caption = strings.TrimSpace(row.Caption)
		if row.URL == "" && row.AssetID == nil {
			continue
		}
		if row.AssetID != nil {
			row.URL = fmt.Sprintf("/api/media/%d", *row.AssetID)
		}
		if !allowed[row.Kind] {
			return nil, fmt.Errorf("图片类型只能是 factory、equipment 或 certificate")
		}
		if row.URL != "" && !validURL(row.URL) {
			return nil, fmt.Errorf("图片地址格式不正确")
		}
		row.ID = 0
		row.VendorID = 0
		clean = append(clean, row)
	}
	return clean, nil
}

func (h AdminHandler) validateSubmissionAssets(c *gin.Context, vendorID uint, vendor model.Vendor) error {
	assetIDs := []*uint{vendor.LogoAssetID, vendor.CoverAssetID}
	for _, row := range vendor.Media {
		assetIDs = append(assetIDs, row.AssetID)
	}
	for _, assetID := range assetIDs {
		if assetID == nil {
			continue
		}
		var asset model.MediaAsset
		if err := h.DB.First(&asset, *assetID).Error; err != nil {
			return fmt.Errorf("草稿图片无权使用或已失效")
		}
		if asset.Status == "staged" && asset.OwnerUsername == c.GetString("username") && asset.VendorID != nil && *asset.VendorID == vendorID {
			continue
		}
		if asset.Status == "published" {
			var count int64
			h.DB.Model(&model.Vendor{}).Where("id = ? AND (logo_asset_id = ? OR cover_asset_id = ?)", vendorID, asset.ID, asset.ID).Count(&count)
			if count == 0 {
				h.DB.Model(&model.VendorMedia{}).Where("vendor_id = ? AND asset_id = ?", vendorID, asset.ID).Count(&count)
			}
			if count > 0 {
				continue
			}
		}
		return fmt.Errorf("草稿图片无权使用或已失效")
	}
	return nil
}

func replaceVendorMedia(db *gorm.DB, vendorID uint, rows []model.VendorMedia) error {
	clean, err := sanitizeVendorMedia(rows)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("vendor_id = ?", vendorID).Delete(&model.VendorMedia{}).Error; err != nil {
			return err
		}
		for i := range clean {
			clean[i].VendorID = vendorID
			if err := tx.Create(&clean[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (h AdminHandler) ListCMSUsers(c *gin.Context) {
	var users []model.AdminUser
	query := h.DB.Model(&model.AdminUser{}).Preload("Vendor")
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query = query.Where("username LIKE ?", "%"+search+"%")
	}
	if role := strings.TrimSpace(c.Query("role")); role != "" {
		query = query.Where("role = ?", role)
	}
	if adminPaginationRequested(c) {
		page, pageSize := pageParams(c, 20)
		result, err := paginate(query.Order("id asc"), &users, page, pageSize)
		if err != nil {
			Fail(c, 500, 500, "账号列表加载失败")
			return
		}
		OK(c, result)
		return
	}
	if err := query.Order("id asc").Find(&users).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "账号列表加载失败")
		return
	}
	OK(c, users)
}

func (h AdminHandler) CreateCMSUser(c *gin.Context) { h.saveCMSUser(c, 0) }
func (h AdminHandler) UpdateCMSUser(c *gin.Context) { h.saveCMSUser(c, idParam(c)) }

func (h AdminHandler) saveCMSUser(c *gin.Context, id uint) {
	var req struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Role      string `json:"role"`
		VendorID  *uint  `json:"vendorId"`
		IsEnabled *bool  `json:"isEnabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "账号内容格式不正确")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || (req.Role != "admin" && req.Role != "vendor") {
		Fail(c, http.StatusBadRequest, 400, "用户名不能为空，角色只能是 admin 或 vendor")
		return
	}
	if req.Role == "vendor" && (req.VendorID == nil || *req.VendorID == 0 || !recordExists[model.Vendor](h.DB, *req.VendorID)) {
		Fail(c, http.StatusBadRequest, 400, "厂商账号必须绑定有效厂商")
		return
	}
	var user model.AdminUser
	if id > 0 {
		if err := h.DB.First(&user, id).Error; err != nil {
			Fail(c, http.StatusNotFound, 404, "账号不存在")
			return
		}
	} else if len(req.Password) < 6 {
		Fail(c, http.StatusBadRequest, 400, "新账号密码至少 6 位")
		return
	}
	user.Username = req.Username
	user.Role = req.Role
	if req.Role == "vendor" {
		user.VendorID = req.VendorID
	} else {
		user.VendorID = nil
	}
	if req.IsEnabled != nil {
		user.IsEnabled = *req.IsEnabled
	} else if id == 0 {
		user.IsEnabled = true
	}
	if req.Password != "" {
		if len(req.Password) < 6 {
			Fail(c, http.StatusBadRequest, 400, "密码至少 6 位")
			return
		}
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			Fail(c, http.StatusInternalServerError, 500, "密码保存失败")
			return
		}
		user.PasswordHash = hash
	}
	if err := h.DB.Save(&user).Error; err != nil {
		Fail(c, http.StatusBadRequest, 400, "账号保存失败，用户名可能已存在")
		return
	}
	h.DB.Preload("Vendor").First(&user, user.ID)
	logOperation(h.DB, c.GetString("username"), upsertAction(id), "users", user.ID)
	OK(c, user)
}

func (h AdminHandler) DeleteCMSUser(c *gin.Context) {
	var user model.AdminUser
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "账号不存在")
		return
	}
	if user.Username == c.GetString("username") {
		Fail(c, http.StatusBadRequest, 400, "不能删除当前登录账号")
		return
	}
	if err := h.DB.Delete(&user).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "账号删除失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "delete", "users", user.ID)
	OK(c, gin.H{"deleted": true})
}
