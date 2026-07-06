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

func TestVendorDetailPayloadIncludesProcessingFields(t *testing.T) {
	payload, err := json.Marshal(Response{
		Code:    0,
		Message: "ok",
		Data: model.Vendor{
			Name:                "测试加工厂商",
			ProvidesProcessing:  true,
			ProcessingServices:  "来图来样加工、数控车削",
			ProcessingMaterials: "钢件、铸铁件",
			ProcessingEquipment: "数控车床、加工中心",
			ProcessingCapacity:  "小批量 3 天交付，批量订单按图报价",
			ProcessingRegions:   "华北、华东",
			ProcessingNotes:     "支持图纸、样件和批量代工",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, want := range []string{
		`"providesProcessing":true`,
		`"processingServices":"来图来样加工、数控车削"`,
		`"processingMaterials":"钢件、铸铁件"`,
		`"processingEquipment":"数控车床、加工中心"`,
		`"processingCapacity":"小批量 3 天交付，批量订单按图报价"`,
		`"processingRegions":"华北、华东"`,
		`"processingNotes":"支持图纸、样件和批量代工"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, want %s", body, want)
		}
	}
}

func TestVendorDetailPayloadIncludesPublicImportReviewFields(t *testing.T) {
	payload, err := json.Marshal(Response{
		Code:    0,
		Message: "ok",
		Data: model.Vendor{
			Name:         "河北冀农农机具有限公司",
			WebsiteURL:   "https://www.hbjinong.com/",
			SourceURL:    "https://www.hbjinong.com/",
			SourceNote:   "公开官网首页采集，人工复核前不标记平台认证。",
			ReviewStatus: "pending",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, want := range []string{
		`"sourceUrl":"https://www.hbjinong.com/"`,
		`"sourceNote":"公开官网首页采集，人工复核前不标记平台认证。"`,
		`"reviewStatus":"pending"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, want %s", body, want)
		}
	}
}

func TestPublicRoutesExposeProcessingEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterPublicRoutes(router, nil)

	for _, want := range []string{"/api/processing-vendors", "/api/processing-filter-options"} {
		if !routeExists(router.Routes(), http.MethodGet, want) {
			t.Fatalf("GET %s route is not registered", want)
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

func routeExists(routes gin.RoutesInfo, method string, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
