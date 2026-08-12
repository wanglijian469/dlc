package database

import "testing"

func TestNormalizeVendorSiteSlug(t *testing.T) {
	if got, err := normalizeVendorSiteSlug("  ABCParts2026 "); err != nil || got != "abcparts2026" {
		t.Fatalf("normalize valid slug = %q, %v", got, err)
	}
	for _, value := range []string{"", "-abc", "abc-", "abc-parts", "abc_parts", "ab", "abcdefghijklmnopq", "厂商站"} {
		if _, err := normalizeVendorSiteSlug(value); err == nil {
			t.Fatalf("normalizeVendorSiteSlug(%q) accepted an invalid address", value)
		}
	}
}
