package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
)

func TestStaticRouteOnlyAcceptsCanonicalDetailPaths(t *testing.T) {
	tests := []struct {
		path, resourceType, slug string
		ok                       bool
	}{
		{"/v/abc-parts", "vendor", "abc-parts", true},
		{"/products/hydraulic-pump", "product", "hydraulic-pump", true},
		{"/vendors/abc-parts", "", "", false},
		{"/products", "", "", false},
		{"/products/a/child", "", "", false},
	}
	for _, test := range tests {
		resourceType, slug, ok := staticRoute(test.path)
		if resourceType != test.resourceType || slug != test.slug || ok != test.ok {
			t.Fatalf("staticRoute(%q) = (%q, %q, %t)", test.path, resourceType, slug, ok)
		}
	}
}

func TestStaticSemanticFallbackOnlyIncludesPublicWechatQRCode(t *testing.T) {
	vendor := model.Vendor{Name: "公开厂商", Wechat: "public-wechat", WechatQRCode: "/api/media/88", WechatPublic: true}
	redactVendor(&vendor)
	html := staticSemanticFallback(StaticPagePayload{Vendor: &vendor}, seoDocument{})
	if !strings.Contains(html, "public-wechat") || !strings.Contains(html, "/api/media/88") {
		t.Fatalf("public WeChat contact missing from static HTML: %s", html)
	}

	privateVendor := model.Vendor{Name: "隐私厂商", Wechat: "secret-wechat", WechatQRCode: "/api/media/99", WechatQRCodeAssetID: uintPointer(99)}
	redactVendor(&privateVendor)
	privateHTML := staticSemanticFallback(StaticPagePayload{Vendor: &privateVendor}, seoDocument{})
	if strings.Contains(privateHTML, "secret-wechat") || strings.Contains(privateHTML, "/api/media/99") {
		t.Fatalf("private WeChat contact leaked into static HTML: %s", privateHTML)
	}
}

func TestStaticPageWritesAtomicallyInsideConfiguredDirectory(t *testing.T) {
	root := t.TempDir()
	service := &StaticPageService{cfg: config.Config{StaticPageDir: root}}
	relative, absolute, err := service.writePage("product", 12, strings.Repeat("a", 64), []byte("<html>ready</html>"))
	if err != nil {
		t.Fatal(err)
	}
	if relative != "product/12-aaaaaaaaaaaaaaaa.html" {
		t.Fatalf("relative path = %q", relative)
	}
	content, err := os.ReadFile(absolute)
	if err != nil || string(content) != "<html>ready</html>" {
		t.Fatalf("content = %q, err = %v", content, err)
	}
	if _, safe := service.safeAbsolutePath(filepath.Join("..", "outside.html")); safe {
		t.Fatal("safeAbsolutePath accepted traversal")
	}
	matches, err := filepath.Glob(filepath.Join(root, "product", "*.tmp"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary files remain: %v, err = %v", matches, err)
	}
}
