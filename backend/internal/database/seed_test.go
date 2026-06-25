package database

import "testing"

func TestDefaultSeedContainsTransmissionChildren(t *testing.T) {
	seed := DefaultSeed()
	children := 0
	for _, menu := range seed.Menus {
		if menu.ParentKey == "transmission" {
			children++
		}
	}
	if children != 7 {
		t.Fatalf("transmission children = %d, want 7", children)
	}
	if len(seed.Vendors) < 12 {
		t.Fatalf("vendors = %d, want at least 12", len(seed.Vendors))
	}
}
