package database

import (
	"reflect"
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestMatchVendorCategoryNamesPrefersProvidedLevel(t *testing.T) {
	categories := []model.VendorCategory{
		{ID: 10, Name: "变速箱齿轮"},
		{ID: 11, Name: "液压系统配件"},
		{ID: 12, Name: "播种施肥配件"},
	}
	got := matchVendorCategoryNames("主营液压件、液压油泵及变速箱齿轮", categories)
	if !reflect.DeepEqual(got, []uint{10, 11}) {
		t.Fatalf("matched categories = %v, want [10 11]", got)
	}
}

func TestAddVendorCategoryAssignmentDeduplicates(t *testing.T) {
	assignments := map[uint]map[uint]bool{}
	addVendorCategoryAssignment(assignments, 7, 3)
	addVendorCategoryAssignment(assignments, 7, 3)
	if len(assignments[7]) != 1 || !assignments[7][3] {
		t.Fatalf("assignments = %#v", assignments)
	}
}

func TestDisabledVendorCategoryDeleteOrderPlacesChildrenBeforeRoots(t *testing.T) {
	children, roots := disabledVendorCategoryDeleteOrder([]model.VendorCategory{
		{ID: 14, Name: "农机易损件"},
		{ID: 24, Name: "变速箱齿轮", ParentID: 15},
		{ID: 15, Name: "传动配件"},
		{ID: 25, Name: "旋耕机/耕作机械", ParentID: 22},
	})
	if !reflect.DeepEqual(children, []uint{24, 25}) {
		t.Fatalf("child delete order = %v, want [24 25]", children)
	}
	if !reflect.DeepEqual(roots, []uint{14, 15}) {
		t.Fatalf("root delete order = %v, want [14 15]", roots)
	}
}

func TestVendorCategoryTaxonomyV2HasExpectedOrderAndParents(t *testing.T) {
	gotNames := make([]string, 0, len(vendorCategoryTaxonomyV2))
	parents := make(map[string]string, len(vendorCategoryTaxonomyV2))
	for _, definition := range vendorCategoryTaxonomyV2 {
		gotNames = append(gotNames, definition.Name)
		parents[definition.Name] = definition.ParentKey
	}
	wantNames := []string{
		"大陆村农机配件市场厂商", "庞口农机配件市场厂商", "农机整机厂", "配件生产厂",
		"传动系统厂商", "行走底盘厂商", "液压系统厂商", "发动机配套厂商", "制动换挡厂商", "电气照明厂商",
		"加工服务商", "铸造锻造", "机械加工", "热处理", "表面处理", "模具制造",
		"原材料供应商", "经销商与服务商",
	}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("taxonomy names = %v, want %v", gotNames, wantNames)
	}
	for _, name := range []string{"传动系统厂商", "行走底盘厂商", "液压系统厂商", "发动机配套厂商", "制动换挡厂商", "电气照明厂商"} {
		if parents[name] != "parts-manufacturer" {
			t.Fatalf("parent of %s = %q, want parts-manufacturer", name, parents[name])
		}
	}
	for _, name := range []string{"铸造锻造", "机械加工", "热处理", "表面处理", "模具制造"} {
		if parents[name] != "processing-provider" {
			t.Fatalf("parent of %s = %q, want processing-provider", name, parents[name])
		}
	}
}

func TestLegacyVendorCategoriesMapToNewTaxonomy(t *testing.T) {
	cases := map[string][]string{
		"传动配件":     {"parts-transmission"},
		"液压系统配件":   {"parts-hydraulic"},
		"旋耕机/耕作机械": {"machine-manufacturer"},
		"加工服务":     {"processing-provider"},
	}
	for legacyName, want := range cases {
		if got := legacyVendorCategoryTargets[legacyName]; !reflect.DeepEqual(got, want) {
			t.Fatalf("legacy mapping %s = %v, want %v", legacyName, got, want)
		}
	}
}

func TestClassifyVendorForTaxonomyV2SupportsMultipleFacets(t *testing.T) {
	vendor := model.Vendor{
		Name: "大陆村齿轮制造有限公司", Address: "河北省宁晋县大陆省级工业园区",
		MainProducts: "变速箱齿轮、链轮及传动部件", ProvidesProcessing: true,
		ProcessingServices: "数控车削、锻造、热处理及表面喷涂",
	}
	want := []string{"market-dalucun", "parts-transmission", "processing-cast-forge", "processing-machining", "processing-heat", "processing-surface"}
	if got := classifyVendorForTaxonomyV2(vendor); !reflect.DeepEqual(got, want) {
		t.Fatalf("classified keys = %v, want %v", got, want)
	}
}

func TestClassifyVendorForTaxonomyV2RecognizesMarketsAndBusinessTypes(t *testing.T) {
	vendor := model.Vendor{
		Name: "庞口农机服务中心", Address: "河北省高阳县庞口镇",
		MainProducts: "旋耕机、深松机整机", Description: "农机经销商并提供农机维修服务",
	}
	want := []string{"market-pangkou", "machine-manufacturer", "dealer-service"}
	if got := classifyVendorForTaxonomyV2(vendor); !reflect.DeepEqual(got, want) {
		t.Fatalf("classified keys = %v, want %v", got, want)
	}
}
