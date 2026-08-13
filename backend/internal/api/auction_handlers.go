package api

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const auctionFeatureKey = "auction.enabled"

type auctionInput struct {
	Title                string
	CategoryID           *uint
	Specification        string
	CompatibleModels     string
	Quantity             int64
	Unit                 string
	DeliveryProvince     string
	DeliveryCity         string
	ExpectedDeliveryNote string
	Description          string
	Image                string
	ImageAssetID         *uint
	MaxBudgetCents       int64
	EndAt                time.Time
	ExpectedVersion      uint
}

type bidInput struct {
	UnitPriceCents  int64
	TaxIncluded     bool
	FreightNote     string
	DeliveryDays    int
	SupplyNote      string
	PromiseText     string
	ExpectedVersion uint
}

func auctionFeatureEnabled(db *gorm.DB) bool {
	var row model.SiteConfig
	if db.Where("config_key = ?", auctionFeatureKey).First(&row).Error != nil {
		return false
	}
	value := strings.TrimSpace(strings.ToLower(row.ConfigValue))
	return value == "true" || value == "1" || value == "{\"enabled\":true}"
}

func requireAuctionFeature(c *gin.Context, db *gorm.DB) bool {
	if auctionFeatureEnabled(db) {
		return true
	}
	Fail(c, http.StatusServiceUnavailable, 503, "采购竞价功能尚未启用")
	return false
}

func validateAuctionInput(db *gorm.DB, input *auctionInput, publishing bool) string {
	input.Title = strings.TrimSpace(input.Title)
	input.Unit = strings.TrimSpace(input.Unit)
	input.DeliveryProvince = strings.TrimSpace(input.DeliveryProvince)
	input.DeliveryCity = strings.TrimSpace(input.DeliveryCity)
	if input.Title == "" || len([]rune(input.Title)) > 160 || input.Quantity <= 0 || input.Unit == "" || input.MaxBudgetCents < 0 {
		return "请填写有效的采购名称、数量、单位和预算"
	}
	if input.CategoryID != nil {
		var count int64
		if db.Model(&model.Category{}).Where("id = ? AND is_enabled = ?", *input.CategoryID, true).Count(&count).Error != nil || count == 0 {
			return "产品分类无效"
		}
	}
	if publishing {
		if input.DeliveryProvince == "" {
			return "发布竞价前请填写收货地区"
		}
		if input.EndAt.Before(time.Now().Add(10*time.Minute)) || input.EndAt.After(time.Now().Add(30*24*time.Hour)) {
			return "截止时间须在 10 分钟至 30 天以内"
		}
	}
	return ""
}

func applyAuctionInput(row *model.ProcurementAuction, input auctionInput) {
	row.Title = input.Title
	row.CategoryID = input.CategoryID
	row.Specification = strings.TrimSpace(input.Specification)
	row.CompatibleModels = strings.TrimSpace(input.CompatibleModels)
	row.Quantity = input.Quantity
	row.Unit = input.Unit
	row.DeliveryProvince = input.DeliveryProvince
	row.DeliveryCity = input.DeliveryCity
	row.ExpectedDeliveryNote = strings.TrimSpace(input.ExpectedDeliveryNote)
	row.Description = strings.TrimSpace(input.Description)
	row.Image = strings.TrimSpace(input.Image)
	row.ImageAssetID = input.ImageAssetID
	row.MaxBudgetCents = input.MaxBudgetCents
	row.EndAt = input.EndAt
}

func (h AdminHandler) CreateOwnAuction(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) {
		return
	}
	if c.GetString("role") != "buyer" {
		Fail(c, 403, 403, "仅采购商可以创建竞价")
		return
	}
	var input auctionInput
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, 400, 400, "填写内容无效")
		return
	}
	if message := validateAuctionInput(h.DB, &input, false); message != "" {
		Fail(c, 400, 400, message)
		return
	}
	row := model.ProcurementAuction{BuyerUserID: c.GetUint("userId"), BuyerUsername: c.GetString("username"), Status: model.AuctionStatusDraft, Currency: "CNY", Version: 1}
	applyAuctionInput(&row, input)
	if h.DB.Create(&row).Error != nil {
		Fail(c, 500, 500, "竞价草稿保存失败")
		return
	}
	recordAuctionEvent(h.DB, row.ID, "created", c.GetUint("userId"), c.GetString("username"), nil)
	OK(c, row)
}

