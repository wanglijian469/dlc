package api

import (
	"net/http"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type vendorPostInput struct {
	PostType     string `json:"postType"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Content      string `json:"content"`
	CoverImage   string `json:"coverImage"`
	CoverAssetID *uint  `json:"coverAssetId"`
}

func validateVendorPost(input *vendorPostInput) string {
	input.PostType = strings.TrimSpace(input.PostType)
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Content = strings.TrimSpace(input.Content)
	if input.PostType != "update" && input.PostType != "case" {
		return "内容类型只能是企业动态或案例"
	}
	if input.Title == "" || len([]rune(input.Title)) > 160 || input.Content == "" {
		return "请填写标题和正文，标题不超过 160 字"
	}
	if len([]rune(input.Summary)) > 500 {
		return "摘要不能超过 500 字"
	}
	return ""
}

func applyVendorPostInput(row *model.VendorPost, input vendorPostInput) {
	row.PostType = input.PostType
	row.Title = input.Title
	row.Summary = input.Summary
	row.Content = input.Content
	row.CoverImage = strings.TrimSpace(input.CoverImage)
	row.CoverAssetID = input.CoverAssetID
}

func (h PublicHandler) VendorPosts(c *gin.Context) {
	var vendor model.Vendor
	if h.DB.Where("id = ? AND is_visible = ? AND publication_status = ?", c.Param("id"), true, "published").First(&vendor).Error != nil {
		Fail(c, http.StatusNotFound, 404, "厂商不存在")
		return
	}
	var rows []model.VendorPost
	query := h.DB.Where("vendor_id = ? AND status = ? AND published_at <= ?", vendor.ID, "approved", time.Now()).Order("published_at desc, id desc")
	if postType := strings.TrimSpace(c.Query("type")); postType != "" {
		query = query.Where("post_type = ?", postType)
	}
	if query.Limit(50).Find(&rows).Error != nil {
		Fail(c, 500, 500, "动态加载失败")
		return
	}
	OK(c, rows)
}

func (h AdminHandler) ListOwnVendorPosts(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var rows []model.VendorPost
	if h.DB.Where("vendor_id = ?", vendorID).Order("id desc").Find(&rows).Error != nil {
		Fail(c, 500, 500, "动态加载失败")
		return
	}
	OK(c, rows)
}

func (h AdminHandler) ListVendorPostsForRole(c *gin.Context) {
	if c.GetString("role") == "vendor" {
		h.ListOwnVendorPosts(c)
		return
	}
	h.AdminListVendorPosts(c)
}

func (h AdminHandler) CreateOwnVendorPost(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var input vendorPostInput
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, 400, 400, "填写内容无效")
		return
	}
	if message := validateVendorPost(&input); message != "" {
		Fail(c, 400, 400, message)
		return
	}
	row := model.VendorPost{VendorID: vendorID, Status: "draft", SubmittedBy: c.GetString("username")}
	applyVendorPostInput(&row, input)
	if h.DB.Create(&row).Error != nil {
		Fail(c, 500, 500, "保存失败")
		return
	}
	OK(c, row)
}

func (h AdminHandler) UpdateOwnVendorPost(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	var row model.VendorPost
	if h.DB.Where("id = ? AND vendor_id = ?", c.Param("id"), vendorID).First(&row).Error != nil {
		Fail(c, 404, 404, "内容不存在")
		return
	}
	if row.Status == "pending" {
		Fail(c, 409, 409, "审核中的内容请先撤回")
		return
	}
	var input vendorPostInput
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, 400, 400, "填写内容无效")
		return
	}
	if message := validateVendorPost(&input); message != "" {
		Fail(c, 400, 400, message)
		return
	}
	applyVendorPostInput(&row, input)
	row.Status = "draft"
	row.ReviewNote = ""
	row.PublishedAt = nil
	if h.DB.Save(&row).Error != nil {
		Fail(c, 500, 500, "保存失败")
		return
	}
	OK(c, row)
}

func (h AdminHandler) SubmitOwnVendorPost(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	result := h.DB.Model(&model.VendorPost{}).Where("id = ? AND vendor_id = ? AND status IN ?", c.Param("id"), vendorID, []string{"draft", "rejected"}).Updates(map[string]any{"status": "pending", "review_note": "", "submitted_by": c.GetString("username")})
	if result.Error != nil || result.RowsAffected != 1 {
		Fail(c, 409, 409, "仅草稿或被驳回内容可以提交")
		return
	}
	OK(c, gin.H{"status": "pending"})
}

func (h AdminHandler) WithdrawOwnVendorPost(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	result := h.DB.Model(&model.VendorPost{}).Where("id = ? AND vendor_id = ? AND status IN ?", c.Param("id"), vendorID, []string{"pending", "approved"}).Updates(map[string]any{"status": "withdrawn", "published_at": nil})
	if result.Error != nil || result.RowsAffected != 1 {
		Fail(c, 409, 409, "该内容当前不能撤回")
		return
	}
	OK(c, gin.H{"status": "withdrawn"})
}

func (h AdminHandler) AdminListVendorPosts(c *gin.Context) {
	var rows []model.VendorPost
	query := h.DB.Order("id desc")
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if query.Limit(200).Find(&rows).Error != nil {
		Fail(c, 500, 500, "内容加载失败")
		return
	}
	OK(c, rows)
}

func (h AdminHandler) ReviewVendorPost(c *gin.Context) {
	var input struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if c.ShouldBindJSON(&input) != nil || (input.Status != "approved" && input.Status != "rejected") {
		Fail(c, 400, 400, "审核状态无效")
		return
	}
	var row model.VendorPost
	if h.DB.Where("id = ? AND status = ?", c.Param("id"), "pending").First(&row).Error != nil {
		Fail(c, 404, 404, "待审核内容不存在")
		return
	}
	now := time.Now()
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{"status": input.Status, "review_note": strings.TrimSpace(input.Note), "reviewed_by": c.GetString("username"), "reviewed_at": &now}
		if input.Status == "approved" {
			updates["published_at"] = &now
		}
		if err := tx.Model(&row).Updates(updates).Error; err != nil {
			return err
		}
		if input.Status == "approved" && row.CoverAssetID != nil {
			return tx.Model(&model.MediaAsset{}).Where("id = ? AND vendor_id = ?", *row.CoverAssetID, row.VendorID).Updates(map[string]any{"status": "published", "published_at": &now}).Error
		}
		return nil
	})
	if err != nil {
		Fail(c, 500, 500, "审核保存失败")
		return
	}
	h.DB.First(&row, row.ID)
	OK(c, row)
}
