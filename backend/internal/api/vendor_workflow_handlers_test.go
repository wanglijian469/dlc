package api

import (
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestApplyVendorEditableFieldsPreservesAdministratorFields(t *testing.T) {
	destination := model.Vendor{
		ID:             7,
		Name:           "旧名称",
		ReviewStatus:   "verified",
		IsVisible:      true,
		IsVerified:     true,
		IsRecommended:  true,
		SortOrder:      12,
		SEOTitle:       "人工 SEO 标题",
		SEOTitleManual: true,
	}
	source := model.Vendor{
		Slug:          "new-vendor-site",
		Name:          " 新名称 ",
		MainProducts:  "齿轮、轴承",
		Phone:         "13800000000",
		WebsiteURL:    " https://vendor.example.com ",
		ReviewStatus:  "rejected",
		IsVisible:     false,
		IsVerified:    false,
		IsRecommended: false,
		SortOrder:     0,
		SEOTitle:      "厂商提交的 SEO 标题",
	}

	applyVendorEditableFields(&destination, source)

	if destination.Name != "新名称" || destination.Slug != "new-vendor-site" || destination.WebsiteURL != "https://vendor.example.com" || destination.MainProducts != "齿轮、轴承" {
		t.Fatalf("editable fields were not copied: %#v", destination)
	}
	if !destination.IsVisible || !destination.IsVerified || !destination.IsRecommended || destination.SortOrder != 12 || destination.ReviewStatus != "verified" {
		t.Fatalf("administrator fields were overwritten: %#v", destination)
	}
	if destination.SEOTitle != "人工 SEO 标题" || !destination.SEOTitleManual {
		t.Fatalf("vendor submission overwrote CMS SEO fields: %#v", destination)
	}
}
