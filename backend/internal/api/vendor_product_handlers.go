package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type vendorProductRecord struct {
	Product          model.Product            `json:"product"`
	Supplier         model.ProductSupplier    `json:"supplier"`
	LatestSubmission *model.ProductSubmission `json:"latestSubmission,omitempty"`
}

func vendorScope(c *gin.Context) (uint, bool) {
	vendorID := c.GetUint("vendorId")
	return vendorID, c.GetString("role") == "vendor" && vendorID > 0
}

func (h AdminHandler) ListOwnProducts(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var suppliers []model.ProductSupplier
	if err := h.DB.Preload("Product.Category").Where("vendor_id = ? AND status <> ?", vendorID, "disabled").Order("id desc").Find(&suppliers).Error; err != nil {
		Fail(c, 500, 500, "产品资料加载失败")
		return
	}
	rows := make([]vendorProductRecord, 0, len(suppliers))
	for _, supplier := range suppliers {
		var latest model.ProductSubmission
		var latestPtr *model.ProductSubmission
		if h.DB.Where("vendor_id = ? AND supplier_id = ?", vendorID, supplier.ID).Order("id desc").First(&latest).Error == nil {
			latestPtr = &latest
		}
		rows = append(rows, vendorProductRecord{Product: supplier.Product, Supplier: supplier, LatestSubmission: latestPtr})
	}
	OK(c, rows)
}

func (h AdminHandler) SearchVendorProductCatalog(c *gin.Context) {
	if _, ok := vendorScope(c); !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	var products []model.Product
	query := h.DB.Preload("Category").Where("publication_status = ?", "published")
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR compatible_models LIKE ? OR description LIKE ?", like, like, like)
	}
	if err := query.Order("is_recommended desc, sort_order asc, id asc").Limit(20).Find(&products).Error; err != nil {
		Fail(c, 500, 500, "产品目录搜索失败")
		return
	}
	enrichProductSummaries(h.DB, products, 0)
	OK(c, products)
}

func supplierFromProduct(input model.Product, productID, vendorID uint) model.ProductSupplier {
	return model.ProductSupplier{ProductID: productID, VendorID: vendorID, VendorProductName: input.Name, VendorModel: input.CompatibleModels, Image: input.Image, GalleryRaw: input.GalleryRaw, CompatibleModels: input.CompatibleModels, Description: input.Description, PriceNote: input.PriceNote, InquiryText: input.InquiryText, InquiryPath: input.InquiryPath, Status: "pending", SourceType: "vendor", ContentVersion: 1}
}

func (h AdminHandler) CreateOwnProduct(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var input model.Product
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Name) == "" {
		Fail(c, 400, 400, "请填写产品名称")
		return
	}
	input.ID = 0
	input.VendorID = vendorID
	input.Status = 2
	input.PublicationStatus = "draft"
	input.ContentVersion = 1
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&input).Error; err != nil {
			return err
		}
		supplier := supplierFromProduct(input, input.ID, vendorID)
		if err := tx.Create(&supplier).Error; err != nil {
			return err
		}
		return createProductSubmission(tx, c.GetString("username"), "new_product", input, supplier)
	})
	if err != nil {
		Fail(c, 500, 500, "产品资料保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "submit", "product-submissions", input.ID)
	var supplier model.ProductSupplier
	h.DB.Where("product_id = ? AND vendor_id = ?", input.ID, vendorID).First(&supplier)
	OK(c, vendorProductRecord{Product: input, Supplier: supplier})
}

func (h AdminHandler) LinkOwnProduct(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var input model.ProductSupplier
	if err := c.ShouldBindJSON(&input); err != nil || input.ProductID == 0 {
		Fail(c, 400, 400, "请选择平台产品")
		return
	}
	var product model.Product
	if err := h.DB.First(&product, "id = ? AND publication_status = ?", input.ProductID, "published").Error; err != nil {
		Fail(c, 404, 404, "平台产品不存在")
		return
	}
	var existing model.ProductSupplier
	foundExisting := h.DB.Where("product_id = ? AND vendor_id = ?", product.ID, vendorID).First(&existing).Error == nil
	if foundExisting && existing.Status != "disabled" {
		Fail(c, 409, 409, "该产品已关联当前厂商")
		return
	}
	input.ID = 0
	input.VendorID = vendorID
	input.ProductID = product.ID
	input.Status = "pending"
	input.SourceType = "vendor"
	input.ContentVersion = 1
	if input.VendorProductName == "" {
		input.VendorProductName = product.Name
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if foundExisting {
			input.ID = existing.ID
			input.ContentVersion = existing.ContentVersion
			if err := tx.Model(&existing).Updates(map[string]interface{}{"status": "pending", "vendor_product_name": input.VendorProductName, "vendor_model": input.VendorModel, "image": input.Image, "gallery": input.GalleryRaw, "compatible_models": input.CompatibleModels, "description": input.Description, "price_note": input.PriceNote, "inquiry_text": input.InquiryText, "inquiry_path": input.InquiryPath, "review_note": ""}).Error; err != nil {
				return err
			}
		} else if err := tx.Create(&input).Error; err != nil {
			return err
		}
		return createProductSubmission(tx, c.GetString("username"), "link_supplier", product, input)
	})
	if err != nil {
		Fail(c, 500, 500, "供应关系提交失败")
		return
	}
	OK(c, vendorProductRecord{Product: product, Supplier: input})
}

