package api

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxRemoteImageSize = 5 * 1024 * 1024

// DownloadRemoteImage fetches an authorised public image and stores a
// normalized copy in the platform media store.
func (h AdminHandler) DownloadRemoteImage(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		Fail(c, http.StatusBadRequest, 400, "请求格式不正确")
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
	asset, err := h.downloadAndStoreRemoteImage(h.DB, request.URL, c.GetString("username"), vendorID, status)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, "下载图片失败："+err.Error())
		return
	}
	logOperation(h.DB, c.GetString("username"), "download", "media-assets", asset.ID)
	OK(c, mediaAssetResponse(asset))
}

func (h AdminHandler) downloadAndStoreRemoteImage(db *gorm.DB, rawURL, username string, vendorID *uint, status string) (model.MediaAsset, error) {
	data, originalName, err := downloadRemoteImage(rawURL)
	if err != nil {
		return model.MediaAsset{}, err
	}
	return h.storeRemoteImage(db, data, originalName, username, vendorID, status)
}

func (h AdminHandler) storeRemoteImage(db *gorm.DB, data []byte, originalName, username string, vendorID *uint, status string) (model.MediaAsset, error) {
	if len(data) == 0 || len(data) > maxRemoteImageSize {
		return model.MediaAsset{}, fmt.Errorf("图片大小必须在 5MB 以内")
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png" && format != "webp") {
		return model.MediaAsset{}, fmt.Errorf("文件必须是可正常解码的 JPEG、PNG 或 WebP 图片")
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width < 1 || height < 1 || int64(width)*int64(height) > 40_000_000 {
		return model.MediaAsset{}, fmt.Errorf("图片尺寸无效或像素超过 4000 万")
	}

	now := time.Now()
	ownerDir := "shared"
	if vendorID != nil && *vendorID > 0 {
		ownerDir = fmt.Sprintf("vendor-%d", *vendorID)
	}
	relativeDir := filepath.Join(now.Format("2006"), now.Format("01"), ownerDir)
	storageDir := filepath.Join(h.Config.MediaDir, relativeDir)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return model.MediaAsset{}, fmt.Errorf("创建媒体目录失败")
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return model.MediaAsset{}, fmt.Errorf("生成安全文件名失败")
	}
	name := hex.EncodeToString(random) + ".png"
	storageKey := filepath.Join(relativeDir, name)
	filePath := filepath.Join(h.Config.MediaDir, storageKey)
	target, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return model.MediaAsset{}, fmt.Errorf("保存图片失败")
	}
	encodeErr := png.Encode(target, img)
	closeErr := target.Close()
	if encodeErr != nil || closeErr != nil {
		_ = os.Remove(filePath)
		return model.MediaAsset{}, fmt.Errorf("保存图片失败")
	}
	stored, err := os.Open(filePath)
	if err != nil {
		_ = os.Remove(filePath)
		return model.MediaAsset{}, fmt.Errorf("校验图片失败")
	}
	hash := sha256.New()
	size, copyErr := io.Copy(hash, stored)
	_ = stored.Close()
	if copyErr != nil {
		_ = os.Remove(filePath)
		return model.MediaAsset{}, fmt.Errorf("校验图片失败")
	}
	asset := model.MediaAsset{
		OwnerUsername: username,
		VendorID:      vendorID,
		StorageKey:    storageKey,
		OriginalName:  originalName,
		MIME:          "image/png",
		Size:          size,
		Width:         width,
		Height:        height,
		SHA256:        hex.EncodeToString(hash.Sum(nil)),
		Status:        status,
	}
	if status == "published" {
		asset.PublishedAt = &now
	}
	if err := db.Create(&asset).Error; err != nil {
		_ = os.Remove(filePath)
		return model.MediaAsset{}, fmt.Errorf("登记图片资源失败")
	}
	return asset, nil
}

func mediaAssetResponse(asset model.MediaAsset) gin.H {
	return gin.H{
		"assetId":    asset.ID,
		"status":     asset.Status,
		"url":        fmt.Sprintf("/api/media/%d", asset.ID),
		"previewUrl": fmt.Sprintf("/api/admin/media/%d", asset.ID),
		"width":      asset.Width,
		"height":     asset.Height,
		"size":       asset.Size,
		"mime":       asset.MIME,
		"sha256":     asset.SHA256,
	}
}

func downloadRemoteImage(rawURL string) ([]byte, string, error) {
	parsed, err := validateRemoteImageURL(rawURL)
	if err != nil {
		return nil, "", err
	}
	if err := ensurePublicRemoteHost(parsed); err != nil {
		return nil, "", err
	}
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("重定向次数过多")
			}
			redirectURL, err := validateRemoteImageURL(request.URL.String())
			if err != nil {
				return err
			}
			return ensurePublicRemoteHost(redirectURL)
		},
	}
	request, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("图片地址无效")
	}
	request.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/*;q=0.8")
	request.Header.Set("User-Agent", "DaluNongjiPartsMediaFetcher/1.0")
	response, err := client.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("无法访问图片地址（可能是防盗链、TLS 证书异常或站点不可访问）")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, "", fmt.Errorf("图片地址返回 HTTP %d", response.StatusCode)
	}
	if response.ContentLength > maxRemoteImageSize {
		return nil, "", fmt.Errorf("图片超过 5MB")
	}
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]))
	if contentType != "" && !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("远端返回的不是图片")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxRemoteImageSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("读取远端图片失败")
	}
	if len(data) > maxRemoteImageSize {
		return nil, "", fmt.Errorf("图片超过 5MB")
	}
	name := path.Base(parsed.Path)
	if name == "." || name == "/" || name == "" {
		name = "remote-image"
	}
	return data, name, nil
}

func validateRemoteImageURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil {
		return nil, fmt.Errorf("仅支持公开的 HTTP 或 HTTPS 图片地址")
	}
	if ip := net.ParseIP(parsed.Hostname()); ip != nil && !isPublicIPAddress(ip) {
		return nil, fmt.Errorf("不允许访问内网图片地址")
	}
	return parsed, nil
}

func ensurePublicRemoteHost(parsed *url.URL) error {
	if ip := net.ParseIP(parsed.Hostname()); ip != nil {
		if !isPublicIPAddress(ip) {
			return fmt.Errorf("不允许访问内网图片地址")
		}
		return nil
	}
	addresses, err := net.LookupIP(parsed.Hostname())
	if err != nil || len(addresses) == 0 {
		return fmt.Errorf("无法解析图片服务器地址")
	}
	for _, address := range addresses {
		if !isPublicIPAddress(address) {
			return fmt.Errorf("不允许访问内网图片地址")
		}
	}
	return nil
}

func isPublicIPAddress(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsUnspecified() && !ip.IsMulticast()
}

func (h AdminHandler) localizeImportImage(db *gorm.DB, value, username string, vendorID *uint, status string) (string, *uint, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return value, nil, nil
	}
	asset, err := h.downloadAndStoreRemoteImage(db, value, username, vendorID, status)
	if err != nil {
		return value, nil, err
	}
	return fmt.Sprintf("/api/media/%d", asset.ID), &asset.ID, nil
}

func vendorIDPointer(id uint) *uint {
	if id == 0 {
		return nil
	}
	return &id
}

func (h AdminHandler) localizeImportImageList(db *gorm.DB, value, username string, vendorID *uint, status string) ([]string, []error) {
	items := splitImportList(value)
	errors := make([]error, 0)
	for index, item := range items {
		localized, _, err := h.localizeImportImage(db, item, username, vendorID, status)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		items[index] = localized
	}
	return items, errors
}
