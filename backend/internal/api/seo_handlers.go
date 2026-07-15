package api

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterSEORoutes exposes crawler discovery documents without coupling them
// to the SPA fallback route.
func RegisterSEORoutes(router *gin.Engine, db *gorm.DB, cfg config.Config) {
	h := SEOHandler{DB: db, Config: cfg, HomeService: service.HomeService{DB: db}}
	router.GET("/robots.txt", h.Robots)
	router.GET("/sitemap.xml", h.Sitemap)
}

type SEOHandler struct {
	DB          *gorm.DB
	Config      config.Config
	HomeService service.HomeService
}

type seoDocument struct {
	Title       string
	Description string
	Canonical   string
	NoIndex     bool
	BodyTitle   string
	BodyText    string
	Breadcrumbs []seoBreadcrumb
	Schema      any
}

type seoBreadcrumb struct {
	Name string
	URL  string
}

func (h SEOHandler) Robots(c *gin.Context) {
	base := h.baseURL(c)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, "User-agent: *\nAllow: /\nDisallow: /admin/\nDisallow: /api/admin/\nDisallow: /search\n\nSitemap: %s/sitemap.xml\n", base)
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemapDocument struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

func (h SEOHandler) Sitemap(c *gin.Context) {
	base := h.baseURL(c)
	urls := []sitemapURL{{Loc: base + "/"}, {Loc: base + "/vendors"}, {Loc: base + "/products"}, {Loc: base + "/service"}, {Loc: base + "/guides"}}
	appendURL := func(path string, updatedAt time.Time) {
		entry := sitemapURL{Loc: base + path}
		if !updatedAt.IsZero() {
			entry.LastMod = updatedAt.UTC().Format("2006-01-02")
		}
		urls = append(urls, entry)
	}
	var vendors []model.Vendor
	h.DB.Select("id, updated_at").Where("is_visible = ? AND publication_status = ?", true, "published").Find(&vendors)
	for _, item := range vendors {
		appendURL(fmt.Sprintf("/vendors/%d", item.ID), item.UpdatedAt)
	}
	var products []model.Product
	visibleProductQuery(h.DB).Select("products.id, products.updated_at").Find(&products)
	for _, item := range products {
		appendURL(fmt.Sprintf("/products/%d", item.ID), item.UpdatedAt)
	}
	var categories []model.Category
	h.DB.Where("is_enabled = ?", true).Find(&categories)
	categoryByID := make(map[uint]model.Category, len(categories))
	for _, item := range categories {
		categoryByID[item.ID] = item
	}
	for _, item := range categories {
		if item.ParentID != 0 {
			if parent, ok := categoryByID[item.ParentID]; !ok || !parent.IsEnabled || parent.ParentID != 0 {
				continue
			}
		}
		appendURL(fmt.Sprintf("/products?categoryId=%d", item.ID), item.UpdatedAt)
	}
	var pages []model.ContentPage
	h.DB.Where("is_enabled = ?", true).Find(&pages)
	for _, item := range pages {
		path := "/" + item.Slug
		if item.PageType == "article" {
			path = "/guides/" + item.Slug
		}
		appendURL(path, item.UpdatedAt)
	}
	payload, err := xml.MarshalIndent(sitemapDocument{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls}, "", "  ")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, xml.Header+string(payload))
}

// RenderSEOApp decorates the SPA entry response with route-specific metadata,
// JSON-LD, and a no-JavaScript semantic fallback. It is returned to every
// visitor at the same canonical URL, never based on crawler user agent.
func RenderSEOApp(c *gin.Context, db *gorm.DB, cfg config.Config, publicDir string) bool {
	if c.Request.Method != http.MethodGet || strings.HasPrefix(c.Request.URL.Path, "/admin") || c.Request.URL.Path == "/robots.txt" || c.Request.URL.Path == "/sitemap.xml" {
		return false
	}
	h := SEOHandler{DB: db, Config: cfg, HomeService: service.HomeService{DB: db}}
	doc, ok := h.document(c)
	if !ok {
		return false
	}
	index, err := os.ReadFile(filepath.Join(publicDir, "index.html"))
	if err != nil {
		return false
	}
	page := injectSEOHTML(string(index), doc, h.meta(c.Request.Context()))
	c.Header("Cache-Control", "public, no-cache")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
	return true
}

