package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeAccountCreateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   accountCreateInput
		wantErr bool
	}{
		{name: "ordinary user", input: accountCreateInput{Username: " member ", Password: "secret1", Role: "user"}},
		{name: "vendor with company", input: accountCreateInput{Username: "vendor", Password: "secret1", Role: "vendor", CompanyName: " 测试公司 "}},
		{name: "admin", input: accountCreateInput{Username: "admin2", Password: "secret1", Role: "admin"}},
		{name: "short password", input: accountCreateInput{Username: "member", Password: "123", Role: "user"}, wantErr: true},
		{name: "public shape cannot omit vendor", input: accountCreateInput{Username: "vendor", Password: "secret1", Role: "vendor"}, wantErr: true},
		{name: "vendor cannot specify both company sources", input: accountCreateInput{Username: "vendor", Password: "secret1", Role: "vendor", VendorID: uintPtr(2), CompanyName: "测试公司"}, wantErr: true},
		{name: "ordinary user cannot bind company", input: accountCreateInput{Username: "member", Password: "secret1", Role: "user", VendorID: uintPtr(2)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAccountCreateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && (got.Username != "member" && tt.name == "ordinary user") {
				t.Fatalf("username was not normalized: %q", got.Username)
			}
			if err == nil && tt.name == "vendor with company" && got.CompanyName != "测试公司" {
				t.Fatalf("company name was not normalized: %q", got.CompanyName)
			}
		})
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
