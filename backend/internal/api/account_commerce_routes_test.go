package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func TestAppSessionLifetimesStayShortAndRotatable(t *testing.T) {
	if appAccessTTL > 15*time.Minute || appAccessTTL >= appRefreshTTL {
		t.Fatalf("unsafe app session TTLs: access=%s refresh=%s", appAccessTTL, appRefreshTTL)
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

func TestAccountCommerceRoutesAreComplete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAccountCommerceRoutes(router, nil, config.Config{})
	want := map[string]bool{
		"POST /api/v1/app/auth/register": false,
		"POST /api/v1/app/auth/login":    false,
		"POST /api/v1/app/auth/refresh":  false,
		"POST /api/v1/media":             false,
		"GET /api/v1/auctions":           false,
		"GET /api/v1/auctions/:id":       false,
		"POST /api/v1/auctions/:id/bids": false,
		"GET /api/v1/me/auctions":        false,
		"POST /api/v1/me/auctions":       false,
		"GET /api/v1/me/notifications":   false,
	}
	for _, route := range router.Routes() {
		if strings.Contains(route.Path, "market-posts") {
			t.Fatalf("retired route remains: %s", route.Path)
		}
		key := route.Method + " " + route.Path
		if _, exists := want[key]; exists {
			want[key] = true
		}
	}
	for route, found := range want {
		if !found {
			t.Errorf("missing account commerce route %s", route)
		}
	}
}
