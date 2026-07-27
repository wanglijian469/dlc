package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/database"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxImportWorkbookSize = 10 * 1024 * 1024

type BulkImportIssue struct {
	Sheet   string `json:"sheet"`
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type BulkImportResult struct {
	Resource         string            `json:"resource"`
	TotalRows        int               `json:"totalRows"`
	Created          int               `json:"created"`
	Updated          int               `json:"updated"`
	RelationsCreated int               `json:"relationsCreated"`
	RelationsUpdated int               `json:"relationsUpdated"`
	Imported         bool              `json:"imported"`
	Issues           []BulkImportIssue `json:"issues"`
	Warnings         []BulkImportIssue `json:"warnings"`
}

type xlsxRow struct {
	Number int
	Cells  []string
}

type xlsxSheet struct {
	Name string
	Rows []xlsxRow
}

type vendorImportRow struct {
	Row                int
	Values             map[string]string
	ProvidesProcessing *bool
	IsRecommended      *bool
	IsVerified         *bool
	SortOrder          *int
	PublicationStatus  string
	LogoAssetID        *uint
	CoverAssetID       *uint
}

type productImportRow struct {
	Row                int
	Values             map[string]string
	CategoryID         uint
	VendorID           uint
	IsHot              *bool
	IsRecommended      *bool
	SortOrder          *int
	PublicationStatus  string
	SpecsRaw           string
	GalleryRaw         string
	SupplierGalleryRaw string
}

func (h AdminHandler) ImportWorkbook(c *gin.Context) {
	resource := strings.TrimSpace(c.Param("resource"))
	if resource != "vendors" && resource != "products" {
		Fail(c, http.StatusBadRequest, 400, "仅支持导入厂商信息或配件产品")
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, "请选择 XLSX 文件")
		return
	}
	if !strings.EqualFold(path.Ext(fileHeader.Filename), ".xlsx") {
		Fail(c, http.StatusBadRequest, 400, "仅支持 .xlsx 文件")
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxImportWorkbookSize {
		Fail(c, http.StatusBadRequest, 400, "导入文件大小必须在 10MB 以内")
		return
	}
	opened, err := fileHeader.Open()
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, "无法读取导入文件")
		return
	}
	defer opened.Close()
	data, err := io.ReadAll(io.LimitReader(opened, maxImportWorkbookSize+1))
	if err != nil || len(data) > maxImportWorkbookSize {
		Fail(c, http.StatusBadRequest, 400, "导入文件读取失败或超过 10MB")
		return
	}
	sheets, err := parseXLSX(data)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, "XLSX 文件格式无效："+err.Error())
		return
	}

	var result BulkImportResult
	if resource == "vendors" {
		result, err = h.importVendors(sheets, c.GetString("username"))
	} else {
		result, err = h.importProducts(sheets, c.GetString("username"))
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "批量导入失败："+err.Error())
		return
	}
	OK(c, result)
}