func (h AdminHandler) UpdateOwnAuction(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) || c.GetString("role") != "buyer" {
		return
	}
	var row model.ProcurementAuction
	if h.DB.Where("id = ? AND buyer_user_id = ?", c.Param("id"), c.GetUint("userId")).First(&row).Error != nil {
		Fail(c, 404, 404, "竞价单不存在")
		return
	}
	if row.Status != model.AuctionStatusDraft {
		Fail(c, 409, 409, "仅草稿可以修改")
		return
	}
	var input auctionInput
	if c.ShouldBindJSON(&input) != nil || input.ExpectedVersion != row.Version {
		Fail(c, 409, 409, "内容版本已变化，请刷新后重试")
		return
	}
	if message := validateAuctionInput(h.DB, &input, false); message != "" {
		Fail(c, 400, 400, message)
		return
	}
	applyAuctionInput(&row, input)
	row.Version++
	if h.DB.Save(&row).Error != nil {
		Fail(c, 500, 500, "保存失败")
		return
	}
	OK(c, row)
}

func (h AdminHandler) PublishOwnAuction(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) || c.GetString("role") != "buyer" {
		return
	}
	userID := c.GetUint("userId")
	var profile model.BuyerProfile
	if h.DB.Where("user_id = ?", userID).First(&profile).Error != nil || strings.TrimSpace(profile.ContactName) == "" || strings.TrimSpace(profile.Phone) == "" || strings.TrimSpace(profile.Province) == "" {
		Fail(c, 409, 409, "请先完善联系人、电话和所在地区")
		return
	}
	var row model.ProcurementAuction
	if h.DB.Where("id = ? AND buyer_user_id = ? AND status = ?", c.Param("id"), userID, model.AuctionStatusDraft).First(&row).Error != nil {
		Fail(c, 404, 404, "竞价草稿不存在")
		return
	}
	input := auctionInput{Title: row.Title, CategoryID: row.CategoryID, Quantity: row.Quantity, Unit: row.Unit, DeliveryProvince: row.DeliveryProvince, MaxBudgetCents: row.MaxBudgetCents, EndAt: row.EndAt}
	if message := validateAuctionInput(h.DB, &input, true); message != "" {
		Fail(c, 400, 400, message)
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if row.ImageAssetID != nil {
			var count int64
			if err := tx.Model(&model.MediaAsset{}).Where("id = ? AND owner_username = ? AND status IN ?", *row.ImageAssetID, c.GetString("username"), []string{"staged", "published"}).Count(&count).Error; err != nil || count != 1 {
				return errAuctionImage
			}
			now := time.Now()
			if err := tx.Model(&model.MediaAsset{}).Where("id = ?", *row.ImageAssetID).Updates(map[string]any{"status": "published", "published_at": &now}).Error; err != nil {
				return err
			}
		}
		result := tx.Model(&model.ProcurementAuction{}).Where("id = ? AND status = ?", row.ID, model.AuctionStatusDraft).Updates(map[string]any{"status": model.AuctionStatusOpen, "original_end_at": row.EndAt, "version": gorm.Expr("version + 1")})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errAuctionVersion
		}
		return nil
	})
	if err != nil {
		Fail(c, 409, 409, "竞价单已被其他操作更新")
		return
	}
	recordAuctionEvent(h.DB, row.ID, "published", userID, c.GetString("username"), gin.H{"endAt": row.EndAt})
	notifyEligibleVendors(h.DB, row.ID, "新采购竞价已发布", row.Title+" 已开始竞价")
	OK(c, gin.H{"status": model.AuctionStatusOpen})
}

