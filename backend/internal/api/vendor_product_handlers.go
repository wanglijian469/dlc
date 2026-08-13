package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type vendorProductDraft struct {
	VendorProductName string     `json:"vendorProductName"`
	VendorModel       string     `json:"vendorModel"`
	CategoryID        uint       `json:"categoryId"`
	CompatibleModels  string     `json:"compatibleModels"`
	Description       string     `json:"description"`
	DetailContent     string     `json:"detailContent"`
	Image             string     `json:"image"`
	GalleryRaw        string     `json:"galleryRaw"`
	SpecsRaw          string     `json:"specsRaw"`
	PriceNote         string     `json:"priceNote"`
	UnitPriceCents    int64      `json:"unitPriceCents"`
	PriceUnit         string     `json:"priceUnit"`
	MinOrderQuantity  int64      `json:"minOrderQuantity"`
	TaxIncluded       bool       `json:"taxIncluded"`
	FreightNote       string     `json:"freightNote"`
	AvailableQuantity int64      `json:"availableQuantity"`
	LeadTime          string     `json:"leadTime"`
	PriceValidUntil   *time.Time `json:"priceValidUntil"`
	Negotiable        bool       `json:"negotiable"`
	SupplyAbility     string     `json:"supplyAbility"`
	InquiryText       string     `json:"inquiryText"`
}

type vendorProductRecord struct {
	ID                uint       `json:"id"`
	RecordType        string     `json:"recordType"`
	VendorProductName string     `json:"vendorProductName"`
	VendorModel       string     `json:"vendorModel"`
	CategoryID        uint       `json:"categoryId,omitempty"`
	CategoryName      string     `json:"categoryName,omitempty"`
	CompatibleModels  string     `json:"compatibleModels"`
	Description       string     `json:"description"`
	DetailContent     string     `json:"detailContent"`
	Image             string     `json:"image"`
	GalleryRaw        string     `json:"galleryRaw,omitempty"`
	SpecsRaw          string     `json:"specsRaw,omitempty"`
	PriceNote         string     `json:"priceNote"`
	UnitPriceCents    int64      `json:"unitPriceCents"`
	PriceUnit         string     `json:"priceUnit"`
	MinOrderQuantity  int64      `json:"minOrderQuantity"`
	TaxIncluded       bool       `json:"taxIncluded"`
	FreightNote       string     `json:"freightNote"`
	AvailableQuantity int64      `json:"availableQuantity"`
	LeadTime          string     `json:"leadTime"`
	PriceValidUntil   *time.Time `json:"priceValidUntil,omitempty"`
	Negotiable        bool       `json:"negotiable"`
	PriceVersion      uint       `json:"priceVersion"`
	PriceUpdatedAt    *time.Time `json:"priceUpdatedAt,omitempty"`
	SupplyAbility     string     `json:"supplyAbility"`
	InquiryText       string     `json:"inquiryText"`
	Status            string     `json:"status"`
	ReviewNote        string     `json:"reviewNote,omitempty"`
	SubmissionID      uint       `json:"submissionId,omitempty"`
	SupplierID        uint       `json:"supplierId,omitempty"`
}

type vendorProductDuplicateItem struct {
	ID         uint   `json:"id"`
	RecordType string `json:"recordType"`
	Name       string `json:"name"`
	Model      string `json:"model"`
	Status     string `json:"status"`
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
	rows, err := h.vendorProductRecords(vendorID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "产品资料加载失败")
		return
	}
	OK(c, rows)
}

func (h AdminHandler) CheckOwnProductDuplicate(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	excludeType := strings.TrimSpace(c.Query("excludeType"))
	excludeID := parseUint(c.Query("excludeId"))
	exact, similar, err := h.checkVendorProductDuplicate(vendorID, c.Query("name"), c.Query("model"), excludeType, excludeID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "重复产品检查失败")
		return
	}
	OK(c, gin.H{"exact": exact, "similar": similar})
}

