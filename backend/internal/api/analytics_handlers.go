package api

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/netip"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const analyticsCookieName = "dlc_visitor"

var analyticsEvents = map[string]bool{
	"page_view":               true,
	"search_submit":           true,
	"join_cta_click":          true,
	"vendor_register_success": true,
	"contact_phone_click":     true,
	"contact_wechat_copy":     true,
	"contact_wechat_qr_view":  true,
	"vendor_website_click":    true,
	"search_zero_results":     true,
	"not_found":               true,
}

type AnalyticsHandler struct {
	DB     *gorm.DB
	Config config.Config
}

type analyticsRequest struct {
	Source      string `json:"source"`
	EventType   string `json:"eventType"`
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
	ContentID   uint   `json:"contentId"`
}

func (h AnalyticsHandler) RecordEvent(c *gin.Context) {
	var req analyticsRequest
	if c.ShouldBindJSON(&req) != nil || !analyticsEvents[req.EventType] {
		Fail(c, http.StatusBadRequest, 400, "访问统计事件格式不正确")
		return
	}
	path, contentType, contentID, ok := "", "", uint(0), false
	if req.EventType == "search_zero_results" || req.EventType == "not_found" {
		path = strings.TrimSpace(req.Path)
		ok = path != "" && len(path) <= 255 && strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//") && !strings.Contains(path, "?")
	} else {
		path, contentType, contentID, ok = h.resolveTarget(strings.TrimSpace(req.Path), strings.TrimSpace(req.ContentType), req.ContentID)
	}
	if !ok {
		Fail(c, http.StatusBadRequest, 400, "访问内容不可统计")
		return
	}
	if req.EventType != "page_view" && req.EventType != "search_zero_results" && req.EventType != "not_found" && !isConversionTargetAllowed(req.EventType, path, contentType) {
		Fail(c, http.StatusBadRequest, 400, "转化事件与页面不匹配")
		return
	}
	vendorID, supplierID := uint(0), uint(0)
	if contentType == "vendor" {
		vendorID = contentID
	}
	if contentType == "supplier" {
		var supplier model.ProductSupplier
		if h.DB.First(&supplier, contentID).Error != nil {
			Fail(c, 400, 400, "访问内容不可统计")
			return
		}
		vendorID, supplierID = supplier.VendorID, supplier.ID
	}
	if c.GetString("role") == "admin" || (vendorID > 0 && c.GetString("role") == "vendor" && c.GetUint("vendorId") == vendorID) {
		OK(c, gin.H{"recorded": false})
		return
	}
	source := "unknown"
	if req.Source == "share" || req.Source == "qr" || req.Source == "site" {
		source = req.Source
	}
	visitor := h.visitorHash(c)
	window := analyticsWindow(req.EventType, time.Now())
	event := model.AnalyticsEvent{
		EventType: req.EventType, Path: path, ContentType: contentType, ContentID: contentID,
		VendorID: vendorID, SupplierID: supplierID, Source: source,
		Province: provinceForIP(c.ClientIP(), h.Config.GeoIPDBPath), VisitorHash: visitor, EventWindow: window,
	}
	if err := h.DB.Create(&event).Error; err != nil {
		// A duplicate index hit means the same action was already received in its
		// short reporting window; it is intentionally treated as a successful no-op.
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			OK(c, gin.H{"recorded": false})
			return
		}
		Fail(c, http.StatusInternalServerError, 500, "访问统计保存失败")
		return
	}
	OK(c, gin.H{"recorded": true})
}

