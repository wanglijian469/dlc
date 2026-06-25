package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"dalu-nongji-parts/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func TestAdminAuthMiddlewareRequiresToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	protected := router.Group("/api/admin")
	protected.Use(AdminAuth("secret"))
	protected.GET("/profile", func(c *gin.Context) { OK(c, gin.H{"username": c.GetString("username")}) })

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/admin/profile", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestAdminAuthMiddlewareAcceptsToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token, err := auth.IssueToken("admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	protected := router.Group("/api/admin")
	protected.Use(AdminAuth("secret"))
	protected.GET("/profile", func(c *gin.Context) { OK(c, gin.H{"username": c.GetString("username")}) })

	req := httptest.NewRequest(http.MethodGet, "/api/admin/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
