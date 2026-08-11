package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeAccountCreateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   accountCreateInput
		wantErr bool
	}{
		{name: "ordinary user is retired", input: accountCreateInput{Username: " member ", Password: "secret1", Role: "user"}, wantErr: true},
		{name: "vendor with company", input: accountCreateInput{Username: "vendor", Password: "secret1", Role: "vendor", CompanyName: " 测试公司 "}},
		{name: "admin", input: accountCreateInput{Username: "admin2", Password: "secret1", Role: "admin"}},
		{name: "editor", input: accountCreateInput{Username: "editor1", Password: "secret1", Role: "editor"}},
		{name: "reviewer", input: accountCreateInput{Username: "reviewer1", Password: "secret1", Role: "reviewer"}},
		{name: "short password", input: accountCreateInput{Username: "vendor", Password: "123", Role: "vendor", CompanyName: "测试公司"}, wantErr: true},
		{name: "public shape cannot omit vendor", input: accountCreateInput{Username: "vendor", Password: "secret1", Role: "vendor"}, wantErr: true},
		{name: "vendor cannot specify both company sources", input: accountCreateInput{Username: "vendor", Password: "secret1", Role: "vendor", VendorID: uintPtr(2), CompanyName: "测试公司"}, wantErr: true},
		{name: "retired role cannot bind company", input: accountCreateInput{Username: "member", Password: "secret1", Role: "user", VendorID: uintPtr(2)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAccountCreateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.name == "vendor with company" && got.CompanyName != "测试公司" {
				t.Fatalf("company name was not normalized: %q", got.CompanyName)
			}
		})
	}
}

func TestPublicRegistrationRejectsOrdinaryUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := AdminHandler{}
	router.POST("/api/auth/register", handler.Register)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username":"member","password":"secret1","role":"user"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "采购商注册或厂商入驻") {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestLoginAudienceAllowsOnlyMarketplaceAndStaffRoles(t *testing.T) {
	if !loginAudienceAllows("account", "vendor") || !loginAudienceAllows("account", "buyer") || loginAudienceAllows("account", "user") || loginAudienceAllows("account", "admin") {
		t.Fatal("account login audience accepted an invalid role")
	}
	if !loginAudienceAllows("staff", "admin") || !loginAudienceAllows("staff", "editor") || !loginAudienceAllows("staff", "reviewer") || loginAudienceAllows("staff", "vendor") {
		t.Fatal("staff login audience roles are incorrect")
	}
}

func TestRequireAnyRoleRejectsOrdinaryUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("role", "user"); c.Next() })
	router.GET("/cms", RequireAnyRole("admin", "vendor"), func(c *gin.Context) { c.Status(http.StatusOK) })

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/cms", nil))
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func uintPtr(value uint) *uint { return &value }