func (h AdminHandler) CreateOwnProduct(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var input vendorProductDraft
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, http.StatusBadRequest, 400, "请填写本厂产品名称")
		return
	}
	if err := validateVendorProductDraft(h.DB, input); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if exact, _, err := h.checkVendorProductDuplicate(vendorID, input.VendorProductName, input.VendorModel, "", 0); err != nil {
		Fail(c, http.StatusInternalServerError, 500, "重复产品检查失败")
		return
	} else if exact {
		Fail(c, http.StatusConflict, 409, "本厂已存在名称和型号相同的产品，请编辑原记录")
		return
	}
	row, err := createStandaloneProductSubmission(h.DB, c.GetString("username"), vendorID, input)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "产品资料提交失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "submit", "product-submissions", row.ID)
	OK(c, recordFromSubmission(row))
}

func (h AdminHandler) UpdateOwnProductSubmission(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var input vendorProductDraft
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, http.StatusBadRequest, 400, "请填写本厂产品名称")
		return
	}
	if err := validateVendorProductDraft(h.DB, input); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	id := parseUint(c.Param("id"))
	var current model.ProductSubmission
	if id == 0 || h.DB.Where("id = ? AND vendor_id = ? AND submission_type = ? AND supplier_id IS NULL AND status IN ?", id, vendorID, "new_product", []string{"pending", "rejected"}).First(&current).Error != nil {
		Fail(c, http.StatusNotFound, 404, "待审核产品不存在")
		return
	}
	if exact, _, err := h.checkVendorProductDuplicate(vendorID, input.VendorProductName, input.VendorModel, "submission", id); err != nil {
		Fail(c, http.StatusInternalServerError, 500, "重复产品检查失败")
		return
	} else if exact {
		Fail(c, http.StatusConflict, 409, "本厂已存在名称和型号相同的产品，请编辑原记录")
		return
	}
	var next model.ProductSubmission
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&current).Update("status", "superseded").Error; err != nil {
			return err
		}
		var err error
		next, err = createStandaloneProductSubmission(tx, c.GetString("username"), vendorID, input)
		return err
	})
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "产品资料重新提交失败")
		return
	}
	OK(c, recordFromSubmission(next))
}

func (h AdminHandler) WithdrawOwnProductSubmission(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	result := h.DB.Model(&model.ProductSubmission{}).Where("id = ? AND vendor_id = ? AND submission_type = ? AND supplier_id IS NULL AND status IN ?", c.Param("id"), vendorID, "new_product", []string{"pending", "rejected"}).Update("status", "superseded")
	if result.Error != nil {
		Fail(c, http.StatusInternalServerError, 500, "撤回产品失败")
		return
	}
	if result.RowsAffected == 0 {
		Fail(c, http.StatusNotFound, 404, "待审核产品不存在")
		return
	}
	OK(c, gin.H{"withdrawn": true})
}

func (h AdminHandler) UpdateOwnProduct(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var supplier model.ProductSupplier
	if h.DB.Preload("Product").Where("id = ? AND vendor_id = ?", c.Param("id"), vendorID).First(&supplier).Error != nil {
		Fail(c, http.StatusNotFound, 404, "产品供应信息不存在")
		return
	}
	var input vendorProductDraft
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, http.StatusBadRequest, 400, "产品供应信息格式错误")
		return
	}
	if err := validateVendorProductDraft(h.DB, input); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if exact, _, err := h.checkVendorProductDuplicate(vendorID, input.VendorProductName, input.VendorModel, "supplier", supplier.ID); err != nil {
		Fail(c, http.StatusInternalServerError, 500, "重复产品检查失败")
		return
	} else if exact {
		Fail(c, http.StatusConflict, 409, "本厂已存在名称和型号相同的产品，请编辑原记录")
		return
	}
	draft := supplier
	applyVendorProductDraft(&draft, input)
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if supplier.Status == "pending" {
			applyVendorProductDraft(&supplier, input)
			if err := tx.Save(&supplier).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.ProductSubmission{}).Where("supplier_id = ? AND status = ?", supplier.ID, "pending").Update("status", "superseded").Error; err != nil {
			return err
		}
		return createProductSubmission(tx, c.GetString("username"), "update_offer", supplier.Product, draft)
	})
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "产品供应信息提交失败")
		return
	}
	OK(c, recordFromSupplier(draft, nil))
}

