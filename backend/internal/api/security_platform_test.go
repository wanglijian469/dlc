package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireCSRFAcceptsMatchingDoubleSubmitToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("csrfHash", credentialHash("csrf-value"))
		c.Next()
	}, RequireCSRF())
	router.POST("/save", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodPost, "/save", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "csrf-value"})
	req.Header.Set("X-CSRF-Token", "csrf-value")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestRequireCSRFRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequireCSRF())
	router.POST("/save", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/save", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestPublicETagReturnsNotModified(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(PublicETag())
	router.GET("/api/home", func(c *gin.Context) { OK(c, gin.H{"version": 1}) })

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/home", nil))
	etag := first.Header().Get("ETag")
	if first.Code != http.StatusOK || etag == "" {
		t.Fatalf("first status = %d, etag = %q", first.Code, etag)
	}
	if cacheControl := first.Header().Get("Cache-Control"); cacheControl != "public, no-cache" {
		t.Fatalf("cache-control = %q, want public, no-cache", cacheControl)
	}
	secondReq := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	secondReq.Header.Set("If-None-Match", etag)
	second := httptest.NewRecorder()
	router.ServeHTTP(second, secondReq)
	if second.Code != http.StatusNotModified || second.Body.Len() != 0 {
		t.Fatalf("second status = %d body = %q, want 304 empty", second.Code, second.Body.String())
	}
}

func TestRevisionTypesCoverAllPublicResources(t *testing.T) {
	for _, resource := range []string{"vendor", "product", "category", "page", "article"} {
		if !validRevisionType(resource) {
			t.Fatalf("resource %q is not accepted", resource)
		}
	}
	if validRevisionType("config") {
		t.Fatal("configuration must not enter the public-content workflow")
	}
}
