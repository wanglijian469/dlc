package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	stddraw "image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
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
	keyHash := sha256.Sum256([]byte(fmt.Sprintf("v2|%s|%d|%s|%s", asset.SHA256, cfg.WatermarkOpacity, cfg.WatermarkText, vendorLabel)))
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
	keyHash := sha256.Sum256([]byte(fmt.Sprintf("v2|%s|%d|%d|%s|%s", publicURL, info.Size(), info.ModTime().UnixNano(), cfg.WatermarkText, vendorLabel)))
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
	target, err := filepath.Abs(filepath.Join(root, filepath.Clean(storageKey)))
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
		label = "DALU PARTS"
	}
	scale := 1
	if canvas.Bounds().Dx() >= 600 {
		scale = 2
	}
	if canvas.Bounds().Dx() >= 1400 {
		scale = 3
	}
	label = fitWatermarkLabel(label, canvas.Bounds().Dx()/scale-24)
	textWidth := font.MeasureString(basicfont.Face7x13, label).Ceil()
	stamp := image.NewRGBA(image.Rect(0, 0, textWidth+20, 26))
	alpha := uint8(255 * opacity / 100)
	stddraw.Draw(stamp, stamp.Bounds(), &image.Uniform{C: color.RGBA{0, 0, 0, alpha}}, image.Point{}, stddraw.Src)
	drawer := font.Drawer{Dst: stamp, Src: image.NewUniform(color.RGBA{255, 255, 255, 235}), Face: basicfont.Face7x13, Dot: fixed.P(10, 18)}
	drawer.DrawString(label)
	targetWidth, targetHeight := stamp.Bounds().Dx()*scale, stamp.Bounds().Dy()*scale
	x := canvas.Bounds().Max.X - targetWidth - 12
	y := canvas.Bounds().Max.Y - targetHeight - 12
	if x < canvas.Bounds().Min.X {
		x = canvas.Bounds().Min.X
	}
	if y < canvas.Bounds().Min.Y {
		y = canvas.Bounds().Min.Y
	}
	xdraw.NearestNeighbor.Scale(canvas, image.Rect(x, y, x+targetWidth, y+targetHeight), stamp, stamp.Bounds(), stddraw.Over, nil)

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

func fitWatermarkLabel(label string, maxWidth int) string {
	if maxWidth < 35 {
		return "DALU"
	}
	runes := []rune(label)
	for len(runes) > 4 && font.MeasureString(basicfont.Face7x13, string(runes)).Ceil() > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return strings.TrimSpace(string(runes))
}

func watermarkLabel(platform, vendor string) string {
	parts := make([]string, 0, 2)
	for _, value := range []string{platform, vendor} {
		if value = asciiWatermarkText(value); value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, " | ")
}

func asciiWatermarkText(value string) string {
	var builder strings.Builder
	for _, r := range strings.TrimSpace(value) {
		if r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) || strings.ContainsRune("-_.", r)) {
			builder.WriteRune(r)
			continue
		}
		if unicode.Is(unicode.Han, r) {
			parts := pinyin.LazyPinyin(string(r), pinyin.NewArgs())
			if len(parts) > 0 {
				if builder.Len() > 0 && !strings.HasSuffix(builder.String(), " ") {
					builder.WriteByte(' ')
				}
				builder.WriteString(parts[0])
			}
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
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
