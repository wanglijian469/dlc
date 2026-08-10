import { cleanup, fireEvent, render, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { HomePayload } from "../../types/api";
import { getHome, getLayoutConfig, getVendorCategories } from "../../api/public";
import { SiteProvider } from "../../contexts/SiteContext";
import { PageFrame } from "./PageFrame";

vi.mock("../../api/public", () => ({
  getHome: vi.fn(),
	getLayoutConfig: vi.fn(),
	getVendorCategories: vi.fn(),
}));

const homePayload: HomePayload = {
  topMenus: [],
  sidebarMenus: [
    {
      id: 1,
      name: "首页",
      icon: "home",
      path: "/",
    },
    {
      id: 2,
      name: "传动配件",
      icon: "cog",
      path: "/products?keyword=传动配件",
      children: [{ id: 3, name: "变速箱齿轮", icon: "dot", path: "/products?keyword=变速箱齿轮" }],
    },
  ],
  auxiliaryMenus: [{ id: 10, name: "提交厂商", path: "/submit" }],
  mobileMenus: [],
  banner: { title: "" },
  recommendedVendors: [],
  moreVendors: [],
  stats: [],
  safeguards: [],
  join: { text: "", buttonText: "", path: "/" },
};

describe("PageFrame", () => {
  beforeEach(() => {
    vi.mocked(getHome).mockResolvedValue(homePayload);
	vi.mocked(getLayoutConfig).mockResolvedValue({
	  siteMeta: { siteName: "大陆农机配件", brandMark: "农", submitVendorText: "厂商入驻", adminLoginText: "厂商登录", mobileBrandName: "大陆农机配件", mobileBrandMark: "农" },
	  theme: { primaryColor: "#1559c7", accentColor: "#0d8b6f" },
	  topMenus: [{ id: 1, name: "首页", path: "/" }, { id: 2, name: "厂商资源", path: "/vendors" }, { id: 3, name: "配件货源", path: "/products" }, { id: 4, name: "加工服务", path: "/service" }],
	  sidebarMenus: homePayload.sidebarMenus,
	  auxiliaryMenus: [{ id: 10, name: "厂商入驻", path: "/join" }],
	  mobileMenus: [], mobileBottomMenus: [], version: "test",
	});
	vi.mocked(getVendorCategories).mockResolvedValue([
	  { id: 10, name: "配件生产厂", vendorCount: 3, children: [{ id: 11, parentId: 10, name: "传动系统厂商", vendorCount: 1 }] },
	  { id: 20, name: "加工服务商", vendorCount: 0 },
	]);
  });

  afterEach(() => cleanup());

  it.each([
    ["/products", "配件货源"],
    ["/vendors", "厂商资源"],
    ["/service", "加工服务"],
    ["/purchase", "采购信息"],
  ])("renders Chinese main navigation and highlights %s", (path, activeLabel) => {
    const { container } = render(
      <MemoryRouter initialEntries={[path]}>
        <PageFrame title="配件产品">
          <div>列表内容</div>
        </PageFrame>
      </MemoryRouter>,
    );

    const topNav = container.querySelector(".top-nav") as HTMLElement;
    expect(topNav.querySelector("a.active")).toHaveTextContent(activeLabel);
    expect(topNav).toHaveTextContent("配件货源");
    expect(topNav).toHaveTextContent("厂商资源");
    expect(topNav).toHaveTextContent("加工服务");
    expect(topNav).toHaveTextContent("采购信息");
  });

  it("renders the same collapsed sidebar categories on desktop subpages and lets parent rows toggle", async () => {
    const { container } = render(
      <MemoryRouter initialEntries={["/products"]}>
        <PageFrame title="配件产品">
          <div>产品列表</div>
        </PageFrame>
      </MemoryRouter>,
    );

    await waitFor(() => expect(getHome).toHaveBeenCalled());

    const sidebar = container.querySelector(".sidebar") as HTMLElement;
    const parent = within(sidebar).getByRole("button", { name: "传动配件" });
    expect(parent).toBeInTheDocument();
    expect(within(sidebar).queryByRole("link", { name: /变速箱齿轮/ })).not.toBeInTheDocument();
    fireEvent.click(parent);
    expect(within(sidebar).getByRole("link", { name: /变速箱齿轮/ })).toBeInTheDocument();
    expect(within(sidebar).queryByRole("link", { name: "提交厂商" })).not.toBeInTheDocument();
  });

  it("renders clickable parent levels before the current page", () => {
    const { container } = render(
      <MemoryRouter initialEntries={["/products/7"]}>
        <PageFrame breadcrumbs={[{ label: "配件产品", path: "/products" }, { label: "液压系统配件", path: "/products?categoryId=5" }]} title="液压油泵总成">
          <div>产品详情</div>
        </PageFrame>
      </MemoryRouter>,
    );

    const breadcrumbs = within(container).getByRole("navigation", { name: "面包屑" });
    expect(within(breadcrumbs).getByRole("link", { name: "首页" })).toHaveAttribute("href", "/");
    expect(within(breadcrumbs).getByRole("link", { name: "配件产品" })).toHaveAttribute("href", "/products");
    expect(within(breadcrumbs).getByRole("link", { name: "液压系统配件" })).toHaveAttribute("href", "/products?categoryId=5");
    expect(within(breadcrumbs).getByText("液压油泵总成")).toHaveAttribute("aria-current", "page");
  });

  it("uses vendor navigation on ordinary pages and marks every assigned category", async () => {
	const { container } = render(
	  <MemoryRouter initialEntries={["/service"]}>
		<SiteProvider>
		  <PageFrame navigationCategoryIds={[11, 20]} title="加工服务"><div>服务内容</div></PageFrame>
		</SiteProvider>
	  </MemoryRouter>,
	);
	const sidebar = container.querySelector(".sidebar") as HTMLElement;
	expect(await within(sidebar).findByRole("link", { name: /传动系统厂商/ })).toHaveClass("context-active");
	expect(within(sidebar).getByText("厂商分类")).toBeInTheDocument();
	expect(within(sidebar).getByText("加工服务商").closest(".menu-row")).toHaveClass("selected");
	expect(within(sidebar).getAllByText("所属")).toHaveLength(2);
	expect(within(sidebar).queryByRole("link", { name: "厂商入驻" })).not.toBeInTheDocument();
	expect(container.querySelector(".mobile-drawer-join")).toHaveTextContent("厂商入驻");
  });

  it("keeps product navigation on product detail routes", async () => {
	const { container } = render(
	  <MemoryRouter initialEntries={["/products/7"]}>
		<SiteProvider><PageFrame title="产品详情"><div>产品内容</div></PageFrame></SiteProvider>
	  </MemoryRouter>,
	);
	await waitFor(() => expect(within(container.querySelector(".sidebar") as HTMLElement).getByRole("button", { name: "传动配件" })).toBeInTheDocument());
	const sidebar = container.querySelector(".sidebar") as HTMLElement;
	expect(within(sidebar).getByText("配件分类")).toBeInTheDocument();
	expect(within(sidebar).getByRole("button", { name: "传动配件" })).toBeInTheDocument();
	expect(within(sidebar).queryByText("全部厂商")).not.toBeInTheDocument();
  });
});