func (h AdminHandler) DeleteOwnProduct(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var supplier model.ProductSupplier
	if h.DB.Where("id = ? AND vendor_id = ?", c.Param("id"), vendorID).First(&supplier).Error != nil {
		Fail(c, http.StatusNotFound, 404, "产品供应信息不存在")
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&supplier).Updates(map[string]any{"status": "disabled", "review_note": "厂商停止供应"}).Error; err != nil {
			return err
		}
		return tx.Model(&model.ProductSubmission{}).Where("supplier_id = ? AND status = ?", supplier.ID, "pending").Update("status", "superseded").Error
	})
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "停止供应失败")
		return
	}
	OK(c, gin.H{"disabled": true})
}

func createStandaloneProductSubmission(tx *gorm.DB, username string, vendorID uint, input vendorProductDraft) (model.ProductSubmission, error) {
	product := model.Product{Name: strings.TrimSpace(input.VendorProductName), CategoryID: input.CategoryID, CompatibleModels: input.CompatibleModels, Description: input.Description, DetailContent: input.DetailContent, Image: input.Image, GalleryRaw: input.GalleryRaw, SpecsRaw: input.SpecsRaw, PublicationStatus: "draft", Status: 2, ContentVersion: 1}
	supplier := model.ProductSupplier{VendorID: vendorID, VendorProductName: strings.TrimSpace(input.VendorProductName), VendorModel: strings.TrimSpace(input.VendorModel), CompatibleModels: input.CompatibleModels, Description: input.Description, DetailContent: input.DetailContent, Image: input.Image, GalleryRaw: input.GalleryRaw, SpecsRaw: input.SpecsRaw, PriceNote: input.PriceNote, UnitPriceCents: input.UnitPriceCents, Currency: "CNY", PriceUnit: input.PriceUnit, MinOrderQuantity: input.MinOrderQuantity, TaxIncluded: input.TaxIncluded, FreightNote: input.FreightNote, AvailableQuantity: input.AvailableQuantity, LeadTime: input.LeadTime, PriceValidUntil: input.PriceValidUntil, Negotiable: input.Negotiable, PriceVersion: 1, SupplyAbility: input.SupplyAbility, InquiryText: input.InquiryText, Status: "pending", SourceType: "vendor", ContentVersion: 1}
	productPayload, err := json.Marshal(product)
	if err != nil {
		return model.ProductSubmission{}, err
	}
	supplierPayload, err := json.Marshal(supplier)
	if err != nil {
		return model.ProductSubmission{}, err
	}
	row := model.ProductSubmission{VendorID: vendorID, SubmissionType: "new_product", ProductPayload: string(productPayload), SupplierPayload: string(supplierPayload), Status: "pending", SubmittedBy: username}
	return row, tx.Create(&row).Error
}

func validateVendorProductDraft(db *gorm.DB, input vendorProductDraft) error {
	if strings.TrimSpace(input.VendorProductName) == "" {
		return fmt.Errorf("请填写本厂产品名称")
	}
	if input.CategoryID == 0 {
		return fmt.Errorf("请选择一级产品大类")
	}
	if input.UnitPriceCents < 0 || input.MinOrderQuantity < 0 || input.AvailableQuantity < 0 || (!input.Negotiable && input.UnitPriceCents > 0 && strings.TrimSpace(input.PriceUnit) == "") {
		return fmt.Errorf("请填写有效价格、计价单位、起订量和库存")
	}
	var categoryCount int64
	if err := db.Model(&model.Category{}).Where("id = ? AND parent_id = ? AND is_enabled = ?", input.CategoryID, 0, true).Count(&categoryCount).Error; err != nil {
		return err
	}
	if categoryCount == 0 {
		return fmt.Errorf("产品大类无效，请重新选择一级分类")
	}
	if err := validateJSONStringArray(input.GalleryRaw, "产品图集"); err != nil {
		return err
	}
	if err := validateJSON(input.SpecsRaw, "产品参数"); err != nil {
		return err
	}
	return nil
}

