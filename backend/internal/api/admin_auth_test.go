package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dalu-nongji-parts/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func TestLoginMessagesAreReadableChinese(t *testing.T) {
	if loginInvalidCredentialsMessage != "用户名或密码错误" {
		t.Fatalf("invalid credentials message = %q", loginInvalidCredentialsMessage)
	}
	if loginSessionFailureMessage != "登录失败" {
		t.Fatalf("session failure message = %q", loginSessionFailureMessage)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/admin/login", (AdminHandler{}).StaffLogin)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader("{"))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), loginInvalidRequestMessage) {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}

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

func TestValidateNewPasswordUsesBcryptByteLimitAndRejectsReuse(t *testing.T) {
	hash, err := auth.HashPassword("current-password")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name     string
		password string
		want     string
	}{
		{name: "too short", password: "short", want: "新密码长度需为 8–72 字节"},
		{name: "multibyte over bcrypt limit", password: strings.Repeat("密", 25), want: "新密码长度需为 8–72 字节"},
		{name: "same password", password: "current-password", want: "新密码不能与当前密码相同"},
		{name: "valid", password: "new-password-2026", want: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := validateNewPassword(hash, test.password); got != test.want {
				t.Fatalf("message = %q, want %q", got, test.want)
			}
		})
	}
}
