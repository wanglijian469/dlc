package api

import (
	"testing"
)

func TestProtectionTracksDistinctCanonicalDetailResources(t *testing.T) {
	tests := []struct {
		path, kind, key string
		tracked         bool
	}{
		{"/v/abc-parts", "vendor", "abc-parts", true},
		{"/api/vendors/slug/abc-parts", "vendor", "abc-parts", true},
		{"/products/rotary-tiller", "product", "rotary-tiller", true},
		{"/api/products/18", "product", "18", true},
		{"/api/vendors", "", "", false},
		{"/search", "", "", false},
	}
	for _, tt := range tests {
		kind, key, tracked := protectedDetailResource(tt.path)
		if kind != tt.kind || key != tt.key || tracked != tt.tracked {
			t.Fatalf("%s = (%q, %q, %t), want (%q, %q, %t)", tt.path, kind, key, tracked, tt.kind, tt.key, tt.tracked)
		}
	}
}

func TestProtectionWindowCountsDistinctResourcesAndResetsAfterThreshold(t *testing.T) {
	service := &AccessProtectionService{windows: map[string]*accessWindow{}}
	cfg := DefaultProtectionConfig()
	cfg.DistinctResourceLimit = 2
	if service.observe("client", "vendor|one", cfg) {
		t.Fatal("first resource triggered threshold")
	}
	if service.observe("client", "vendor|one", cfg) {
		t.Fatal("repeat resource triggered threshold")
	}
	if service.observe("client", "vendor|two", cfg) {
		t.Fatal("second distinct resource triggered threshold")
	}
	if !service.observe("client", "vendor|three", cfg) {
		t.Fatal("third distinct resource did not trigger threshold")
	}
	if service.observe("client", "vendor|four", cfg) {
		t.Fatal("window did not reset after threshold event")
	}
}

func TestProtectionConfigRejectsHeaderInjectionAndInvalidCIDR(t *testing.T) {
	cfg := DefaultProtectionConfig()
	cfg.BlockedAIAgents = []string{"GPTBot\nUser-agent: *"}
	if validateProtectionConfig(cfg) == nil {
		t.Fatal("crawler rule injection was accepted")
	}
	cfg = DefaultProtectionConfig()
	cfg.AllowCIDRs = []string{"not-an-ip"}
	if validateProtectionConfig(cfg) == nil {
		t.Fatal("invalid allowlist entry was accepted")
	}
}