func (h AdminHandler) importVendors(sheets map[string]xlsxSheet, username string) (BulkImportResult, error) {
	result := BulkImportResult{Resource: "vendors", Issues: []BulkImportIssue{}, Warnings: []BulkImportIssue{}}
	sheet, ok := sheets["厂商信息"]
	if !ok {
		result.Issues = append(result.Issues, BulkImportIssue{Sheet: "厂商信息", Row: 1, Message: "缺少“厂商信息”工作表"})
		return result, nil
	}
	headers, issues := importHeaders(sheet, []string{"厂商名称*"})
	result.Issues = append(result.Issues, issues...)
	if len(result.Issues) > 0 {
		return result, nil
	}

	rows := make([]vendorImportRow, 0)
	for _, row := range sheet.Rows[1:] {
		values := importValues(headers, row)
		name := values["厂商名称*"]
		if isImportRowEmpty(values) || strings.HasPrefix(name, "示例：") {
			continue
		}
		result.TotalRows++
		if name == "" {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "厂商名称不能为空"})
			continue
		}
		if website := values["官网 URL"]; website != "" {
			normalizedWebsite, websiteErr := normalizeImportWebsiteURL(website)
			if websiteErr != nil {
				result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "官网 URL 格式不正确"})
			} else {
				values["官网 URL"] = normalizedWebsite
			}
		}
		provides, boolErr := parseOptionalBool(values["是否提供加工"])
		if boolErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "是否提供加工：" + boolErr.Error()})
		}
		recommended, boolErr := parseOptionalBool(values["推荐厂商"])
		if boolErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "推荐厂商：" + boolErr.Error()})
		}
		verified, boolErr := parseOptionalBool(values["平台认证"])
		if boolErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "平台认证：" + boolErr.Error()})
		}
		sortOrder, intErr := parseOptionalInt(values["排序"])
		if intErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "排序必须是整数"})
		}
		publication, statusErr := parsePublicationStatus(values["发布状态"])
		if statusErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: statusErr.Error()})
		}
		rows = append(rows, vendorImportRow{Row: row.Number, Values: values, ProvidesProcessing: provides, IsRecommended: recommended, IsVerified: verified, SortOrder: sortOrder, PublicationStatus: publication})
	}
	if result.TotalRows == 0 {
		result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: 2, Message: "没有可导入的数据行"})
	}
	if len(result.Issues) > 0 {
		return result, nil
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			var vendor model.Vendor
			lookup := tx.Where("name = ?", row.Values["厂商名称*"]).First(&vendor)
			created := lookup.Error == gorm.ErrRecordNotFound
			if lookup.Error != nil && lookup.Error != gorm.ErrRecordNotFound {
				return lookup.Error
			}
			if created {
				vendor = model.Vendor{Name: row.Values["厂商名称*"], ReviewStatus: "pending", DataOrigin: "admin", PublicationStatus: "hidden", ContentVersion: 1}
			} else {
				vendor.ContentVersion++
			}
			if row.Values["Logo URL"] != "" {
				localized, assetID, localizeErr := h.localizeImportImage(tx, row.Values["Logo URL"], username, vendorIDPointer(vendor.ID), "published")
				if localizeErr != nil {
					result.Warnings = append(result.Warnings, BulkImportIssue{Sheet: sheet.Name, Row: row.Row, Message: "Logo URL 未能下载到本地，保留原地址：" + localizeErr.Error()})
				} else {
					row.Values["Logo URL"], row.LogoAssetID = localized, assetID
				}
			}
			if row.Values["封面图 URL"] != "" {
				localized, assetID, localizeErr := h.localizeImportImage(tx, row.Values["封面图 URL"], username, vendorIDPointer(vendor.ID), "published")
				if localizeErr != nil {
					result.Warnings = append(result.Warnings, BulkImportIssue{Sheet: sheet.Name, Row: row.Row, Message: "封面图 URL 未能下载到本地，保留原地址：" + localizeErr.Error()})
				} else {
					row.Values["封面图 URL"], row.CoverAssetID = localized, assetID
				}
			}
			applyVendorImport(&vendor, row)
			service.ApplyVendorSEO(&vendor)
			if err := database.EnsureVendorSlug(tx, &vendor); err != nil {
				return fmt.Errorf("厂商信息第 %d 行页面标识生成失败: %w", row.Row, err)
			}
			if row.LogoAssetID != nil {
				vendor.LogoAssetID = row.LogoAssetID
			}
			if row.CoverAssetID != nil {
				vendor.CoverAssetID = row.CoverAssetID
			}
			if err := tx.Omit("Tags", "Media").Save(&vendor).Error; err != nil {
				return err
			}
			if created {
				result.Created++
			} else {
				result.Updated++
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	result.Imported = true
	logOperation(h.DB, username, "import", "vendors", 0)
	return result, nil
}

func (h AdminHandler) importProducts(sheets map[string]xlsxSheet, username string) (BulkImportResult, error) {
	result := BulkImportResult{Resource: "products", Issues: []BulkImportIssue{}, Warnings: []BulkImportIssue{}}
	sheet, ok := sheets["配件产品"]
	if !ok {
		sheet, ok = sheets["产品信息"]
	}
	if !ok {
		result.Issues = append(result.Issues, BulkImportIssue{Sheet: "配件产品", Row: 1, Message: "缺少“配件产品”工作表"})
		return result, nil
	}
	headers, issues := importHeaders(sheet, []string{"产品名称*", "产品分类*", "关联厂商名称*"})
	result.Issues = append(result.Issues, issues...)
	if len(result.Issues) > 0 {
		return result, nil
	}

	var categories []model.Category
	if err := h.DB.Find(&categories).Error; err != nil {
		return result, err
	}
	categoryIDs := map[string]uint{}
	for _, category := range categories {
		categoryIDs[strings.TrimSpace(category.Name)] = category.ID
	}
	var vendors []model.Vendor
	if err := h.DB.Find(&vendors).Error; err != nil {
		return result, err
	}
	vendorIDs := map[string]uint{}
	for _, vendor := range vendors {
		vendorIDs[strings.TrimSpace(vendor.Name)] = vendor.ID
	}

	rows := make([]productImportRow, 0)
	productCategories := map[string]string{}
	for _, row := range sheet.Rows[1:] {
		values := importValues(headers, row)
		name := values["产品名称*"]
		if isImportRowEmpty(values) || strings.HasPrefix(name, "示例：") {
			continue
		}
		result.TotalRows++
		categoryName, vendorName := values["产品分类*"], values["关联厂商名称*"]
		if name == "" {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "产品名称不能为空"})
		}
		categoryID := categoryIDs[categoryName]
		if categoryName == "" || categoryID == 0 {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "产品分类不存在，请使用“字段说明”中的分类名称"})
		}
		vendorID := vendorIDs[vendorName]
		if vendorName == "" || vendorID == 0 {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "关联厂商不存在，请先导入厂商信息"})
		}
		if previous, exists := productCategories[name]; exists && previous != categoryName {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "同一产品重复行的产品分类不一致"})
		} else if name != "" {
			productCategories[name] = categoryName
		}
		isHot, boolErr := parseOptionalBool(values["热门"])
		if boolErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "热门：" + boolErr.Error()})
		}
		isRecommended, boolErr := parseOptionalBool(values["推荐"])
		if boolErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "推荐：" + boolErr.Error()})
		}
		sortOrder, intErr := parseOptionalInt(values["排序"])
		if intErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: "排序必须是整数"})
		}
		publication, statusErr := parsePublicationStatus(values["发布状态"])
		if statusErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: statusErr.Error()})
		}
		specsRaw, specsErr := importSpecsJSON(values["关键参数（参数名=参数值）"])
		if specsErr != nil {
			result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: row.Number, Message: specsErr.Error()})
		}
		galleryRaw, _ := json.Marshal(splitImportList(values["产品图库 URL（多个用换行）"]))
		supplierGalleryRaw, _ := json.Marshal(splitImportList(values["厂商产品图库 URL（多个用换行）"]))
		rows = append(rows, productImportRow{Row: row.Number, Values: values, CategoryID: categoryID, VendorID: vendorID, IsHot: isHot, IsRecommended: isRecommended, SortOrder: sortOrder, PublicationStatus: publication, SpecsRaw: specsRaw, GalleryRaw: string(galleryRaw), SupplierGalleryRaw: string(supplierGalleryRaw)})
	}
	if result.TotalRows == 0 {
		result.Issues = append(result.Issues, BulkImportIssue{Sheet: sheet.Name, Row: 2, Message: "没有可导入的数据行"})
	}
	if len(result.Issues) > 0 {
		return result, nil
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		processed := map[string]model.Product{}
		for _, row := range rows {
			product, done := processed[row.Values["产品名称*"]]
			if !done {
				lookup := tx.Where("name = ?", row.Values["产品名称*"]).First(&product)
				created := lookup.Error == gorm.ErrRecordNotFound
				if lookup.Error != nil && lookup.Error != gorm.ErrRecordNotFound {
					return lookup.Error
				}
				if created {
					product = model.Product{Name: row.Values["产品名称*"], CategoryID: row.CategoryID, PublicationStatus: "hidden", Status: 2, ContentVersion: 1}
				} else {
					product.ContentVersion++
				}
				if row.Values["产品主图 URL"] != "" {
					localized, _, localizeErr := h.localizeImportImage(tx, row.Values["产品主图 URL"], username, vendorIDPointer(row.VendorID), "published")
					if localizeErr != nil {
						result.Warnings = append(result.Warnings, BulkImportIssue{Sheet: sheet.Name, Row: row.Row, Message: "产品主图 URL 未能下载到本地，保留原地址：" + localizeErr.Error()})
					} else {
						row.Values["产品主图 URL"] = localized
					}
				}
				if row.Values["产品图库 URL（多个用换行）"] != "" {
					gallery, localizeErrors := h.localizeImportImageList(tx, row.Values["产品图库 URL（多个用换行）"], username, vendorIDPointer(row.VendorID), "published")
					for _, localizeErr := range localizeErrors {
						result.Warnings = append(result.Warnings, BulkImportIssue{Sheet: sheet.Name, Row: row.Row, Message: "产品图库图片未能下载到本地，保留原地址：" + localizeErr.Error()})
					}
					row.Values["产品图库 URL（多个用换行）"] = strings.Join(gallery, "\n")
					galleryRaw, _ := json.Marshal(gallery)
					row.GalleryRaw = string(galleryRaw)
				}
				applyProductImport(&product, row)
				if err := database.EnsureProductSlug(tx, &product); err != nil {
					return fmt.Errorf("配件产品第 %d 行页面标识生成失败: %w", row.Row, err)
				}
				if err := tx.Save(&product).Error; err != nil {
					return err
				}
				processed[product.Name] = product
				if created {
					result.Created++
				} else {
					result.Updated++
				}
			}

			var supplier model.ProductSupplier
			lookup := tx.Where("product_id = ? AND vendor_id = ?", product.ID, row.VendorID).First(&supplier)
			createdSupplier := lookup.Error == gorm.ErrRecordNotFound
			if lookup.Error != nil && lookup.Error != gorm.ErrRecordNotFound {
				return lookup.Error
			}
			if createdSupplier {
				supplier = model.ProductSupplier{ProductID: product.ID, VendorID: row.VendorID, ContentVersion: 1}
			} else {
				supplier.ContentVersion++
			}
			if row.Values["厂商产品图片 URL"] != "" {
				localized, _, localizeErr := h.localizeImportImage(tx, row.Values["厂商产品图片 URL"], username, vendorIDPointer(row.VendorID), "published")
				if localizeErr != nil {
					result.Warnings = append(result.Warnings, BulkImportIssue{Sheet: sheet.Name, Row: row.Row, Message: "厂商产品图片 URL 未能下载到本地，保留原地址：" + localizeErr.Error()})
				} else {
					row.Values["厂商产品图片 URL"] = localized
				}
			}
			if row.Values["厂商产品图库 URL（多个用换行）"] != "" {
				gallery, localizeErrors := h.localizeImportImageList(tx, row.Values["厂商产品图库 URL（多个用换行）"], username, vendorIDPointer(row.VendorID), "published")
				for _, localizeErr := range localizeErrors {
					result.Warnings = append(result.Warnings, BulkImportIssue{Sheet: sheet.Name, Row: row.Row, Message: "厂商产品图库图片未能下载到本地，保留原地址：" + localizeErr.Error()})
				}
				row.Values["厂商产品图库 URL（多个用换行）"] = strings.Join(gallery, "\n")
				galleryRaw, _ := json.Marshal(gallery)
				row.SupplierGalleryRaw = string(galleryRaw)
			}
			applySupplierImport(&supplier, product, row, username)
			if err := tx.Save(&supplier).Error; err != nil {
				return err
			}
			if createdSupplier {
				result.RelationsCreated++
			} else {
				result.RelationsUpdated++
			}
		}
		for _, product := range processed {
			if product.PublicationStatus == "published" && !hasPublishableSupplier(tx, product.ID, nil) {
				return fmt.Errorf("产品“%s”发布前必须关联至少一家已发布且前台可见的厂商", product.Name)
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	result.Imported = true
	logOperation(h.DB, username, "import", "products", 0)
	return result, nil
}

func applyVendorImport(vendor *model.Vendor, row vendorImportRow) {
	values := row.Values
	vendor.Name = values["厂商名称*"]
	setImportedString(&vendor.ShortName, values["简称"])
	setImportedString(&vendor.Province, values["省份"])
	setImportedString(&vendor.City, values["城市"])
	setImportedString(&vendor.County, values["区县"])
	setImportedString(&vendor.Address, values["详细地址"])
	setImportedString(&vendor.WebsiteURL, values["官网 URL"])
	setImportedString(&vendor.ContactName, values["联系人"])
	setImportedString(&vendor.Phone, values["联系电话"])
	setImportedString(&vendor.Wechat, values["微信/其他联系方式"])
	setImportedString(&vendor.EstablishedYear, values["成立年份"])
	setImportedString(&vendor.FactoryArea, values["厂房面积"])
	setImportedString(&vendor.EmployeeCount, values["员工人数"])
	setImportedString(&vendor.MainProducts, values["主营产品"])
	setImportedString(&vendor.ServiceModels, values["服务/适配机型"])
	setImportedString(&vendor.Description, values["企业简介"])
	setImportedString(&vendor.ServiceAdvantages, values["服务优势"])
	setImportedString(&vendor.AnnualCapacity, values["年产能/供货能力"])
	setImportedString(&vendor.Equipment, values["主要设备"])
	setImportedString(&vendor.Certifications, values["资质认证"])
	setImportedString(&vendor.AfterSalesService, values["售后服务"])
	setImportedString(&vendor.ProcessingServices, values["加工服务"])
	setImportedString(&vendor.ProcessingMaterials, values["加工材料"])
	setImportedString(&vendor.ProcessingEquipment, values["加工设备"])
	setImportedString(&vendor.ProcessingCapacity, values["加工能力/精度"])
	setImportedString(&vendor.ProcessingRegions, values["服务区域"])
	setImportedString(&vendor.ProcessingNotes, values["加工说明"])
	setImportedString(&vendor.Logo, values["Logo URL"])
	setImportedString(&vendor.CoverImage, values["封面图 URL"])
	if row.ProvidesProcessing != nil {
		vendor.ProvidesProcessing = *row.ProvidesProcessing
	}
	if row.IsRecommended != nil {
		vendor.IsRecommended = *row.IsRecommended
	}
	if row.IsVerified != nil {
		vendor.IsVerified = *row.IsVerified
	}
	if row.SortOrder != nil {
		vendor.SortOrder = *row.SortOrder
	}
	if row.PublicationStatus != "" {
		vendor.PublicationStatus = row.PublicationStatus
	}
	vendor.IsVisible = vendor.PublicationStatus == "published"
}

func applyProductImport(product *model.Product, row productImportRow) {
	values := row.Values
	product.Name = values["产品名称*"]
	product.CategoryID = row.CategoryID
	setImportedString(&product.CompatibleModels, values["适配机型"])
	setImportedString(&product.Description, values["产品简介"])
	setImportedString(&product.DetailContent, values["详细说明"])
	setImportedString(&product.PriceNote, values["价格说明"])
	setImportedString(&product.Image, values["产品主图 URL"])
	setImportedString(&product.InquiryText, values["询价文案"])
	setImportedString(&product.InquiryPath, values["询价链接"])
	if values["关键参数（参数名=参数值）"] != "" {
		product.SpecsRaw = row.SpecsRaw
	}
	if values["产品图库 URL（多个用换行）"] != "" {
		product.GalleryRaw = row.GalleryRaw
	}
	if row.IsHot != nil {
		product.IsHot = *row.IsHot
	}
	if row.IsRecommended != nil {
		product.IsRecommended = *row.IsRecommended
	}
	if row.SortOrder != nil {
		product.SortOrder = *row.SortOrder
	}
	if row.PublicationStatus != "" {
		product.PublicationStatus = row.PublicationStatus
	}
	if product.PublicationStatus == "published" {
		product.Status = 1
	} else {
		product.Status = 2
	}
}

func applySupplierImport(supplier *model.ProductSupplier, product model.Product, row productImportRow, username string) {
	values := row.Values
	setImportedString(&supplier.VendorProductName, values["厂商产品名称"])
	setImportedString(&supplier.VendorModel, values["厂商型号"])
	setImportedString(&supplier.CompatibleModels, values["厂商适配机型"])
	setImportedString(&supplier.Image, values["厂商产品图片 URL"])
	setImportedString(&supplier.Description, values["厂商描述"])
	setImportedString(&supplier.PriceNote, values["厂商价格说明"])
	setImportedString(&supplier.InquiryText, values["询价文案"])
	setImportedString(&supplier.InquiryPath, values["询价链接"])
	if values["厂商产品图库 URL（多个用换行）"] != "" {
		supplier.GalleryRaw = row.SupplierGalleryRaw
	}
	if supplier.VendorProductName == "" {
		supplier.VendorProductName = product.Name
	}
	if supplier.CompatibleModels == "" {
		supplier.CompatibleModels = product.CompatibleModels
	}
	if supplier.Image == "" {
		supplier.Image = product.Image
	}
	supplier.Status = "approved"
	supplier.SourceType = "admin"
	supplier.ReviewedBy = username
	now := time.Now()
	supplier.ReviewedAt = &now
}

func importHeaders(sheet xlsxSheet, required []string) (map[string]int, []BulkImportIssue) {
	issues := []BulkImportIssue{}
	if len(sheet.Rows) == 0 {
		return nil, []BulkImportIssue{{Sheet: sheet.Name, Row: 1, Message: "工作表为空"}}
	}
	headers := map[string]int{}
	for index, value := range sheet.Rows[0].Cells {
		if normalized := canonicalImportHeader(normalizeImportText(value)); normalized != "" {
			headers[normalized] = index
		}
	}
	for _, name := range required {
		if _, ok := headers[name]; !ok {
			issues = append(issues, BulkImportIssue{Sheet: sheet.Name, Row: 1, Message: "缺少必需列：“" + name + "”"})
		}
	}
	return headers, issues
}

func importValues(headers map[string]int, row xlsxRow) map[string]string {
	values := make(map[string]string, len(headers))
	for name, index := range headers {
		if index < len(row.Cells) {
			values[name] = normalizeImportText(row.Cells[index])
		} else {
			values[name] = ""
		}
	}
	return values
}

func isImportRowEmpty(values map[string]string) bool {
	for _, value := range values {
		if value != "" {
			return false
		}
	}
	return true
}

func setImportedString(target *string, value string) {
	if value != "" {
		*target = value
	}
}

func parseOptionalBool(value string) (*bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(strings.NewReplacer("\u00a0", " ", "\u200b", "", "\ufeff", "").Replace(value)))
	if normalized == "" {
		return nil, nil
	}
	switch normalized {
	case "是", "true", "1", "yes", "y":
		parsed := true
		return &parsed, nil
	case "否", "false", "0", "no", "n":
		parsed := false
		return &parsed, nil
	case "待核实", "待确认", "待补充", "暂无", "未知", "不详", "n/a", "na", "-", "—":
		// 采集资料中的占位值表示暂未取得该字段，不覆盖已有配置。
		return nil, nil
	default:
		return nil, fmt.Errorf("请填写“是”或“否”")
	}
}

func parseOptionalInt(value string) (*int, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parsePublicationStatus(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "草稿", "draft":
		return "draft", nil
	case "发布", "已发布", "published":
		return "published", nil
	case "隐藏", "hidden":
		return "hidden", nil
	default:
		return "", fmt.Errorf("发布状态只能填写“草稿”“已发布”或“隐藏”")
	}
}

func importSpecsJSON(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	items := splitImportList(value)
	specs := make([]model.ProductSpec, 0, len(items))
	for _, item := range items {
		separator := "="
		if !strings.Contains(item, separator) {
			separator = "："
		}
		parts := strings.SplitN(item, separator, 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return "", fmt.Errorf("关键参数格式应为“参数名=参数值”，多个参数用分号或换行分隔")
		}
		specs = append(specs, model.ProductSpec{Name: strings.TrimSpace(parts[0]), Value: strings.TrimSpace(parts[1])})
	}
	payload, _ := json.Marshal(specs)
	return string(payload), nil
}

func splitImportList(value string) []string {
	replacer := strings.NewReplacer("\r\n", "\n", "\r", "\n", "；", "\n", ";", "\n", "，", "\n", ",", "\n")
	parts := strings.Split(replacer.Replace(value), "\n")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func normalizeImportText(value string) string {
	return strings.TrimSpace(strings.ReplaceAll(value, "\u00a0", " "))
}

func normalizeImportWebsiteURL(value string) (string, error) {
	normalized := normalizeImportText(value)
	if normalized == "" {
		return "", nil
	}
	lower := strings.ToLower(normalized)
	if strings.HasPrefix(normalized, "//") || (!strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") && strings.Contains(normalized, "://")) {
		return "", fmt.Errorf("unsupported URL scheme")
	}
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		normalized = "https://" + normalized
	}
	if !validURL(normalized) {
		return "", fmt.Errorf("invalid URL")
	}
	return normalized, nil
}

func canonicalImportHeader(value string) string {
	aliases := map[string]string{
		"Logo 文件名或 URL":        "Logo URL",
		"厂房/门店封面图 URL":         "封面图 URL",
		"产品主图文件名或 URL":         "产品主图 URL",
		"产品图片/图库 URL（多个用换行）":   "产品图库 URL（多个用换行）",
		"厂商产品图片/图库 URL（多个用换行）": "厂商产品图库 URL（多个用换行）",
		"厂商补充说明":               "厂商描述",
	}
	if canonical := aliases[value]; canonical != "" {
		return canonical
	}
	return value
}

func parseXLSX(data []byte) (map[string]xlsxSheet, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	entries := map[string]*zip.File{}
	for _, entry := range reader.File {
		name := strings.TrimPrefix(strings.ReplaceAll(entry.Name, "\\", "/"), "./")
		entries[name] = entry
	}
	readEntry := func(name string) ([]byte, error) {
		entry := entries[name]
		if entry == nil {
			return nil, fmt.Errorf("缺少 %s", name)
		}
		opened, openErr := entry.Open()
		if openErr != nil {
			return nil, openErr
		}
		defer opened.Close()
		content, readErr := io.ReadAll(io.LimitReader(opened, maxImportWorkbookSize*4+1))
		if readErr != nil {
			return nil, readErr
		}
		if len(content) > maxImportWorkbookSize*4 {
			return nil, fmt.Errorf("工作表内容过大")
		}
		return content, nil
	}

	workbookData, err := readEntry("xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	var workbook struct {
		Sheets []struct {
			Name string `xml:"name,attr"`
			RID  string `xml:"id,attr"`
		} `xml:"sheets>sheet"`
	}
	if err := xml.Unmarshal(workbookData, &workbook); err != nil {
		return nil, err
	}
	relsData, err := readEntry("xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}
	var relationships struct {
		Items []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(relsData, &relationships); err != nil {
		return nil, err
	}
	targets := map[string]string{}
	for _, relationship := range relationships.Items {
		target := strings.ReplaceAll(relationship.Target, "\\", "/")
		if strings.HasPrefix(target, "/") {
			target = strings.TrimPrefix(target, "/")
		} else {
			target = path.Join("xl", target)
		}
		targets[relationship.ID] = target
	}

	sharedStrings := []string{}
	if entries["xl/sharedStrings.xml"] != nil {
		sharedData, readErr := readEntry("xl/sharedStrings.xml")
		if readErr != nil {
			return nil, readErr
		}
		var shared struct {
			Items []xlsxInlineString `xml:"si"`
		}
		if err := xml.Unmarshal(sharedData, &shared); err != nil {
			return nil, err
		}
		for _, item := range shared.Items {
			sharedStrings = append(sharedStrings, item.Value())
		}
	}

	result := map[string]xlsxSheet{}
	for _, sheetRef := range workbook.Sheets {
		target := targets[sheetRef.RID]
		if target == "" {
			continue
		}
		sheetData, readErr := readEntry(target)
		if readErr != nil {
			return nil, readErr
		}
		var worksheet struct {
			Rows []struct {
				Number int `xml:"r,attr"`
				Cells  []struct {
					Ref    string           `xml:"r,attr"`
					Type   string           `xml:"t,attr"`
					Raw    string           `xml:"v"`
					Inline xlsxInlineString `xml:"is"`
				} `xml:"c"`
			} `xml:"sheetData>row"`
		}
		if err := xml.Unmarshal(sheetData, &worksheet); err != nil {
			return nil, err
		}
		rows := make([]xlsxRow, 0, len(worksheet.Rows))
		for rowIndex, xmlRow := range worksheet.Rows {
			rowNumber := xmlRow.Number
			if rowNumber == 0 {
				rowNumber = rowIndex + 1
			}
			cells := []string{}
			for _, cell := range xmlRow.Cells {
				column := xlsxColumnIndex(cell.Ref)
				for len(cells) <= column {
					cells = append(cells, "")
				}
				value := cell.Raw
				switch cell.Type {
				case "inlineStr":
					value = cell.Inline.Value()
				case "s":
					index, parseErr := strconv.Atoi(cell.Raw)
					if parseErr == nil && index >= 0 && index < len(sharedStrings) {
						value = sharedStrings[index]
					}
				}
				cells[column] = value
			}
			rows = append(rows, xlsxRow{Number: rowNumber, Cells: cells})
		}
		result[sheetRef.Name] = xlsxSheet{Name: sheetRef.Name, Rows: rows}
	}
	return result, nil
}

type xlsxInlineString struct {
	Text string `xml:"t"`
	Runs []struct {
		Text string `xml:"t"`
	} `xml:"r"`
}

func (value xlsxInlineString) Value() string {
	if value.Text != "" {
		return value.Text
	}
	var builder strings.Builder
	for _, run := range value.Runs {
		builder.WriteString(run.Text)
	}
	return builder.String()
}

func xlsxColumnIndex(reference string) int {
	index := 0
	found := false
	for _, char := range strings.ToUpper(reference) {
		if char < 'A' || char > 'Z' {
			break
		}
		found = true
		index = index*26 + int(char-'A'+1)
	}
	if !found {
		return 0
	}
	return index - 1
}
