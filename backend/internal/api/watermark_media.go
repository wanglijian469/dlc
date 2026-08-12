package api

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	stddraw "image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
)

//go:embed assets/NotoSansCJKsc-Regular.otf
var watermarkFontData []byte

var (
	watermarkFontOnce sync.Once
	watermarkFont     *opentype.Font
	watermarkFontErr  error
)

type WatermarkService struct {
	db         *gorm.DB
	appCfg     config.Config
	protection *AccessProtectionService
	mu         sync.Mutex
	running    bool
}

func NewWatermarkService(db *gorm.DB, appCfg config.Config, protection *AccessProtectionService) *WatermarkService {
	return &WatermarkService{db: db, appCfg: appCfg, protection: protection}
}

// PublicPath returns an original only when the asset type is intentionally
// exempt (logo/certificate/banner) or watermarking is disabled.
func (s *WatermarkService) PublicPath(asset *model.MediaAsset) (string, string, error) {
	source, err := safeMediaPath(s.appCfg.MediaDir, asset.StorageKey)
	if err != nil {
		return "", "", err
	}
	cfg := DefaultProtectionConfig()
	if s.protection != nil {
		cfg = s.protection.Config()
	}
	eligible, vendorLabel := s.watermarkContext(asset.ID)
	if !cfg.WatermarkEnabled || !eligible {
		return source, asset.MIME, nil
	}
	keyHash := sha256.Sum256([]byte(fmt.Sprintf("v3|%s|%d|%s|%s", asset.SHA256, cfg.WatermarkOpacity, cfg.WatermarkText, vendorLabel)))
	extension := ".png"
	outputMIME := "image/png"
	if asset.MIME == "image/jpeg" {
		extension = ".jpg"
		outputMIME = "image/jpeg"
	}
	storageKey := filepath.Join("public-watermarks", fmt.Sprintf("%d-%s%s", asset.ID, hex.EncodeToString(keyHash[:8]), extension))
	target, err := safeMediaPath(s.appCfg.MediaDir, storageKey)
	if err != nil {
		return "", "", err
	}
	if _, err := os.Stat(target); err == nil {
		if asset.PublicStorageKey != storageKey {
			s.db.Model(asset).Update("public_storage_key", storageKey)
		}
		return target, outputMIME, nil
	}
	if err := generateWatermarkedImage(source, target, watermarkLabel(cfg.WatermarkText, vendorLabel), cfg.WatermarkOpacity, outputMIME); err != nil {
		return "", "", err
	}
	if err := s.db.Model(asset).Update("public_storage_key", storageKey).Error; err != nil {
		return "", "", err
	}
	asset.PublicStorageKey = storageKey
	return target, outputMIME, nil
}

func (s *WatermarkService) watermarkContext(assetID uint) (bool, string) {
	if s == nil || s.db == nil {
		return false, ""
	}
	return s.watermarkContextForURL(fmt.Sprintf("/api/media/%d", assetID), assetID)
}

func (s *WatermarkService) watermarkContextForURL(url string, assetID uint) (bool, string) {
	var excluded int64
	s.db.Model(&model.Vendor{}).Where("logo = ? OR cover_image = ? OR wechat_qr_code = ? OR logo_asset_id = ? OR cover_asset_id = ? OR wechat_qr_code_asset_id = ?", url, url, url, assetID, assetID, assetID).Count(&excluded)
	if excluded == 0 && assetID > 0 {
		s.db.Model(&model.VendorMedia{}).Where("asset_id = ? AND kind = ?", assetID, "certificate").Count(&excluded)
	}
	if excluded == 0 {
		s.db.Model(&model.VendorMedia{}).Where("url = ? AND kind = ?", url, "certificate").Count(&excluded)
	}
	if excluded == 0 {
		s.db.Model(&model.Banner{}).Where("background_image = ?", url).Count(&excluded)
	}
	if excluded == 0 {
		s.db.Model(&model.SiteConfig{}).Where("config_value LIKE ?", "%"+url+"%").Count(&excluded)
	}
	var vendor model.Vendor
	err := s.db.Joins("JOIN vendor_media vm ON vm.vendor_id = vendors.id").
		Where("(vm.asset_id = ? OR vm.url = ?) AND vm.kind IN ?", assetID, url, []string{"factory", "equipment"}).
		Order("vendors.id").First(&vendor).Error
	if err == nil {
		return true, preferredVendorLabel(vendor)
	}
	var product model.Product
	if s.db.Where("image = ? OR gallery LIKE ? OR specs LIKE ?", url, "%"+url+"%", "%"+url+"%").First(&product).Error == nil {
		if s.db.Joins("JOIN product_suppliers ps ON ps.vendor_id = vendors.id").
			Where("ps.product_id = ? AND ps.status = ?", product.ID, "approved").
			Order("vendors.id").First(&vendor).Error == nil {
			return true, preferredVendorLabel(vendor)
		}
		return true, ""
	}
	var supplier model.ProductSupplier
	if s.db.Where("status = ? AND (image = ? OR gallery LIKE ?)", "approved", url, "%"+url+"%").First(&supplier).Error == nil {
		if s.db.First(&vendor, supplier.VendorID).Error == nil {
			return true, preferredVendorLabel(vendor)
		}
		return true, ""
	}
	if excluded > 0 {
		return false, ""
	}
	return false, ""
}

