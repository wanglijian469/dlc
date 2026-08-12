package api

import (
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestNormalizeVendorProductIdentity(t *testing.T) {
	if got := normalizeVendorProductIdentity(" 1LF-260 / 液压犁 "); got != "1lf260液压犁" {
		t.Fatalf("normalized identity = %q", got)
	}
}

func TestScoreProductMatchIsDeterministic(t *testing.T) {
	view := model.ProductSubmissionView{ProductDraft: model.Product{CategoryID: 10}, SupplierDraft: model.ProductSupplier{VendorProductName: "液压翻转犁", VendorModel: "1LF-260", CompatibleModels: "拖拉机"}}
	product := model.Product{Name: "液压翻转犁 1LF-260", CategoryID: 11, CompatibleModels: "1LF-260、拖拉机"}
	score, reasons := scoreProductMatch(view, product, []uint{10, 11}, "")
	if score < 50 || len(reasons) < 2 {
		t.Fatalf("score = %d, reasons = %#v", score, reasons)
	}
}