func (h AnalyticsHandler) resolveTarget(path, contentType string, contentID uint) (string, string, uint, bool) {
	if path == "" || len(path) > 255 || strings.Contains(path, "?") || !strings.HasPrefix(path, "/") {
		return "", "", 0, false
	}
	if path == "/" || map[string]bool{"/vendors": true, "/products": true, "/service": true, "/join": true, "/about": true, "/purchase": true, "/links": true, "/guides": true, "/search": true}[path] {
		if contentType == "category" && path == "/products" && contentID > 0 {
			var category model.Category
			if h.DB.First(&category, "id = ? AND is_enabled = ?", contentID, true).Error != nil {
				return "", "", 0, false
			}
			return path, contentType, contentID, true
		}
		if contentType != "" || contentID != 0 {
			return "", "", 0, false
		}
		return path, "", 0, true
	}
	if slug, id, ok := parseShowroomPath(path); ok {
		supplier, err := loadShowroomSupplier(h.DB, slug, id)
		if err != nil {
			return "", "", 0, false
		}
		return showroomPath(supplier), "supplier", supplier.ID, true
	}
	if id, ok := numericRoute(path, "/vendors/"); ok {
		var vendor model.Vendor
		if h.DB.First(&vendor, "id = ? AND is_visible = ? AND publication_status = ?", id, true, "published").Error != nil {
			return "", "", 0, false
		}
		return path, "vendor", id, true
	}
	if slug, ok := slugRoute(path, "/vendors/"); ok {
		var vendor model.Vendor
		if publishedVendorQuery(h.DB).First(&vendor, "slug = ?", slug).Error != nil {
			return "", "", 0, false
		}
		return path, "vendor", vendor.ID, true
	}
	if id, ok := numericRoute(path, "/v/"); ok {
		var vendor model.Vendor
		if h.DB.First(&vendor, "id = ? AND is_visible = ? AND publication_status = ?", id, true, "published").Error != nil {
			return "", "", 0, false
		}
		return path, "vendor", id, true
	}
	if slug, ok := slugRoute(path, "/v/"); ok {
		var vendor model.Vendor
		if publishedVendorQuery(h.DB).First(&vendor, "slug = ?", slug).Error != nil {
			return "", "", 0, false
		}
		return path, "vendor", vendor.ID, true
	}
	if slug, ok := slugRoute(path, "/products/category/"); ok {
		var category model.Category
		if publishedCategoryQuery(h.DB).First(&category, "slug = ?", slug).Error != nil {
			return "", "", 0, false
		}
		return path, "category", category.ID, true
	}
	if id, ok := numericRoute(path, "/products/"); ok {
		var product model.Product
		if visibleProductQuery(h.DB).First(&product, id).Error != nil {
			return "", "", 0, false
		}
		return path, "product", id, true
	}
	if slug, ok := slugRoute(path, "/products/"); ok {
		var product model.Product
		if visibleProductQuery(h.DB).First(&product, "slug = ?", slug).Error != nil {
			return "", "", 0, false
		}
		return path, "product", product.ID, true
	}
	if strings.HasPrefix(path, "/guides/") && len(strings.TrimPrefix(path, "/guides/")) > 0 {
		slug := strings.TrimPrefix(path, "/guides/")
		if strings.Contains(slug, "/") {
			return "", "", 0, false
		}
		var article model.ContentPage
		if h.DB.Where("slug = ? AND page_type = ? AND is_enabled = ? AND (published_at IS NULL OR published_at <= ?)", slug, "article", true, time.Now()).First(&article).Error != nil {
			return "", "", 0, false
		}
		return path, "article", article.ID, true
	}
	return "", "", 0, false
}

func numericRoute(path, prefix string) (uint, bool) {
	raw := strings.TrimPrefix(path, prefix)
	if raw == path || raw == "" || strings.Contains(raw, "/") {
		return 0, false
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	return uint(id), err == nil && id > 0
}

func slugRoute(path, prefix string) (string, bool) {
	raw := strings.TrimPrefix(path, prefix)
	if raw == path || raw == "" || strings.Contains(raw, "/") {
		return "", false
	}
	if _, err := strconv.ParseUint(raw, 10, 64); err == nil {
		return "", false
	}
	return raw, true
}

func isConversionTargetAllowed(eventType, path, contentType string) bool {
	switch eventType {
	case "search_submit":
		return path == "/search"
	case "join_cta_click", "vendor_register_success":
		return path == "/join"
	case "contact_phone_click", "contact_wechat_copy", "contact_wechat_qr_view", "vendor_website_click":
		return contentType == "vendor" || contentType == "supplier"
	default:
		return false
	}
}

func analyticsWindow(eventType string, now time.Time) int64 {
	interval := int64(30 * time.Minute / time.Second)
	if eventType != "page_view" {
		interval = int64(5 * time.Minute / time.Second)
	}
	return now.Unix() / interval
}

func (h AnalyticsHandler) visitorHash(c *gin.Context) string {
	visitorID := ""
	if cookie, err := c.Request.Cookie(analyticsCookieName); err == nil {
		visitorID = cookie.Value
	}
	if len(visitorID) < 24 {
		bytes := make([]byte, 24)
		if _, err := rand.Read(bytes); err == nil {
			visitorID = hex.EncodeToString(bytes)
		} else {
			visitorID = strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		http.SetCookie(c.Writer, &http.Cookie{Name: analyticsCookieName, Value: visitorID, Path: "/", MaxAge: 90 * 24 * 60 * 60, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: strings.EqualFold(h.Config.Environment, "production")})
	}
	secret := h.Config.AuthSecret
	if secret == "" {
		secret = "analytics-development-secret"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(visitorID))
	return hex.EncodeToString(mac.Sum(nil))
}

type provinceNetwork struct {
	prefix   netip.Prefix
	province string
}

var provinceCache = struct {
	sync.Mutex
	paths map[string][]provinceNetwork
}{paths: map[string][]provinceNetwork{}}

func provinceForIP(rawIP, databasePath string) string {
	ip, err := netip.ParseAddr(rawIP)
	if err != nil || !ip.IsGlobalUnicast() || databasePath == "" {
		return "未知"
	}
	provinceCache.Lock()
	rows, found := provinceCache.paths[databasePath]
	if !found {
		rows = loadProvinceNetworks(databasePath)
		provinceCache.paths[databasePath] = rows
	}
	provinceCache.Unlock()
	for _, row := range rows {
		if row.prefix.Contains(ip) {
			return row.province
		}
	}
	return "未知"
}

func loadProvinceNetworks(path string) []provinceNetwork {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	rows := []provinceNetwork{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), ",", 2)
		if len(parts) != 2 || strings.HasPrefix(parts[0], "#") {
			continue
		}
		prefix, err := netip.ParsePrefix(strings.TrimSpace(parts[0]))
		province := strings.TrimSpace(parts[1])
		if err == nil && province != "" {
			rows = append(rows, provinceNetwork{prefix, province})
		}
	}
	return rows
}

