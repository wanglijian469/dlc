package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errLastPublishableSupplier = errors.New("cannot remove the last publishable supplier")

type productReviewRequest struct {
	Status          string         `json:"status"`
	ReviewNote      string         `json:"reviewNote"`
	Resolution      string         `json:"resolution"`
	TargetProductID uint           `json:"targetProductId"`
	CatalogProduct  *model.Product `json:"catalogProduct"`
	MergeProductID  uint           `json:"mergeProductId"`
}

type productMatchSuggestion struct {
	Product             model.Product `json:"product"`
	Score               int           `json:"score"`
	Reasons             []string      `json:"reasons"`
	VendorAlreadyLinked bool          `json:"vendorAlreadyLinked"`
}

func decodeProductSubmission(row model.ProductSubmission) model.ProductSubmissionView {
	view := model.ProductSubmissionView{ProductSubmission: row}
	_ = json.Unmarshal([]byte(row.ProductPayload), &view.ProductDraft)
	_ = json.Unmarshal([]byte(row.SupplierPayload), &view.SupplierDraft)
	return view
}

func (h AdminHandler) ListProductSubmissions(c *gin.Context) {
	var rows []model.ProductSubmission
	query := h.DB.Preload("Vendor").Preload("Product.Category").Order("created_at desc, id desc")
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	page, pageSize := pageParams(c, 20)
	var total int64
	if err := query.Model(&model.ProductSubmission{}).Count(&total).Error; err != nil {
		Fail(c, 500, 500, "产品审核列表加载失败")
		return
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		Fail(c, 500, 500, "产品审核列表加载失败")
		return
	}
	views := make([]model.ProductSubmissionView, 0, len(rows))
	for _, row := range rows {
		views = append(views, decodeProductSubmission(row))
	}
	OK(c, PageResult{Items: views, Page: page, PageSize: pageSize, Total: total})
}

func (h AdminHandler) ProductSubmissionMatches(c *gin.Context) {
	var submission model.ProductSubmission
	if h.DB.Where("id = ? AND submission_type = ?", c.Param("id"), "new_product").First(&submission).Error != nil {
		Fail(c, http.StatusNotFound, 404, "产品提交不存在")
		return
	}
	view := decodeProductSubmission(submission)
	keyword := strings.TrimSpace(c.Query("keyword"))
	query := h.DB.Preload("Category").Where("publication_status = ?", "published")
	categoryIDs := categoryDescendantIDs(h.DB, view.ProductDraft.CategoryID)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR compatible_models LIKE ? OR description LIKE ?", like, like, like)
	} else if len(categoryIDs) > 0 {
		nameLike := "%" + strings.TrimSpace(view.SupplierDraft.VendorProductName) + "%"
		modelLike := "%" + strings.TrimSpace(view.SupplierDraft.VendorModel) + "%"
		if strings.TrimSpace(view.SupplierDraft.VendorModel) != "" {
			query = query.Where("category_id IN ? OR name LIKE ? OR name LIKE ? OR compatible_models LIKE ?", categoryIDs, nameLike, modelLike, modelLike)
		} else {
			query = query.Where("category_id IN ? OR name LIKE ?", categoryIDs, nameLike)
		}
	}
	var products []model.Product
	if query.Order("is_recommended desc, sort_order asc, id asc").Limit(100).Find(&products).Error != nil {
		Fail(c, http.StatusInternalServerError, 500, "匹配产品加载失败")
		return
	}
	enrichProductSummaries(h.DB, products, 0)
	items := make([]productMatchSuggestion, 0, len(products))
	for _, product := range products {
		score, reasons := scoreProductMatch(view, product, categoryIDs, keyword)
		if keyword == "" && score == 0 {
			continue
		}
		var linked int64
		h.DB.Model(&model.ProductSupplier{}).Where("product_id = ? AND vendor_id = ? AND status <> ?", product.ID, submission.VendorID, "disabled").Count(&linked)
		items = append(items, productMatchSuggestion{Product: product, Score: score, Reasons: reasons, VendorAlreadyLinked: linked > 0})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Score > items[j].Score })
	if len(items) > 10 {
		items = items[:10]
	}
	OK(c, items)
}

