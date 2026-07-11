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
