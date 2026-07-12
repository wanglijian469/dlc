package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

func (h AdminHandler) ReviewProductSubmission(c *gin.Context) {
	var req struct {
		Status         string `json:"status"`
		ReviewNote     string `json:"reviewNote"`
		MergeProductID uint   `json:"mergeProductId"`
	}
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
		if err := tx.First(&result, c.Param("id")).Error; err != nil {
			return err
		}
		if result.Status != "pending" {
			return errSubmissionReviewed
		}
		view := decodeProductSubmission(result)
		var supplier model.ProductSupplier
		if result.SupplierID == nil || tx.First(&supplier, *result.SupplierID).Error != nil {
			return gorm.ErrRecordNotFound
		}
		if result.BaseVersion != supplier.ContentVersion {
			return errSubmissionConflict
		}
		now := time.Now()
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
	dst.CompatibleModels = src.CompatibleModels
	dst.Description = src.Description
	dst.PriceNote = src.PriceNote
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
	if h.DB.Preload("Vendor").Where("product_id = ?", c.Param("id")).Order("id asc").Find(&rows).Error != nil {
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
	err := h.DB.Where("product_id = ? AND vendor_id = ?", uint(productID), input.VendorID).First(&row).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		Fail(c, 500, 500, "供应关系保存失败")
		return
	}
	if err == gorm.ErrRecordNotFound {
		row = model.ProductSupplier{ProductID: uint(productID), VendorID: input.VendorID, ContentVersion: 1}
	}
	applySupplierDraft(&row, input)
	row.Status = "approved"
	row.SourceType = "admin"
	row.ReviewedBy = c.GetString("username")
	now := time.Now()
	row.ReviewedAt = &now
	if h.DB.Save(&row).Error != nil {
		Fail(c, 500, 500, "供应关系保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "upsert", "product-suppliers", row.ID)
	OK(c, row)
}

func (h AdminHandler) DisableAdminProductSupplier(c *gin.Context) {
	result := h.DB.Model(&model.ProductSupplier{}).Where("id = ? AND product_id = ?", c.Param("supplierId"), c.Param("id")).Update("status", "disabled")
	if result.Error != nil || result.RowsAffected == 0 {
		Fail(c, 404, 404, "供应关系不存在")
		return
	}
	OK(c, gin.H{"disabled": true})
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
