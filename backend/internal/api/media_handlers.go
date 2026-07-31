package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
)

func (h AdminHandler) SecureUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, "请选择上传文件")
		return
	}
	if file.Size > 5*1024*1024 {
		Fail(c, 400, 400, "图片不能超过 5MB")
		return
	}
	source, err := file.Open()
	if err != nil {
		Fail(c, 400, 400, "无法读取上传文件")
		return
	}
	defer source.Close()
	img, format, err := image.Decode(io.LimitReader(source, 5*1024*1024+1))
	if err != nil || (format != "jpeg" && format != "png" && format != "webp") {
		Fail(c, 400, 400, "文件必须是可正常解码的 JPEG、PNG 或 WebP 图片")
		return
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width < 1 || height < 1 || int64(width)*int64(height) > 40_000_000 {
		Fail(c, 400, 400, "图片尺寸无效或像素超过 4000 万")
		return
	}
	now := time.Now()
	ownerDir := "shared"
	if id := c.GetUint("vendorId"); id > 0 {
		ownerDir = fmt.Sprintf("vendor-%d", id)
	}
	relativeDir := filepath.Join(now.Format("2006"), now.Format("01"), ownerDir)
	storageDir := filepath.Join(h.Config.MediaDir, relativeDir)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		Fail(c, 500, 500, "创建媒体目录失败")
		return
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		Fail(c, 500, 500, "生成安全文件名失败")
		return
	}
	name := hex.EncodeToString(random) + ".png"
	storageKey := filepath.Join(relativeDir, name)
	path := filepath.Join(h.Config.MediaDir, storageKey)
	target, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		Fail(c, 500, 500, "保存上传文件失败")
		return
	}
	encodeErr := png.Encode(target, img) // re-encoding removes EXIF and active metadata
	closeErr := target.Close()
	if encodeErr != nil || closeErr != nil {
		_ = os.Remove(path)
		Fail(c, 500, 500, "保存上传文件失败")
		return
	}
	stored, err := os.Open(path)
	if err != nil {
		_ = os.Remove(path)
		Fail(c, 500, 500, "校验上传文件失败")
		return
	}
	hash := sha256.New()
	size, copyErr := io.Copy(hash, stored)
	_ = stored.Close()
	if copyErr != nil {
		_ = os.Remove(path)
		Fail(c, 500, 500, "校验上传文件失败")
		return
	}
	status := "published"
	if c.GetString("role") == "vendor" {
		status = "staged"
	}
	var vendorID *uint
	if id := c.GetUint("vendorId"); id > 0 {
		vendorID = &id
	}
	asset := model.MediaAsset{OwnerUsername: c.GetString("username"), VendorID: vendorID, StorageKey: storageKey, OriginalName: filepath.Base(file.Filename), MIME: "image/png", Size: size, Width: width, Height: height, SHA256: hex.EncodeToString(hash.Sum(nil)), AltText: strings.TrimSpace(c.PostForm("altText")), Caption: strings.TrimSpace(c.PostForm("caption")), Status: status}
	if err := h.DB.Create(&asset).Error; err != nil {
		_ = os.Remove(path)
		Fail(c, 500, 500, "登记上传文件失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "upload", "media-assets", asset.ID)
	OK(c, gin.H{"assetId": asset.ID, "status": asset.Status, "url": fmt.Sprintf("/api/media/%d", asset.ID), "previewUrl": fmt.Sprintf("/api/admin/media/%d", asset.ID), "width": width, "height": height, "size": size, "mime": asset.MIME, "sha256": asset.SHA256})
}

func (h AdminHandler) PreviewMedia(c *gin.Context) {
	asset, ok := h.authorizedAsset(c)
	if !ok {
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("Content-Type", asset.MIME)
	c.File(filepath.Join(h.Config.MediaDir, asset.StorageKey))
}

func (h AdminHandler) DeleteMedia(c *gin.Context) {
	asset, ok := h.authorizedAsset(c)
	if !ok {
		return
	}
	if asset.Status == "published" {
		Fail(c, 409, 409, "已发布图片不能直接删除")
		return
	}
	if err := h.DB.Model(&asset).Update("status", "orphaned").Error; err != nil {
		Fail(c, 500, 500, "删除图片失败")
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (h AdminHandler) UpdateMediaMetadata(c *gin.Context) {
	asset, ok := h.authorizedAsset(c)
	if !ok {
		return
	}
	var req struct {
		AltText string `json:"altText"`
		Caption string `json:"caption"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, http.StatusBadRequest, 400, "媒体说明格式不正确")
		return
	}
	req.AltText, req.Caption = strings.TrimSpace(req.AltText), strings.TrimSpace(req.Caption)
	if len([]rune(req.AltText)) > 255 || len([]rune(req.Caption)) > 500 {
		Fail(c, http.StatusBadRequest, 400, "图片替代文本或图注过长")
		return
	}
	if err := h.DB.Model(&asset).Updates(map[string]any{"alt_text": req.AltText, "caption": req.Caption}).Error; err != nil {
		Fail(c, 500, 500, "媒体说明保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "update-metadata", "media-assets", asset.ID)
	h.DB.First(&asset, asset.ID)
	OK(c, asset)
}

func (h AdminHandler) authorizedAsset(c *gin.Context) (model.MediaAsset, bool) {
	var asset model.MediaAsset
	if err := h.DB.First(&asset, c.Param("id")).Error; err != nil {
		Fail(c, 404, 404, "图片不存在")
		return asset, false
	}
	if c.GetString("role") != "admin" && asset.OwnerUsername != c.GetString("username") {
		vendorID := c.GetUint("vendorId")
		if vendorID == 0 || !h.assetBelongsToVendor(asset, vendorID) {
			Fail(c, 403, 403, "无权访问该图片")
			return asset, false
		}
	}
	return asset, true
}

func (h AdminHandler) assetBelongsToVendor(asset model.MediaAsset, vendorID uint) bool {
	if asset.VendorID != nil && *asset.VendorID == vendorID {
		return true
	}
	var references int64
	h.DB.Model(&model.Vendor{}).Where("id = ? AND (logo_asset_id = ? OR cover_asset_id = ? OR wechat_qr_code_asset_id = ?)", vendorID, asset.ID, asset.ID, asset.ID).Count(&references)
	if references > 0 {
		return true
	}
	h.DB.Model(&model.VendorMedia{}).Where("vendor_id = ? AND asset_id = ?", vendorID, asset.ID).Count(&references)
	if references > 0 {
		return true
	}
	url := fmt.Sprintf("/api/media/%d", asset.ID)
	h.DB.Model(&model.ProductSupplier{}).Where("vendor_id = ? AND (image = ? OR gallery LIKE ?)", vendorID, url, "%"+url+"%").Count(&references)
	if references > 0 {
		return true
	}
	h.DB.Model(&model.ProductSupplier{}).
		Joins("JOIN products ON products.id = product_suppliers.product_id").
		Where("product_suppliers.vendor_id = ? AND (products.image = ? OR products.gallery LIKE ?)", vendorID, url, "%"+url+"%").Count(&references)
	return references > 0
}

func (h AdminHandler) PublicMedia(c *gin.Context) {
	var asset model.MediaAsset
	if err := h.DB.First(&asset, c.Param("id")).Error; err != nil {
		Fail(c, 404, 404, "图片不存在")
		return
	}
	var references int64
	h.DB.Model(&model.Vendor{}).Where("(logo_asset_id = ? OR cover_asset_id = ?) AND is_visible = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)", asset.ID, asset.ID, true, "published", time.Now()).Count(&references)
	if references == 0 {
		h.DB.Model(&model.Vendor{}).Where("wechat_qr_code_asset_id = ? AND wechat_public = ? AND is_visible = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)", asset.ID, true, true, "published", time.Now()).Count(&references)
	}
	if references == 0 {
		h.DB.Model(&model.VendorMedia{}).Joins("JOIN vendors ON vendors.id = vendor_media.vendor_id").Where("vendor_media.asset_id = ? AND vendors.is_visible = ? AND vendors.publication_status = ?", asset.ID, true, "published").Count(&references)
	}
	url := fmt.Sprintf("/api/media/%d", asset.ID)
	if references == 0 {
		h.DB.Model(&model.Product{}).Where("products.publication_status = ? AND (products.published_at IS NULL OR products.published_at <= ?) AND (products.image = ? OR products.gallery LIKE ? OR products.specs LIKE ?) AND EXISTS (SELECT 1 FROM product_suppliers ps JOIN vendors v ON v.id = ps.vendor_id WHERE ps.product_id = products.id AND ps.status = 'approved' AND v.is_visible = 1 AND v.publication_status = 'published' AND (v.published_at IS NULL OR v.published_at <= ?))", "published", time.Now(), url, "%"+url+"%", "%"+url+"%", time.Now()).Count(&references)
		if references == 0 {
			h.DB.Model(&model.ProductSupplier{}).Joins("JOIN vendors ON vendors.id = product_suppliers.vendor_id").Where("product_suppliers.status = ? AND vendors.is_visible = ? AND vendors.publication_status = ? AND (product_suppliers.image = ? OR product_suppliers.gallery LIKE ?)", "approved", true, "published", url, "%"+url+"%").Count(&references)
		}
	}
	if references == 0 {
		h.DB.Model(&model.Banner{}).Where("is_enabled = ? AND background_image = ?", true, url).Count(&references)
	}
	if references == 0 {
		h.DB.Model(&model.SiteConfig{}).Where("config_key = ? AND config_value LIKE ?", "site.meta", "%"+url+"%").Count(&references)
	}
	if references == 0 {
		Fail(c, 404, 404, "图片不存在")
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	publicPath := filepath.Join(h.Config.MediaDir, asset.StorageKey)
	publicMIME := asset.MIME
	if h.Watermarks != nil {
		var err error
		publicPath, publicMIME, err = h.Watermarks.PublicPath(&asset)
		if err != nil {
			Fail(c, http.StatusServiceUnavailable, 503, "公开图片生成失败，请稍后重试")
			return
		}
	}
	c.Header("Content-Type", publicMIME)
	c.File(publicPath)
}