func (h AdminHandler) ReviewProductSubmission(c *gin.Context) {
	var req productReviewRequest
	if c.ShouldBindJSON(&req) != nil || (req.Status != "approved" && req.Status != "rejected") {
		Fail(c, 400, 400, "审核状态无效")
		return
	}
	if req.Status == "rejected" && len([]rune(strings.TrimSpace(req.ReviewNote))) < 2 {
		Fail(c, 400, 400, "驳回时请填写审核意见")
		return
	}
	var result model.ProductSubmission
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&result, c.Param("id")).Error; err != nil {
			return err
		}
		if result.Status != "pending" {
			return errSubmissionReviewed
		}
		view := decodeProductSubmission(result)
		now := time.Now()
		if result.SubmissionType == "new_product" && result.SupplierID == nil {
			if req.Status == "approved" {
				if err := approveStandaloneProductSubmission(tx, &result, view, req, c.GetString("username"), now); err != nil {
					return err
				}
			}
			result.Status = req.Status
			result.ReviewNote = strings.TrimSpace(req.ReviewNote)
			result.ReviewedBy = c.GetString("username")
			result.ReviewedAt = &now
			return tx.Save(&result).Error
		}
		var supplier model.ProductSupplier
		if result.SupplierID == nil || tx.First(&supplier, *result.SupplierID).Error != nil {
			return gorm.ErrRecordNotFound
		}
		if result.BaseVersion != supplier.ContentVersion {
			return errSubmissionConflict
		}
		if req.Status == "approved" {
			targetID := uint(0)
			if result.ProductID != nil {
				targetID = *result.ProductID
			}
			if req.MergeProductID > 0 {
				targetID = req.MergeProductID
			}
			var target model.Product
			if targetID == 0 || tx.First(&target, targetID).Error != nil {
				return gorm.ErrRecordNotFound
			}
			if result.SubmissionType == "new_product" && req.MergeProductID == 0 {
				applyProductDraft(&target, view.ProductDraft)
				target.PublicationStatus = "published"
				target.Status = 1
				target.ContentVersion++
				if err := tx.Save(&target).Error; err != nil {
					return err
				}
			}
			if req.MergeProductID > 0 && result.ProductID != nil && *result.ProductID != req.MergeProductID {
				var duplicate model.ProductSupplier
				if err := tx.Where("product_id = ? AND vendor_id = ? AND id <> ?", req.MergeProductID, result.VendorID, supplier.ID).First(&duplicate).Error; err == nil {
					if duplicate.Status != "disabled" {
						return errSupplierConflict
					}
					if err := tx.Model(&supplier).Update("status", "disabled").Error; err != nil {
						return err
					}
					supplier = duplicate
					result.SupplierID = &supplier.ID
				}
				if err := tx.Model(&model.Product{}).Where("id = ?", *result.ProductID).Updates(map[string]interface{}{"publication_status": "hidden", "status": 2}).Error; err != nil {
					return err
				}
			}
			applySupplierDraft(&supplier, view.SupplierDraft)
			supplier.ProductID = target.ID
			supplier.Status = "approved"
			supplier.ReviewNote = strings.TrimSpace(req.ReviewNote)
			supplier.ReviewedBy = c.GetString("username")
			supplier.ReviewedAt = &now
			supplier.ContentVersion++
			if err := tx.Save(&supplier).Error; err != nil {
				return err
			}
			publishProductAndSupplierMedia(tx, target, supplier)
		} else if supplier.Status == "pending" {
			supplier.Status = "rejected"
			supplier.ReviewNote = strings.TrimSpace(req.ReviewNote)
			supplier.ReviewedBy = c.GetString("username")
			supplier.ReviewedAt = &now
			if err := tx.Save(&supplier).Error; err != nil {
				return err
			}
		}
		result.Status = req.Status
		result.ReviewNote = strings.TrimSpace(req.ReviewNote)
		result.ReviewedBy = c.GetString("username")
		result.ReviewedAt = &now
		return tx.Save(&result).Error
	})
	if err != nil {
		switch err {
		case errSubmissionReviewed:
			Fail(c, 409, 409, "该产品提交已审核")
		case errSubmissionConflict:
			Fail(c, 409, 409, "供应信息版本已变化，请重新审核")
		case errSupplierConflict:
			Fail(c, 409, 409, "目标产品已关联该厂商")
		case errInvalidProductResolution:
			Fail(c, 400, 400, "新产品审核必须选择关联已有产品或创建标准产品，并完善标准产品名称")
		case gorm.ErrRecordNotFound:
			Fail(c, 404, 404, "产品提交或关联记录不存在")
		default:
			Fail(c, 500, 500, "产品审核失败")
		}
		return
	}
	logOperation(h.DB, c.GetString("username"), req.Status, "product-submissions", result.ID)
	OK(c, result)
}

