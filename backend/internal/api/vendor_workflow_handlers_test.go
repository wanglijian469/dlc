package api

import (
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestApplyVendorEditableFieldsPreservesAdministratorFields(t *testing.T) {
	destination := model.Vendor{
		ID:            7,
		Name:          "旧名称",
		ReviewStatus:  "verified",
		IsVisible:     true,
		IsVerified:    true,
		IsRecommended: true,
		SortOrder:     12,
		SourceURL:     "https://source.example.com",
	}
	source := model.Vendor{
		Name:          " 新名称 ",
		MainProducts:  "齿轮、轴承",
		Phone:         "13800000000",
		WebsiteURL:    " https://vendor.example.com ",
		ReviewStatus:  "rejected",
		IsVisible:     false,
		IsVerified:    false,
		IsRecommended: false,
		SortOrder:     0,
	}

	applyVendorEditableFields(&destination, source)

	if destination.Name != "新名称" || destination.WebsiteURL != "https://vendor.example.com" || destination.MainProducts != "齿轮、轴承" {
		t.Fatalf("editable fields were not copied: %#v", destination)
	}
	if !destination.IsVisible || !destination.IsVerified || !destination.IsRecommended || destination.SortOrder != 12 || destination.ReviewStatus != "verified" || destination.SourceURL == "" {
		t.Fatalf("administrator fields were overwritten: %#v", destination)
	}
}
