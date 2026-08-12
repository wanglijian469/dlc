import type { Menu, VendorCategory } from "../types/api";

type VendorNavigationOptions = {
  categoryPath?: (categoryId?: number) => string;
  newlyJoinedPath?: () => string;
  activeCategoryIds?: Iterable<number>;
};

export function buildVendorNavigationMenus(categories: VendorCategory[], options: VendorNavigationOptions = {}): Menu[] {
  const categoryPath = options.categoryPath || ((categoryId?: number) => categoryId ? `/vendors?vendorCategoryId=${categoryId}` : "/vendors");
  const newlyJoinedPath = options.newlyJoinedPath || (() => "/vendors?newlyJoined=true&sort=latest");
  const activeCategoryIds = new Set(options.activeCategoryIds || []);

  return [
    { id: 900000, name: "全部厂商", icon: "factory", path: categoryPath() },
    ...categories.map((category) => {
      const children = category.children || [];
      const hasActiveChild = children.some((child) => activeCategoryIds.has(child.id));
      return {
        id: 900000 + category.id,
        name: category.name,
        icon: category.icon || "factory",
        path: categoryPath(category.id),
        contextActive: activeCategoryIds.has(category.id),
        isDefaultOpen: activeCategoryIds.has(category.id) || hasActiveChild,
        children: children.length ? [
          { id: 1900000 + category.id, name: `全部${category.name}`, icon: "dot", path: categoryPath(category.id) },
          ...children.map((child) => ({
            id: 2900000 + child.id,
            name: child.name,
            icon: child.icon || "dot",
            path: categoryPath(child.id),
            contextActive: activeCategoryIds.has(child.id),
          })),
        ] : undefined,
      } satisfies Menu;
    }),
    { id: 890000, name: "新入驻厂商", icon: "clipboard-plus", path: newlyJoinedPath() },
  ];
}
