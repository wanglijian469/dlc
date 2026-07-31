package api

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const dailyVendorContactLimit int64 = 50

var errDailyVendorContactLimit = errors.New("daily vendor contact limit reached")

type vendorContactResponse struct {
	VendorID        uint   `json:"vendorId"`
	Phone           string `json:"phone,omitempty"`
	Wechat          string `json:"wechat,omitempty"`
	WechatQRCodeURL string `json:"wechatQrCodeUrl,omitempty"`
	ContactName     string `json:"contactName,omitempty"`
}

func (h PublicHandler) VendorContact(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	vendor, ok := authorizeVendorContact(h.DB, c)
	if !ok {
		return
	}
	qrURL := ""
	if vendor.WechatQRCodeAssetID != nil {
		qrURL = fmt.Sprintf("/api/vendors/%d/contact-qr", vendor.ID)
	} else if vendor.WechatQRCode != "" {
		qrURL = vendor.WechatQRCode
	}
	logOperation(h.DB, c.GetString("username"), "view-contact", "vendors", vendor.ID)
	OK(c, vendorContactResponse{
		VendorID: vendor.ID, Phone: vendor.Phone, Wechat: vendor.Wechat,
		WechatQRCodeURL: qrURL, ContactName: vendor.ContactName,
	})
}

// VendorContactQRCode serves the original QR image only after the same
// authenticated, per-account contact access check used by VendorContact.
func (h AdminHandler) VendorContactQRCode(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	vendor, ok := authorizeVendorContact(h.DB, c)
	if !ok {
		return
	}
	if vendor.WechatQRCodeAssetID == nil {
		Fail(c, http.StatusNotFound, 404, "微信二维码不存在")
		return
	}
	var asset model.MediaAsset
	if err := h.DB.Where("id = ? AND status = ?", *vendor.WechatQRCodeAssetID, "published").First(&asset).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "微信二维码不存在")
		return
	}
	path, err := safeMediaPath(h.Config.MediaDir, asset.StorageKey)
	if err != nil {
		Fail(c, http.StatusNotFound, 404, "微信二维码不存在")
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="vendor-%d-wechat-qr.png"`, vendor.ID))
	c.Header("Content-Type", asset.MIME)
	c.File(filepath.Clean(path))
}

func authorizeVendorContact(db *gorm.DB, c *gin.Context) (model.Vendor, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, http.StatusBadRequest, 400, "厂商 ID 无效")
		return model.Vendor{}, false
	}
	var vendor model.Vendor
	if err := publishedVendorQuery(db).First(&vendor, uint(id)).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "厂商不存在")
		return model.Vendor{}, false
	}
	userID := c.GetUint("userId")
	if userID == 0 {
		Fail(c, http.StatusUnauthorized, 401, "请登录后查看完整联系方式")
		return model.Vendor{}, false
	}
	hasProtectedContact := (!vendor.PhonePublic && vendor.Phone != "") ||
		(!vendor.WechatPublic && (vendor.Wechat != "" || vendor.WechatQRCode != "")) ||
		(!vendor.ContactNamePublic && vendor.ContactName != "")
	exempt := !hasProtectedContact || c.GetString("role") == "admin" || (c.GetString("role") == "vendor" && c.GetUint("vendorId") == vendor.ID)
	if exempt {
		return vendor, true
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	resetAt := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	err = db.Transaction(func(tx *gorm.DB) error {
		// Lock the account row so concurrent requests cannot both consume
		// the last remaining distinct-vendor slot.
		var account model.AdminUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account, userID).Error; err != nil {
			return err
		}
		var existing model.ContactAccessLog
		err := tx.Where("user_id = ? AND vendor_id = ? AND access_date = ?", userID, vendor.ID, today).First(&existing).Error
		if err == nil {
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		var used int64
		if err := tx.Model(&model.ContactAccessLog{}).Where("user_id = ? AND access_date = ?", userID, today).Count(&used).Error; err != nil {
			return err
		}
		if used >= dailyVendorContactLimit {
			return errDailyVendorContactLimit
		}
		row := model.ContactAccessLog{UserID: userID, VendorID: vendor.ID, AccessDate: today}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
	})
	if errors.Is(err, errDailyVendorContactLimit) {
		retryAfter := int64(time.Until(resetAt).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
		c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "访问过于频繁，请稍后再试", "data": nil})
		return model.Vendor{}, false
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "联系方式访问记录保存失败")
		return model.Vendor{}, false
	}
	return vendor, true
}
