package api

import (
	"dalu-nongji-parts/backend/internal/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

func (h AnalyticsHandler) showroomAnalytics(vendorID uint, start time.Time) (gin.H, error) {
	var queryError error
	check := func(result *gorm.DB) {
		if result.Error != nil && queryError == nil {
			queryError = result.Error
		}
	}
	base := func() *gorm.DB {
		return h.DB.Model(&model.AnalyticsEvent{}).Where("created_at >= ? AND vendor_id = ?", start, vendorID)
	}
	var pv, uv, productPV, productUV, contactUV int64
	check(base().Where("event_type = ?", "page_view").Count(&pv))
	check(base().Where("event_type = ?", "page_view").Distinct("visitor_hash").Count(&uv))
	check(base().Where("event_type = ? AND supplier_id > 0", "page_view").Count(&productPV))
	check(base().Where("event_type = ? AND supplier_id > 0", "page_view").Distinct("visitor_hash").Count(&productUV))
	contacts := []analyticsCount{}
	contactTypes := []string{"contact_phone_click", "contact_wechat_copy", "contact_wechat_qr_view", "vendor_website_click"}
	check(base().Select("event_type AS label, COUNT(*) AS count").Where("event_type IN ?", contactTypes).Group("event_type").Scan(&contacts))
	// Conversion numerator is a subset of visitors who also viewed this vendor in
	// the same range. Website jumps are not contact intent.
	check(base().Where("event_type IN ? AND visitor_hash IN (SELECT visitor_hash FROM analytics_events WHERE vendor_id = ? AND created_at >= ? AND event_type = 'page_view')", contactTypes[:3], vendorID, start).Distinct("visitor_hash").Count(&contactUV))
	sources := []analyticsCount{}
	check(base().Select("source AS label, COUNT(*) AS count").Where("event_type = ?", "page_view").Group("source").Scan(&sources))
	var counts []struct {
		SupplierID uint
		PV         int64
		UV         int64
	}
	check(base().Select("supplier_id, COUNT(*) AS pv, COUNT(DISTINCT visitor_hash) AS uv").Where("event_type = ? AND supplier_id > 0", "page_view").Group("supplier_id").Order("pv desc, supplier_id asc").Limit(20).Scan(&counts))
	items := []gin.H{}
	for _, count := range counts {
		var s model.ProductSupplier
		if h.DB.Preload("Vendor").Where("id = ? AND vendor_id = ?", count.SupplierID, vendorID).First(&s).Error != nil {
			continue
		}
		items = append(items, gin.H{"supplierId": s.ID, "name": s.VendorProductName, "path": fmt.Sprintf("/v/%s/products/%d", s.Vendor.Slug, s.ID), "pv": count.PV, "uv": count.UV})
	}
	rate := 0.0
	if uv > 0 {
		rate = float64(contactUV) / float64(uv)
	}
	return gin.H{"pv": pv, "uv": uv, "productPV": productPV, "productUV": productUV, "contactUV": contactUV, "conversionRate": rate, "contactEvents": contacts, "sources": sources, "items": items}, queryError
}