var errSupplierConflict = &workflowError{"supplier conflict"}
var errInvalidProductResolution = errors.New("invalid product resolution")

func approveStandaloneProductSubmission(tx *gorm.DB, submission *model.ProductSubmission, view model.ProductSubmissionView, req productReviewRequest, username string, now time.Time) error {
	var target model.Product
	switch req.Resolution {
	case "link_product":
		if req.TargetProductID == 0 || tx.Where("id = ? AND publication_status = ?", req.TargetProductID, "published").First(&target).Error != nil {
			return gorm.ErrRecordNotFound
		}
		var linked int64
		if err := tx.Model(&model.ProductSupplier{}).Where("product_id = ? AND vendor_id = ? AND status <> ?", target.ID, submission.VendorID, "disabled").Count(&linked).Error; err != nil {
			return err
		}
		if linked > 0 {
			return errSupplierConflict
		}
	case "create_product":
		if req.CatalogProduct == nil || strings.TrimSpace(req.CatalogProduct.Name) == "" {
			return errInvalidProductResolution
		}
		input := req.CatalogProduct
		target = model.Product{
			Name:             strings.TrimSpace(input.Name),
			Image:            input.Image,
			CategoryID:       input.CategoryID,
			CompatibleModels: input.CompatibleModels,
			Description:      input.Description,
			DetailContent:    input.DetailContent,
			GalleryRaw:       input.GalleryRaw,
			SpecsRaw:         input.SpecsRaw,
		}
		target.PublicationStatus = "published"
		target.Status = 1
		target.ContentVersion = 1
		target.PublishedAt = &now
		target.CreatedAt = time.Time{}
		target.UpdatedAt = time.Time{}
		target.DeletedAt = gorm.DeletedAt{}
		if target.CategoryID > 0 {
			var categoryCount int64
			if err := tx.Model(&model.Category{}).Where("id = ?", target.CategoryID).Count(&categoryCount).Error; err != nil || categoryCount == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		if err := database.EnsureProductSlug(tx, &target); err != nil {
			return err
		}
		if err := tx.Create(&target).Error; err != nil {
			return err
		}
	default:
		return errInvalidProductResolution
	}

	supplier := view.SupplierDraft
	supplier.ID = 0
	supplier.ProductID = target.ID
	supplier.VendorID = submission.VendorID
	supplier.Product = model.Product{}
	supplier.Vendor = model.Vendor{}
	supplier.Status = "approved"
	supplier.SourceType = "vendor"
	supplier.ReviewNote = strings.TrimSpace(req.ReviewNote)
	supplier.ReviewedBy = username
	supplier.ReviewedAt = &now
	supplier.ContentVersion = 1
	supplier.CreatedAt = time.Time{}
	supplier.UpdatedAt = time.Time{}
	supplier.DeletedAt = gorm.DeletedAt{}
	if strings.TrimSpace(supplier.VendorProductName) == "" {
		supplier.VendorProductName = view.ProductDraft.Name
	}
	if err := tx.Create(&supplier).Error; err != nil {
		var linked int64
		if countErr := tx.Unscoped().Model(&model.ProductSupplier{}).Where("product_id = ? AND vendor_id = ?", target.ID, submission.VendorID).Count(&linked).Error; countErr == nil && linked > 0 {
			return errSupplierConflict
		}
		return err
	}
	productID, supplierID := target.ID, supplier.ID
	submission.ProductID, submission.SupplierID = &productID, &supplierID
	publishProductAndSupplierMedia(tx, target, supplier)
	return nil
}

func categoryDescendantIDs(db *gorm.DB, root uint) []uint {
	if root == 0 {
		return nil
	}
	var categories []model.Category
	if db.Select("id", "parent_id").Find(&categories).Error != nil {
		return []uint{root}
	}
	result, seen := []uint{root}, map[uint]bool{root: true}
	for changed := true; changed; {
		changed = false
		for _, category := range categories {
			if !seen[category.ID] && seen[category.ParentID] {
				seen[category.ID] = true
				result = append(result, category.ID)
				changed = true
			}
		}
	}
	return result
}

func scoreProductMatch(view model.ProductSubmissionView, product model.Product, categoryIDs []uint, keyword string) (int, []string) {
	score, reasons := 0, make([]string, 0, 4)
	wantName := normalizeVendorProductIdentity(view.SupplierDraft.VendorProductName)
	if wantName == "" {
		wantName = normalizeVendorProductIdentity(view.ProductDraft.Name)
	}
	productName := normalizeVendorProductIdentity(product.Name)
	if wantName != "" && wantName == productName {
		score += 60
		reasons = append(reasons, "产品名称一致")
	} else if wantName != "" && (strings.Contains(productName, wantName) || strings.Contains(wantName, productName)) {
		score += 35
		reasons = append(reasons, "产品名称相近")
	}
	modelValue := normalizeVendorProductIdentity(view.SupplierDraft.VendorModel)
	compatible := normalizeVendorProductIdentity(product.CompatibleModels)
	if modelValue != "" && (strings.Contains(productName, modelValue) || strings.Contains(compatible, modelValue)) {
		score += 25
		reasons = append(reasons, "型号匹配")
	}
	wantCompatible := normalizeVendorProductIdentity(view.SupplierDraft.CompatibleModels)
	if wantCompatible != "" && compatible != "" && (strings.Contains(compatible, wantCompatible) || strings.Contains(wantCompatible, compatible)) {
		score += 20
		reasons = append(reasons, "适配信息相近")
	}
	for _, id := range categoryIDs {
		if product.CategoryID == id {
			score += 15
			reasons = append(reasons, "分类相符")
			break
		}
	}
	if keyword != "" && score == 0 {
		score = 1
		reasons = append(reasons, "搜索结果")
	}
	return score, reasons
}

func applyProductDraft(dst *model.Product, src model.Product) {
	dst.Name = strings.TrimSpace(src.Name)
	dst.Image = src.Image
	dst.CategoryID = src.CategoryID
	dst.CompatibleModels = src.CompatibleModels
	dst.Description = src.Description
	dst.DetailContent = src.DetailContent
	dst.GalleryRaw = src.GalleryRaw
	dst.SpecsRaw = src.SpecsRaw
}

func applySupplierDraft(dst *model.ProductSupplier, src model.ProductSupplier) {
	dst.VendorProductName = src.VendorProductName
	dst.VendorModel = src.VendorModel
	dst.Image = src.Image
	dst.GalleryRaw = src.GalleryRaw
	dst.SpecsRaw = src.SpecsRaw
	dst.CompatibleModels = src.CompatibleModels
	dst.Description = src.Description
	dst.DetailContent = src.DetailContent
	dst.PriceNote = src.PriceNote
	dst.SupplyAbility = src.SupplyAbility
	dst.InquiryText = src.InquiryText
	dst.InquiryPath = src.InquiryPath
}

func publishProductAndSupplierMedia(db *gorm.DB, product model.Product, supplier model.ProductSupplier) {
	urls := []string{product.Image, supplier.Image}
	var gallery []string
	_ = json.Unmarshal([]byte(product.GalleryRaw), &gallery)
	urls = append(urls, gallery...)
	gallery = nil
	_ = json.Unmarshal([]byte(supplier.GalleryRaw), &gallery)
	urls = append(urls, gallery...)
	for _, spec := range product.Specs() {
		urls = append(urls, spec.Image)
	}
	for _, spec := range supplier.Specs() {
		urls = append(urls, spec.Image)
	}
	ids := make([]uint, 0, len(urls))
	for _, value := range urls {
		if strings.HasPrefix(value, "/api/media/") {
			if id, err := strconv.ParseUint(strings.TrimPrefix(value, "/api/media/"), 10, 64); err == nil {
				ids = append(ids, uint(id))
			}
		}
	}
	if len(ids) > 0 {
		now := time.Now()
		db.Model(&model.MediaAsset{}).Where("id IN ?", uniqueUintIDs(ids)).Updates(map[string]interface{}{"status": "published", "published_at": &now})
	}
}

func (h AdminHandler) ListAdminProductSuppliers(c *gin.Context) {
	var rows []model.ProductSupplier
	if h.DB.Preload("Vendor").Where("product_id = ? AND status <> ?", c.Param("id"), "disabled").Order("id asc").Find(&rows).Error != nil {
		Fail(c, 500, 500, "供应商列表加载失败")
		return
	}
	OK(c, rows)
}

func (h AdminHandler) SaveAdminProductSupplier(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input model.ProductSupplier
	if productID == 0 || c.ShouldBindJSON(&input) != nil || input.VendorID == 0 {
		Fail(c, 400, 400, "请选择产品和厂商")
		return
	}
	var productCount, vendorCount int64
	h.DB.Model(&model.Product{}).Where("id = ?", uint(productID)).Count(&productCount)
	h.DB.Model(&model.Vendor{}).Where("id = ?", input.VendorID).Count(&vendorCount)
	if productCount == 0 || vendorCount == 0 {
		Fail(c, 404, 404, "产品或厂商不存在")
		return
	}
	var row model.ProductSupplier
	err := h.DB.Unscoped().Where("product_id = ? AND vendor_id = ?", uint(productID), input.VendorID).First(&row).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		Fail(c, 500, 500, "供应关系保存失败")
		return
	}
	if err == gorm.ErrRecordNotFound {
		row = model.ProductSupplier{ProductID: uint(productID), VendorID: input.VendorID, ContentVersion: 1}
	}
	row.DeletedAt = gorm.DeletedAt{}
	applySupplierDraft(&row, input)
	row.Status = "approved"
	row.SourceType = "admin"
	row.ReviewedBy = c.GetString("username")
	now := time.Now()
	row.ReviewedAt = &now
	if h.DB.Unscoped().Save(&row).Error != nil {
		Fail(c, 500, 500, "供应关系保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "upsert", "product-suppliers", row.ID)
	OK(c, row)
}

func (h AdminHandler) DisableAdminProductSupplier(c *gin.Context) {
	var row model.ProductSupplier
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND product_id = ?", c.Param("supplierId"), c.Param("id")).First(&row).Error; err != nil {
			return err
		}
		var product model.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, row.ProductID).Error; err != nil {
			return err
		}
		if product.PublicationStatus == "published" {
			var currentQualifies, remaining int64
			publishable := tx.Model(&model.ProductSupplier{}).
				Joins("JOIN vendors ON vendors.id = product_suppliers.vendor_id AND vendors.deleted_at IS NULL").
				Where("product_suppliers.product_id = ? AND product_suppliers.status = ? AND vendors.is_visible = ? AND vendors.publication_status = ? AND (vendors.published_at IS NULL OR vendors.published_at <= ?)",
					product.ID, "approved", true, "published", time.Now())
			if err := publishable.Where("product_suppliers.id = ?", row.ID).Count(&currentQualifies).Error; err != nil {
				return err
			}
			if currentQualifies > 0 {
				if err := publishable.Where("product_suppliers.id <> ?", row.ID).Count(&remaining).Error; err != nil {
					return err
				}
				if remaining == 0 {
					return errLastPublishableSupplier
				}
			}
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return tx.Model(&model.ProductSubmission{}).Where("supplier_id = ? AND status = ?", row.ID, "pending").Update("status", "superseded").Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		Fail(c, http.StatusNotFound, 404, "供应关系或产品不存在")
		return
	}
	if errors.Is(err, errLastPublishableSupplier) {
		Fail(c, http.StatusConflict, 409, "已发布产品必须保留至少一家已发布且前台可见的厂商")
		return
	}
	if err != nil {
		Fail(c, 500, 500, "删除供应关系失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "delete", "product-suppliers", row.ID)
	OK(c, gin.H{"deleted": true})
}

func (h AdminHandler) BatchSaveAdminProductSuppliers(c *gin.Context) {
	var req struct {
		ProductIDs []uint `json:"productIds"`
		VendorID   uint   `json:"vendorId"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, http.StatusBadRequest, 400, "批量关联参数格式错误")
		return
	}
	req.ProductIDs = uniqueUintIDs(req.ProductIDs)
	if len(req.ProductIDs) == 0 || req.VendorID == 0 {
		Fail(c, http.StatusBadRequest, 400, "请选择产品和厂商")
		return
	}
	var vendor model.Vendor
	if err := h.DB.First(&vendor, req.VendorID).Error; err != nil {
		Fail(c, http.StatusNotFound, 404, "厂商不存在或已删除")
		return
	}
	var products []model.Product
	if err := h.DB.Where("id IN ?", req.ProductIDs).Find(&products).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 500, "产品校验失败")
		return
	}
	if len(products) != len(req.ProductIDs) {
		Fail(c, http.StatusBadRequest, 400, "所选产品中包含不存在或已删除的记录")
		return
	}
	created, existing := 0, 0
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		for _, product := range products {
			var current model.ProductSupplier
			lookup := tx.Unscoped().Where("product_id = ? AND vendor_id = ?", product.ID, vendor.ID).First(&current)
			wasActive := lookup.Error == nil && !current.DeletedAt.Valid && current.Status != "disabled"
			supplier, err := upsertAdminProductSupplier(tx, product, vendor.ID, c.GetString("username"))
			if err != nil {
				return err
			}
			if wasActive {
				existing++
			} else {
				created++
			}
			logOperation(tx, c.GetString("username"), "upsert", "product-suppliers", supplier.ID)
		}
		return nil
	})
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "批量关联失败，未写入任何关系")
		return
	}
	OK(c, gin.H{"created": created, "existing": existing, "total": len(products)})
}

func (h AdminHandler) MergeProducts(c *gin.Context) {
	targetID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		SourceProductID uint `json:"sourceProductId"`
	}
	if c.ShouldBindJSON(&req) != nil || targetID == 0 || req.SourceProductID == 0 || uint(targetID) == req.SourceProductID {
		Fail(c, 400, 400, "合并产品参数无效")
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var target, source model.Product
		if tx.First(&target, uint(targetID)).Error != nil || tx.First(&source, req.SourceProductID).Error != nil {
			return gorm.ErrRecordNotFound
		}
		var rows []model.ProductSupplier
		if tx.Where("product_id = ?", source.ID).Find(&rows).Error != nil {
			return gorm.ErrInvalidData
		}
		for _, row := range rows {
			var existing model.ProductSupplier
			if err := tx.Where("product_id = ? AND vendor_id = ?", target.ID, row.VendorID).First(&existing).Error; err == nil {
				if existing.Status == "disabled" && row.Status != "disabled" {
					applySupplierDraft(&existing, row)
					existing.Status = row.Status
					existing.ContentVersion++
					if err := tx.Save(&existing).Error; err != nil {
						return err
					}
				}
				if err := tx.Model(&row).Update("status", "disabled").Error; err != nil {
					return err
				}
			} else if err := tx.Model(&row).Update("product_id", target.ID).Error; err != nil {
				return err
			}
		}
		return tx.Model(&source).Updates(map[string]interface{}{"publication_status": "hidden", "status": 2}).Error
	})
	if err != nil {
		Fail(c, 500, 500, "产品合并失败")
		return
	}
	OK(c, gin.H{"merged": true})
}
