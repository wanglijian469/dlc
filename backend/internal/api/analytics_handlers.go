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
	"vendor_website_click":    true,
	"search_zero_results":     true,
	"not_found":               true,
}

type AnalyticsHandler struct {
	DB     *gorm.DB
	Config config.Config
}

type analyticsRequest struct {
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
	visitor := h.visitorHash(c)
	window := analyticsWindow(req.EventType, time.Now())
	event := model.AnalyticsEvent{
		EventType: req.EventType, Path: path, ContentType: contentType, ContentID: contentID,
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
	case "contact_phone_click", "contact_wechat_copy", "vendor_website_click":
		return contentType == "vendor"
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

func (h AnalyticsHandler) Summary(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days != 7 && days != 30 && days != 90 {
		days = 30
	}
	start := time.Now().AddDate(0, 0, -days)
	base := h.DB.Model(&model.AnalyticsEvent{}).Where("created_at >= ?", start)
	var pv, uv, conversions int64
	base.Where("event_type = ?", "page_view").Count(&pv)
	base.Where("event_type = ?", "page_view").Distinct("visitor_hash").Count(&uv)
	base.Where("event_type <> ?", "page_view").Count(&conversions)
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
