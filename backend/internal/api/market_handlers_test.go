package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
)

func TestMarketDTOOmitsPrivateContactAndOwnerID(t *testing.T) {
	payload, err := json.Marshal(marketDTO(model.MarketPost{
		ID: 9, PostType: "demand", Title: "求购收割机链条", Description: "需要长期采购收割机链条配件",
		ContactName: "王经理", ContactPhone: "13812345678", OwnerUserID: 7, OwnerUsername: "buyer-one",
		Status: "published", ExpiresAt: time.Now().Add(24 * time.Hour),
	}))
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, secret := range []string{"王经理", "13812345678", "ownerUserId", "contactPhone"} {
		if strings.Contains(body, secret) {
			t.Fatalf("private field leaked in public DTO: %s", body)
		}
	}
	if !strings.Contains(body, `"publisherName":"buyer-one"`) {
		t.Fatalf("publisher display name missing: %s", body)
	}
}

func TestMarketPostRoleAndExpiryValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Set("role", "buyer")
	context.Set("userId", uint(4))
	context.Set("username", "buyer-one")
	handler := AdminHandler{}
	base := marketPostInput{Type: "demand", Title: "求购收割机链条", Description: "长期采购收割机链条和相关配件", ContactName: "王经理", ContactPhone: "13812345678"}
	item, ok := handler.marketPostFromInput(context, base, model.MarketPost{})
	if !ok || item.Status != "published" || item.OwnerUserID != 4 {
		t.Fatalf("valid buyer demand rejected: %#v", item)
	}
	if remaining := time.Until(item.ExpiresAt); remaining < 29*24*time.Hour || remaining > 31*24*time.Hour {
		t.Fatalf("default expiry = %s", remaining)
	}
	base.Type = "supply"
	if _, ok := handler.marketPostFromInput(context, base, model.MarketPost{}); ok {
		t.Fatal("buyer was allowed to publish a supply post")
	}
}

func TestAppSessionLifetimesStayShortAndRotatable(t *testing.T) {
	if appAccessTTL > 15*time.Minute || appAccessTTL >= appRefreshTTL {
		t.Fatalf("unsafe app session TTLs: access=%s refresh=%s", appAccessTTL, appRefreshTTL)
	}
}

func TestMarketPostRulesCapMediaAndExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Set("role", "vendor")
	context.Set("userId", uint(8))
	context.Set("vendorId", uint(3))
	context.Set("username", "vendor-one")
	handler := AdminHandler{}
	input := marketPostInput{Type: "supply", Title: "供应收割机链条", Description: "稳定供应收割机链条和相关配件", ContactName: "李经理", ContactPhone: "13912345678", ExpiresInDays: 90, AssetIDs: []uint{1, 2, 3, 4, 5, 6}}
	item, ok := handler.marketPostFromInput(context, input, model.MarketPost{})
	if !ok || item.VendorID == nil || *item.VendorID != 3 {
		t.Fatalf("valid vendor supply rejected: %#v", item)
	}
	if remaining := time.Until(item.ExpiresAt); remaining < 89*24*time.Hour || remaining > 91*24*time.Hour {
		t.Fatalf("maximum expiry = %s", remaining)
	}
	input.ExpiresInDays = 91
	if _, ok := handler.marketPostFromInput(context, input, model.MarketPost{}); ok {
		t.Fatal("post longer than 90 days was accepted")
	}
	input.ExpiresInDays = 30
	input.AssetIDs = append(input.AssetIDs, 7)
	if _, ok := handler.marketPostFromInput(context, input, model.MarketPost{}); ok {
		t.Fatal("post with more than six images was accepted")
	}
}

func TestExpiredMarketPostIsReportedWithoutContact(t *testing.T) {
	dto := marketDTO(model.MarketPost{Status: "published", ContactName: "秘密联系人", ContactPhone: "13800000000", ExpiresAt: time.Now().Add(-time.Minute)})
	if dto.Status != "expired" {
		t.Fatalf("status = %q, want expired", dto.Status)
	}
	payload, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "秘密联系人") || strings.Contains(string(payload), "13800000000") {
		t.Fatalf("expired DTO leaked contact: %s", payload)
	}
}

func TestAppCredentialHashNeverStoresPlainToken(t *testing.T) {
	token, err := randomCredential(32)
	if err != nil {
		t.Fatal(err)
	}
	hash := credentialHash(token)
	if hash == token || len(hash) != 64 || credentialHash(token) != hash {
		t.Fatalf("credential hashing is not deterministic and one-way: token=%q hash=%q", token, hash)
	}
}

func TestBuyerProfileContactIsOnlyAddedByAuthenticatedDTO(t *testing.T) {
	payload, err := json.Marshal(buyerProfileDTO(model.BuyerProfile{ID: 2, UserID: 4, DisplayName: "王经理", Phone: "13812345678"}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), "13812345678") {
		t.Fatalf("authenticated buyer profile omitted its own phone: %s", payload)
	}
	modelPayload, err := json.Marshal(model.BuyerProfile{Phone: "13812345678"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(modelPayload), "13812345678") {
		t.Fatalf("raw buyer profile leaked phone: %s", modelPayload)
	}
}

func TestMarketplaceRoutesAreAdditiveAndComplete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterMarketplaceRoutes(router, nil, config.Config{})
	want := map[string]bool{
		"POST /api/v1/app/auth/register":       false,
		"POST /api/v1/app/auth/login":          false,
		"POST /api/v1/app/auth/refresh":        false,
		"GET /api/v1/market-posts":             false,
		"GET /api/v1/market-posts/:id/contact": false,
		"POST /api/v1/me/market-posts":         false,
		"PUT /api/v1/me/market-posts/:id":      false,
		"DELETE /api/v1/me/market-posts/:id":   false,
		"POST /api/v1/media":                   false,
		"GET /api/v1/auctions":                 false,
		"GET /api/v1/auctions/:id":             false,
		"POST /api/v1/auctions/:id/bids":       false,
		"GET /api/v1/me/auctions":              false,
		"POST /api/v1/me/auctions":             false,
		"GET /api/v1/me/notifications":          false,
	}
	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := want[key]; exists {
			want[key] = true
		}
	}
	for route, found := range want {
		if !found {
			t.Errorf("missing marketplace route %s", route)
		}
	}
}
