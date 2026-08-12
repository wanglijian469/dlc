import { describe, expect, it } from "vitest";
import { buildVendorNavigationMenus } from "./vendorNavigation";

describe("buildVendorNavigationMenus", () => {
  const categories = [
	{ id: 1, name: "大陆村农机配件市场厂商", vendorCount: 2 },
	{ id: 2, name: "配件生产厂", vendorCount: 3, children: [{ id: 21, parentId: 2, name: "传动系统厂商", vendorCount: 0 }] },
  ];

  it("builds the fixed vendor-first order without category badges", () => {
	const menus = buildVendorNavigationMenus(categories);
	expect(menus.map((menu) => menu.name)).toEqual(["全部厂商", "大陆村农机配件市场厂商", "配件生产厂", "新入驻厂商"]);
	expect(menus[1].badge).toBeUndefined();
	expect(menus[2].children?.map((menu) => menu.name)).toEqual(["全部配件生产厂", "传动系统厂商"]);
	expect(menus[2].children?.every((menu) => menu.badge === undefined)).toBe(true);
	expect(menus.at(-1)?.path).toBe("/vendors?newlyJoined=true&sort=latest");
  });

  it("marks all direct assignments and opens every related parent", () => {
	const menus = buildVendorNavigationMenus(categories, { activeCategoryIds: [1, 21] });
	expect(menus[1].contextActive).toBe(true);
	expect(menus[2].isDefaultOpen).toBe(true);
	expect(menus[2].contextActive).toBe(false);
	expect(menus[2].children?.[1].contextActive).toBe(true);
  });

  it("supports filter-preserving links on the vendor directory", () => {
	const menus = buildVendorNavigationMenus(categories, {
	  categoryPath: (id) => `/vendors?province=河北${id ? `&vendorCategoryId=${id}` : ""}`,
	  newlyJoinedPath: () => "/vendors?province=河北&newlyJoined=true&sort=latest",
	});
	expect(menus[1].path).toContain("province=河北");
	expect(menus[2].children?.[1].path).toContain("vendorCategoryId=21");
	expect(menus.at(-1)?.path).toContain("newlyJoined=true");
  });
});
