package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func TestCMSRoutesAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := config.Config{AuthSecret: "secret", PublicDir: t.TempDir()}

	RegisterPublicRoutes(router, nil)
	RegisterAdminRoutes(router, nil, cfg)

	routes := map[string]bool{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	for _, want := range []string{
		"POST /api/auth/register",
		"PUT /api/auth/password",
		"GET /api/site-meta",
		"GET /api/pages/:slug",
		"GET /api/products/:id",
		"GET /api/products/:id/suppliers",
		"GET /api/admin/pages",
		"POST /api/admin/pages",
		"PUT /api/admin/pages/:id",
		"DELETE /api/admin/pages/:id",
		"GET /api/admin/friend-links",
		"POST /api/admin/friend-links",
		"PUT /api/admin/friend-links/:id",
		"DELETE /api/admin/friend-links/:id",
		"POST /api/admin/uploads",
		"GET /api/admin/vendor-profile",
		"PUT /api/admin/vendor-profile",
		"GET /api/admin/vendor-submissions",
		"PUT /api/admin/vendor-submissions/:id/review",
		"GET /api/admin/vendor-product-catalog",
		"POST /api/admin/vendor-products/link",
		"GET /api/admin/vendor-analytics",
		"GET /api/admin/product-submissions",
		"PUT /api/admin/product-submissions/:id/review",
		"GET /api/admin/products/:id/suppliers",
		"POST /api/admin/products/:id/suppliers",
		"DELETE /api/admin/products/:id/suppliers/:supplierId",
		"GET /api/admin/vendor-options",
		"POST /api/admin/product-suppliers/batch",
		"PUT /api/admin/products/:id/merge",
		"GET /api/admin/users",
		"POST /api/admin/users",
	} {
		if !routes[want] {
			t.Fatalf("route %q not registered; routes=%v", want, routes)
		}
	}
	if routes["POST /api/admin/register"] {
		t.Fatal("retired public registration route remains registered")
	}
}

func TestAdminUploadRejectsUnsupportedFileType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := config.Config{AuthSecret: "secret", PublicDir: t.TempDir()}
	RegisterAdminRoutes(router, nil, cfg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "notes.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("not an image")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	token, err := auth.IssueToken("admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/uploads", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "图片") {
		t.Fatalf("body = %s, want image type message", rec.Body.String())
	}
}