func (h AdminHandler) ListPublicAuctions(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) {
		return
	}
	_ = AdvanceAuctions(h.DB)
	var rows []model.ProcurementAuction
	query := h.DB.Preload("Category").Where("status IN ?", []string{model.AuctionStatusOpen, model.AuctionStatusAwaiting, model.AuctionStatusAwarded, model.AuctionStatusUnawarded}).Order("end_at asc").Limit(100)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR specification LIKE ?", like, like)
	}
	if query.Find(&rows).Error != nil {
		Fail(c, 500, 500, "竞价列表加载失败")
		return
	}
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.auctionView(row, c, false))
	}
	OK(c, items)
}

func (h AdminHandler) PublicAuctionDetail(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) {
		return
	}
	_ = AdvanceAuctions(h.DB)
	var row model.ProcurementAuction
	if h.DB.Preload("Category").Where("id = ? AND status <> ?", c.Param("id"), model.AuctionStatusDraft).First(&row).Error != nil {
		Fail(c, 404, 404, "竞价单不存在")
		return
	}
	OK(c, h.auctionView(row, c, false))
}

func (h AdminHandler) ListOwnAuctions(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) || c.GetString("role") != "buyer" {
		return
	}
	_ = AdvanceAuctions(h.DB)
	var rows []model.ProcurementAuction
	if h.DB.Preload("Category").Where("buyer_user_id = ?", c.GetUint("userId")).Order("id desc").Find(&rows).Error != nil {
		Fail(c, 500, 500, "竞价单加载失败")
		return
	}
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.auctionView(row, c, true))
	}
	OK(c, items)
}

func (h AdminHandler) OwnAuctionDetail(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) || c.GetString("role") != "buyer" {
		return
	}
	_ = AdvanceAuctions(h.DB)
	var row model.ProcurementAuction
	if h.DB.Preload("Category").Where("id = ? AND buyer_user_id = ?", c.Param("id"), c.GetUint("userId")).First(&row).Error != nil {
		Fail(c, 404, 404, "竞价单不存在")
		return
	}
	var bids []model.AuctionBid
	h.DB.Preload("Vendor").Where("auction_id = ?", row.ID).Order("unit_price_cents asc, created_at asc, id asc").Find(&bids)
	latestIDs := make(map[uint]uint, len(bids))
	for _, bid := range bids {
		if bid.ID > latestIDs[bid.VendorID] {
			latestIDs[bid.VendorID] = bid.ID
		}
	}
	bidViews := make([]gin.H, 0, len(bids))
	for _, bid := range bids {
		view := fullBidDTO(bid)
		view["isLatest"] = latestIDs[bid.VendorID] == bid.ID
		bidViews = append(bidViews, view)
	}
	OK(c, gin.H{"auction": h.auctionView(row, c, true), "bids": bidViews})
}

