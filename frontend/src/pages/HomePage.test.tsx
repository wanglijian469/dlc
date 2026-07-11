import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { HomeView } from "./HomePage";

vi.mock("../api/public", async (importOriginal) => ({ ...(await importOriginal<typeof import("../api/public")>()), getFriendLinks: vi.fn().mockResolvedValue([]) }));

describe("HomeView", () => {
  it("keeps the homepage focused on recommended, regular and processing vendors", () => {
    render(<MemoryRouter><HomeView home={{ topMenus: [{ id: 1, name: "首页", path: "/" }], sidebarMenus: [], auxiliaryMenus: [], mobileMenus: [], banner: { title: "查农机配件，找公开厂商资料" }, recommendedVendors: [{ id: 1, name: "推荐厂商甲" }], moreVendors: [{ id: 2, name: "普通厂商乙" }], processingVendors: [{ id: 3, name: "加工厂商丙" }], featuredProducts: [{ id: 8, name: "不应出现在首页的产品" }], popularCategories: [{ id: 9, name: "不应出现在首页的品类" }], stats: [], safeguards: [], join: { text: "", buttonText: "", path: "/join" } }} /></MemoryRouter>);
    expect(screen.getByText("查农机配件，找公开厂商资料")).toBeInTheDocument();
    expect(screen.getByText("推荐厂商甲")).toBeInTheDocument();
    expect(screen.getByText("普通厂商乙")).toBeInTheDocument();
    expect(screen.getByText("加工厂商丙")).toBeInTheDocument();
    expect(screen.queryByText("不应出现在首页的产品")).not.toBeInTheDocument();
    expect(screen.queryByText("不应出现在首页的品类")).not.toBeInTheDocument();
  });
});