func (h SEOHandler) document(c *gin.Context) (seoDocument, bool) {
	meta := h.meta(c.Request.Context())
	base := h.baseURL(c)
	path := c.Request.URL.Path
	canonical := base + path
	doc := seoDocument{Title: meta.DefaultSEOTitle, Description: meta.DefaultSEODescription, Canonical: canonical, BodyTitle: meta.SiteName, BodyText: meta.DefaultSEODescription}
	doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "WebSite", "name": meta.SiteName, "url": base}
	if path == "/search" {
		doc.NoIndex = true
		doc.Title = "站内搜索｜" + meta.SiteName
		doc.BodyTitle = "站内搜索"
		return doc, true
	}
	if path == "/" {
		doc.Schema = []any{doc.Schema, map[string]any{"@context": "https://schema.org", "@type": "Organization", "name": meta.SiteName, "url": base}}
		return doc, true
	}
	if path == "/vendors" {
		doc.Title = "农机配件厂家目录｜" + meta.SiteName
		doc.BodyTitle = "农机配件厂家目录"
		doc.NoIndex = len(c.Request.URL.Query()) > 0
		return doc, true
	}
	if path == "/products" {
		if categoryID, err := strconv.ParseUint(c.Query("categoryId"), 10, 64); err == nil && categoryID > 0 {
			var category model.Category
			if h.DB.First(&category, uint(categoryID)).Error == nil && category.IsEnabled {
				doc.Canonical = base + "/products?categoryId=" + strconv.FormatUint(categoryID, 10)
				doc.Title = fallbackSEO(category.SEOTitle, category.Name+"｜农机配件分类") + "｜" + meta.SiteName
				doc.Description = fallbackSEO(category.SEODescription, "查看"+category.Name+"相关农机配件、适配信息与供应厂商资料。")
				doc.BodyTitle, doc.BodyText = category.Name, doc.Description
				doc.Breadcrumbs = []seoBreadcrumb{{Name: "配件产品", URL: base + "/products"}}
				if category.ParentID != 0 {
					var parent model.Category
					if h.DB.First(&parent, category.ParentID).Error == nil && parent.IsEnabled {
						doc.Breadcrumbs = append(doc.Breadcrumbs, seoBreadcrumb{Name: parent.Name, URL: base + "/products?categoryId=" + strconv.FormatUint(uint64(parent.ID), 10)})
					}
				}
				doc.Breadcrumbs = append(doc.Breadcrumbs, seoBreadcrumb{Name: category.Name, URL: doc.Canonical})
				doc.Schema = withBreadcrumbSchema(doc.Schema, doc.Breadcrumbs)
				for key := range c.Request.URL.Query() {
					if key != "categoryId" {
						doc.NoIndex = true
						break
					}
				}
				return doc, true
			}
		}
		doc.Title = "农机配件产品目录｜" + meta.SiteName
		doc.BodyTitle = "农机配件产品目录"
		doc.NoIndex = len(c.Request.URL.Query()) > 0
		return doc, true
	}
	if path == "/service" {
		doc.Title = "农机配件加工服务｜" + meta.SiteName
		doc.BodyTitle = "农机配件加工服务"
		doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "Service", "name": doc.BodyTitle, "provider": map[string]any{"@type": "Organization", "name": meta.SiteName}}
		return doc, true
	}
	if path == "/guides" {
		doc.Title = "农机配件行业指南｜" + meta.SiteName
		doc.BodyTitle = "农机配件行业指南"
		doc.BodyText = "围绕选型、适配、保养与采购验收整理真实业务知识。"
		return doc, true
	}
	if id, ok := routeID(path, "/vendors/"); ok {
		var vendor model.Vendor
		if h.DB.First(&vendor, "id = ? AND is_visible = ? AND publication_status = ?", id, true, "published").Error != nil {
			return seoDocument{}, false
		}
		doc.Title = fallbackSEO(vendor.SEOTitle, vendor.Name+"｜农机配件厂家") + "｜" + meta.SiteName
		doc.Description = fallbackSEO(vendor.SEODescription, firstNonEmpty(vendor.Description, vendor.MainProducts, "查看厂商主营产品与服务能力。"))
		doc.BodyTitle, doc.BodyText = vendor.Name, doc.Description
		doc.Breadcrumbs = []seoBreadcrumb{{Name: "厂商目录", URL: base + "/vendors"}, {Name: vendor.Name, URL: canonical}}
		doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "Organization", "name": vendor.Name, "url": canonical, "description": doc.Description, "address": map[string]any{"@type": "PostalAddress", "addressRegion": vendor.Province, "addressLocality": vendor.City}}
		doc.Schema = withBreadcrumbSchema(doc.Schema, doc.Breadcrumbs)
		return doc, true
	}
	if id, ok := routeID(path, "/products/"); ok {
		var product model.Product
		if visibleProductQuery(h.DB).Preload("Category").First(&product, "products.id = ?", id).Error != nil {
			return seoDocument{}, false
		}
		doc.Title = fallbackSEO(product.SEOTitle, product.Name+"｜"+product.Category.Name) + "｜" + meta.SiteName
		doc.Description = fallbackSEO(product.SEODescription, firstNonEmpty(product.DetailContent, product.Description, product.CompatibleModels, "查看产品规格与适配信息。"))
		doc.BodyTitle, doc.BodyText = product.Name, doc.Description
		doc.Breadcrumbs = []seoBreadcrumb{{Name: "配件产品", URL: base + "/products"}, {Name: product.Category.Name, URL: base + "/products?categoryId=" + strconv.FormatUint(uint64(product.CategoryID), 10)}, {Name: product.Name, URL: canonical}}
		doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "Product", "name": product.Name, "description": doc.Description, "url": canonical, "category": product.Category.Name, "image": absoluteURL(base, product.Image)}
		doc.Schema = withBreadcrumbSchema(doc.Schema, doc.Breadcrumbs)
		return doc, true
	}
	if slug, ok := routeSlug(path, "/guides/"); ok {
		return h.contentDocument(slug, "article", base)
	}
	if slug, ok := routeSlug(path, "/"); ok && !strings.Contains(slug, "/") {
		return h.contentDocument(slug, "page", base)
	}
	return seoDocument{}, false
}