func (h AdminHandler) PlaceAuctionBid(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) || c.GetString("role") != "vendor" {
		return
	}
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	var vendor model.Vendor
	if h.DB.Where("id = ? AND is_visible = ? AND publication_status = ? AND review_status IN ?", vendorID, true, "published", []string{"verified", "approved"}).First(&vendor).Error != nil {
		Fail(c, 403, 403, "只有已审核并公开的厂商可以报价")
		return
	}
	var input bidInput
	if c.ShouldBindJSON(&input) != nil || input.UnitPriceCents <= 0 || input.DeliveryDays < 0 {
		Fail(c, 400, 400, "报价内容无效")
		return
	}
	var newBid model.AuctionBid
	var extended bool
	var becameLeader bool
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var row model.ProcurementAuction
		if err := tx.Where("id = ?", c.Param("id")).First(&row).Error; err != nil {
			return err
		}
		if row.Status == model.AuctionStatusOpen && !row.EndAt.After(time.Now()) {
			return errAuctionEnded
		}
		if row.Status != model.AuctionStatusOpen {
			return errAuctionNotOpen
		}
		if input.ExpectedVersion != row.Version {
			return errAuctionVersion
		}
		if row.Quantity <= 0 || input.UnitPriceCents > math.MaxInt64/row.Quantity {
			return errAuctionBidInvalid
		}
		if row.MaxBudgetCents > 0 && input.UnitPriceCents*row.Quantity > row.MaxBudgetCents {
			return errAuctionBudget
		}
		var previous model.AuctionBid
		if err := tx.Where("auction_id = ? AND vendor_id = ?", row.ID, vendorID).Order("id desc").First(&previous).Error; err == nil && input.UnitPriceCents >= previous.UnitPriceCents {
			return errAuctionMustLower
		}
		leadingBefore, _, _ := latestAuctionBidStats(tx, row.ID)
		newBid = model.AuctionBid{AuctionID: row.ID, VendorID: vendorID, UnitPriceCents: input.UnitPriceCents, TotalPriceCents: input.UnitPriceCents * row.Quantity, TaxIncluded: input.TaxIncluded, FreightNote: strings.TrimSpace(input.FreightNote), DeliveryDays: input.DeliveryDays, SupplyNote: strings.TrimSpace(input.SupplyNote), PromiseText: strings.TrimSpace(input.PromiseText), SubmittedBy: c.GetString("username")}
		if err := tx.Create(&newBid).Error; err != nil {
			return err
		}
		newLow := leadingBefore == 0 || input.UnitPriceCents < leadingBefore
		becameLeader = newLow
		updates := map[string]any{"version": gorm.Expr("version + 1")}
		if newLow && time.Until(row.EndAt) <= 5*time.Minute && row.ExtensionMinutes < 30 {
			updates["end_at"] = row.EndAt.Add(5 * time.Minute)
			updates["extension_minutes"] = row.ExtensionMinutes + 5
			extended = true
		}
		result := tx.Model(&model.ProcurementAuction{}).Where("id = ? AND status = ? AND version = ?", row.ID, model.AuctionStatusOpen, input.ExpectedVersion).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errAuctionVersion
		}
		details, _ := json.Marshal(gin.H{"bidId": newBid.ID, "unitPriceCents": newBid.UnitPriceCents})
		if err := tx.Create(&model.AuctionEvent{AuctionID: row.ID, EventType: "bid", ActorID: c.GetUint("userId"), ActorName: c.GetString("username"), Details: string(details)}).Error; err != nil {
			return err
		}
		if extended {
			return tx.Create(&model.AuctionEvent{AuctionID: row.ID, EventType: "extended", Details: "{\"minutes\":5}"}).Error
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			Fail(c, 404, 404, "竞价单不存在")
		case errors.Is(err, errAuctionEnded), errors.Is(err, errAuctionNotOpen):
			Fail(c, 409, 409, "竞价已截止或不可报价")
		case errors.Is(err, errAuctionVersion):
			Fail(c, 409, 409, "竞价状态已变化，请刷新后重试")
		case errors.Is(err, errAuctionMustLower):
			Fail(c, 409, 409, "新报价必须低于本厂上一次报价")
		case errors.Is(err, errAuctionBudget):
			Fail(c, 400, 400, "报价总额超过采购方最高预算")
		default:
			Fail(c, 400, 400, "报价提交失败")
		}
		return
	}
	var row model.ProcurementAuction
	h.DB.First(&row, newBid.AuctionID)
	notifyUser(h.DB, row.BuyerUserID, "auction", row.ID, "竞价收到新报价", row.Title+" 收到新的厂商报价")
	if becameLeader {
		notifyAuctionVendors(h.DB, row.ID, "竞价出现新低价", row.Title+" 出现新低价，您的排名可能发生变化")
	}
	if extended {
		notifyAuctionVendors(h.DB, row.ID, "竞价自动延时", row.Title+" 因最后五分钟出现新低价，截止时间自动延长 5 分钟")
	}
	OK(c, gin.H{"bid": anonymousBidDTO(newBid), "auctionVersion": row.Version, "endAt": row.EndAt, "extended": extended})
}

