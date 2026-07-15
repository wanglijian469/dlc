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

func TestNormalizeHomeModulesFiltersDuplicatesAndSorts(t *testing.T) {
	modules := []HomeModule{{Type: "join", Visible: true, SortOrder: 30}, {Type: "categories", Visible: true, Limit: -2, SortOrder: 10}, {Type: "categories", Visible: true, SortOrder: 20}, {Type: "unknown", Visible: true}}
	normalizeHomeModules(&modules)
	if len(modules) != 2 || modules[0].Type != "categories" || modules[1].Type != "join" {
		t.Fatalf("normalized modules = %#v", modules)
	}
	if modules[0].Limit != 0 {
		t.Fatalf("negative limit was not normalized: %d", modules[0].Limit)
	}
}

func TestDefaultHomeModulesContainsAllSupportedTypes(t *testing.T) {
	modules := DefaultHomeModules()
	if len(modules) != 7 {
		t.Fatalf("default modules = %d, want 7", len(modules))
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
