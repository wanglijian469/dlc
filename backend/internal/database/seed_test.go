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

func TestDefaultSeedContainsExpandedSidebarMenuTree(t *testing.T) {
	seed := DefaultSeed()
	wantChildren := map[string]int{
		"wearing":      5,
		"transmission": 7,
		"chassis":      5,
		"hydraulic":    5,
		"engine":       5,
		"brake":        5,
		"electrical":   5,
		"harvester":    5,
		"seeding":      5,
	}
	actualChildren := map[string]int{}
	seenParents := map[string]bool{}

	for _, menu := range seed.Menus {
		if _, ok := wantChildren[menu.Key]; ok {
			seenParents[menu.Key] = true
		}
		if menu.ParentKey != "" {
			actualChildren[menu.ParentKey]++
		}
		if menu.Key == "future" || menu.Name == "待扩展菜单" {
			t.Fatalf("seed should not include placeholder menu %q", menu.Key)
		}
	}

	for key, want := range wantChildren {
		if !seenParents[key] {
			t.Fatalf("missing parent menu %q", key)
		}
		if actualChildren[key] != want {
			t.Fatalf("children for %s = %d, want %d", key, actualChildren[key], want)
		}
	}
}
