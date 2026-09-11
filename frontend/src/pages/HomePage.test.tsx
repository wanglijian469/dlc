import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { dedupeHome, HomeView } from "./HomePage";
import type { HomePayload } from "../types/api";

vi.mock("../api/public", async (importOriginal) => ({ ...(await importOriginal<typeof import("../api/public")>()), getFriendLinks: vi.fn().mockResolvedValue([]) }));

describe("HomeView", () => {
  const moduleFixture: HomePayload = {
    topMenus: [], sidebarMenus: [], auxiliaryMenus: [], mobileMenus: [], banner: { title: "目录入口" },
    recommendedVendors: [{ id: 81, name: "推荐甲" }, { id: 82, name: "推荐乙" }],
    moreVendors: [{ id: 83, name: "普通甲" }], processingVendors: [], stats: [], safeguards: [],
    join: { text: "", buttonText: "", path: "/join" },
  };
  it("preserves configured module visibility, titles and limits", () => {
    render(<MemoryRouter><HomeView home={{ ...moduleFixture, modules: [
      { type: "recommendedVendors", title: "精选合作伙伴", visible: true, limit: 1, path: "/vendors?sort=recommended", sortOrder: 10 },
      { type: "moreVendors", title: "普通厂商", visible: false, limit: 5, sortOrder: 20 },
      { type: "processingServices", title: "加工服务", visible: false, limit: 4, sortOrder: 30 },
    ] }} /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "精选合作伙伴" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "推荐甲" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "推荐乙" })).not.toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "普通甲" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /浏览全部厂商/ })).toHaveAttribute("href", "/vendors?sort=recommended");
  });
  it("provides a directory entry when no vendors are available", () => {
    render(<MemoryRouter><HomeView home={{ ...moduleFixture, recommendedVendors: [], moreVendors: [] }} /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "公开厂商资料正在完善" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看厂商目录" })).toHaveAttribute("href", "/vendors");
  });
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
