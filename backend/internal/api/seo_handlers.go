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
	router.GET("/sitemaps/:section", h.SitemapSection)
}

type SEOHandler struct {
	DB          *gorm.DB
	Config      config.Config
	HomeService service.HomeService
}

type seoDocument struct {
	Title       string
	Description string
	Keywords    string
	Canonical   string
	Image       string
	OGType      string
	NoIndex     bool
	NoFollow    bool
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
	cfg := DefaultProtectionConfig()
	if h.DB != nil {
		var row model.SiteConfig
		if h.DB.Where("config_key = ?", protectionConfigKey).First(&row).Error == nil {
			_ = json.Unmarshal([]byte(row.ConfigValue), &cfg)
		}
	}
	var body strings.Builder
	seen := map[string]bool{}
	for _, agent := range cfg.BlockedAIAgents {
		agent = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(agent, "\r", ""), "\n", ""))
		if agent == "" || seen[strings.ToLower(agent)] {
			continue
		}
		seen[strings.ToLower(agent)] = true
		body.WriteString("User-agent: " + agent + "\nDisallow: /\n\n")
	}
	searchRules := "Allow: /\nDisallow: /admin/\nDisallow: /account/\nDisallow: /api/admin/\nDisallow: /search\n"
	body.WriteString("User-agent: Googlebot\n" + searchRules + "\nUser-agent: Baiduspider\n" + searchRules + "\n")
	body.WriteString("User-agent: *\n" + searchRules + "\n")
	body.WriteString("Sitemap: " + base + "/sitemap.xml\n")
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, body.String())
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

type sitemapIndexEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemapIndexDocument struct {
	XMLName xml.Name            `xml:"sitemapindex"`
	Xmlns   string              `xml:"xmlns,attr"`
	Items   []sitemapIndexEntry `xml:"sitemap"`
}

const sitemapChunkSize = 45000

func (h SEOHandler) Sitemap(c *gin.Context) {
	base := h.baseURL(c)
	sections := h.sitemapSections(base)
	index := sitemapIndexDocument{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, name := range []string{"static", "categories", "vendors", "products", "content"} {
		rows := sections[name]
		pages := (len(rows) + sitemapChunkSize - 1) / sitemapChunkSize
		if pages == 0 {
			pages = 1
		}
		for page := 1; page <= pages; page++ {
			entry := sitemapIndexEntry{Loc: fmt.Sprintf("%s/sitemaps/%s-%d.xml", base, name, page)}
			if len(rows) > 0 {
				entry.LastMod = rows[len(rows)-1].LastMod
			}
			index.Items = append(index.Items, entry)
		}
	}
	payload, err := xml.MarshalIndent(index, "", "  ")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, xml.Header+string(payload))
}

func (h SEOHandler) SitemapSection(c *gin.Context) {
	section := strings.TrimSuffix(c.Param("section"), ".xml")
	cut := strings.LastIndex(section, "-")
	if cut < 1 {
		c.Status(http.StatusNotFound)
		return
	}
	page, err := strconv.Atoi(section[cut+1:])
	if err != nil || page < 1 {
		c.Status(http.StatusNotFound)
		return
	}
	name := section[:cut]
	rows, ok := h.sitemapSections(h.baseURL(c))[name]
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	start := (page - 1) * sitemapChunkSize
	if start > len(rows) {
		c.Status(http.StatusNotFound)
		return
	}
	end := start + sitemapChunkSize
	if end > len(rows) {
		end = len(rows)
	}
	payload, err := xml.MarshalIndent(sitemapDocument{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: rows[start:end]}, "", "  ")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, xml.Header+string(payload))
}

