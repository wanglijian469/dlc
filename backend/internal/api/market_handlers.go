package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type marketPostInput struct {
	Type             string `json:"type"`
	Title            string `json:"title"`
	CategoryID       *uint  `json:"categoryId"`
	CompatibleModels string `json:"compatibleModels"`
	Province         string `json:"province"`
	City             string `json:"city"`
	Quantity         string `json:"quantity"`
	DeliveryNote     string `json:"deliveryNote"`
	Description      string `json:"description"`
	ContactName      string `json:"contactName"`
	ContactPhone     string `json:"contactPhone"`
	ExpiresInDays    int    `json:"expiresInDays"`
	AssetIDs         []uint `json:"assetIds"`
}

type marketPostDTO struct {
	ID               uint            `json:"id"`
	Type             string          `json:"type"`
	Title            string          `json:"title"`
	CategoryID       *uint           `json:"categoryId,omitempty"`
	Category         *model.Category `json:"category,omitempty"`
	CompatibleModels string          `json:"compatibleModels"`
	Province         string          `json:"province"`
	City             string          `json:"city"`
	Quantity         string          `json:"quantity"`
	DeliveryNote     string          `json:"deliveryNote"`
	Description      string          `json:"description"`
	PublisherName    string          `json:"publisherName"`
	VendorID         *uint           `json:"vendorId,omitempty"`
	Status           string          `json:"status"`
	Images           []string        `json:"images"`
	ExpiresAt        time.Time       `json:"expiresAt"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

func (h AdminHandler) ListMarketPosts(c *gin.Context) {
	query := h.DB.Model(&model.MarketPost{}).Where("status = ? AND expires_at > ?", "published", time.Now())
	if value := strings.TrimSpace(c.Query("type")); value == "supply" || value == "demand" {
		query = query.Where("post_type = ?", value)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR compatible_models LIKE ?", like, like, like)
	}
	if id, err := strconv.ParseUint(c.Query("categoryId"), 10, 64); err == nil && id > 0 {
		query = query.Where("category_id = ?", id)
	}
	if province := strings.TrimSpace(c.Query("province")); province != "" {
		query = query.Where("province = ?", province)
	}
	if value, err := time.Parse(time.RFC3339, strings.TrimSpace(c.Query("publishedAfter"))); err == nil {
		query = query.Where("created_at >= ?", value)
	}
	if value, err := time.Parse(time.RFC3339, strings.TrimSpace(c.Query("publishedBefore"))); err == nil {
		query = query.Where("created_at <= ?", value)
	}
	h.writeMarketPostPage(c, query, false)
}

func (h AdminHandler) MarketPostDetail(c *gin.Context) {
	var item model.MarketPost
	if h.marketPostQuery().Where("market_posts.id = ? AND market_posts.status = ? AND market_posts.expires_at > ?", c.Param("id"), "published", time.Now()).First(&item).Error != nil {
		Fail(c, http.StatusNotFound, 404, "供求信息不存在或已下架")
		return
	}
	OK(c, marketDTO(item))
}

func (h AdminHandler) MarketPostContact(c *gin.Context) {
	var item model.MarketPost
	if h.DB.Where("id = ? AND status = ? AND expires_at > ?", c.Param("id"), "published", time.Now()).First(&item).Error != nil {
		Fail(c, http.StatusNotFound, 404, "供求信息不存在或已下架")
		return
	}
	userID := c.GetUint("userId")
	ipAddress := c.ClientIP()
	var recent int64
	h.DB.Model(&model.MarketContactAccessLog{}).
		Where("(user_id = ? OR ip_address = ?) AND created_at > ?", userID, ipAddress, time.Now().Add(-10*time.Minute)).
		Count(&recent)
	if recent >= 30 {
		Fail(c, http.StatusTooManyRequests, 429, "联系方式查看过于频繁，请稍后再试")
		return
	}
	h.DB.Create(&model.MarketContactAccessLog{MarketPostID: item.ID, UserID: userID, IPAddress: ipAddress})
	OK(c, gin.H{"marketPostId": item.ID, "contactName": item.ContactName, "phone": item.ContactPhone})
}

func (h AdminHandler) ListOwnMarketPosts(c *gin.Context) {
	query := h.DB.Model(&model.MarketPost{}).Where("owner_user_id = ?", c.GetUint("userId"))
	h.writeMarketPostPage(c, query, true)
}

func (h AdminHandler) OwnMarketPost(c *gin.Context) {
	var item model.MarketPost
	if h.marketPostQuery().Where("market_posts.id = ? AND market_posts.owner_user_id = ?", c.Param("id"), c.GetUint("userId")).First(&item).Error != nil {
		Fail(c, http.StatusNotFound, 404, "供求信息不存在")
		return
	}
	OK(c, gin.H{"post": marketDTO(item), "contactName": item.ContactName, "contactPhone": item.ContactPhone, "assetIds": marketAssetIDs(item)})
}

func (h AdminHandler) CreateOwnMarketPost(c *gin.Context) {
	var input marketPostInput
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, http.StatusBadRequest, 400, "供求信息格式不正确")
		return
	}
	item, ok := h.marketPostFromInput(c, input, model.MarketPost{})
	if !ok {
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		return replaceMarketPostMedia(tx, item.ID, c.GetString("username"), input.AssetIDs)
	}); err != nil {
		Fail(c, http.StatusInternalServerError, 500, "供求信息发布失败")
		return
	}
	h.marketPostQuery().First(&item, item.ID)
	logOperation(h.DB, c.GetString("username"), "publish", "market-posts", item.ID)
	OK(c, marketDTO(item))
}

func (h AdminHandler) UpdateOwnMarketPost(c *gin.Context) {
	var item model.MarketPost
	if h.DB.Where("id = ? AND owner_user_id = ?", c.Param("id"), c.GetUint("userId")).First(&item).Error != nil {
		Fail(c, http.StatusNotFound, 404, "供求信息不存在")
		return
	}
	if item.Status == "removed" {
		Fail(c, http.StatusConflict, 409, "已被平台下架的信息不能编辑")
		return
	}
	var input marketPostInput
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, http.StatusBadRequest, 400, "供求信息格式不正确")
		return
	}
	updated, ok := h.marketPostFromInput(c, input, item)
	if !ok {
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&updated).Error; err != nil {
			return err
		}
		return replaceMarketPostMedia(tx, updated.ID, c.GetString("username"), input.AssetIDs)
	}); err != nil {
		Fail(c, 500, 500, "供求信息保存失败")
		return
	}
	h.marketPostQuery().First(&updated, updated.ID)
	logOperation(h.DB, c.GetString("username"), "update", "market-posts", updated.ID)
	OK(c, marketDTO(updated))
}

func (h AdminHandler) WithdrawOwnMarketPost(c *gin.Context) {
	result := h.DB.Model(&model.MarketPost{}).Where("id = ? AND owner_user_id = ? AND status <> ?", c.Param("id"), c.GetUint("userId"), "removed").Update("status", "withdrawn")
	if result.Error != nil || result.RowsAffected == 0 {
		Fail(c, http.StatusNotFound, 404, "供求信息不存在或不能撤回")
		return
	}
	OK(c, gin.H{"withdrawn": true})
}

func (h AdminHandler) AdminListMarketPosts(c *gin.Context) {
	query := h.DB.Model(&model.MarketPost{})
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		if status == "expired" {
			query = query.Where("status = ? OR (status = ? AND expires_at <= ?)", "expired", "published", time.Now())
		} else {
			query = query.Where("status = ?", status)
		}
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		query = query.Where("title LIKE ? OR owner_username LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	h.writeMarketPostPage(c, query, true)
}

func (h AdminHandler) AdminUpdateMarketPostStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if c.ShouldBindJSON(&req) != nil || (req.Status != "published" && req.Status != "removed") {
		Fail(c, http.StatusBadRequest, 400, "状态只能是 published 或 removed")
		return
	}
	updates := map[string]any{"status": req.Status, "removal_reason": clean(req.Reason, 500)}
	if req.Status == "published" {
		updates["removal_reason"] = ""
		updates["expires_at"] = gorm.Expr("IF(expires_at > NOW(), expires_at, DATE_ADD(NOW(), INTERVAL 30 DAY))")
	}
	result := h.DB.Model(&model.MarketPost{}).Where("id = ?", c.Param("id")).Updates(updates)
	if result.Error != nil || result.RowsAffected == 0 {
		Fail(c, http.StatusNotFound, 404, "供求信息不存在")
		return
	}
	var item model.MarketPost
	h.marketPostQuery().First(&item, c.Param("id"))
	logOperation(h.DB, c.GetString("username"), req.Status, "market-posts", item.ID)
	OK(c, marketDTO(item))
}

func (h AdminHandler) marketPostFromInput(c *gin.Context, input marketPostInput, item model.MarketPost) (model.MarketPost, bool) {
	input.Type, input.Title, input.Description = clean(input.Type, 20), clean(input.Title, 160), clean(input.Description, 5000)
	input.ContactName, input.ContactPhone = clean(input.ContactName, 80), clean(input.ContactPhone, 40)
	if input.Type != "supply" && input.Type != "demand" {
		Fail(c, 400, 400, "供求类型必须是 supply 或 demand")
		return item, false
	}
	role := c.GetString("role")
	if (role == "buyer" && input.Type != "demand") || (role == "vendor" && input.Type != "supply") {
		Fail(c, http.StatusForbidden, 403, "采购商只能发布求购，厂商只能发布供应")
		return item, false
	}
	if len([]rune(input.Title)) < 4 || len([]rune(input.Description)) < 10 || input.ContactName == "" || len(input.ContactPhone) < 6 {
		Fail(c, 400, 400, "标题至少 4 字、正文至少 10 字，并请填写有效联系人和电话")
		return item, false
	}
	if len(input.AssetIDs) > 6 {
		Fail(c, 400, 400, "每条供求信息最多上传 6 张图片")
		return item, false
	}
	seenAssets := make(map[uint]struct{}, len(input.AssetIDs))
	for _, assetID := range input.AssetIDs {
		if assetID == 0 {
			Fail(c, 400, 400, "图片编号无效")
			return item, false
		}
		if _, exists := seenAssets[assetID]; exists {
			Fail(c, 400, 400, "不能重复选择同一张图片")
			return item, false
		}
		seenAssets[assetID] = struct{}{}
	}
	days := input.ExpiresInDays
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		Fail(c, 400, 400, "有效期最长 90 天")
		return item, false
	}
	if input.CategoryID != nil {
		var count int64
		h.DB.Model(&model.Category{}).Where("id = ? AND is_enabled = ?", *input.CategoryID, true).Count(&count)
		if count == 0 {
			Fail(c, 400, 400, "所选分类不存在")
			return item, false
		}
	}
	item.PostType, item.Title, item.CategoryID = input.Type, input.Title, input.CategoryID
	item.CompatibleModels, item.Province, item.City = clean(input.CompatibleModels, 500), clean(input.Province, 50), clean(input.City, 50)
	item.Quantity, item.DeliveryNote, item.Description = clean(input.Quantity, 100), clean(input.DeliveryNote, 255), input.Description
	item.ContactName, item.ContactPhone = input.ContactName, input.ContactPhone
	item.ExpiresAt = time.Now().Add(time.Duration(days) * 24 * time.Hour)
	if item.ID == 0 {
		item.OwnerUserID, item.OwnerUsername, item.Status = c.GetUint("userId"), c.GetString("username"), "published"
		if id := c.GetUint("vendorId"); id > 0 {
			item.VendorID = &id
		}
	} else if item.Status == "withdrawn" || item.Status == "expired" {
		item.Status = "published"
	}
	return item, true
}

func (h AdminHandler) marketPostQuery() *gorm.DB {
	return h.DB.Preload("Category").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") })
}

func (h AdminHandler) writeMarketPostPage(c *gin.Context, query *gorm.DB, includePrivate bool) {
	page, pageSize := pageParams(c, 20)
	var total int64
	query.Count(&total)
	var rows []model.MarketPost
	if err := query.Preload("Category").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Order("created_at desc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		Fail(c, 500, 500, "供求信息加载失败")
		return
	}
	items := make([]marketPostDTO, 0, len(rows))
	for _, row := range rows {
		dto := marketDTO(row)
		if !includePrivate {
			dto.Status = "published"
		}
		items = append(items, dto)
	}
	OK(c, PageResult{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func marketDTO(item model.MarketPost) marketPostDTO {
	images := make([]string, 0, len(item.Media))
	for _, media := range item.Media {
		images = append(images, "/api/media/"+strconv.FormatUint(uint64(media.AssetID), 10))
	}
	status := item.Status
	if status == "published" && item.ExpiresAt.Before(time.Now()) {
		status = "expired"
	}
	return marketPostDTO{ID: item.ID, Type: item.PostType, Title: item.Title, CategoryID: item.CategoryID, Category: item.Category, CompatibleModels: item.CompatibleModels, Province: item.Province, City: item.City, Quantity: item.Quantity, DeliveryNote: item.DeliveryNote, Description: item.Description, PublisherName: item.OwnerUsername, VendorID: item.VendorID, Status: status, Images: images, ExpiresAt: item.ExpiresAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func marketAssetIDs(item model.MarketPost) []uint {
	ids := make([]uint, 0, len(item.Media))
	for _, media := range item.Media {
		ids = append(ids, media.AssetID)
	}
	return ids
}

func replaceMarketPostMedia(tx *gorm.DB, postID uint, username string, assetIDs []uint) error {
	if len(assetIDs) > 0 {
		var count int64
		if err := tx.Model(&model.MediaAsset{}).Where("id IN ? AND owner_username = ? AND status IN ?", assetIDs, username, []string{"staged", "published"}).Count(&count).Error; err != nil || count != int64(len(assetIDs)) {
			return gorm.ErrRecordNotFound
		}
	}
	if err := tx.Where("market_post_id = ?", postID).Delete(&model.MarketPostMedia{}).Error; err != nil {
		return err
	}
	for index, assetID := range assetIDs {
		if err := tx.Create(&model.MarketPostMedia{MarketPostID: postID, AssetID: assetID, SortOrder: index}).Error; err != nil {
			return err
		}
	}
	if len(assetIDs) > 0 {
		now := time.Now()
		return tx.Model(&model.MediaAsset{}).Where("id IN ?", assetIDs).Updates(map[string]any{"status": "published", "published_at": &now}).Error
	}
	return nil
}