func (h AdminHandler) CancelOwnAuction(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) || c.GetString("role") != "buyer" {
		return
	}
	var input struct{ Reason string }
	_ = c.ShouldBindJSON(&input)
	input.Reason = strings.TrimSpace(input.Reason)
	var row model.ProcurementAuction
	if h.DB.Where("id = ? AND buyer_user_id = ?", c.Param("id"), c.GetUint("userId")).First(&row).Error != nil {
		Fail(c, 404, 404, "竞价单不存在")
		return
	}
	if row.Status != model.AuctionStatusDraft && row.Status != model.AuctionStatusOpen {
		Fail(c, 409, 409, "当前状态不能取消")
		return
	}
	var bids int64
	h.DB.Model(&model.AuctionBid{}).Where("auction_id = ?", row.ID).Count(&bids)
	if bids > 0 && input.Reason == "" {
		Fail(c, 400, 400, "已有报价，取消时必须填写原因")
		return
	}
	result := h.DB.Model(&model.ProcurementAuction{}).Where("id = ? AND status = ?", row.ID, row.Status).Updates(map[string]any{"status": model.AuctionStatusCancelled, "cancel_reason": input.Reason, "version": gorm.Expr("version + 1")})
	if result.RowsAffected != 1 {
		Fail(c, 409, 409, "状态已变化")
		return
	}
	recordAuctionEvent(h.DB, row.ID, "cancelled", c.GetUint("userId"), c.GetString("username"), gin.H{"reason": input.Reason})
	notifyAuctionVendors(h.DB, row.ID, "竞价已取消", row.Title+" 已取消："+input.Reason)
	OK(c, gin.H{"status": model.AuctionStatusCancelled})
}

func (h AdminHandler) AwardOwnAuction(c *gin.Context) {
	if !requireAuctionFeature(c, h.DB) || c.GetString("role") != "buyer" {
		return
	}
	var input struct{ BidID uint }
	if c.ShouldBindJSON(&input) != nil || input.BidID == 0 {
		Fail(c, 400, 400, "请选择中标报价")
		return
	}
	var row model.ProcurementAuction
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND buyer_user_id = ?", c.Param("id"), c.GetUint("userId")).First(&row).Error; err != nil {
			return err
		}
		if row.Status != model.AuctionStatusAwaiting {
			return errAuctionNotAwardable
		}
		var bid model.AuctionBid
		if err := tx.Where("id = ? AND auction_id = ?", input.BidID, row.ID).First(&bid).Error; err != nil {
			return err
		}
		var latestBid model.AuctionBid
		if err := tx.Where("auction_id = ? AND vendor_id = ?", row.ID, bid.VendorID).Order("id desc").First(&latestBid).Error; err != nil || latestBid.ID != bid.ID {
			return errAuctionBidSuperseded
		}
		now := time.Now()
		result := tx.Model(&model.ProcurementAuction{}).Where("id = ? AND status = ?", row.ID, model.AuctionStatusAwaiting).Updates(map[string]any{"status": model.AuctionStatusAwarded, "awarded_bid_id": bid.ID, "awarded_at": &now, "version": gorm.Expr("version + 1")})
		if result.RowsAffected != 1 {
			return errAuctionVersion
		}
		return tx.Model(&model.AuctionBid{}).Where("id = ?", bid.ID).Update("is_winning", true).Error
	})
	if err != nil {
		Fail(c, 409, 409, "定标失败、状态不允许或已被处理")
		return
	}
	recordAuctionEvent(h.DB, row.ID, "awarded", c.GetUint("userId"), c.GetString("username"), gin.H{"bidId": input.BidID})
	notifyAuctionResult(h.DB, row.ID, input.BidID, row.Title)
	OK(c, gin.H{"status": model.AuctionStatusAwarded, "awardedBidId": input.BidID})
}

