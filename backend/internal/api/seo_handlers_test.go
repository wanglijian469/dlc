package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func TestRobotsPublishesDiscoveryAndPrivatePathRules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := SEOHandler{HomeService: service.HomeService{}}
	router.GET("/robots.txt", handler.Robots)

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	req.Host = "parts.example.cn"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	for _, want := range []string{"Disallow: /admin/", "Disallow: /api/admin/", "Disallow: /search", "Sitemap: http://parts.example.cn/sitemap.xml"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("robots = %q, want %q", rec.Body.String(), want)
		}
	}
}

func TestInjectSEOHTMLUsesCanonicalMetadataAndNoIndex(t *testing.T) {
	doc := seoDocument{
		Title:       "旋耕机配件厂家｜大陆农机配件",
		Description: "真实厂商与适配信息",
		Canonical:   "https://parts.example.cn/products?categoryId=8",
		NoIndex:     true,
		BodyTitle:   "旋耕机配件",
		BodyText:    "查看真实产品与厂商资料。",
		Schema:      map[string]any{"@context": "https://schema.org", "@type": "Product", "name": "旋耕机配件"},
	}
	html := injectSEOHTML("<html><head></head><body><div id=\"root\"></div></body></html>", doc, service.DefaultSiteMeta())
	for _, want := range []string{
		"<title>旋耕机配件厂家｜大陆农机配件</title>",
		"<link rel=\"canonical\" href=\"https://parts.example.cn/products?categoryId=8\">",
		"<meta name=\"robots\" content=\"noindex,follow\">",
		"application/ld+json",
		"<main id=\"seo-fallback\" data-server-rendered=\"true\">",
		"<h1>旋耕机配件</h1>",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered html missing %q: %s", want, html)
		}
	}
}

func TestSEOHelperNormalizesCanonicalPathsAndDescriptions(t *testing.T) {
	if id, ok := routeID("/products/24", "/products/"); !ok || id != 24 {
		t.Fatalf("routeID valid route = (%d, %t), want (24, true)", id, ok)
	}
	if id, ok := routeID("/v/24", "/v/"); !ok || id != 24 {
		t.Fatalf("routeID vendor site = (%d, %t), want (24, true)", id, ok)
	}
	if _, ok := routeID("/products/24/extra", "/products/"); ok {
		t.Fatal("routeID accepted a nested path")
	}
	if got := trimSEO("  一段   可抓取的\n摘要  ", 100); got != "一段 可抓取的 摘要" {
		t.Fatalf("trimSEO = %q", got)
	}
	if got := absoluteURL("https://parts.example.cn/", "/media/product.png"); got != "https://parts.example.cn/media/product.png" {
		t.Fatalf("absoluteURL = %q", got)
	}
	data, err := json.Marshal(withBreadcrumbSchema(map[string]any{"@type": "Product"}, []seoBreadcrumb{{Name: "产品", URL: "https://parts.example.cn/products"}}))
	if err != nil || !strings.Contains(string(data), `"BreadcrumbList"`) || !strings.Contains(string(data), `"ListItem"`) {
		t.Fatalf("breadcrumb JSON-LD = %s, err = %v", data, err)
	}
}

func TestAppendSiteNameAvoidsDuplicateBrand(t *testing.T) {
	if got := appendSiteName("河北冀农｜旋耕机配件厂家", "大陆农机配件"); got != "河北冀农｜旋耕机配件厂家｜大陆农机配件" {
		t.Fatalf("appended title = %q", got)
	}
	if got := appendSiteName("河北冀农｜大陆农机配件", "大陆农机配件"); got != "河北冀农｜大陆农机配件" {
		t.Fatalf("duplicate brand title = %q", got)
	}
}

func TestStaticSPAResponseAddsNoIndexForSearchURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	publicDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html><head></head><body><div id=\"root\"></div></body></html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	RegisterStaticRoutes(router, nil, publicDir)

	req := httptest.NewRequest(http.MethodGet, "/search?keyword=旋耕机", nil)
	req.Host = "parts.example.cn"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	for _, want := range []string{"<meta name=\"robots\" content=\"noindex,follow\">", "<link rel=\"canonical\" href=\"http://parts.example.cn/search\">", "<main id=\"seo-fallback\" data-server-rendered=\"true\">"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("static response missing %q: %s", want, rec.Body.String())
		}
	}
}