func (h SEOHandler) contentDocument(slug, pageType, base string) (seoDocument, bool) {
	var page model.ContentPage
	if h.DB.Where("slug = ? AND page_type = ? AND is_enabled = ?", slug, pageType, true).First(&page).Error != nil {
		return seoDocument{}, false
	}
	path := "/" + slug
	if pageType == "article" {
		path = "/guides/" + slug
	}
	description := fallbackSEO(page.SEODescription, firstNonEmpty(page.Summary, page.Content, "农机配件行业指南。"))
	meta := h.meta(context.Background())
	doc := seoDocument{Title: fallbackSEO(page.SEOTitle, page.Title) + "｜" + meta.SiteName, Description: description, Canonical: base + path, BodyTitle: page.Title, BodyText: firstNonEmpty(page.Content, page.Summary), Breadcrumbs: []seoBreadcrumb{{Name: pageTypeLabel(pageType), URL: base + "/" + map[bool]string{true: "guides", false: ""}[pageType == "article"]}}}
	doc.Schema = map[string]any{"@context": "https://schema.org", "@type": map[bool]string{true: "Article", false: "WebPage"}[pageType == "article"], "headline": page.Title, "description": description, "url": doc.Canonical, "datePublished": page.PublishedAt, "dateModified": page.UpdatedAt, "author": map[string]any{"@type": "Organization", "name": firstNonEmpty(page.AuthorName, meta.SiteName)}}
	doc.Schema = withBreadcrumbSchema(doc.Schema, doc.Breadcrumbs)
	return doc, true
}