func (h SEOHandler) sitemapSections(base string) map[string][]sitemapURL {
	sections := map[string][]sitemapURL{"static": {{Loc: base + "/"}, {Loc: base + "/vendors"}, {Loc: base + "/products"}, {Loc: base + "/service"}, {Loc: base + "/guides"}}, "categories": {}, "vendors": {}, "products": {}, "content": {}}
	appendURL := func(section, path string, publishedAt *time.Time, updatedAt time.Time) {
		entry := sitemapURL{Loc: base + path}
		stamp := updatedAt
		if publishedAt != nil {
			stamp = *publishedAt
		}
		if !stamp.IsZero() {
			entry.LastMod = stamp.UTC().Format("2006-01-02")
		}
		sections[section] = append(sections[section], entry)
	}
	var vendors []model.Vendor
	publishedVendorQuery(h.DB).Select("id, slug, published_at, updated_at").Order("id asc").Find(&vendors)
	for _, item := range vendors {
		if item.Slug != "" {
			appendURL("vendors", "/v/"+item.Slug, item.PublishedAt, item.UpdatedAt)
		}
	}
	var products []model.Product
	visibleProductQuery(h.DB).Select("products.id, products.slug, products.published_at, products.updated_at").Order("products.id asc").Find(&products)
	for _, item := range products {
		if item.Slug != "" {
			appendURL("products", "/products/"+item.Slug, item.PublishedAt, item.UpdatedAt)
		}
	}
	var categories []model.Category
	publishedCategoryQuery(h.DB).Order("id asc").Find(&categories)
	for _, item := range categories {
		if item.Slug != "" {
			appendURL("categories", "/products/category/"+item.Slug, item.PublishedAt, item.UpdatedAt)
		}
	}
	var pages []model.ContentPage
	h.DB.Where("is_enabled = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)", true, "published", time.Now()).Order("id asc").Find(&pages)
	for _, item := range pages {
		path := "/" + item.Slug
		if item.PageType == "article" {
			path = "/guides/" + item.Slug
		}
		appendURL("content", path, item.PublishedAt, item.UpdatedAt)
	}
	return sections
}

// RenderSEOApp decorates the SPA entry response with route-specific metadata,
// JSON-LD, and a no-JavaScript semantic fallback. It is returned to every
// visitor at the same canonical URL, never based on crawler user agent.
func RenderSEOApp(c *gin.Context, db *gorm.DB, cfg config.Config, publicDir string) bool {
	if (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) || strings.HasPrefix(c.Request.URL.Path, "/api/") || c.Request.URL.Path == "/robots.txt" || c.Request.URL.Path == "/sitemap.xml" || strings.HasPrefix(c.Request.URL.Path, "/sitemaps/") {
		return false
	}
	h := SEOHandler{DB: db, Config: cfg, HomeService: service.HomeService{DB: db}}
	if destination, status, ok := h.redirect(c); ok {
		c.Redirect(status, destination)
		return true
	}
	status := http.StatusOK
	doc, ok := h.document(c)
	if strings.HasPrefix(c.Request.URL.Path, "/admin") || strings.HasPrefix(c.Request.URL.Path, "/account/") {
		doc, ok = seoDocument{Title: "管理中心", Description: "平台内容管理中心", Canonical: h.baseURL(c) + c.Request.URL.Path, NoIndex: true, NoFollow: true, BodyTitle: "管理中心", BodyText: "此页面需要登录。"}, true
	}
	if !ok {
		status = http.StatusNotFound
		doc = seoDocument{Title: "页面不存在", Description: "请求的页面不存在或内容尚未发布。", Canonical: h.baseURL(c) + c.Request.URL.Path, NoIndex: true, NoFollow: true, BodyTitle: "页面不存在", BodyText: "请返回首页、产品目录或厂商目录继续浏览。"}
	}
	index, err := os.ReadFile(filepath.Join(publicDir, "index.html"))
	if err != nil {
		return false
	}
	page := injectSEOHTML(string(index), doc, h.meta(c.Request.Context()))
	c.Header("Cache-Control", "public, no-cache")
	if doc.NoIndex {
		directive := "noindex, follow"
		if doc.NoFollow {
			directive = "noindex, nofollow"
		}
		c.Header("X-Robots-Tag", directive)
	}
	if c.Request.Method == http.MethodHead {
		c.Data(status, "text/html; charset=utf-8", nil)
	} else {
		c.Data(status, "text/html; charset=utf-8", []byte(page))
	}
	return true
}