func applyVendorProductDraft(dst *model.ProductSupplier, input vendorProductDraft) {
	dst.VendorProductName = strings.TrimSpace(input.VendorProductName)
	dst.VendorModel = strings.TrimSpace(input.VendorModel)
	dst.CompatibleModels = input.CompatibleModels
	dst.Description = input.Description
	dst.DetailContent = input.DetailContent
	dst.Image = input.Image
	dst.GalleryRaw = input.GalleryRaw
	dst.SpecsRaw = input.SpecsRaw
	dst.PriceNote = input.PriceNote
	dst.UnitPriceCents = input.UnitPriceCents
	dst.Currency = "CNY"
	dst.PriceUnit = input.PriceUnit
	dst.MinOrderQuantity = input.MinOrderQuantity
	dst.TaxIncluded = input.TaxIncluded
	dst.FreightNote = input.FreightNote
	dst.AvailableQuantity = input.AvailableQuantity
	dst.LeadTime = input.LeadTime
	dst.PriceValidUntil = input.PriceValidUntil
	dst.Negotiable = input.Negotiable
	dst.SupplyAbility = input.SupplyAbility
	dst.InquiryText = input.InquiryText
}

func supplierFromProduct(input model.Product, productID, vendorID uint) model.ProductSupplier {
	return model.ProductSupplier{ProductID: productID, VendorID: vendorID, VendorProductName: input.Name, VendorModel: input.CompatibleModels, Image: input.Image, GalleryRaw: input.GalleryRaw, SpecsRaw: input.SpecsRaw, CompatibleModels: input.CompatibleModels, Description: input.Description, DetailContent: input.DetailContent, PriceNote: input.PriceNote, InquiryText: input.InquiryText, InquiryPath: input.InquiryPath, Status: "pending", SourceType: "vendor", ContentVersion: 1}
}

func (h AdminHandler) vendorProductRecords(vendorID uint) ([]vendorProductRecord, error) {
	var suppliers []model.ProductSupplier
	if err := h.DB.Preload("Product.Category").Where("vendor_id = ? AND status <> ?", vendorID, "disabled").Order("id desc").Find(&suppliers).Error; err != nil {
		return nil, err
	}
	rows := make([]vendorProductRecord, 0, len(suppliers))
	for _, supplier := range suppliers {
		var latest model.ProductSubmission
		var latestPtr *model.ProductSubmission
		if h.DB.Where("vendor_id = ? AND supplier_id = ?", vendorID, supplier.ID).Order("id desc").First(&latest).Error == nil {
			latestPtr = &latest
		}
		rows = append(rows, recordFromSupplier(supplier, latestPtr))
	}
	var submissions []model.ProductSubmission
	if err := h.DB.Where("vendor_id = ? AND submission_type = ? AND supplier_id IS NULL AND status IN ?", vendorID, "new_product", []string{"pending", "rejected"}).Order("id desc").Find(&submissions).Error; err != nil {
		return nil, err
	}
	for _, submission := range submissions {
		rows = append(rows, recordFromSubmission(submission))
	}
	return rows, nil
}