func (s *WatermarkService) PublicLegacyPath(source, publicURL string) (string, string, error) {
	cfg := DefaultProtectionConfig()
	if s.protection != nil {
		cfg = s.protection.Config()
	}
	eligible, vendorLabel := s.watermarkContextForURL(publicURL, 0)
	if !cfg.WatermarkEnabled || !eligible {
		return source, "", nil
	}
	info, err := os.Stat(source)
	if err != nil {
		return "", "", err
	}
	keyHash := sha256.Sum256([]byte(fmt.Sprintf("v3|%s|%d|%d|%s|%s", publicURL, info.Size(), info.ModTime().UnixNano(), cfg.WatermarkText, vendorLabel)))
	extension := strings.ToLower(filepath.Ext(source))
	outputMIME := "image/png"
	if extension == ".jpg" || extension == ".jpeg" {
		extension = ".jpg"
		outputMIME = "image/jpeg"
	} else {
		extension = ".png"
	}
	storageKey := filepath.Join("public-watermarks", "legacy-"+hex.EncodeToString(keyHash[:8])+extension)
	target, err := safeMediaPath(s.appCfg.MediaDir, storageKey)
	if err != nil {
		return "", "", err
	}
	if _, err := os.Stat(target); err == nil {
		return target, outputMIME, nil
	}
	if err := generateWatermarkedImage(source, target, watermarkLabel(cfg.WatermarkText, vendorLabel), cfg.WatermarkOpacity, outputMIME); err != nil {
		return "", "", err
	}
	return target, outputMIME, nil
}

func (s *WatermarkService) StartBuildAll() (model.WatermarkBuildJob, error) {
	if s.protection != nil && !s.protection.Config().WatermarkEnabled {
		return model.WatermarkBuildJob{}, fmt.Errorf("请先启用公开图片水印")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		var job model.WatermarkBuildJob
		if s.db.Where("status IN ?", []string{"queued", "running"}).Order("id desc").First(&job).Error == nil {
			return job, nil
		}
	}
	var total int64
	if err := s.db.Model(&model.MediaAsset{}).Where("status = ?", "published").Count(&total).Error; err != nil {
		return model.WatermarkBuildJob{}, err
	}
	job := model.WatermarkBuildJob{Status: "queued", Total: int(total)}
	if err := s.db.Create(&job).Error; err != nil {
		return job, err
	}
	s.running = true
	go s.runBuild(job.ID)
	return job, nil
}

func (s *WatermarkService) runBuild(jobID uint) {
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()
	now := time.Now()
	s.db.Model(&model.WatermarkBuildJob{}).Where("id = ?", jobID).Updates(map[string]any{"status": "running", "started_at": &now})
	var assets []model.MediaAsset
	if err := s.db.Where("status = ?", "published").Order("id").Find(&assets).Error; err != nil {
		finished := time.Now()
		s.db.Model(&model.WatermarkBuildJob{}).Where("id = ?", jobID).Updates(map[string]any{"status": "failed", "error": err.Error(), "finished_at": &finished})
		return
	}
	succeeded, skipped, failed := 0, 0, 0
	var errorsList []string
	for index := range assets {
		eligible, _ := s.watermarkContext(assets[index].ID)
		if !eligible {
			skipped++
		} else if _, _, err := s.PublicPath(&assets[index]); err != nil {
			failed++
			if len(errorsList) < 10 {
				errorsList = append(errorsList, fmt.Sprintf("#%d: %v", assets[index].ID, err))
			}
		} else {
			succeeded++
		}
		s.db.Model(&model.WatermarkBuildJob{}).Where("id = ?", jobID).Updates(map[string]any{
			"processed": index + 1, "succeeded": succeeded, "skipped": skipped, "failed": failed,
		})
	}
	finished := time.Now()
	status := "completed"
	if failed > 0 {
		status = "completed_with_error"
	}
	s.db.Model(&model.WatermarkBuildJob{}).Where("id = ?", jobID).Updates(map[string]any{
		"status": status, "error": strings.Join(errorsList, "\n"), "finished_at": &finished,
	})
}

