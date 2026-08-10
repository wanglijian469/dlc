package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestPublicRoutesExposeVendorCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterPublicRoutes(router, nil)
	if !routeExists(router.Routes(), http.MethodGet, "/api/vendor-categories") {
		t.Fatal("GET /api/vendor-categories route is not registered")
	}
}

func TestNewlyJoinedVendorCutoffUsesAnExactNinetyDayWindow(t *testing.T) {
	now := time.Date(2026, time.August, 10, 14, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	want := time.Date(2026, time.May, 12, 14, 30, 0, 0, now.Location())
	if got := newlyJoinedVendorCutoff(now); !got.Equal(want) {
		t.Fatalf("newlyJoinedVendorCutoff() = %s, want %s", got, want)
	}
}

func TestBuildVendorCategoryTreeUsesRecursiveRootCounts(t *testing.T) {
	categories := []model.VendorCategory{
		{ID: 1, Name: "传动配件", IsEnabled: true},
		{ID: 2, Name: "变速箱齿轮", ParentID: 1, IsEnabled: true},
		{ID: 3, Name: "液压系统", IsEnabled: true},
	}
	tree := buildVendorCategoryTree(categories, map[uint]int64{1: 4}, map[uint]int64{2: 2})
	if len(tree) != 2 || tree[0].VendorCount != 4 || len(tree[0].Children) != 1 || tree[0].Children[0].VendorCount != 2 {
		t.Fatalf("vendor category tree = %#v", tree)
	}
	if tree[1].VendorCount != 0 {
		t.Fatalf("empty category count = %d, want 0", tree[1].VendorCount)
	}
}

func TestPublicRoutesExposeIndustryGuideEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterPublicRoutes(router, nil)

	for _, want := range []string{"/api/articles", "/api/articles/:slug"} {
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

func TestMaskPhoneKeepsOnlyEnoughDigitsForRecognition(t *testing.T) {
	if got, want := maskPhone("0319-5666294"), "031*-*****94"; got != want {
		t.Fatalf("maskPhone() = %q, want %q", got, want)
	}
	if got, want := maskPhone("13812345678"), "138******78"; got != want {
		t.Fatalf("maskPhone() = %q, want %q", got, want)
	}
}

func TestRedactVendorRespectsPerFieldPublicSettings(t *testing.T) {
	vendor := model.Vendor{
		Phone: "13812345678", Wechat: "secret-wechat", WechatQRCode: "/api/media/12", WechatQRCodeAssetID: uintPointer(12), ContactName: "王经理",
		PhonePublic: false, WechatPublic: true, ContactNamePublic: false,
	}
	redactVendor(&vendor)
	if vendor.Phone != "138******78" || vendor.Wechat != "secret-wechat" || vendor.WechatQRCode != "/api/media/12" || vendor.ContactName != "" {
		t.Fatalf("unexpected public contact payload: %#v", vendor)
	}
	if !vendor.PhoneAvailable || !vendor.WechatAvailable || !vendor.ContactNameAvailable {
		t.Fatalf("availability flags not preserved: %#v", vendor)
	}

	vendor.WechatPublic = false
	redactVendor(&vendor)
	if vendor.Wechat != "" || vendor.WechatQRCode != "" || vendor.WechatQRCodeAssetID != nil || !vendor.WechatAvailable {
		t.Fatalf("private WeChat data leaked or availability was lost: %#v", vendor)
	}
}

func uintPointer(value uint) *uint { return &value }

func routeExists(routes gin.RoutesInfo, method string, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
