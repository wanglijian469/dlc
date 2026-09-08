package api

import (
	"dalu-nongji-parts/backend/internal/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

func showroomSupplierQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.ProductSupplier{}).Where("product_suppliers.status = ?", "approved").
		Where("EXISTS (SELECT 1 FROM products p WHERE p.id = product_suppliers.product_id AND p.deleted_at IS NULL AND p.publication_status = 'published' AND (p.published_at IS NULL OR p.published_at <= ?))", time.Now()).
		Where("EXISTS (SELECT 1 FROM vendors v WHERE v.id = product_suppliers.vendor_id AND v.deleted_at IS NULL AND v.publication_status = 'published' AND v.is_visible = 1 AND (v.published_at IS NULL OR v.published_at <= ?))", time.Now())
}

func showroomPath(s model.ProductSupplier) string {
	return fmt.Sprintf("/v/%s/products/%d", s.Vendor.Slug, s.ID)
}

func parseShowroomPath(path string) (string, uint, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[0] != "v" || parts[1] == "" || parts[2] != "products" {
		return "", 0, false
	}
	id := parseUint(parts[3])
	return parts[1], id, id > 0
}

func loadShowroomSupplier(db *gorm.DB, slug string, id uint) (model.ProductSupplier, error) {
	var row model.ProductSupplier
	err := showroomSupplierQuery(db).Preload("Vendor").Preload("Product.Category").
		Where("product_suppliers.id = ? AND EXISTS (SELECT 1 FROM vendors owner WHERE owner.id = product_suppliers.vendor_id AND owner.slug = ?)", id, slug).First(&row).Error
	if err == nil {
		redactVendor(&row.Vendor)
		row.ReviewNote = ""
		row.ReviewedBy = ""
		row.SubmittedBy = ""
		row.ReviewedAt = nil
	}
	return row, err
}

func (h PublicHandler) ShowroomProduct(c *gin.Context) {
	row, err := loadShowroomSupplier(h.DB, c.Param("slug"), parseUint(c.Param("supplierId")))
	if err != nil {
		Fail(c, 404, 404, "本厂产品不存在或暂未公开")
		return
	}
	OK(c, row)
}

func (h PublicHandler) ShowroomProducts(c *gin.Context) {
	var vendor model.Vendor
	if publishedVendorQuery(h.DB).Where("slug = ? OR id = ?", c.Param("slug"), parseUint(c.Param("slug"))).First(&vendor).Error != nil {
		Fail(c, 404, 404, "企业展厅暂未公开")
		return
	}
	query := showroomSupplierQuery(h.DB).Preload("Product.Category").Where("product_suppliers.vendor_id = ?", vendor.ID)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("vendor_product_name LIKE ? OR vendor_model LIKE ? OR product_suppliers.compatible_models LIKE ?", like, like, like)
	}
	if id := queryUint(c, "categoryId"); id > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM products p WHERE p.id = product_suppliers.product_id AND p.category_id IN ?)", productCategoryIDs(h.DB, id))
	}
	rows := []model.ProductSupplier{}
	page, size := pageParams(c, 12)
	result, err := paginate(query.Order("showroom_featured desc, showroom_order asc, product_suppliers.id desc"), &rows, page, size)
	if err != nil {
		Fail(c, 500, 500, "本厂产品加载失败")
		return
	}
	redactVendor(&vendor)
	products := make([]model.Product, 0, len(rows))
	for _, s := range rows {
		s.Vendor = vendor
		s.ReviewNote = ""
		s.ReviewedBy = ""
		s.SubmittedBy = ""
		s.ReviewedAt = nil
		p := s.Product
		s.Product = model.Product{}
		p.Supplier = &s
		products = append(products, p)
	}
	result.Items = products
	OK(c, result)
}

func (h AdminHandler) UpdateShowroomOrder(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, 403, 403, "账号未绑定厂商")
		return
	}
	var req struct {
		Featured bool `json:"featured"`
		Order    int  `json:"order"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Order < 0 || req.Order > 9999 {
		Fail(c, 400, 400, "排序范围为 0–9999")
		return
	}
	var row model.ProductSupplier
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND vendor_id = ? AND status = ?", c.Param("id"), vendorID, "approved").First(&row).Error != nil {
			return gorm.ErrRecordNotFound
		}
		row.ShowroomFeatured = req.Featured
		row.ShowroomOrder = req.Order
		return tx.Save(&row).Error
	})
	if err != nil {
		Fail(c, 404, 404, "可排序的已发布本厂产品不存在")
		return
	}
	logOperation(h.DB, c.GetString("username"), "showroom-order", "product-suppliers", row.ID)
	OK(c, gin.H{"featured": row.ShowroomFeatured, "order": row.ShowroomOrder})
}