func (h AdminHandler) AdminListAuctions(c *gin.Context) {
	_ = AdvanceAuctions(h.DB)
	var rows []model.ProcurementAuction
	query := h.DB.Preload("Category").Order("id desc")
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if query.Limit(200).Find(&rows).Error != nil {
		Fail(c, 500, 500, "竞价加载失败")
		return
	}
	OK(c, rows)
}

func (h AdminHandler) AdminAuctionDetail(c *gin.Context) {
	_ = AdvanceAuctions(h.DB)
	var row model.ProcurementAuction
	if h.DB.Preload("Category").First(&row, c.Param("id")).Error != nil {
		Fail(c, 404, 404, "竞价单不存在")
		return
	}
	var bids []model.AuctionBid
	var events []model.AuctionEvent
	h.DB.Preload("Vendor").Where("auction_id = ?", row.ID).Order("id desc").Find(&bids)
	h.DB.Where("auction_id = ?", row.ID).Order("id desc").Find(&events)
	bidViews := make([]gin.H, 0, len(bids))
	for _, bid := range bids {
		bidViews = append(bidViews, fullBidDTO(bid))
	}
	OK(c, gin.H{"auction": row, "bids": bidViews, "events": events})
}

func (h AdminHandler) AdminUpdateAuctionStatus(c *gin.Context) {
	var input struct{ Status, Reason string }
	if c.ShouldBindJSON(&input) != nil || (input.Status != model.AuctionStatusSuspended && input.Status != model.AuctionStatusCancelled && input.Status != model.AuctionStatusUnawarded) {
		Fail(c, 400, 400, "监管状态无效")
		return
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if input.Reason == "" {
		Fail(c, 400, 400, "请填写操作原因")
		return
	}
	var row model.ProcurementAuction
	if h.DB.First(&row, c.Param("id")).Error != nil {
		Fail(c, 404, 404, "竞价单不存在")
		return
	}
	if row.Status == model.AuctionStatusAwarded || row.Status == model.AuctionStatusCancelled {
		Fail(c, 409, 409, "竞价已结束，不能重复处理")
		return
	}
	if h.DB.Model(&row).Updates(map[string]any{"status": input.Status, "cancel_reason": input.Reason, "version": gorm.Expr("version + 1")}).Error != nil {
		Fail(c, 500, 500, "状态更新失败")
		return
	}
	recordAuctionEvent(h.DB, row.ID, "admin_"+input.Status, c.GetUint("userId"), c.GetString("username"), gin.H{"reason": input.Reason})
	notifyUser(h.DB, row.BuyerUserID, "auction", row.ID, "竞价状态变更", row.Title+"："+input.Reason)
	notifyAuctionVendors(h.DB, row.ID, "竞价状态变更", row.Title+"："+input.Reason)
	OK(c, gin.H{"status": input.Status})
}

func (h AdminHandler) auctionView(row model.ProcurementAuction, c *gin.Context, full bool) gin.H {
	leading, count, latest := latestAuctionBidStats(h.DB, row.ID)
	view := gin.H{"id": row.ID, "title": row.Title, "categoryId": row.CategoryID, "category": row.Category, "specification": row.Specification, "compatibleModels": row.CompatibleModels, "quantity": row.Quantity, "unit": row.Unit, "deliveryProvince": row.DeliveryProvince, "deliveryCity": row.DeliveryCity, "expectedDeliveryNote": row.ExpectedDeliveryNote, "description": row.Description, "image": row.Image, "maxBudgetCents": row.MaxBudgetCents, "currency": row.Currency, "status": row.Status, "endAt": row.EndAt, "originalEndAt": row.OriginalEndAt, "extensionMinutes": row.ExtensionMinutes, "bidCount": count, "leadingPriceCents": leading, "cancelReason": row.CancelReason, "version": row.Version, "createdAt": row.CreatedAt, "updatedAt": row.UpdatedAt}
	if vendorID, ok := vendorScope(c); ok {
		if bid, exists := latest[vendorID]; exists {
			view["myLatestBid"] = anonymousBidDTO(bid)
			view["myRank"] = rankForBid(latest, bid)
		}
	}
	if row.AwardedBidID != nil {
		var bid model.AuctionBid
		if h.DB.Preload("Vendor").First(&bid, *row.AwardedBidID).Error == nil {
			result := publicAwardBidDTO(bid)
			if full || c.GetUint("userId") == row.BuyerUserID || bid.VendorID == c.GetUint("vendorId") {
				result = fullBidDTO(bid)
			}
			view["awardedBid"] = result
		}
	}
	return view
}

func latestAuctionBidStats(db *gorm.DB, auctionID uint) (int64, int64, map[uint]model.AuctionBid) {
	var rows []model.AuctionBid
	db.Raw("SELECT b.* FROM auction_bids b JOIN (SELECT vendor_id, MAX(id) AS max_id FROM auction_bids WHERE auction_id = ? GROUP BY vendor_id) x ON x.max_id = b.id ORDER BY b.unit_price_cents ASC, b.created_at ASC, b.id ASC", auctionID).Scan(&rows)
	latest := make(map[uint]model.AuctionBid, len(rows))
	for _, row := range rows {
		latest[row.VendorID] = row
	}
	if len(rows) == 0 {
		return 0, 0, latest
	}
	return rows[0].UnitPriceCents, int64(len(rows)), latest
}

func rankForBid(rows map[uint]model.AuctionBid, target model.AuctionBid) int {
	rank := 1
	for _, row := range rows {
		if row.UnitPriceCents < target.UnitPriceCents || (row.UnitPriceCents == target.UnitPriceCents && (row.CreatedAt.Before(target.CreatedAt) || (row.CreatedAt.Equal(target.CreatedAt) && row.ID < target.ID))) {
			rank++
		}
	}
	return rank
}

func anonymousBidDTO(b model.AuctionBid) gin.H {
	return gin.H{"id": b.ID, "auctionId": b.AuctionID, "unitPriceCents": b.UnitPriceCents, "totalPriceCents": b.TotalPriceCents, "taxIncluded": b.TaxIncluded, "freightNote": b.FreightNote, "deliveryDays": b.DeliveryDays, "supplyNote": b.SupplyNote, "promiseText": b.PromiseText, "isWinning": b.IsWinning, "createdAt": b.CreatedAt}
}

func publicAwardBidDTO(b model.AuctionBid) gin.H {
	return gin.H{"unitPriceCents": b.UnitPriceCents, "totalPriceCents": b.TotalPriceCents, "createdAt": b.CreatedAt}
}

func fullBidDTO(b model.AuctionBid) gin.H {
	view := anonymousBidDTO(b)
	view["vendorId"] = b.VendorID
	if b.Vendor != nil {
		view["vendorName"] = b.Vendor.Name
	}
	return view
}

func recordAuctionEvent(db *gorm.DB, auctionID uint, event string, actorID uint, actor string, details any) {
	raw, _ := json.Marshal(details)
	db.Create(&model.AuctionEvent{AuctionID: auctionID, EventType: event, ActorID: actorID, ActorName: actor, Details: string(raw)})
}

func notifyUser(db *gorm.DB, userID uint, business string, businessID uint, title, content string) {
	if userID > 0 {
		db.Create(&model.UserNotification{UserID: userID, BusinessType: business, BusinessID: businessID, Title: title, Content: content})
	}
}

func notifyAuctionVendors(db *gorm.DB, auctionID uint, title, content string) {
	var ids []uint
	db.Model(&model.AdminUser{}).Joins("JOIN auction_bids ON auction_bids.vendor_id = admin_users.vendor_id").Where("auction_bids.auction_id = ? AND admin_users.is_enabled = ?", auctionID, true).Distinct("admin_users.id").Pluck("admin_users.id", &ids)
	for _, id := range ids {
		notifyUser(db, id, "auction", auctionID, title, content)
	}
}

func notifyEligibleVendors(db *gorm.DB, auctionID uint, title, content string) {
	var ids []uint
	db.Model(&model.AdminUser{}).
		Joins("JOIN vendors ON vendors.id = admin_users.vendor_id").
		Where("admin_users.role = ? AND admin_users.is_enabled = ? AND vendors.is_visible = ? AND vendors.publication_status = ? AND vendors.review_status IN ?", "vendor", true, true, "published", []string{"verified", "approved"}).
		Pluck("admin_users.id", &ids)
	for _, id := range ids {
		notifyUser(db, id, "auction", auctionID, title, content)
	}
}

func notifyAuctionResult(db *gorm.DB, auctionID, bidID uint, title string) {
	var winning model.AuctionBid
	db.First(&winning, bidID)
	var users []model.AdminUser
	db.Joins("JOIN auction_bids ON auction_bids.vendor_id = admin_users.vendor_id").Where("auction_bids.auction_id = ? AND admin_users.is_enabled = ?", auctionID, true).Group("admin_users.id").Find(&users)
	for _, user := range users {
		message := "您参与的竞价未中标"
		if user.VendorID != nil && *user.VendorID == winning.VendorID {
			message = "恭喜，您的报价已中标"
		}
		notifyUser(db, user.ID, "auction", auctionID, "竞价定标结果", title+"："+message)
	}
}

func AdvanceAuctions(db *gorm.DB) error {
	now := time.Now()
	var endingSoon []model.ProcurementAuction
	if err := db.Where("status = ? AND end_at > ? AND end_at <= ?", model.AuctionStatusOpen, now, now.Add(30*time.Minute)).Find(&endingSoon).Error; err == nil {
		for _, row := range endingSoon {
			var count int64
			db.Model(&model.AuctionEvent{}).Where("auction_id = ? AND event_type = ?", row.ID, "ending_soon").Count(&count)
			if count == 0 {
				recordAuctionEvent(db, row.ID, "ending_soon", 0, "system", gin.H{"endAt": row.EndAt})
				notifyUser(db, row.BuyerUserID, "auction", row.ID, "竞价即将截止", row.Title+" 将在 30 分钟内截止")
				notifyAuctionVendors(db, row.ID, "竞价即将截止", row.Title+" 将在 30 分钟内截止")
			}
		}
	}
	var rows []model.ProcurementAuction
	if err := db.Where("status = ? AND end_at <= ?", model.AuctionStatusOpen, now).Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		nextStatus := model.AuctionStatusAwaiting
		var bidCount int64
		db.Model(&model.AuctionBid{}).Where("auction_id = ?", row.ID).Count(&bidCount)
		if bidCount == 0 {
			nextStatus = model.AuctionStatusUnawarded
		}
		result := db.Model(&model.ProcurementAuction{}).Where("id = ? AND status = ? AND end_at <= ?", row.ID, model.AuctionStatusOpen, now).Updates(map[string]any{"status": nextStatus, "version": gorm.Expr("version + 1")})
		if result.RowsAffected == 1 {
			recordAuctionEvent(db, row.ID, "closed", 0, "system", nil)
			if nextStatus == model.AuctionStatusUnawarded {
				notifyUser(db, row.BuyerUserID, "auction", row.ID, "竞价未成交", row.Title+" 截止时没有有效报价")
			} else {
				notifyUser(db, row.BuyerUserID, "auction", row.ID, "竞价已截止待定标", row.Title+" 已截止，请选择中标报价")
				notifyAuctionVendors(db, row.ID, "竞价已截止", row.Title+" 已截止，等待采购方定标")
			}
		}
	}
	return nil
}

var (
	errAuctionEnded         = errors.New("auction ended")
	errAuctionNotOpen       = errors.New("auction not open")
	errAuctionVersion       = errors.New("auction version conflict")
	errAuctionBidInvalid    = errors.New("invalid bid")
	errAuctionMustLower     = errors.New("bid must lower")
	errAuctionBudget        = errors.New("over budget")
	errAuctionNotAwardable  = errors.New("not awardable")
	errAuctionBidSuperseded = errors.New("bid superseded")
	errAuctionImage         = errors.New("invalid auction image")
)