type analyticsTrend struct {
	Date string `json:"date"`
	PV   int64  `json:"pv"`
	UV   int64  `json:"uv"`
}
type analyticsCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}
type analyticsContentCount struct {
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
	ContentID   uint   `json:"contentId"`
	Count       int64  `json:"count"`
}

type vendorAnalyticsProduct struct {
	ProductID uint   `json:"productId"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	PV        int64  `json:"pv"`
	UV        int64  `json:"uv"`
}

type vendorAnalyticsProductCount struct {
	ProductID uint  `gorm:"column:product_id"`
	PV        int64 `gorm:"column:pv"`
	UV        int64 `gorm:"column:uv"`
}

func analyticsDays(raw string) int {
	days, _ := strconv.Atoi(raw)
	if days != 7 && days != 30 && days != 90 {
		return 30
	}
	return days
}

func (h AnalyticsHandler) VendorSummary(c *gin.Context) {
	vendorID := c.GetUint("vendorId")
	if c.GetString("role") != "vendor" || vendorID == 0 {
		Fail(c, http.StatusForbidden, 403, "厂商账号未绑定厂商资料")
		return
	}
	var vendor model.Vendor
	if err := h.DB.Select("id", "name").First(&vendor, vendorID).Error; err != nil {
		Fail(c, http.StatusForbidden, 403, "厂商账号未绑定有效厂商资料")
		return
	}

	days := analyticsDays(c.DefaultQuery("days", "30"))
	start := time.Now().AddDate(0, 0, -days)
	vendorEvents := h.DB.Model(&model.AnalyticsEvent{}).
		Where("created_at >= ? AND content_type = ? AND content_id = ?", start, "vendor", vendorID)
	var vendorPV, vendorUV, contacts int64
	vendorEvents.Session(&gorm.Session{}).Where("event_type = ?", "page_view").Count(&vendorPV)
	vendorEvents.Session(&gorm.Session{}).Where("event_type = ?", "page_view").Distinct("visitor_hash").Count(&vendorUV)
	contactTypes := []string{"contact_phone_click", "contact_wechat_copy", "contact_wechat_qr_view", "vendor_website_click"}
	vendorEvents.Session(&gorm.Session{}).Where("event_type IN ?", contactTypes).Count(&contacts)
	contactEvents := []analyticsCount{}
	h.DB.Model(&model.AnalyticsEvent{}).
		Select("event_type AS label, COUNT(*) AS count").
		Where("created_at >= ? AND content_type = ? AND content_id = ? AND event_type IN ?", start, "vendor", vendorID, contactTypes).
		Group("event_type").Order("event_type asc").Scan(&contactEvents)
	vendorTrend := []analyticsTrend{}
	h.DB.Model(&model.AnalyticsEvent{}).
		Select("DATE(created_at) AS date, COUNT(*) AS pv, COUNT(DISTINCT visitor_hash) AS uv").
		Where("created_at >= ? AND event_type = ? AND content_type = ? AND content_id = ?", start, "page_view", "vendor", vendorID).
		Group("DATE(created_at)").Order("date asc").Scan(&vendorTrend)

	var products []model.Product
	h.DB.Model(&model.Product{}).
		Select("products.id", "products.name", "products.slug").
		Joins("JOIN product_suppliers ON product_suppliers.product_id = products.id AND product_suppliers.deleted_at IS NULL").
		Where("product_suppliers.vendor_id = ? AND product_suppliers.status = ?", vendorID, "approved").
		Order("products.name asc").Find(&products)
	productIDs := make([]uint, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	productPV, productUV := int64(0), int64(0)
	productTrend := []analyticsTrend{}
	productCounts := []vendorAnalyticsProductCount{}
	if len(productIDs) > 0 {
		productEvents := h.DB.Model(&model.AnalyticsEvent{}).
			Where("created_at >= ? AND event_type = ? AND content_type = ? AND content_id IN ?", start, "page_view", "product", productIDs)
		productEvents.Session(&gorm.Session{}).Count(&productPV)
		productEvents.Session(&gorm.Session{}).Distinct("visitor_hash").Count(&productUV)
		h.DB.Model(&model.AnalyticsEvent{}).
			Select("DATE(created_at) AS date, COUNT(*) AS pv, COUNT(DISTINCT visitor_hash) AS uv").
			Where("created_at >= ? AND event_type = ? AND content_type = ? AND content_id IN ?", start, "page_view", "product", productIDs).
			Group("DATE(created_at)").Order("date asc").Scan(&productTrend)
		h.DB.Model(&model.AnalyticsEvent{}).
			Select("content_id AS product_id, COUNT(*) AS pv, COUNT(DISTINCT visitor_hash) AS uv").
			Where("created_at >= ? AND event_type = ? AND content_type = ? AND content_id IN ?", start, "page_view", "product", productIDs).
			Group("content_id").Order("pv desc, content_id asc").Scan(&productCounts)
	}
	countsByID := make(map[uint]vendorAnalyticsProductCount, len(productCounts))
	for _, count := range productCounts {
		countsByID[count.ProductID] = count
	}
	items := make([]vendorAnalyticsProduct, 0, len(products))
	for _, product := range products {
		path := "/products/" + product.Slug
		if strings.TrimSpace(product.Slug) == "" {
			path = "/products/" + strconv.FormatUint(uint64(product.ID), 10)
		}
		count := countsByID[product.ID]
		items = append(items, vendorAnalyticsProduct{ProductID: product.ID, Name: product.Name, Path: path, PV: count.PV, UV: count.UV})
	}
	// Keep the public ranking stable by traffic, then product name for empty/tied data.
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].PV != items[j].PV {
			return items[i].PV > items[j].PV
		}
		if items[i].UV != items[j].UV {
			return items[i].UV > items[j].UV
		}
		return items[i].Name < items[j].Name
	})

	showroom, err := h.showroomAnalytics(vendorID, start)
	if err != nil {
		Fail(c, 500, 500, "推广统计暂时不可用，请稍后重试")
		return
	}
	OK(c, gin.H{
		"showroom": showroom,
		"days":     days,
		"vendor":   gin.H{"id": vendor.ID, "name": vendor.Name, "pv": vendorPV, "uv": vendorUV, "contacts": contacts, "contactEvents": contactEvents, "trend": vendorTrend},
		"products": gin.H{"pv": productPV, "uv": productUV, "trend": productTrend, "items": items},
	})
}

func (h AnalyticsHandler) Summary(c *gin.Context) {
	days := analyticsDays(c.DefaultQuery("days", "30"))
	start := time.Now().AddDate(0, 0, -days)
	base := h.DB.Model(&model.AnalyticsEvent{}).Where("created_at >= ?", start)
	var pv, uv, conversions int64
	base.Session(&gorm.Session{}).Where("event_type = ?", "page_view").Count(&pv)
	base.Session(&gorm.Session{}).Where("event_type = ?", "page_view").Distinct("visitor_hash").Count(&uv)
	base.Session(&gorm.Session{}).Where("event_type <> ?", "page_view").Count(&conversions)
	trend := []analyticsTrend{}
	h.DB.Model(&model.AnalyticsEvent{}).Select("DATE(created_at) AS date, COUNT(*) AS pv, COUNT(DISTINCT visitor_hash) AS uv").Where("created_at >= ? AND event_type = ?", start, "page_view").Group("DATE(created_at)").Order("date asc").Scan(&trend)
	provinces := []analyticsCount{}
	h.DB.Model(&model.AnalyticsEvent{}).Select("province AS label, COUNT(*) AS count").Where("created_at >= ? AND event_type = ?", start, "page_view").Group("province").Order("count desc, province asc").Limit(12).Scan(&provinces)
	contents := []analyticsContentCount{}
	h.DB.Model(&model.AnalyticsEvent{}).Select("path, content_type, content_id, COUNT(*) AS count").Where("created_at >= ? AND event_type = ? AND content_type <> ''", start, "page_view").Group("path, content_type, content_id").Order("count desc, path asc").Limit(12).Scan(&contents)
	conversionRows := []analyticsCount{}
	h.DB.Model(&model.AnalyticsEvent{}).Select("event_type AS label, COUNT(*) AS count").Where("created_at >= ? AND event_type <> ?", start, "page_view").Group("event_type").Order("count desc, event_type asc").Scan(&conversionRows)
	rate := 0.0
	if pv > 0 {
		rate = float64(conversions) / float64(pv)
	}
	OK(c, gin.H{"days": days, "pv": pv, "uv": uv, "conversions": conversions, "conversionRate": rate, "trend": trend, "provinces": provinces, "contents": contents, "conversionEvents": conversionRows})
}
