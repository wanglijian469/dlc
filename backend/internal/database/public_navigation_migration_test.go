package database

import (
	"reflect"
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestOrderMenusByPathPrioritizesVendorsAndPreservesCustomRows(t *testing.T) {
	menus := []model.Menu{
		{ID: 1, Name: "首页", Path: "/"},
		{ID: 2, Name: "配件货源", Path: "/products"},
		{ID: 8, Name: "行业指南", Path: "/guides"},
		{ID: 3, Name: "厂商资源", Path: "/vendors"},
		{ID: 4, Name: "加工服务", Path: "/service"},
		{ID: 9, Name: "采购信息", Path: "/purchase"},
	}
	ordered := orderMenusByPath(menus, []string{"/", "/vendors", "/products", "/service"})
	paths := make([]string, 0, len(ordered))
	for _, menu := range ordered {
		paths = append(paths, menu.Path)
	}
	want := []string{"/", "/vendors", "/products", "/service", "/guides", "/purchase"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("ordered paths = %v, want %v", paths, want)
	}
	if ordered[1].Name != "厂商资源" || ordered[2].Name != "配件货源" {
		t.Fatalf("menu labels changed: %#v", ordered)
	}
}

func TestOrderMenusByPathIsIdempotent(t *testing.T) {
	menus := []model.Menu{{ID: 1, Path: "/"}, {ID: 3, Path: "/vendors"}, {ID: 2, Path: "/products"}, {ID: 4, Path: "/service"}}
	preferred := []string{"/", "/vendors", "/products", "/service"}
	once := orderMenusByPath(menus, preferred)
	twice := orderMenusByPath(once, preferred)
	if !reflect.DeepEqual(once, twice) {
		t.Fatalf("second ordering changed rows: once=%v twice=%v", once, twice)
	}
}

func TestOrderMenusByPathPreservesDuplicateAndCustomRows(t *testing.T) {
	menus := []model.Menu{
		{ID: 1, Name: "自定义厂商入口", Path: "/vendors", SortOrder: 1},
		{ID: 2, Name: "首页", Path: "/", SortOrder: 2},
		{ID: 3, Name: "厂商资源", Path: "/vendors", SortOrder: 3},
		{ID: 4, Name: "帮助", Path: "/help", SortOrder: 4},
	}
	ordered := orderMenusByPath(menus, []string{"/", "/vendors", "/products"})
	if len(ordered) != len(menus) {
		t.Fatalf("expected every menu row to be preserved, got %d of %d", len(ordered), len(menus))
	}
	if got := []uint{ordered[0].ID, ordered[1].ID, ordered[2].ID, ordered[3].ID}; !reflect.DeepEqual(got, []uint{2, 1, 3, 4}) {
		t.Fatalf("unexpected order with duplicate paths: %v", got)
	}
}
