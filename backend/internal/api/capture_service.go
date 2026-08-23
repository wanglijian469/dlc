package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/gorm"
)

type CaptureService struct {
	db       *gorm.DB
	config   config.Config
	provider CaptureAIProvider
	wake     chan struct{}
	once     sync.Once
	mu       sync.Mutex
}

func NewCaptureService(db *gorm.DB, cfg config.Config) *CaptureService {
	return &CaptureService{db: db, config: cfg, wake: make(chan struct{}, 1)}
}

func (s *CaptureService) Start() {
	if s == nil || s.db == nil {
		return
	}
	s.once.Do(func() {
		cutoff := time.Now().Add(-15 * time.Minute)
		s.db.Model(&model.CapturePackage{}).Where("status = ? AND started_at < ?", model.CaptureStatusProcessing, cutoff).
			Updates(map[string]any{"status": model.CaptureStatusQueued, "error_message": "服务重启后自动恢复", "started_at": nil})
		go s.loop()
	})
}

func (s *CaptureService) currentProvider() (CaptureAIProvider, config.Config, error) {
	if s == nil {
		return nil, config.Config{}, fmt.Errorf("智能识别服务不可用")
	}
	if s.provider != nil {
		return s.provider, s.config, nil
	}
	effective, err := effectiveCaptureAIConfig(s.db, s.config)
	if err != nil {
		return nil, effective, err
	}
	if !effective.CaptureAIEnabled || strings.TrimSpace(effective.TencentSecretID) == "" || strings.TrimSpace(effective.TencentSecretKey) == "" || strings.TrimSpace(effective.TencentTokenHubKey) == "" {
		return nil, effective, fmt.Errorf("智能识别尚未启用或 OCR/TokenHub 密钥未配置")
	}
	return NewTencentCaptureProvider(effective), effective, nil
}

func (s *CaptureService) Enabled() bool {
	provider, _, err := s.currentProvider()
	return err == nil && provider != nil
}

func (s *CaptureService) Notify() {
	if !s.Enabled() {
		return
	}
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *CaptureService) loop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.wake:
		case <-ticker.C:
		}
		for s.processNext() {
		}
	}
}

func (s *CaptureService) processNext() bool {
	if !s.Enabled() {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var item model.CapturePackage
	if s.db.Where("status = ?", model.CaptureStatusQueued).Order("queued_at asc, id asc").First(&item).Error != nil {
		return false
	}
	now := time.Now()
	result := s.db.Model(&model.CapturePackage{}).Where("id = ? AND status = ?", item.ID, model.CaptureStatusQueued).
		Updates(map[string]any{"status": model.CaptureStatusProcessing, "started_at": &now, "attempts": gorm.Expr("attempts + 1"), "error_message": ""})
	if result.RowsAffected != 1 {
		return true
	}
	if err := s.process(item.ID); err != nil {
		completed := time.Now()
		s.db.Model(&model.CapturePackage{}).Where("id = ?", item.ID).Updates(map[string]any{"status": model.CaptureStatusFailed, "error_message": err.Error(), "completed_at": &completed})
	}
	return true
}

func (s *CaptureService) process(packageID uint) error {
	var item model.CapturePackage
	if err := s.db.Preload("Documents.Asset").First(&item, packageID).Error; err != nil {
		return err
	}
	if len(item.Documents) == 0 {
		return fmt.Errorf("资料包中没有图片")
	}
	sort.Slice(item.Documents, func(i, j int) bool { return item.Documents[i].SortOrder < item.Documents[j].SortOrder })
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	provider, effective, err := s.currentProvider()
	if err != nil {
		return err
	}
	recognition, err := recognizeCaptureWithRetry(ctx, provider, item.Documents, time.Sleep)
	if err != nil {
		return err
	}
	matchCaptureDraft(s.db, &recognition.Draft)
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, document := range recognition.Documents {
			if err := tx.Model(&model.CaptureDocument{}).Where("id = ? AND capture_package_id = ?", document.DocumentID, packageID).Updates(map[string]any{
				"document_type": document.Type, "ocr_json": document.OCRJSON, "ocr_text": document.OCRText,
				"ocr_confidence": document.Confidence, "provider_request_id": document.RequestID,
			}).Error; err != nil {
				return err
			}
		}
		var oldCropAssetIDs []uint
		if err := tx.Model(&model.CaptureProductCrop{}).Where("capture_package_id = ?", packageID).Pluck("asset_id", &oldCropAssetIDs).Error; err != nil {
			return err
		}
		if err := tx.Where("capture_package_id = ?", packageID).Delete(&model.CaptureProductCrop{}).Error; err != nil {
			return err
		}
		if len(oldCropAssetIDs) > 0 {
			if err := tx.Model(&model.MediaAsset{}).Where("id IN ? AND purpose = ?", oldCropAssetIDs, "capture_crop").Update("status", "orphaned").Error; err != nil {
				return err
			}
		}
		if err := s.createCrops(tx, item, &recognition.Draft); err != nil {
			recognition.Draft.Warnings = append(recognition.Draft.Warnings, "部分产品候选图片裁切失败")
		}
		draftJSON, _ := json.Marshal(recognition.Draft)
		version := item.RecognitionVersion + 1
		captureResult := model.CaptureResult{CapturePackageID: packageID, Version: version, Provider: "tencent", Model: effective.CaptureVisionModel, ResultJSON: recognition.RawResult}
		if err := tx.Create(&captureResult).Error; err != nil {
			return err
		}
		completed := time.Now()
		status := model.CaptureStatusReady
		if len(recognition.Draft.unconfirmedFields()) > 0 {
			status = model.CaptureStatusNeedsReview
		}
		return tx.Model(&model.CapturePackage{}).Where("id = ?", packageID).Updates(map[string]any{
			"status": status, "draft_json": string(draftJSON), "recognition_version": version, "completed_at": &completed, "error_message": "",
		}).Error
	})
}