func (h SEOHandler) document(c *gin.Context) (seoDocument, bool) {
	meta := h.meta(c.Request.Context())
	base := h.baseURL(c)
	path := c.Request.URL.Path
	canonical := base + path
	doc := seoDocument{Title: meta.DefaultSEOTitle, Description: meta.DefaultSEODescription, Canonical: canonical, BodyTitle: meta.SiteName, BodyText: meta.DefaultSEODescription}
	doc.OGType = "website"
	doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "WebSite", "name": meta.SiteName, "url": base}
	if path == "/search" {
		doc.NoIndex = true
		doc.Title = "站内搜索｜" + meta.SiteName
		doc.BodyTitle = "站内搜索"
		return doc, true
	}
	if path == "/" {
		doc.Schema = []any{doc.Schema, map[string]any{"@context": "https://schema.org", "@type": "Organization", "name": meta.SiteName, "url": base, "logo": absoluteURL(base, meta.BrandLogo)}}
		return doc, true
	}
	if path == "/vendors" {
		doc.Title = "农机配件厂家目录｜" + meta.SiteName
		doc.BodyTitle = "农机配件厂家目录"
		doc.NoIndex = len(c.Request.URL.Query()) > 0
		doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "ItemList", "name": doc.BodyTitle, "url": doc.Canonical}
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
		doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "ItemList", "name": doc.BodyTitle, "url": doc.Canonical}
		return doc, true
	}
	if slug, ok := routeSlug(path, "/products/category/"); ok {
		var category model.Category
		if publishedCategoryQuery(h.DB).First(&category, "slug = ?", slug).Error != nil {
			return seoDocument{}, false
		}
		doc.Canonical = base + "/products/category/" + category.Slug
		doc.Title = fallbackSEO(category.SEOTitle, category.Name+"｜农机配件分类") + "｜" + meta.SiteName
		doc.Description = fallbackSEO(category.SEODescription, "查看"+category.Name+"相关农机配件、适配信息与供应厂商资料。")
		doc.BodyTitle, doc.BodyText = category.Name, doc.Description
		doc.Breadcrumbs = []seoBreadcrumb{{Name: "配件产品", URL: base + "/products"}}
		if category.ParentID != 0 {
			var parent model.Category
			if publishedCategoryQuery(h.DB).First(&parent, category.ParentID).Error == nil {
				doc.Breadcrumbs = append(doc.Breadcrumbs, seoBreadcrumb{Name: parent.Name, URL: base + "/products/category/" + parent.Slug})
			}
		}
		doc.Breadcrumbs = append(doc.Breadcrumbs, seoBreadcrumb{Name: category.Name, URL: doc.Canonical})
		doc.Schema = withBreadcrumbSchema(map[string]any{"@context": "https://schema.org", "@type": "CollectionPage", "name": category.Name, "description": doc.Description, "url": doc.Canonical}, doc.Breadcrumbs)
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
	if path == "/contact" || path == "/feedback" {
		label := map[string]string{"/contact": "联系我们", "/feedback": "反馈建议"}[path]
		doc.Title = label + "｜" + meta.SiteName
		doc.BodyTitle = label
		doc.Description = "联系平台运营方，反馈厂商资料、产品信息或平台使用问题。"
		doc.BodyText = doc.Description
		return doc, true
	}
	if slug, ok := routeSlug(path, "/v/"); ok {
		var vendor model.Vendor
		if publishedVendorQuery(h.DB).First(&vendor, "slug = ?", slug).Error != nil {
			return seoDocument{}, false
		}
		doc.Title = appendSiteName(fallbackSEO(vendor.SEOTitle, vendor.Name+"｜农机配件厂家"), meta.SiteName)
		doc.Description = fallbackSEO(vendor.SEODescription, firstNonEmpty(vendor.Description, vendor.MainProducts, "查看厂商主营产品与服务能力。"))
		doc.BodyTitle, doc.BodyText = vendor.Name, doc.Description
		doc.Breadcrumbs = []seoBreadcrumb{{Name: "厂商目录", URL: base + "/vendors"}, {Name: vendor.Name, URL: canonical}}
		doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "Organization", "name": vendor.Name, "url": canonical, "description": doc.Description, "address": map[string]any{"@type": "PostalAddress", "addressRegion": vendor.Province, "addressLocality": vendor.City}}
		doc.Image = absoluteURL(base, firstNonEmpty(vendor.CoverImage, vendor.Logo))
		doc.Schema = withBreadcrumbSchema(doc.Schema, doc.Breadcrumbs)
		return doc, true
	}
	if slug, ok := routeSlug(path, "/products/"); ok {
		var product model.Product
		if visibleProductQuery(h.DB).Preload("Category").First(&product, "products.slug = ?", slug).Error != nil {
			return seoDocument{}, false
		}
		doc.Title = fallbackSEO(product.SEOTitle, product.Name+"｜"+product.Category.Name) + "｜" + meta.SiteName
		doc.Description = fallbackSEO(product.SEODescription, firstNonEmpty(product.DetailContent, product.Description, product.CompatibleModels, "查看产品规格与适配信息。"))
		doc.BodyTitle, doc.BodyText = product.Name, doc.Description
		doc.Breadcrumbs = []seoBreadcrumb{{Name: "配件产品", URL: base + "/products"}, {Name: product.Category.Name, URL: base + "/products/category/" + product.Category.Slug}, {Name: product.Name, URL: canonical}}
		doc.Schema = map[string]any{"@context": "https://schema.org", "@type": "Product", "name": product.Name, "description": doc.Description, "url": canonical, "category": product.Category.Name, "image": absoluteURL(base, product.Image)}
		doc.Image = absoluteURL(base, product.Image)
		doc.OGType = "product"
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
	if publishedPageQuery(h.DB, pageType).Where("slug = ?", slug).First(&page).Error != nil {
		return seoDocument{}, false
	}
	path := "/" + slug
	if pageType == "article" {
		path = "/guides/" + slug
	}
	description := fallbackSEO(page.SEODescription, firstNonEmpty(page.Summary, page.Content, "农机配件行业指南。"))
	meta := h.meta(context.Background())
	doc := seoDocument{Title: fallbackSEO(page.SEOTitle, page.Title) + "｜" + meta.SiteName, Description: description, Keywords: page.SEOKeywords, Canonical: base + path, Image: absoluteURL(base, page.CoverImage), OGType: map[bool]string{true: "article", false: "website"}[pageType == "article"], BodyTitle: page.Title, BodyText: firstNonEmpty(page.Content, page.Summary), Breadcrumbs: []seoBreadcrumb{{Name: pageTypeLabel(pageType), URL: base + "/" + map[bool]string{true: "guides", false: ""}[pageType == "article"]}}}
	doc.Schema = map[string]any{"@context": "https://schema.org", "@type": map[bool]string{true: "Article", false: "WebPage"}[pageType == "article"], "headline": page.Title, "description": description, "url": doc.Canonical, "datePublished": page.PublishedAt, "dateModified": page.UpdatedAt, "author": map[string]any{"@type": "Organization", "name": firstNonEmpty(page.AuthorName, meta.SiteName)}}
	doc.Schema = withBreadcrumbSchema(doc.Schema, doc.Breadcrumbs)
	return doc, true
}