func (h AdminHandler) UpdateOwnProduct(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var supplier model.ProductSupplier
	if err := h.DB.Preload("Product").Where("id = ? AND vendor_id = ?", c.Param("id"), vendorID).First(&supplier).Error; err != nil {
		Fail(c, 404, 404, "产品供应信息不存在")
		return
	}
	var input model.ProductSupplier
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, 400, 400, "产品供应信息格式错误")
		return
	}
	draft := supplier
	draft.VendorProductName = strings.TrimSpace(input.VendorProductName)
	if draft.VendorProductName == "" {
		draft.VendorProductName = supplier.Product.Name
	}
	draft.VendorModel = input.VendorModel
	draft.Image = input.Image
	draft.GalleryRaw = input.GalleryRaw
	draft.CompatibleModels = input.CompatibleModels
	draft.Description = input.Description
	draft.PriceNote = input.PriceNote
	draft.InquiryText = input.InquiryText
	draft.InquiryPath = input.InquiryPath
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if supplier.Status == "pending" {
			if err := tx.Model(&supplier).Updates(map[string]interface{}{"vendor_product_name": draft.VendorProductName, "vendor_model": draft.VendorModel, "image": draft.Image, "gallery": draft.GalleryRaw, "compatible_models": draft.CompatibleModels, "description": draft.Description, "price_note": draft.PriceNote, "inquiry_text": draft.InquiryText, "inquiry_path": draft.InquiryPath}).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.ProductSubmission{}).Where("supplier_id = ? AND status = ?", supplier.ID, "pending").Update("status", "superseded").Error; err != nil {
			return err
		}
		return createProductSubmission(tx, c.GetString("username"), "update_offer", supplier.Product, draft)
	})
	if err != nil {
		Fail(c, 500, 500, "产品供应信息提交失败")
		return
	}
	OK(c, draft)
}

func (h AdminHandler) DeleteOwnProduct(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var supplier model.ProductSupplier
	if err := h.DB.Where("id = ? AND vendor_id = ?", c.Param("id"), vendorID).First(&supplier).Error; err != nil {
		Fail(c, 404, 404, "产品供应信息不存在")
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&supplier).Updates(map[string]interface{}{"status": "disabled", "review_note": "厂商停止供应"}).Error; err != nil {
			return err
		}
		return tx.Model(&model.ProductSubmission{}).Where("supplier_id = ? AND status = ?", supplier.ID, "pending").Update("status", "superseded").Error
	})
	if err != nil {
		Fail(c, 500, 500, "停止供应失败")
		return
	}
	OK(c, gin.H{"disabled": true})
}

func createProductSubmission(tx *gorm.DB, username, typ string, product model.Product, supplier model.ProductSupplier) error {
	productPayload, err := json.Marshal(product)
	if err != nil {
		return err
	}
	supplierPayload, err := json.Marshal(supplier)
	if err != nil {
		return err
	}
	productID, supplierID := product.ID, supplier.ID
	submission := model.ProductSubmission{VendorID: supplier.VendorID, ProductID: &productID, SupplierID: &supplierID, SubmissionType: typ, BaseVersion: supplier.ContentVersion, ProductPayload: string(productPayload), SupplierPayload: string(supplierPayload), Status: "pending", SubmittedBy: username}
	return tx.Create(&submission).Error
}
