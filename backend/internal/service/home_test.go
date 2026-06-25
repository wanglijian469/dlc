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