func recordFromSupplier(supplier model.ProductSupplier, latest *model.ProductSubmission) vendorProductRecord {
	name := supplier.VendorProductName
	if name == "" {
		name = supplier.Product.Name
	}
	status := supplier.Status
	if latest != nil && latest.Status == "pending" && supplier.Status == "approved" {
		status = "pending_update"
	}
	categoryID := supplier.Product.CategoryID
	if supplier.Product.Category.ParentID > 0 {
		categoryID = supplier.Product.Category.ParentID
	}
	return vendorProductRecord{ID: supplier.ID, RecordType: "supplier", SupplierID: supplier.ID, VendorProductName: name, VendorModel: supplier.VendorModel, CategoryID: categoryID, CategoryName: supplier.Product.Category.Name, CompatibleModels: supplier.CompatibleModels, Description: supplier.Description, DetailContent: supplier.DetailContent, Image: supplier.Image, GalleryRaw: supplier.GalleryRaw, SpecsRaw: supplier.SpecsRaw, PriceNote: supplier.PriceNote, UnitPriceCents: supplier.UnitPriceCents, PriceUnit: supplier.PriceUnit, MinOrderQuantity: supplier.MinOrderQuantity, TaxIncluded: supplier.TaxIncluded, FreightNote: supplier.FreightNote, AvailableQuantity: supplier.AvailableQuantity, LeadTime: supplier.LeadTime, PriceValidUntil: supplier.PriceValidUntil, Negotiable: supplier.Negotiable, PriceVersion: supplier.PriceVersion, PriceUpdatedAt: supplier.PriceUpdatedAt, SupplyAbility: supplier.SupplyAbility, InquiryText: supplier.InquiryText, Status: status, ReviewNote: supplier.ReviewNote}
}

func recordFromSubmission(submission model.ProductSubmission) vendorProductRecord {
	view := decodeProductSubmission(submission)
	return vendorProductRecord{ID: submission.ID, RecordType: "submission", SubmissionID: submission.ID, VendorProductName: view.SupplierDraft.VendorProductName, VendorModel: view.SupplierDraft.VendorModel, CategoryID: view.ProductDraft.CategoryID, CompatibleModels: view.SupplierDraft.CompatibleModels, Description: view.SupplierDraft.Description, DetailContent: view.SupplierDraft.DetailContent, Image: view.SupplierDraft.Image, GalleryRaw: view.SupplierDraft.GalleryRaw, SpecsRaw: view.SupplierDraft.SpecsRaw, PriceNote: view.SupplierDraft.PriceNote, UnitPriceCents: view.SupplierDraft.UnitPriceCents, PriceUnit: view.SupplierDraft.PriceUnit, MinOrderQuantity: view.SupplierDraft.MinOrderQuantity, TaxIncluded: view.SupplierDraft.TaxIncluded, FreightNote: view.SupplierDraft.FreightNote, AvailableQuantity: view.SupplierDraft.AvailableQuantity, LeadTime: view.SupplierDraft.LeadTime, PriceValidUntil: view.SupplierDraft.PriceValidUntil, Negotiable: view.SupplierDraft.Negotiable, PriceVersion: view.SupplierDraft.PriceVersion, SupplyAbility: view.SupplierDraft.SupplyAbility, InquiryText: view.SupplierDraft.InquiryText, Status: submission.Status, ReviewNote: submission.ReviewNote}
}

func (h AdminHandler) checkVendorProductDuplicate(vendorID uint, name, productModel, excludeType string, excludeID uint) (bool, []vendorProductDuplicateItem, error) {
	wantName, wantModel := normalizeVendorProductIdentity(name), normalizeVendorProductIdentity(productModel)
	rows, err := h.vendorProductRecords(vendorID)
	if err != nil {
		return false, nil, err
	}
	exact := false
	similar := make([]vendorProductDuplicateItem, 0)
	for _, row := range rows {
		if row.RecordType == excludeType && row.ID == excludeID {
			continue
		}
		rowName, rowModel := normalizeVendorProductIdentity(row.VendorProductName), normalizeVendorProductIdentity(row.VendorModel)
		if rowName == wantName && rowModel == wantModel {
			exact = true
			continue
		}
		if (wantModel != "" && rowModel == wantModel) || (len([]rune(wantName)) >= 3 && (strings.Contains(rowName, wantName) || strings.Contains(wantName, rowName))) {
			similar = append(similar, vendorProductDuplicateItem{ID: row.ID, RecordType: row.RecordType, Name: row.VendorProductName, Model: row.VendorModel, Status: row.Status})
		}
	}
	if len(similar) > 5 {
		similar = similar[:5]
	}
	return exact, similar, nil
}

func normalizeVendorProductIdentity(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, strings.TrimSpace(value))
}

func parseUint(value string) uint {
	var out uint
	_, _ = fmt.Sscan(value, &out)
	return out
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