func (h SEOHandler) meta(ctx context.Context) service.SiteMeta { return h.HomeService.SiteMeta(ctx) }

func (h SEOHandler) redirect(c *gin.Context) (string, int, bool) {
	if h.DB == nil {
		return "", 0, false
	}
	source := c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		source += "?" + c.Request.URL.RawQuery
	}
	var item model.SEORedirect
	if h.DB.Where("source_path = ?", source).First(&item).Error == nil && item.DestinationPath != source {
		status := item.StatusCode
		if status != http.StatusMovedPermanently && status != http.StatusPermanentRedirect {
			status = http.StatusMovedPermanently
		}
		return item.DestinationPath, status, true
	}
	if id, ok := routeID(c.Request.URL.Path, "/vendors/"); ok {
		var vendor model.Vendor
		if publishedVendorQuery(h.DB).Select("id, slug").First(&vendor, id).Error == nil && vendor.Slug != "" {
			return "/v/" + vendor.Slug, http.StatusMovedPermanently, true
		}
	}
	if slug, ok := routeSlug(c.Request.URL.Path, "/vendors/"); ok {
		var vendor model.Vendor
		if publishedVendorQuery(h.DB).Select("id, slug").First(&vendor, "slug = ?", slug).Error == nil && vendor.Slug != "" {
			return "/v/" + vendor.Slug, http.StatusMovedPermanently, true
		}
	}
	if id, ok := routeID(c.Request.URL.Path, "/v/"); ok {
		var vendor model.Vendor
		if publishedVendorQuery(h.DB).Select("id, slug").First(&vendor, id).Error == nil && vendor.Slug != "" {
			return "/v/" + vendor.Slug, http.StatusMovedPermanently, true
		}
	}
	if id, ok := routeID(c.Request.URL.Path, "/products/"); ok {
		var product model.Product
		if visibleProductQuery(h.DB).Select("products.id, products.slug").First(&product, id).Error == nil && product.Slug != "" {
			return "/products/" + product.Slug, http.StatusMovedPermanently, true
		}
	}
	if c.Request.URL.Path == "/products" && len(c.Request.URL.Query()) == 1 {
		if raw := c.Query("categoryId"); raw != "" {
			var category model.Category
			if publishedCategoryQuery(h.DB).First(&category, "id = ?", raw).Error == nil && category.Slug != "" {
				return "/products/category/" + category.Slug, http.StatusMovedPermanently, true
			}
		}
	}
	return "", 0, false
}

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
	keywords := template.HTMLEscapeString(trimSEO(doc.Keywords, 255))
	ogType := template.HTMLEscapeString(firstNonEmpty(doc.OGType, "website"))
	image := template.HTMLEscapeString(doc.Image)
	head := fmt.Sprintf(`<title>%s</title><meta name="description" content="%s"><link rel="canonical" href="%s"><meta property="og:type" content="%s"><meta property="og:site_name" content="%s"><meta property="og:title" content="%s"><meta property="og:description" content="%s"><meta property="og:url" content="%s">`, title, description, canonical, ogType, template.HTMLEscapeString(meta.SiteName), title, description, canonical)
	if keywords != "" {
		head += `<meta name="keywords" content="` + keywords + `">`
	}
	if image != "" {
		head += `<meta property="og:image" content="` + image + `"><meta name="twitter:card" content="summary_large_image">`
	}
	if doc.NoIndex {
		directive := "noindex,follow"
		if doc.NoFollow {
			directive = "noindex,nofollow"
		}
		head += `<meta name="robots" content="` + directive + `">`
	}
	if meta.BaiduVerification != "" {
		head += `<meta name="baidu-site-verification" content="` + template.HTMLEscapeString(meta.BaiduVerification) + `">`
	}
	if meta.GoogleVerification != "" {
		head += `<meta name="google-site-verification" content="` + template.HTMLEscapeString(meta.GoogleVerification) + `">`
	}
	if doc.Schema != nil {
		head += `<script type="application/ld+json">` + strings.ReplaceAll(string(schema), "</", "<\\/") + `</script>`
	}
	if strings.Contains(index, "<title>大陆农机配件</title>") {
		index = strings.Replace(index, "<title>大陆农机配件</title>", head, 1)
	} else {
		index = strings.Replace(index, "</head>", head+"</head>", 1)
	}
	return strings.Replace(index, `<div id="root"></div>`, `<div id="root"></div>`+semanticFallback(doc), 1)
}

func semanticFallback(doc seoDocument) string {
	var content bytes.Buffer
	content.WriteString(`<main id="seo-fallback" data-server-rendered="true"><nav aria-label="面包屑">`)
	for _, item := range doc.Breadcrumbs {
		content.WriteString(`<a href="` + template.HTMLEscapeString(item.URL) + `">` + template.HTMLEscapeString(item.Name) + `</a> / `)
	}
	content.WriteString(`</nav><h1>` + template.HTMLEscapeString(doc.BodyTitle) + `</h1><p>` + template.HTMLEscapeString(doc.BodyText) + `</p><nav aria-label="相关页面"><a href="/">首页</a> · <a href="/products">配件产品</a> · <a href="/vendors">厂商目录</a> · <a href="/guides">行业指南</a> · <a href="/join">厂商入驻</a></nav></main>`)
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
func appendSiteName(title, siteName string) string {
	title, siteName = strings.TrimSpace(title), strings.TrimSpace(siteName)
	if siteName == "" || strings.Contains(title, siteName) {
		return title
	}
	if title == "" {
		return siteName
	}
	return title + "｜" + siteName
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
