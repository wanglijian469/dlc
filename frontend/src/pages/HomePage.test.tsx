import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { dedupeHome, HomeView } from "./HomePage";

vi.mock("../api/public", async (importOriginal) => ({ ...(await importOriginal<typeof import("../api/public")>()), getFriendLinks: vi.fn().mockResolvedValue([]) }));

describe("HomeView", () => {
  it("keeps processing vendors independent while deduplicating general homepage modules", () => {
    const home = dedupeHome({
      topMenus: [], sidebarMenus: [], auxiliaryMenus: [], mobileMenus: [],
      banner: { title: "加工服务" },
      recommendedVendors: [{ id: 1, name: "加工厂商" }, { id: 2, name: "推荐厂商" }],
      moreVendors: [{ id: 1, name: "加工厂商" }, { id: 3, name: "普通厂商" }],
      processingVendors: [{ id: 1, name: "加工厂商" }],
      featuredProducts: [], popularCategories: [], stats: [], safeguards: [], join: { text: "", buttonText: "", path: "/join" },
    });

    expect((home.processingVendors || []).map((item) => item.id)).toEqual([1]);
    expect(home.recommendedVendors.map((item) => item.id)).toEqual([1, 2]);
    expect(home.moreVendors.map((item) => item.id)).toEqual([3]);
  });

  it("keeps the homepage focused on recommended, regular and processing vendors", () => {
    render(<MemoryRouter><HomeView home={{ topMenus: [{ id: 1, name: "首页", path: "/" }], sidebarMenus: [], auxiliaryMenus: [], mobileMenus: [], banner: { title: "查农机配件，找公开厂商资料" }, recommendedVendors: [{ id: 1, name: "推荐厂商甲" }], moreVendors: [{ id: 2, name: "普通厂商乙" }], processingVendors: [{ id: 3, name: "加工厂商丙" }], featuredProducts: [{ id: 8, name: "不应出现在首页的产品" }], popularCategories: [{ id: 9, name: "不应出现在首页的品类" }], stats: [], safeguards: [], join: { text: "", buttonText: "", path: "/join" } }} /></MemoryRouter>);
    expect(screen.getByText("查农机配件，找公开厂商资料")).toBeInTheDocument();
    expect(screen.getByText("推荐厂商甲")).toBeInTheDocument();
    expect(screen.getByText("普通厂商乙")).toBeInTheDocument();
    expect(screen.getByText("加工厂商丙")).toBeInTheDocument();
    expect(screen.queryByText("不应出现在首页的产品")).not.toBeInTheDocument();
    expect(screen.queryByText("不应出现在首页的品类")).not.toBeInTheDocument();
  });

  it("renders a processing vendor in both the general and processing sections", () => {
    render(<MemoryRouter><HomeView home={dedupeHome({ topMenus: [], sidebarMenus: [], auxiliaryMenus: [], mobileMenus: [], banner: { title: "首页" }, recommendedVendors: [{ id: 1, name: "加工推荐厂商" }], moreVendors: [], processingVendors: [{ id: 1, name: "加工推荐厂商" }], featuredProducts: [], popularCategories: [], stats: [], safeguards: [], join: { text: "", buttonText: "", path: "/join" } })} /></MemoryRouter>);

    expect(screen.getAllByRole("heading", { name: "加工推荐厂商" })).toHaveLength(2);
  });
});
