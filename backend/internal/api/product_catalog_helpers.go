package api

import (
	"strings"

	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/gorm"
)

const approvedSupplierExists = `EXISTS (
	SELECT 1 FROM product_suppliers ps
	JOIN vendors v ON v.id = ps.vendor_id AND v.deleted_at IS NULL
	WHERE ps.product_id = products.id AND ps.deleted_at IS NULL AND ps.status = 'approved'
	AND v.is_visible = 1 AND v.publication_status = 'published'
)`

func visibleProductQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.Product{}).Where("products.publication_status = ?", "published").Where(approvedSupplierExists)
}

func enrichProductSummaries(db *gorm.DB, products []model.Product, scopedVendorID uint) {
	for i := range products {
		var suppliers []model.ProductSupplier
		query := db.Preload("Vendor").Where("product_id = ? AND status = ?", products[i].ID, "approved").
			Where("EXISTS (SELECT 1 FROM vendors v WHERE v.id = product_suppliers.vendor_id AND v.deleted_at IS NULL AND v.is_visible = 1 AND v.publication_status = 'published')")
		if scopedVendorID > 0 {
			query = query.Where("vendor_id = ?", scopedVendorID)
		}
		_ = query.Order("id asc").Find(&suppliers).Error
		products[i].SupplierCount = int64(len(suppliers))
		regions, seen := make([]string, 0, len(suppliers)), map[string]bool{}
		for j := range suppliers {
			region := strings.TrimSpace(strings.Join(filterNonEmpty([]string{suppliers[j].Vendor.Province, suppliers[j].Vendor.City}), " · "))
			if region != "" && !seen[region] {
				seen[region] = true
				regions = append(regions, region)
			}
		}
		products[i].SupplierRegions = regions
		if scopedVendorID > 0 && len(suppliers) > 0 {
			supplier := suppliers[0]
			products[i].Supplier = &supplier
		}
	}
}

func filterNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}
	return out
}
