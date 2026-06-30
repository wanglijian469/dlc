package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func TestHealthHandlerReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHealthRoute(router)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("body = %s, want ok", rec.Body.String())
	}
}

func TestVendorDetailPayloadIncludesRichProfileFields(t *testing.T) {
	payload, err := json.Marshal(Response{
		Code:    0,
		Message: "ok",
		Data: model.Vendor{
			Name:              "浙江汉丰农机有限公司",
			WebsiteURL:        "https://vendor.example.com",
			AnnualCapacity:    "年产液压件 20 万套",
			Equipment:         "数控车床、液压测试台",
			Certifications:    "ISO9001 质量管理体系",
			AfterSalesService: "质保 12 个月",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, want := range []string{`"websiteUrl":"https://vendor.example.com"`, `"annualCapacity":"年产液压件 20 万套"`, `"equipment":"数控车床、液压测试台"`, `"certifications":"ISO9001 质量管理体系"`, `"afterSalesService":"质保 12 个月"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, want %s", body, want)
		}
	}
}

func TestPageResultSerializesPaginationShape(t *testing.T) {
	payload, err := json.Marshal(Response{
		Code:    0,
		Message: "ok",
		Data: PageResult{
			Items:    []string{"厂商"},
			Page:     2,
			PageSize: 12,
			Total:    30,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, want := range []string{`"items":["厂商"]`, `"page":2`, `"pageSize":12`, `"total":30`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, want %s", body, want)
		}
	}
}
