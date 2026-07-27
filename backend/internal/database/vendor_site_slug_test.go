package database

import "testing"

func TestNormalizeVendorSiteSlug(t *testing.T) {
	if got, err := normalizeVendorSiteSlug("  ABC-Parts-2026 "); err != nil || got != "abc-parts-2026" {
		t.Fatalf("normalize valid slug = %q, %v", got, err)
	}
	for _, value := range []string{"", "-abc", "abc-", "abc--parts", "abc_parts", "厂商站"} {
		if _, err := normalizeVendorSiteSlug(value); err == nil {
			t.Fatalf("normalizeVendorSiteSlug(%q) accepted an invalid address", value)
		}
	}
}
