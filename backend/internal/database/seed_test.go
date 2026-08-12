package database

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultSeedDoesNotContainDemoHostsOrFakeStats(t *testing.T) {
	payload, err := json.Marshal(DefaultSeed())
	if err != nil {
		t.Fatal(err)
	}
	content := strings.ToLower(string(payload))
	for _, forbidden := range []string{"dummyimage.com", "example.com", "2000+", "10万+", "5000+", "30+"} {
		if strings.Contains(content, strings.ToLower(forbidden)) {
			t.Fatalf("default seed contains forbidden demo value %q", forbidden)
		}
	}
}

func TestDefaultBannerUsesUpdatedCopyAndKeywords(t *testing.T) {
	seed := DefaultSeed()
	if len(seed.Banners) != 1 {
		t.Fatalf("banner count = %d, want 1", len(seed.Banners))
	}
	banner := seed.Banners[0]
	if banner.Title != "农机供应链，查农机配件，厂商信息" {
		t.Fatalf("banner title = %q", banner.Title)
	}
	if banner.HotKeywordsRaw != "收割机链条,齿轮,皮带,液压油泵,刀片,滤芯" {
		t.Fatalf("hot keywords = %q", banner.HotKeywordsRaw)
	}
}

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
	if len(seed.Vendors) != 1 {
		t.Fatalf("default vendors = %d, want only the verified-source vendor", len(seed.Vendors))
	}
}

func TestDefaultSeedContainsExpandedSidebarMenuTree(t *testing.T) {
	seed := DefaultSeedWithDemo(true)
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

func TestDefaultSeedDoesNotDefaultOpenSidebarMenus(t *testing.T) {
	seed := DefaultSeedWithDemo(true)
	for _, menu := range seed.Menus {
		if menu.MenuType == "sidebar" && menu.IsDefaultOpen {
			t.Fatalf("sidebar menu %q should not be default open", menu.Key)
		}
	}
}

func TestDemoSeedVendorsAreQuarantined(t *testing.T) {
	seed := DefaultSeedWithDemo(true)
	demos := 0
	for _, vendor := range seed.Vendors {
		if vendor.DataOrigin != "demo" {
			continue
		}
		demos++
		if vendor.PublicationStatus != "hidden" || vendor.IsVisible || vendor.IsRecommended || vendor.IsVerified {
			t.Fatalf("demo vendor %q is not quarantined", vendor.Name)
		}
	}
	if demos != 2 {
		t.Fatalf("demo vendors = %d, want 2", demos)
	}
}

func TestDemoSeedExcludesRetiredVendors(t *testing.T) {
	retired := map[string]bool{
		"江苏东成农机配件有限公司": true, "河南中联农机制造有限公司": true, "安徽豪华农机配件有限公司": true,
		"山东万鑫农机配件有限公司": true, "宁波动力机械有限公司": true, "浙江汉丰农机有限公司": true,
		"河北力捷机械有限公司": true, "辽宁佳丰农机配件有限公司": true, "四川川沃农机有限公司": true,
		"陕西恒农农机配件有限公司": true,
	}
	for _, vendor := range DefaultSeedWithDemo(true).Vendors {
		if retired[vendor.Name] {
			t.Fatalf("retired vendor %q remains in the seed", vendor.Name)
		}
	}
}

func TestDefaultSeedVendorsDoNotUseFactoryDummyCovers(t *testing.T) {
	seed := DefaultSeed()
	for _, vendor := range seed.Vendors {
		if strings.Contains(vendor.CoverImage, "Factory") {
			t.Fatalf("vendor %q uses old factory dummy cover %q", vendor.Name, vendor.CoverImage)
		}
	}
}

func TestDefaultSeedVendorsContainRichProfileFields(t *testing.T) {
	seed := DefaultSeedWithDemo(true)
	if len(seed.Vendors) == 0 {
		t.Fatal("seed vendors should not be empty")
	}
	vendor := reflect.ValueOf(seed.Vendors[0])
	for _, field := range []string{"EstablishedYear", "FactoryArea", "EmployeeCount", "AnnualCapacity", "Equipment", "Certifications"} {
		value := vendor.FieldByName(field)
		if !value.IsValid() {
			t.Fatalf("Vendor missing rich profile field %s", field)
		}
		if strings.TrimSpace(value.String()) == "" {
			t.Fatalf("seed vendor field %s should not be empty", field)
		}
	}
}

func TestDefaultSeedContainsProcessingTagsAndVendors(t *testing.T) {
	seed := DefaultSeedWithDemo(true)
	processingTags := 0
	for _, tag := range seed.Tags {
		if tag.TagType == "processing" {
			processingTags++
		}
	}
	if processingTags < 4 {
		t.Fatalf("processing tags = %d, want at least 4", processingTags)
	}

	processingVendors := 0
	for _, vendor := range seed.Vendors {
		if !vendor.ProvidesProcessing {
			continue
		}
		processingVendors++
		if strings.TrimSpace(vendor.ProcessingServices) == "" {
			t.Fatalf("processing vendor %q missing services", vendor.Name)
		}
		if strings.TrimSpace(vendor.ProcessingEquipment) == "" {
			t.Fatalf("processing vendor %q missing equipment", vendor.Name)
		}
	}
	if processingVendors != 2 {
		t.Fatalf("processing vendors = %d, want 2", processingVendors)
	}
}

func TestDefaultSeedContainsHBJinongPublicImport(t *testing.T) {
	seed := DefaultSeed()
	var vendorIndex int
	var vendorName string
	var reviewStatus string
	var mainProducts string
	for i, vendor := range seed.Vendors {
		if vendor.WebsiteURL != "https://www.hbjinong.com/" {
			continue
		}
		vendorIndex = i + 1
		vendorName = vendor.Name
		reviewStatus = vendor.ReviewStatus
		mainProducts = vendor.MainProducts
	}
	if vendorIndex == 0 {
		t.Fatal("missing 河北冀农 vendor imported from public website")
	}
	if vendorName != "河北冀农农机具有限公司" {
		t.Fatalf("vendor name = %q", vendorName)
	}
	if reviewStatus != "pending" {
		t.Fatalf("reviewStatus = %q, want pending", reviewStatus)
	}
	for _, want := range []string{"液压翻转犁", "旋耕机", "驱动耙"} {
		if !strings.Contains(mainProducts, want) {
			t.Fatalf("mainProducts = %q, want %q", mainProducts, want)
		}
	}

	productNames := map[string]bool{}
	for _, product := range seed.Products {
		if int(product.VendorIDValue()) == vendorIndex {
			productNames[product.Name] = true
		}
	}
	for _, want := range []string{"液压翻转犁1LF-360", "液压翻转犁1LF-260", "3米驱动耙"} {
		if !productNames[want] {
			t.Fatalf("missing 河北冀农 representative product %q", want)
		}
	}
}