func safeMediaPath(mediaDir, storageKey string) (string, error) {
	root, err := filepath.Abs(mediaDir)
	if err != nil {
		return "", err
	}
	normalizedKey := strings.ReplaceAll(strings.TrimSpace(storageKey), "\\", "/")
	cleanKey := path.Clean(normalizedKey)
	hasWindowsVolume := len(cleanKey) >= 2 && ((cleanKey[0] >= 'a' && cleanKey[0] <= 'z') || (cleanKey[0] >= 'A' && cleanKey[0] <= 'Z')) && cleanKey[1] == ':'
	if normalizedKey == "" || strings.ContainsRune(normalizedKey, '\x00') || path.IsAbs(cleanKey) || hasWindowsVolume || cleanKey == "." || cleanKey == ".." || strings.HasPrefix(cleanKey, "../") {
		return "", fmt.Errorf("媒体文件路径无效")
	}
	target, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(cleanKey)))
	if err != nil {
		return "", err
	}
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("媒体文件路径无效")
	}
	return target, nil
}

func generateWatermarkedImage(source, target, label string, opacity int, outputMIME string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	decoded, _, err := image.Decode(input)
	input.Close()
	if err != nil {
		return err
	}
	canvas := image.NewRGBA(decoded.Bounds())
	stddraw.Draw(canvas, canvas.Bounds(), decoded, decoded.Bounds().Min, stddraw.Src)
	if label == "" {
		label = "大陆农机配件"
	}

	width, height := canvas.Bounds().Dx(), canvas.Bounds().Dy()
	fontSize := float64(minInt(width, height)) * 0.04
	if fontSize < 14 {
		fontSize = 14
	}
	if fontSize > 42 {
		fontSize = 42
	}
	face, err := newWatermarkFace(fontSize)
	if err != nil {
		return err
	}
	defer face.Close()

	marginX := maxInt(6, width*2/100)
	marginY := maxInt(6, height*2/100)
	label = fitWatermarkLabel(face, label, width-marginX*2)
	textWidth := font.MeasureString(face, label).Ceil()
	metrics := face.Metrics()
	baseline := canvas.Bounds().Max.Y - marginY - metrics.Descent.Ceil()
	x := canvas.Bounds().Max.X - marginX - textWidth
	if x < canvas.Bounds().Min.X+marginX {
		x = canvas.Bounds().Min.X + marginX
	}

	alpha := uint8(255 * clampInt(opacity, 5, 90) / 100)
	outlineAlpha := uint8(minInt(255, int(alpha)+24))
	for _, offset := range []image.Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {1, 1}} {
		drawWatermarkText(canvas, face, label, x+offset.X, baseline+offset.Y, color.RGBA{0, 0, 0, outlineAlpha})
	}
	drawWatermarkText(canvas, face, label, x, baseline, color.RGBA{255, 255, 255, alpha})

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(target), ".watermark-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if outputMIME == "image/jpeg" {
		err = jpeg.Encode(temp, canvas, &jpeg.Options{Quality: 88})
	} else {
		err = png.Encode(temp, canvas)
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tempName, target)
}

func newWatermarkFace(size float64) (font.Face, error) {
	watermarkFontOnce.Do(func() {
		watermarkFont, watermarkFontErr = opentype.Parse(watermarkFontData)
	})
	if watermarkFontErr != nil {
		return nil, watermarkFontErr
	}
	return opentype.NewFace(watermarkFont, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
}

func drawWatermarkText(dst stddraw.Image, face font.Face, label string, x, baseline int, textColor color.RGBA) {
	drawer := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  fixed.P(x, baseline),
	}
	drawer.DrawString(label)
}

func fitWatermarkLabel(face font.Face, label string, maxWidth int) string {
	runes := []rune(label)
	for len(runes) > 1 && font.MeasureString(face, string(runes)).Ceil() > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return strings.TrimSpace(string(runes))
}

func watermarkLabel(_ string, vendor string) string {
	platform := "大陆农机配件"
	vendor = strings.TrimSpace(vendor)
	if vendor == "" {
		return platform
	}
	return platform + " · " + vendor
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(value, low, high int) int {
	return maxInt(low, minInt(high, value))
}

func preferredVendorLabel(vendor model.Vendor) string {
	if strings.TrimSpace(vendor.ShortName) != "" {
		return vendor.ShortName
	}
	return vendor.Name
}

func (h AdminHandler) CreateWatermarkBuildJob(c *gin.Context) {
	if h.Watermarks == nil {
		Fail(c, 503, 503, "图片水印服务不可用")
		return
	}
	job, err := h.Watermarks.StartBuildAll()
	if err != nil {
		Fail(c, 500, 500, "水印任务创建失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "generate-all", "media-watermarks", job.ID)
	OK(c, job)
}

func (h AdminHandler) WatermarkBuildJob(c *gin.Context) {
	var job model.WatermarkBuildJob
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		Fail(c, 404, 404, "水印任务不存在")
		return
	}
	OK(c, job)
}
