package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type seoIssueSummary struct {
	MissingTitle       int64 `json:"missingTitle"`
	MissingDescription int64 `json:"missingDescription"`
	MissingImage       int64 `json:"missingImage"`
	DuplicateTitles    int64 `json:"duplicateTitles"`
	DuplicateDescs     int64 `json:"duplicateDescriptions"`
	Drafts             int64 `json:"drafts"`
	InReview           int64 `json:"inReview"`
	Redirects          int64 `json:"redirects"`
	Recent404          int64 `json:"recent404"`
	ZeroResultSearches int64 `json:"zeroResultSearches"`
}

func (h AdminHandler) SEOStatus(c *gin.Context) {
	var result seoIssueSummary
	now := time.Now()
	h.DB.Model(&model.Vendor{}).Where("publication_status = ? AND is_visible = ? AND (seo_title = '' OR seo_title IS NULL)", "published", true).Count(&result.MissingTitle)
	var productMissingTitle, pageMissingTitle int64
	h.DB.Model(&model.Product{}).Where("publication_status = ? AND (seo_title = '' OR seo_title IS NULL)", "published").Count(&productMissingTitle)
	h.DB.Model(&model.ContentPage{}).Where("publication_status = ? AND is_enabled = ? AND (seo_title = '' OR seo_title IS NULL)", "published", true).Count(&pageMissingTitle)
	result.MissingTitle += productMissingTitle + pageMissingTitle

	h.DB.Model(&model.Vendor{}).Where("publication_status = ? AND is_visible = ? AND (seo_description = '' OR seo_description IS NULL)", "published", true).Count(&result.MissingDescription)
	var productMissingDesc, pageMissingDesc int64
	h.DB.Model(&model.Product{}).Where("publication_status = ? AND (seo_description = '' OR seo_description IS NULL)", "published").Count(&productMissingDesc)
	h.DB.Model(&model.ContentPage{}).Where("publication_status = ? AND is_enabled = ? AND (seo_description = '' OR seo_description IS NULL)", "published", true).Count(&pageMissingDesc)
	result.MissingDescription += productMissingDesc + pageMissingDesc

	h.DB.Model(&model.Vendor{}).Where("publication_status = ? AND is_visible = ? AND (logo = '' OR cover_image = '')", "published", true).Count(&result.MissingImage)
	var productMissingImage, pageMissingImage int64
	h.DB.Model(&model.Product{}).Where("publication_status = ? AND (image = '' OR image IS NULL)", "published").Count(&productMissingImage)
	h.DB.Model(&model.ContentPage{}).Where("publication_status = ? AND is_enabled = ? AND page_type = ? AND (cover_image = '' OR cover_image IS NULL)", "published", true, "article").Count(&pageMissingImage)
	result.MissingImage += productMissingImage + pageMissingImage

	result.DuplicateTitles = duplicateMetadataCount(h.DB, "seo_title")
	result.DuplicateDescs = duplicateMetadataCount(h.DB, "seo_description")
	h.DB.Model(&model.ContentRevision{}).Where("status IN ?", []string{"draft", "rejected"}).Count(&result.Drafts)
	h.DB.Model(&model.ContentRevision{}).Where("status = ?", "in_review").Count(&result.InReview)
	h.DB.Model(&model.SEORedirect{}).Count(&result.Redirects)
	h.DB.Model(&model.AnalyticsEvent{}).Where("event_type = ? AND created_at >= ?", "not_found", now.AddDate(0, 0, -30)).Count(&result.Recent404)
	h.DB.Model(&model.AnalyticsEvent{}).Where("event_type = ? AND created_at >= ?", "search_zero_results", now.AddDate(0, 0, -30)).Count(&result.ZeroResultSearches)

	OK(c, gin.H{
		"issues":          result,
		"sitemap":         gin.H{"index": "/sitemap.xml", "sections": []string{"static", "categories", "vendors", "products", "content"}},
		"baiduSubmission": gin.H{"enabled": strings.TrimSpace(h.Config.BaiduPushToken) != "", "configured": strings.TrimSpace(h.Config.BaiduPushToken) != ""},
	})
}

func duplicateMetadataCount(db *gorm.DB, column string) int64 {
	if column != "seo_title" && column != "seo_description" {
		return 0
	}
	var rows []struct{ Value string }
	query := "SELECT " + column + " AS value FROM (SELECT " + column + " FROM vendors WHERE deleted_at IS NULL AND publication_status = 'published' UNION ALL SELECT " + column + " FROM products WHERE deleted_at IS NULL AND publication_status = 'published' UNION ALL SELECT " + column + " FROM content_pages WHERE deleted_at IS NULL AND publication_status = 'published') metadata WHERE " + column + " <> '' GROUP BY " + column + " HAVING COUNT(*) > 1"
	db.Raw(query).Scan(&rows)
	return int64(len(rows))
}

func (h AdminHandler) SEOConsoleHealth(c *gin.Context) {
	if h.DB == nil {
		Fail(c, http.StatusServiceUnavailable, 503, "SEO 控制台不可用")
		return
	}
	h.SEOStatus(c)
}

func (h AdminHandler) SubmitBaiduURLs(c *gin.Context) {
	token := strings.TrimSpace(h.Config.BaiduPushToken)
	meta := (service.HomeService{DB: h.DB}).SiteMeta(c.Request.Context())
	parsed, err := url.Parse(strings.TrimRight(meta.SiteURL, "/"))
	if token == "" || err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		Fail(c, http.StatusBadRequest, 400, "请先配置正式 HTTPS siteUrl 和 BAIDU_PUSH_TOKEN")
		return
	}
	seo := SEOHandler{DB: h.DB, HomeService: service.HomeService{DB: h.DB}}
	sections := seo.sitemapSections(strings.TrimRight(meta.SiteURL, "/"))
	urls := make([]string, 0, 2000)
	for _, section := range []string{"static", "categories", "vendors", "products", "content"} {
		for _, item := range sections[section] {
			urls = append(urls, item.Loc)
			if len(urls) == 2000 {
				break
			}
		}
		if len(urls) == 2000 {
			break
		}
	}
	endpoint := fmt.Sprintf("https://data.zz.baidu.com/urls?site=%s&token=%s", url.QueryEscape(parsed.Scheme+"://"+parsed.Host), url.QueryEscape(token))
	req, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, bytes.NewBufferString(strings.Join(urls, "\n")))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		Fail(c, http.StatusBadGateway, 502, "百度资源提交请求失败")
		return
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		Fail(c, http.StatusBadGateway, 502, "百度资源提交未成功")
		return
	}
	logOperation(h.DB, c.GetString("username"), "submit-baidu", "seo", 0)
	OK(c, gin.H{"submitted": len(urls), "response": string(body)})
}
