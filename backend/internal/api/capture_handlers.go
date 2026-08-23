package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxCaptureDocuments = 20

func (h AdminHandler) ListCapturePackages(c *gin.Context) {
	query := h.DB.Model(&model.CapturePackage{}).Preload("Vendor").Order("updated_at desc, id desc")
	if c.GetString("role") != "admin" {
		query = query.Where("owner_username = ? AND vendor_id = ?", c.GetString("username"), c.GetUint("vendorId"))
	}
	var rows []model.CapturePackage
	if err := query.Find(&rows).Error; err != nil {
		Fail(c, 500, 500, "资料包列表加载失败")
		return
	}
	OK(c, rows)
}

func (h AdminHandler) CreateCapturePackage(c *gin.Context) {
	var req struct {
		Title    string `json:"title"`
		VendorID *uint  `json:"vendorId"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, 400, 400, "资料包内容格式不正确")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || len([]rune(req.Title)) > 150 {
		Fail(c, 400, 400, "请填写 150 字以内的资料包名称")
		return
	}
	if c.GetString("role") == "vendor" {
		id := c.GetUint("vendorId")
		if id == 0 {
			Fail(c, 403, 403, "厂商账号未绑定公司")
			return
		}
		req.VendorID = &id
	}
	if req.VendorID != nil && !recordExists[model.Vendor](h.DB, *req.VendorID) {
		Fail(c, 400, 400, "关联厂商不存在")
		return
	}
	item := model.CapturePackage{Title: req.Title, OwnerUsername: c.GetString("username"), OwnerRole: c.GetString("role"), VendorID: req.VendorID, Status: model.CaptureStatusUploading}
	if err := h.DB.Create(&item).Error; err != nil {
		Fail(c, 500, 500, "资料包创建失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "create", "capture-packages", item.ID)
	h.writeCapturePackage(c, item.ID)
}

func (h AdminHandler) GetCapturePackage(c *gin.Context) {
	item, ok := h.authorizedCapturePackage(c, c.Param("id"))
	if !ok {
		return
	}
	h.writeCapturePackage(c, item.ID)
}

func (h AdminHandler) DeleteCapturePackage(c *gin.Context) {
	item, ok := h.authorizedCapturePackage(c, c.Param("id"))
	if !ok {
		return
	}
	if item.Status == model.CaptureStatusProcessing || item.Status == model.CaptureStatusCommitted {
		Fail(c, 409, 409, "处理中或已提交的资料包不能删除")
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var assetIDs []uint
		if err := tx.Model(&model.CaptureDocument{}).Where("capture_package_id = ?", item.ID).Pluck("asset_id", &assetIDs).Error; err != nil {
			return err
		}
		var cropIDs []uint
		if err := tx.Model(&model.CaptureProductCrop{}).Where("capture_package_id = ?", item.ID).Pluck("asset_id", &cropIDs).Error; err != nil {
			return err
		}
		assetIDs = append(assetIDs, cropIDs...)
		if len(assetIDs) > 0 {
			if err := tx.Model(&model.MediaAsset{}).Where("id IN ?", assetIDs).Update("status", "orphaned").Error; err != nil {
				return err
			}
		}
		return tx.Delete(&model.CapturePackage{}, item.ID).Error
	})
	if err != nil {
		Fail(c, 500, 500, "资料包删除失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "delete", "capture-packages", item.ID)
	OK(c, gin.H{"deleted": true})
}

func (h AdminHandler) UploadCaptureDocuments(c *gin.Context) {
	item, ok := h.authorizedCapturePackage(c, c.Param("id"))
	if !ok {
		return
	}
	if item.Status == model.CaptureStatusProcessing || item.Status == model.CaptureStatusCommitted {
		Fail(c, 409, 409, "当前资料包不能继续上传")
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		Fail(c, 400, 400, "请选择资料图片")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}
	var existing int64
	h.DB.Model(&model.CaptureDocument{}).Where("capture_package_id = ?", item.ID).Count(&existing)
	if len(files) == 0 || int(existing)+len(files) > maxCaptureDocuments {
		Fail(c, 400, 400, "每个资料包最多上传 20 张图片")
		return
	}
	documentType := strings.TrimSpace(c.PostForm("documentType"))
	if documentType != "business_card" && documentType != "brochure" {
		documentType = "unknown"
	}
	created := make([]model.CaptureDocument, 0, len(files))
	for index, header := range files {
		asset, storeErr := h.storeCaptureAsset(header, item)
		if storeErr != nil {
			Fail(c, 400, 400, storeErr.Error())
			return
		}
		if err := h.DB.Create(&asset).Error; err != nil {
			_ = os.Remove(filepath.Join(h.Config.MediaDir, asset.StorageKey))
			Fail(c, 500, 500, "资料图片登记失败")
			return
		}
		document := model.CaptureDocument{CapturePackageID: item.ID, AssetID: asset.ID, DocumentType: documentType, SortOrder: int(existing) + index + 1, Asset: asset}
		if err := h.DB.Create(&document).Error; err != nil {
			h.DB.Model(&asset).Update("status", "orphaned")
			Fail(c, 500, 500, "资料图片登记失败")
			return
		}
		created = append(created, document)
	}
	h.DB.Model(&item).Updates(map[string]any{"status": model.CaptureStatusUploading, "draft_json": "", "error_message": ""})
	logOperation(h.DB, c.GetString("username"), "upload", "capture-packages", item.ID)
	OK(c, created)
}

func (h AdminHandler) DeleteCaptureDocument(c *gin.Context) {
	var document model.CaptureDocument
	if h.DB.First(&document, c.Param("id")).Error != nil {
		Fail(c, 404, 404, "资料图片不存在")
		return
	}
	item, ok := h.authorizedCapturePackage(c, strconv.FormatUint(uint64(document.CapturePackageID), 10))
	if !ok {
		return
	}
	if item.Status == model.CaptureStatusProcessing || item.Status == model.CaptureStatusCommitted {
		Fail(c, 409, 409, "当前资料包不能删除图片")
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&document).Error; err != nil {
			return err
		}
		return tx.Model(&model.MediaAsset{}).Where("id = ?", document.AssetID).Update("status", "orphaned").Error
	}); err != nil {
		Fail(c, 500, 500, "资料图片删除失败")
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (h AdminHandler) UpdateCaptureDocument(c *gin.Context) {
	var document model.CaptureDocument
	if h.DB.First(&document, c.Param("id")).Error != nil {
		Fail(c, 404, 404, "资料图片不存在")
		return
	}
	item, ok := h.authorizedCapturePackage(c, strconv.FormatUint(uint64(document.CapturePackageID), 10))
	if !ok {
		return
	}
	if item.Status == model.CaptureStatusProcessing || item.Status == model.CaptureStatusCommitted {
		Fail(c, 409, 409, "当前资料包不能调整图片")
		return
	}
	var req struct {
		SortOrder    int    `json:"sortOrder"`
		DocumentType string `json:"documentType"`
	}
	if c.ShouldBindJSON(&req) != nil || req.SortOrder < 1 || req.SortOrder > maxCaptureDocuments {
		Fail(c, 400, 400, "资料图片设置格式不正确")
		return
	}
	if req.DocumentType != "unknown" && req.DocumentType != "business_card" && req.DocumentType != "brochure" {
		Fail(c, 400, 400, "资料类型不正确")
		return
	}
	if err := h.DB.Model(&document).Updates(map[string]any{"sort_order": req.SortOrder, "document_type": req.DocumentType}).Error; err != nil {
		Fail(c, 500, 500, "资料图片设置保存失败")
		return
	}
	h.writeCapturePackage(c, item.ID)
}

func (h AdminHandler) RecognizeCapturePackage(c *gin.Context) {
	item, ok := h.authorizedCapturePackage(c, c.Param("id"))
	if !ok {
		return
	}
	if h.Capture == nil || !h.Capture.Enabled() {
		Fail(c, 503, 503, "智能识别尚未启用，请在平台配置中填写 OCR 密钥和 TokenHub API Key")
		return
	}
	if item.Status == model.CaptureStatusProcessing || item.Status == model.CaptureStatusQueued {
		Fail(c, 409, 409, "资料包正在识别中")
		return
	}
	if item.Status == model.CaptureStatusCommitted {
		Fail(c, 409, 409, "已提交的资料包不能重新识别")
		return
	}
	var count int64
	h.DB.Model(&model.CaptureDocument{}).Where("capture_package_id = ?", item.ID).Count(&count)
	if count == 0 {
		Fail(c, 400, 400, "请先上传资料图片")
		return
	}
	now := time.Now()
	if err := h.DB.Model(&item).Updates(map[string]any{"status": model.CaptureStatusQueued, "queued_at": &now, "started_at": nil, "completed_at": nil, "error_message": ""}).Error; err != nil {
		Fail(c, 500, 500, "识别任务创建失败")
		return
	}
	h.Capture.Notify()
	logOperation(h.DB, c.GetString("username"), "recognize", "capture-packages", item.ID)
	h.writeCapturePackage(c, item.ID)
}

func (h AdminHandler) UpdateCaptureDraft(c *gin.Context) {
	item, ok := h.authorizedCapturePackage(c, c.Param("id"))
	if !ok {
		return
	}
	if item.Status == model.CaptureStatusProcessing || item.Status == model.CaptureStatusQueued || item.Status == model.CaptureStatusCommitted {
		Fail(c, 409, 409, "当前资料包不能修改")
		return
	}
	var draft CaptureDraft
	if c.ShouldBindJSON(&draft) != nil {
		Fail(c, 400, 400, "识别草稿格式不正确")
		return
	}
	draft.normalize()
	if err := draft.validateForSave(); err != nil {
		Fail(c, 400, 400, err.Error())
		return
	}
	status := model.CaptureStatusReady
	if len(draft.unconfirmedFields()) > 0 {
		status = model.CaptureStatusNeedsReview
	}
	raw, _ := json.Marshal(draft)
	if err := h.DB.Model(&item).Updates(map[string]any{"draft_json": string(raw), "status": status, "error_message": ""}).Error; err != nil {
		Fail(c, 500, 500, "识别草稿保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "update-draft", "capture-packages", item.ID)
	h.writeCapturePackage(c, item.ID)
}

func (h AdminHandler) CommitCapturePackage(c *gin.Context) {
	item, ok := h.authorizedCapturePackage(c, c.Param("id"))
	if !ok {
		return
	}
	if item.Status == model.CaptureStatusCommitted {
		h.writeCapturePackage(c, item.ID)
		return
	}
	var draft CaptureDraft
	if item.DraftJSON == "" || json.Unmarshal([]byte(item.DraftJSON), &draft) != nil {
		Fail(c, 400, 400, "资料包尚无可提交草稿")
		return
	}
	draft.normalize()
	if err := draft.validateForSave(); err != nil {
		Fail(c, 400, 400, err.Error())
		return
	}
	if missing := draft.unconfirmedFields(); len(missing) > 0 {
		Fail(c, 400, 400, "请先确认所有红色和关键字段")
		return
	}
	if fieldValue(draft.VendorFields, "name") == "" && item.VendorID == nil && draft.VendorMatchID == nil {
		Fail(c, 400, 400, "请确认厂商名称")
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		vendor, err := commitCaptureVendor(tx, item, draft, c.GetString("username"))
		if err != nil {
			return err
		}
		if err := commitCaptureProducts(tx, item, vendor, draft, c.GetString("username")); err != nil {
			return err
		}
		now := time.Now()
		return tx.Model(&model.CapturePackage{}).Where("id = ? AND status <> ?", item.ID, model.CaptureStatusCommitted).Updates(map[string]any{"status": model.CaptureStatusCommitted, "vendor_id": vendor.ID, "committed_at": &now}).Error
	})
	if err != nil {
		Fail(c, 400, 400, "资料包提交失败："+err.Error())
		return
	}
	logOperation(h.DB, c.GetString("username"), "commit", "capture-packages", item.ID)
	h.writeCapturePackage(c, item.ID)
}

func (h AdminHandler) CaptureDocumentContent(c *gin.Context) {
	var document model.CaptureDocument
	if h.DB.Preload("Asset").First(&document, c.Param("id")).Error != nil {
		Fail(c, 404, 404, "资料图片不存在")
		return
	}
	if _, ok := h.authorizedCapturePackage(c, strconv.FormatUint(uint64(document.CapturePackageID), 10)); !ok {
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("Content-Type", document.Asset.MIME)
	c.File(filepath.Join(h.Config.MediaDir, document.Asset.StorageKey))
}

func (h AdminHandler) CaptureCropContent(c *gin.Context) {
	var crop model.CaptureProductCrop
	if h.DB.Preload("Asset").First(&crop, c.Param("id")).Error != nil {
		Fail(c, 404, 404, "候选图片不存在")
		return
	}
	if _, ok := h.authorizedCapturePackage(c, strconv.FormatUint(uint64(crop.CapturePackageID), 10)); !ok {
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("Content-Type", crop.Asset.MIME)
	c.File(filepath.Join(h.Config.MediaDir, crop.Asset.StorageKey))
}

func (h AdminHandler) authorizedCapturePackage(c *gin.Context, id any) (model.CapturePackage, bool) {
	var item model.CapturePackage
	if h.DB.First(&item, id).Error != nil {
		Fail(c, 404, 404, "资料包不存在")
		return item, false
	}
	if c.GetString("role") != "admin" && (item.OwnerUsername != c.GetString("username") || item.VendorID == nil || *item.VendorID != c.GetUint("vendorId")) {
		Fail(c, 403, 403, "无权访问该资料包")
		return item, false
	}
	return item, true
}

func (h AdminHandler) writeCapturePackage(c *gin.Context, id uint) {
	var item model.CapturePackage
	if h.DB.Preload("Vendor").Preload("Documents.Asset").Preload("Crops.Asset").First(&item, id).Error != nil {
		Fail(c, 404, 404, "资料包不存在")
		return
	}
	var draft CaptureDraft
	if item.DraftJSON != "" {
		_ = json.Unmarshal([]byte(item.DraftJSON), &draft)
		draft.normalize()
	}
	var categories []model.Category
	h.DB.Where("parent_id = ? AND is_enabled = ?", 0, true).Order("sort_order asc, id asc").Find(&categories)
	OK(c, gin.H{"package": item, "draft": draft, "aiEnabled": h.Capture != nil && h.Capture.Enabled(), "categories": categories})
}

func (h AdminHandler) storeCaptureAsset(header *multipart.FileHeader, item model.CapturePackage) (model.MediaAsset, error) {
	if header.Size <= 0 || header.Size > 10*1024*1024 {
		return model.MediaAsset{}, fmt.Errorf("单张图片不能超过 10MB")
	}
	source, err := header.Open()
	if err != nil {
		return model.MediaAsset{}, fmt.Errorf("无法读取资料图片")
	}
	defer source.Close()
	img, format, err := image.Decode(io.LimitReader(source, 10*1024*1024+1))
	if err != nil || (format != "jpeg" && format != "png" && format != "webp") {
		return model.MediaAsset{}, fmt.Errorf("资料必须是 JPEG、PNG 或 WebP 图片")
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width < 300 || height < 300 || int64(width)*int64(height) > 50_000_000 {
		return model.MediaAsset{}, fmt.Errorf("图片过小或像素超过 5000 万")
	}
	relativeDir := filepath.Join(time.Now().Format("2006"), time.Now().Format("01"), fmt.Sprintf("capture-%d", item.ID))
	if err := os.MkdirAll(filepath.Join(h.Config.MediaDir, relativeDir), 0755); err != nil {
		return model.MediaAsset{}, fmt.Errorf("创建资料目录失败")
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return model.MediaAsset{}, err
	}
	storageKey := filepath.Join(relativeDir, hex.EncodeToString(random)+".png")
	path := filepath.Join(h.Config.MediaDir, storageKey)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return model.MediaAsset{}, err
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return model.MediaAsset{}, fmt.Errorf("保存资料图片失败")
	}
	_ = file.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		_ = os.Remove(path)
		return model.MediaAsset{}, err
	}
	hash := sha256.Sum256(data)
	return model.MediaAsset{OwnerUsername: item.OwnerUsername, VendorID: item.VendorID, StorageKey: storageKey, OriginalName: filepath.Base(header.Filename), MIME: "image/png", Size: int64(len(data)), Width: width, Height: height, SHA256: hex.EncodeToString(hash[:]), Purpose: "capture_source", Status: "private"}, nil
}

func commitCaptureVendor(tx *gorm.DB, item model.CapturePackage, draft CaptureDraft, username string) (model.Vendor, error) {
	var vendor model.Vendor
	vendorID := uint(0)
	if item.VendorID != nil {
		vendorID = *item.VendorID
	} else if draft.VendorMatchID != nil {
		vendorID = *draft.VendorMatchID
	}
	if vendorID > 0 {
		if err := tx.First(&vendor, vendorID).Error; err != nil {
			return vendor, fmt.Errorf("关联厂商不存在")
		}
	} else {
		vendor = model.Vendor{Name: fieldValue(draft.VendorFields, "name"), DataOrigin: "admin", ReviewStatus: "pending", PublicationStatus: "hidden", IsVisible: false, ContentVersion: 1}
		if err := database.EnsureVendorSlug(tx, &vendor); err != nil {
			return vendor, err
		}
		if err := tx.Create(&vendor).Error; err != nil {
			return vendor, err
		}
		if err := tx.Model(&vendor).Update("is_visible", false).Error; err != nil {
			return vendor, err
		}
		vendor.IsVisible = false
	}
	proposed := vendor
	applyCaptureVendorFields(&proposed, draft.VendorFields)
	payload, err := json.Marshal(proposed)
	if err != nil {
		return vendor, err
	}
	if err := tx.Model(&model.VendorSubmission{}).Where("vendor_id = ? AND status = ?", vendor.ID, "pending").Update("status", "superseded").Error; err != nil {
		return vendor, err
	}
	packageID := item.ID
	submission := model.VendorSubmission{VendorID: vendor.ID, CapturePackageID: &packageID, BaseVersion: vendor.ContentVersion, Payload: string(payload), Status: "pending", SubmittedBy: username}
	if err := tx.Create(&submission).Error; err != nil {
		return vendor, err
	}
	return vendor, nil
}

func applyCaptureVendorFields(vendor *model.Vendor, fields map[string]CaptureField) {
	set := func(key string, target *string) {
		if value := fieldValue(fields, key); value != "" {
			*target = value
		}
	}
	set("name", &vendor.Name)
	set("shortName", &vendor.ShortName)
	set("province", &vendor.Province)
	set("city", &vendor.City)
	set("county", &vendor.County)
	set("address", &vendor.Address)
	set("websiteUrl", &vendor.WebsiteURL)
	set("contactName", &vendor.ContactName)
	set("phone", &vendor.Phone)
	set("wechat", &vendor.Wechat)
	set("mainProducts", &vendor.MainProducts)
	set("description", &vendor.Description)
	set("serviceAdvantages", &vendor.ServiceAdvantages)
	set("equipment", &vendor.Equipment)
	set("certifications", &vendor.Certifications)
}

func commitCaptureProducts(tx *gorm.DB, item model.CapturePackage, vendor model.Vendor, draft CaptureDraft, username string) error {
	for _, source := range draft.Products {
		name := fieldValue(source.Fields, "name")
		if name == "" {
			continue
		}
		categoryID := uint(0)
		if source.TargetProductID != nil {
			var target model.Product
			if tx.First(&target, *source.TargetProductID).Error == nil {
				categoryID = target.CategoryID
			}
		}
		if categoryID == 0 {
			categoryName := fieldValue(source.Fields, "categoryName")
			var category model.Category
			if categoryName != "" && tx.Where("name = ? AND is_enabled = ?", categoryName, true).First(&category).Error == nil {
				categoryID = category.ID
			}
		}
		if categoryID == 0 {
			return fmt.Errorf("产品“%s”尚未匹配有效分类", name)
		}
		selected := append([]uint(nil), source.SelectedCropIDs...)
		sort.Slice(selected, func(i, j int) bool { return selected[i] < selected[j] })
		urls := []string{}
		for _, cropID := range selected {
			var crop model.CaptureProductCrop
			if tx.Where("id = ? AND capture_package_id = ? AND product_key = ?", cropID, item.ID, source.Key).First(&crop).Error != nil {
				return fmt.Errorf("产品候选图片无效")
			}
			if err := tx.Model(&model.MediaAsset{}).Where("id = ?", crop.AssetID).Updates(map[string]any{"purpose": "submission", "status": "staged", "vendor_id": vendor.ID}).Error; err != nil {
				return err
			}
			urls = append(urls, fmt.Sprintf("/api/media/%d", crop.AssetID))
		}
		gallery, _ := json.Marshal(urls)
		specs := make([]model.ProductSpec, 0, len(source.Specs))
		for _, spec := range source.Specs {
			if spec.Name.Value != "" && spec.Value.Value != "" {
				specs = append(specs, model.ProductSpec{Name: spec.Name.Value, Value: spec.Value.Value})
			}
		}
		specsRaw, _ := json.Marshal(specs)
		imageURL := ""
		if len(urls) > 0 {
			imageURL = urls[0]
		}
		product := model.Product{Name: name, CategoryID: categoryID, CompatibleModels: fieldValue(source.Fields, "compatibleModels"), Description: fieldValue(source.Fields, "description"), DetailContent: fieldValue(source.Fields, "detailContent"), Image: imageURL, GalleryRaw: string(gallery), SpecsRaw: string(specsRaw), PriceNote: fieldValue(source.Fields, "priceNote"), PublicationStatus: "draft", Status: 2, ContentVersion: 1}
		supplier := model.ProductSupplier{VendorID: vendor.ID, VendorProductName: name, VendorModel: fieldValue(source.Fields, "model"), CompatibleModels: product.CompatibleModels, Description: product.Description, DetailContent: product.DetailContent, Image: imageURL, GalleryRaw: string(gallery), SpecsRaw: string(specsRaw), PriceNote: product.PriceNote, SupplyAbility: fieldValue(source.Fields, "supplyAbility"), Currency: "CNY", Status: "pending", SourceType: "capture", ContentVersion: 1}
		productPayload, _ := json.Marshal(product)
		supplierPayload, _ := json.Marshal(supplier)
		packageID := item.ID
		submission := model.ProductSubmission{CapturePackageID: &packageID, VendorID: vendor.ID, ProductID: source.TargetProductID, SubmissionType: "new_product", ProductPayload: string(productPayload), SupplierPayload: string(supplierPayload), Status: "pending", SubmittedBy: username}
		if err := tx.Create(&submission).Error; err != nil {
			return err
		}
	}
	return nil
}