func recognizeCaptureWithRetry(ctx context.Context, provider CaptureAIProvider, documents []model.CaptureDocument, wait func(time.Duration)) (CaptureRecognition, error) {
	var recognition CaptureRecognition
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		recognition, err = provider.Recognize(ctx, documents)
		if err == nil {
			return recognition, nil
		}
		if attempt < 2 {
			wait(time.Duration(1<<attempt) * time.Second)
		}
	}
	return recognition, err
}

func (s *CaptureService) createCrops(tx *gorm.DB, item model.CapturePackage, draft *CaptureDraft) error {
	documents := map[uint]model.CaptureDocument{}
	for _, document := range item.Documents {
		documents[document.ID] = document
	}
	for _, suggestion := range draft.CropSuggestions {
		box := normalizeCaptureField(CaptureField{SourceBoxes: []CaptureSourceBox{suggestion.CaptureSourceBox}}).SourceBoxes
		if len(box) == 0 {
			continue
		}
		document, ok := documents[suggestion.DocumentID]
		if !ok {
			continue
		}
		asset, err := s.cropAsset(document.Asset, item, suggestion)
		if err != nil {
			continue
		}
		if err := tx.Create(&asset).Error; err != nil {
			_ = os.Remove(filepath.Join(s.config.MediaDir, asset.StorageKey))
			continue
		}
		crop := model.CaptureProductCrop{CapturePackageID: item.ID, DocumentID: document.ID, AssetID: asset.ID, ProductKey: suggestion.ProductKey, X: suggestion.X, Y: suggestion.Y, Width: suggestion.Width, Height: suggestion.Height}
		if err := tx.Create(&crop).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *CaptureService) cropAsset(source model.MediaAsset, item model.CapturePackage, suggestion CaptureCropSuggestion) (model.MediaAsset, error) {
	opened, err := os.Open(filepath.Join(s.config.MediaDir, source.StorageKey))
	if err != nil {
		return model.MediaAsset{}, err
	}
	img, _, err := image.Decode(opened)
	_ = opened.Close()
	if err != nil {
		return model.MediaAsset{}, err
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	x0, y0 := int(suggestion.X*float64(width)), int(suggestion.Y*float64(height))
	x1, y1 := int((suggestion.X+suggestion.Width)*float64(width)), int((suggestion.Y+suggestion.Height)*float64(height))
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > width {
		x1 = width
	}
	if y1 > height {
		y1 = height
	}
	if x1-x0 < 100 || y1-y0 < 100 {
		return model.MediaAsset{}, fmt.Errorf("crop too small")
	}
	targetImage := image.NewRGBA(image.Rect(0, 0, x1-x0, y1-y0))
	draw.Draw(targetImage, targetImage.Bounds(), img, image.Pt(bounds.Min.X+x0, bounds.Min.Y+y0), draw.Src)
	relativeDir := filepath.Join(time.Now().Format("2006"), time.Now().Format("01"), fmt.Sprintf("capture-%d", item.ID))
	if err := os.MkdirAll(filepath.Join(s.config.MediaDir, relativeDir), 0755); err != nil {
		return model.MediaAsset{}, err
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return model.MediaAsset{}, err
	}
	storageKey := filepath.Join(relativeDir, hex.EncodeToString(random)+".png")
	path := filepath.Join(s.config.MediaDir, storageKey)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return model.MediaAsset{}, err
	}
	if err := png.Encode(file, targetImage); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return model.MediaAsset{}, err
	}
	_ = file.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		_ = os.Remove(path)
		return model.MediaAsset{}, err
	}
	hash := sha256.Sum256(data)
	return model.MediaAsset{OwnerUsername: item.OwnerUsername, VendorID: item.VendorID, StorageKey: storageKey, OriginalName: "product-crop.png", MIME: "image/png", Size: int64(len(data)), Width: x1 - x0, Height: y1 - y0, SHA256: hex.EncodeToString(hash[:]), Purpose: "capture_crop", Status: "private"}, nil
}

func matchCaptureDraft(db *gorm.DB, draft *CaptureDraft) {
	name, phone, website := fieldValue(draft.VendorFields, "name"), fieldValue(draft.VendorFields, "phone"), fieldValue(draft.VendorFields, "websiteUrl")
	if name != "" || phone != "" || website != "" {
		var vendors []model.Vendor
		query := db.Model(&model.Vendor{})
		conditions, args := []string{}, []any{}
		if name != "" {
			conditions = append(conditions, "LOWER(REPLACE(name, ' ', '')) = ?")
			args = append(args, strings.ToLower(strings.ReplaceAll(name, " ", "")))
		}
		if phone != "" {
			conditions = append(conditions, "phone = ?")
			args = append(args, phone)
		}
		if website != "" {
			conditions = append(conditions, "website_url = ?")
			args = append(args, website)
		}
		query.Where(strings.Join(conditions, " OR "), args...).Limit(2).Find(&vendors)
		if len(vendors) == 1 {
			id := vendors[0].ID
			draft.VendorMatchID = &id
		}
	}
	for index := range draft.Products {
		name := fieldValue(draft.Products[index].Fields, "name")
		if name == "" {
			continue
		}
		var products []model.Product
		db.Where("LOWER(REPLACE(name, ' ', '')) = ?", strings.ToLower(strings.ReplaceAll(name, " ", ""))).Limit(2).Find(&products)
		if len(products) == 1 {
			id := products[0].ID
			draft.Products[index].TargetProductID = &id
		}
	}
}
