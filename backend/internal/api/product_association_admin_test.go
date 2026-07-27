package api

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAdminProductInputAcceptsMultipleVendorIDs(t *testing.T) {
	var input adminProductInput
	if err := json.Unmarshal([]byte(`{
		"name":"液压油泵总成",
		"publicationStatus":"published",
		"vendorIds":[9,12,9]
	}`), &input); err != nil {
		t.Fatalf("decode admin product input: %v", err)
	}
	if input.Name != "液压油泵总成" || input.PublicationStatus != "published" {
		t.Fatalf("product fields were not decoded: %#v", input.Product)
	}
	if got, want := uniqueUintIDs(input.VendorIDs), []uint{9, 12}; !reflect.DeepEqual(got, want) {
		t.Fatalf("vendor ids = %v, want %v", got, want)
	}
}

func TestUniqueUintIDsDropsZeroAndPreservesSelectionOrder(t *testing.T) {
	got := uniqueUintIDs([]uint{0, 7, 3, 7, 0, 9})
	want := []uint{7, 3, 9}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unique ids = %v, want %v", got, want)
	}
}
