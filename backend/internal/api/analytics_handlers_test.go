package api

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAnalyticsStaticTargetAndConversions(t *testing.T) {
	h := AnalyticsHandler{}
	path, typ, id, ok := h.resolveTarget("/products", "", 0)
	if !ok || path != "/products" || typ != "" || id != 0 {
		t.Fatalf("static target = %q %q %d %v", path, typ, id, ok)
	}
	if _, _, _, ok := h.resolveTarget("/admin/dashboard", "", 0); ok {
		t.Fatal("admin route must not be tracked")
	}
	if !isConversionTargetAllowed("join_cta_click", "/join", "") || isConversionTargetAllowed("join_cta_click", "/products", "") {
		t.Fatal("conversion target validation is incorrect")
	}
}

func TestProvinceResolverUsesOfflineCIDRFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "province.csv")
	if err := os.WriteFile(path, []byte("203.0.113.0/24,测试省\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := provinceForIP("203.0.113.8", path); got != "测试省" {
		t.Fatalf("province = %q", got)
	}
	if got := provinceForIP("192.168.1.8", path); got != "未知" {
		t.Fatalf("private IP province = %q", got)
	}
}

func TestAnalyticsWindowsLimitRepeatedEvents(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	if analyticsWindow("page_view", now) == analyticsWindow("page_view", now.Add(31*time.Minute)) {
		t.Fatal("page view window should advance after 30 minutes")
	}
	if analyticsWindow("search_submit", now) == analyticsWindow("search_submit", now.Add(6*time.Minute)) {
		t.Fatal("conversion window should advance after 5 minutes")
	}
}

func TestAnalyticsDaysAllowsOnlySupportedRanges(t *testing.T) {
	for raw, want := range map[string]int{"7": 7, "30": 30, "90": 90, "": 30, "8": 30, "invalid": 30} {
		if got := analyticsDays(raw); got != want {
			t.Fatalf("analyticsDays(%q) = %d, want %d", raw, got, want)
		}
	}
}
