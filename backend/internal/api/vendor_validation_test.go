package api

import (
	"strings"
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestValidateVendorServiceAdvantagesUsesUnicodeCharacterLimit(t *testing.T) {
	valid := model.Vendor{ServiceAdvantages: strings.Repeat("农", 80)}
	if err := validateVendorProfileContent(valid); err != nil {
		t.Fatalf("80 characters rejected: %v", err)
	}

	invalid := model.Vendor{ServiceAdvantages: strings.Repeat("农", 81)}
	if err := validateVendorProfileContent(invalid); err == nil {
		t.Fatal("81 characters were accepted")
	}
}
