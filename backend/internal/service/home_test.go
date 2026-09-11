package service

import (
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestBuildMenuTreeSortsAndNestsMenus(t *testing.T) {
	menus := []model.Menu{
		{ID: 3, Name: "变速箱齿轮", ParentID: 2, SortOrder: 1, IsEnabled: true},
		{ID: 1, Name: "首页", ParentID: 0, SortOrder: 1, IsEnabled: true},
		{ID: 2, Name: "传动配件", ParentID: 0, SortOrder: 2, IsEnabled: true, IsDefaultOpen: true},
		{ID: 4, Name: "禁用", ParentID: 0, SortOrder: 0, IsEnabled: false},
	}

	tree := BuildMenuTree(menus)
	if len(tree) != 2 {
		t.Fatalf("len(tree) = %d, want 2", len(tree))
	}
	if tree[0].Name != "首页" {
		t.Fatalf("first menu = %q, want 首页", tree[0].Name)
	}
	if len(tree[1].Children) != 1 || tree[1].Children[0].Name != "变速箱齿轮" {
		t.Fatalf("children = %#v, want nested child", tree[1].Children)
	}
}

func TestNormalizeTopMenusRestoresRequiredEntriesAndOrder(t *testing.T) {
	menus := []model.Menu{
		{ID: 91, Name: "旧供求名称", Path: "/purchase", Icon: "old", MenuType: "top", SortOrder: 1, IsEnabled: true},
		{ID: 92, Name: "自定义入口", Path: "/custom", MenuType: "top", SortOrder: 2, IsEnabled: true},
	}

	got := NormalizeTopMenus(menus)
	wantPaths := []string{"/", "/vendors", "/products", "/service", "/custom"}
	if len(got) != len(wantPaths) {
		t.Fatalf("len(menus) = %d, want %d: %#v", len(got), len(wantPaths), got)
	}
	for index, path := range wantPaths {
		if got[index].Path != path {
			t.Fatalf("menu[%d].Path = %q, want %q", index, got[index].Path, path)
		}
	}
	if got[4].ID != 92 {
		t.Fatalf("custom menu not preserved: %#v", got[4])
	}
}

func TestNormalizeHomeModulesFiltersRetiredTypesDuplicatesAndSorts(t *testing.T) {
	modules := []HomeModule{{Type: "join", Visible: true, SortOrder: 30}, {Type: "recommendedVendors", Visible: true, Limit: -2, SortOrder: 10}, {Type: "recommendedVendors", Visible: true, SortOrder: 20}, {Type: "unknown", Visible: true}}
	normalizeHomeModules(&modules)
	if len(modules) != 3 || modules[0].Type != "recommendedVendors" {
		t.Fatalf("normalized modules = %#v", modules)
	}
	if modules[0].Limit != 0 {
		t.Fatalf("negative limit was not normalized: %d", modules[0].Limit)
	}
}

func TestDefaultHomeModulesContainsOnlyRenderedTypes(t *testing.T) {
	modules := DefaultHomeModules()
	if len(modules) != 3 {
		t.Fatalf("default modules = %d, want 3", len(modules))
	}
	for index := 1; index < len(modules); index++ {
		if modules[index-1].SortOrder >= modules[index].SortOrder {
			t.Fatalf("modules are not sorted: %#v", modules)
		}
	}
}

func TestMergeCategoryMenusAddsChildrenAndKeepsShortcuts(t *testing.T) {
	menus := []model.Menu{
		{ID: 10, Name: "播种施肥配件", MenuType: "sidebar", Path: "/products?keyword=播种", SortOrder: 10, IsEnabled: true},
		{ID: 11, Name: "排种器", ParentID: 10, MenuType: "sidebar", Path: "/products?keyword=排种器", SortOrder: 10, IsEnabled: true},
	}
	categories := []model.Category{
		{ID: 5, Name: "播种施肥配件", SortOrder: 20, IsEnabled: true},
		{ID: 6, Name: "旋耕机/耕作机械", ParentID: 5, SortOrder: 5, IsEnabled: true},
	}

	tree := MergeCategoryMenus(menus, categories)
	if len(tree) != 1 || tree[0].Path != "/products?categoryId=5" {
		t.Fatalf("root menu = %#v, want category-backed root", tree)
	}
	if len(tree[0].Children) != 2 {
		t.Fatalf("children = %#v, want category child and shortcut", tree[0].Children)
	}
	if tree[0].Children[0].Name != "旋耕机/耕作机械" || tree[0].Children[0].Path != "/products?categoryId=6" {
		t.Fatalf("first child = %#v, want category child", tree[0].Children[0])
	}
	if tree[0].Children[1].Name != "排种器" {
		t.Fatalf("shortcut was not preserved: %#v", tree[0].Children)
	}
}

func TestMergeCategoryMenusPrefersCategoryOnDuplicateName(t *testing.T) {
	menus := []model.Menu{
		{ID: 20, Name: "播种施肥配件", MenuType: "sidebar", SortOrder: 10, IsEnabled: true},
		{ID: 21, Name: "开沟器", ParentID: 20, MenuType: "sidebar", Path: "/products?keyword=开沟器", SortOrder: 10, IsEnabled: true},
	}
	categories := []model.Category{
		{ID: 7, Name: "播种施肥配件", SortOrder: 10, IsEnabled: true},
		{ID: 8, Name: "开沟器", ParentID: 7, SortOrder: 10, IsEnabled: true},
	}

	tree := MergeCategoryMenus(menus, categories)
	if len(tree[0].Children) != 1 {
		t.Fatalf("duplicate child was not removed: %#v", tree[0].Children)
	}
	if tree[0].Children[0].Path != "/products?categoryId=8" {
		t.Fatalf("duplicate shortcut won over category: %#v", tree[0].Children[0])
	}
}

func TestMergeCategoryMenusHidesChildrenOfDisabledRoot(t *testing.T) {
	menus := []model.Menu{
		{ID: 40, Name: "已停用一级", MenuType: "sidebar", Path: "/products?categoryId=30", IsEnabled: true},
		{ID: 41, Name: "历史快捷项", ParentID: 40, MenuType: "sidebar", IsEnabled: true},
	}
	categories := []model.Category{
		{ID: 30, Name: "已停用一级", IsEnabled: false},
		{ID: 31, Name: "仍启用的二级", ParentID: 30, IsEnabled: true},
	}
	if tree := MergeCategoryMenus(menus, categories); len(tree) != 0 {
		t.Fatalf("disabled category tree should be hidden: %#v", tree)
	}
}

func TestMergeCategoryMenusUsesStableAnchorWhenCategoryIsRenamed(t *testing.T) {
	menus := []model.Menu{
		{ID: 70, CategoryID: 9, Name: "old name", MenuType: "sidebar", Path: "/products?keyword=old", IsEnabled: true},
		{ID: 71, ParentID: 70, Name: "shortcut", MenuType: "sidebar", Path: "/products?keyword=shortcut", IsEnabled: true},
	}
	categories := []model.Category{{ID: 9, Name: "new name", IsEnabled: true}}

	tree := MergeCategoryMenus(menus, categories)
	if len(tree) != 1 || tree[0].Name != "new name" || tree[0].Path != "/products?categoryId=9" {
		t.Fatalf("renamed category did not replace legacy anchor: %#v", tree)
	}
	if len(tree[0].Children) != 1 || tree[0].Children[0].Name != "shortcut" {
		t.Fatalf("legacy shortcut was not retained: %#v", tree[0].Children)
	}
}

func TestMergeMobileCategoryMenusUsesRootsAndKeepsManualShortcut(t *testing.T) {
	menus := []model.Menu{
		{ID: 80, CategoryID: 9, Name: "old name", MenuType: "mobile", SortOrder: 1, IsEnabled: true},
		{ID: 81, Name: "all", MenuType: "mobile", Path: "/products", SortOrder: 99, IsEnabled: true},
	}
	categories := []model.Category{
		{ID: 9, Name: "new name", Icon: "cog", SortOrder: 1, IsEnabled: true},
		{ID: 10, Name: "child", ParentID: 9, IsEnabled: true},
	}

	menus = MergeMobileCategoryMenus(menus, categories)
	if len(menus) != 2 || menus[0].Name != "new name" || menus[0].Path != "/products?categoryId=9" {
		t.Fatalf("mobile categories = %#v", menus)
	}
	if menus[1].Name != "all" {
		t.Fatalf("manual mobile shortcut was not retained: %#v", menus)
	}
}
