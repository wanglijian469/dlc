package service

import (
	"strings"
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestSuggestVendorSEOUsesRealDirectoryFields(t *testing.T) {
	vendor := model.Vendor{
		Name:          "河北冀农农机具有限公司",
		ShortName:     "河北冀农农机具",
		Province:      "河北省",
		City:          "石家庄市",
		MainProducts:  "旋耕机链条、齿轮、刀片、齿轮",
		ServiceModels: "旋耕机、微耕机",
	}
	got := SuggestVendorSEO(vendor)
	if got.SEOTitle != "河北冀农农机具｜旋耕机链条、齿轮厂家" {
		t.Fatalf("title = %q", got.SEOTitle)
	}
	if runeLen(got.SEOTitle) > VendorSEOTitleLimit || runeLen(got.SEODescription) > VendorSEODescriptionLimit {
		t.Fatalf("suggestion exceeds limits: %#v", got)
	}
	for _, want := range []string{"河北冀农农机具有限公司", "河北省石家庄市", "旋耕机链条", "适配旋耕机"} {
		if !strings.Contains(got.SEODescription, want) {
			t.Fatalf("description %q missing %q", got.SEODescription, want)
		}
	}
}

func TestSuggestVendorSEODescribesProcessingWithoutInventedClaims(t *testing.T) {
	got := SuggestVendorSEO(model.Vendor{
		Name:               "山东精工农机有限公司",
		MainProducts:       "轴套，齿轮坯",
		ProvidesProcessing: true,
		ProcessingServices: "数控车削、焊接加工",
	})
	if !strings.Contains(got.SEOTitle, "轴套、齿轮坯加工厂家") || !strings.Contains(got.SEODescription, "数控车削、焊接加工") {
		t.Fatalf("processing suggestion = %#v", got)
	}
	for _, unsupported := range []string{"领先", "第一", "专业认证"} {
		if strings.Contains(got.SEOTitle+got.SEODescription, unsupported) {
			t.Fatalf("suggestion contains unsupported claim %q", unsupported)
		}
	}
}

func TestApplyVendorSEOPreservesEachManualFieldIndependently(t *testing.T) {
	vendor := model.Vendor{
		Name:                 "测试农机配件有限公司",
		MainProducts:         "链条、齿轮",
		SEOTitle:             "人工标题",
		SEOTitleManual:       true,
		SEODescription:       "旧自动摘要",
		SEODescriptionManual: false,
	}
	ApplyVendorSEO(&vendor)
	if vendor.SEOTitle != "人工标题" {
		t.Fatalf("manual title was overwritten: %q", vendor.SEOTitle)
	}
	if vendor.SEODescription == "旧自动摘要" || !strings.Contains(vendor.SEODescription, "链条、齿轮") {
		t.Fatalf("automatic description was not refreshed: %q", vendor.SEODescription)
	}

	vendor.SEOTitle = ""
	ApplyVendorSEO(&vendor)
	if vendor.SEOTitleManual || vendor.SEOTitle == "" {
		t.Fatalf("blank manual title did not return to automatic mode: %#v", vendor)
	}
}

func TestSuggestVendorSEOFallsBackSafelyForSparseProfile(t *testing.T) {
	got := SuggestVendorSEO(model.Vendor{Name: "测试厂商"})
	if got.SEOTitle != "测试厂商｜农机配件厂家" {
		t.Fatalf("title = %q", got.SEOTitle)
	}
	if !strings.Contains(got.SEODescription, "厂商信息页面") || !strings.Contains(got.SEODescription, "联系方式") {
		t.Fatalf("description = %q", got.SEODescription)
	}
}