func (h SEOHandler) meta(ctx context.Context) service.SiteMeta { return h.HomeService.SiteMeta(ctx) }

func (h SEOHandler) baseURL(c *gin.Context) string {
	meta := h.HomeService.SiteMeta(c.Request.Context())
	if value := strings.TrimRight(strings.TrimSpace(meta.SiteURL), "/"); value != "" {
		return value
	}
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
	}
	return scheme + "://" + c.Request.Host
}

func injectSEOHTML(index string, doc seoDocument, meta service.SiteMeta) string {
	title := template.HTMLEscapeString(doc.Title)
	description := template.HTMLEscapeString(trimSEO(doc.Description, 260))
	canonical := template.HTMLEscapeString(doc.Canonical)
	schema, _ := json.Marshal(doc.Schema)
	head := fmt.Sprintf(`<title>%s</title><meta name="description" content="%s"><link rel="canonical" href="%s"><meta property="og:title" content="%s"><meta property="og:description" content="%s"><meta property="og:url" content="%s">`, title, description, canonical, title, description, canonical)
	if doc.NoIndex {
		head += `<meta name="robots" content="noindex,follow">`
	}
	if meta.BaiduVerification != "" {
		head += `<meta name="baidu-site-verification" content="` + template.HTMLEscapeString(meta.BaiduVerification) + `">`
	}
	if meta.GoogleVerification != "" {
		head += `<meta name="google-site-verification" content="` + template.HTMLEscapeString(meta.GoogleVerification) + `">`
	}
	head += `<script type="application/ld+json">` + strings.ReplaceAll(string(schema), "</", "<\\/") + `</script>`
	if strings.Contains(index, "<title>大陆农机配件</title>") {
		index = strings.Replace(index, "<title>大陆农机配件</title>", head, 1)
	} else {
		index = strings.Replace(index, "</head>", head+"</head>", 1)
	}
	return strings.Replace(index, `<div id="root"></div>`, `<div id="root"></div>`+semanticFallback(doc), 1)
}

func semanticFallback(doc seoDocument) string {
	var content bytes.Buffer
	content.WriteString(`<noscript><main><nav aria-label="面包屑">`)
	for _, item := range doc.Breadcrumbs {
		content.WriteString(`<a href="` + template.HTMLEscapeString(item.URL) + `">` + template.HTMLEscapeString(item.Name) + `</a> / `)
	}
	content.WriteString(`</nav><h1>` + template.HTMLEscapeString(doc.BodyTitle) + `</h1><p>` + template.HTMLEscapeString(doc.BodyText) + `</p><p><a href="/join">厂商入驻</a></p></main></noscript>`)
	return content.String()
}

func routeID(path, prefix string) (uint, bool) {
	value, ok := routeSlug(path, prefix)
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseUint(value, 10, 64)
	return uint(id), err == nil && id > 0
}
func routeSlug(path, prefix string) (string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	value := strings.TrimPrefix(path, prefix)
	return value, value != "" && !strings.Contains(value, "/")
}
func fallbackSEO(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(fallback)
}
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
func trimSEO(value string, max int) string {
	runes := []rune(strings.Join(strings.Fields(value), " "))
	if len(runes) > max {
		return string(runes[:max])
	}
	return string(runes)
}
func absoluteURL(base, value string) string {
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed.IsAbs() {
		return value
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(value, "/")
}
func pageTypeLabel(pageType string) string {
	if pageType == "article" {
		return "行业指南"
	}
	return "平台页面"
}

func withBreadcrumbSchema(schema any, breadcrumbs []seoBreadcrumb) any {
	if len(breadcrumbs) == 0 {
		return schema
	}
	items := make([]map[string]any, 0, len(breadcrumbs))
	for index, item := range breadcrumbs {
		items = append(items, map[string]any{"@type": "ListItem", "position": index + 1, "name": item.Name, "item": item.URL})
	}
	return []any{schema, map[string]any{"@context": "https://schema.org", "@type": "BreadcrumbList", "itemListElement": items}}
}
