import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getFilterOptions, getHome, getLayoutConfig, getVendorCategories, listVendors } from "../api/public";
import { SiteProvider } from "../contexts/SiteContext";
import { VendorsPage } from "./VendorsPage";

vi.mock("../api/public", () => ({
  getFilterOptions: vi.fn(),
  getHome: vi.fn(),
	getLayoutConfig: vi.fn(),
	getVendorCategories: vi.fn(),
  listVendors: vi.fn(),
}));

const mockedGetFilterOptions = vi.mocked(getFilterOptions);
const mockedGetHome = vi.mocked(getHome);
const mockedGetLayoutConfig = vi.mocked(getLayoutConfig);
const mockedGetVendorCategories = vi.mocked(getVendorCategories);
const mockedListVendors = vi.mocked(listVendors);

function renderVendors(path = "/vendors") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <SiteProvider><VendorsPage /></SiteProvider>
    </MemoryRouter>,
  );
}

describe("VendorsPage", () => {
  beforeEach(() => {
	mockedGetLayoutConfig.mockResolvedValue({
	  siteMeta: { siteName: "大陆农机配件", brandMark: "农", submitVendorText: "厂商入驻", adminLoginText: "厂商登录", mobileBrandName: "大陆农机配件", mobileBrandMark: "农" },
	  theme: { primaryColor: "#1559c7", accentColor: "#0d8b6f" },
	  topMenus: [{ id: 1, name: "首页", path: "/" }, { id: 2, name: "厂商资源", path: "/vendors" }, { id: 3, name: "配件货源", path: "/products" }, { id: 4, name: "加工服务", path: "/service" }],
	  sidebarMenus: [], auxiliaryMenus: [{ id: 5, name: "厂商入驻", path: "/join" }], mobileMenus: [], mobileBottomMenus: [], version: "test",
	});
    mockedGetHome.mockResolvedValue({
      topMenus: [],
      sidebarMenus: [],
      auxiliaryMenus: [],
      mobileMenus: [],
      banner: { title: "" },
      recommendedVendors: [],
      moreVendors: [],
      stats: [],
      safeguards: [],
      join: { text: "", buttonText: "", path: "/" },
    });
    mockedGetFilterOptions.mockResolvedValue({
      provinces: ["山东", "河北"],
      categories: [{ id: 1, name: "传动配件" }],
      serviceTags: [{ id: 1, name: "源头厂商" }],
    });
	mockedGetVendorCategories.mockResolvedValue([{ id: 10, name: "配件生产厂", vendorCount: 1, children: [{ id: 11, name: "传动系统厂商", parentId: 10, vendorCount: 0 }] }]);
    mockedListVendors.mockResolvedValue({
      items: [{ id: 1, name: "山东沃得农机配件有限公司", province: "山东", mainProducts: "链条、齿轮", tags: [{ id: 1, name: "源头厂商" }] }],
      page: 1,
      pageSize: 12,
      total: 1,
    });
  });

  it("renders vendor directory results with filters", async () => {
    renderVendors();

    expect(await screen.findByRole("heading", { name: "厂商资源" })).toBeInTheDocument();
    expect(await screen.findByText("山东沃得农机配件有限公司")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "搜索" })).toBeInTheDocument();
	  await waitFor(() => expect(mockedListVendors).toHaveBeenCalledWith(expect.objectContaining({ page: 1, pageSize: 12 })));
	expect(screen.getAllByText("厂商分类").length).toBeGreaterThanOrEqual(2);
	fireEvent.click(screen.getAllByRole("button", { name: /配件生产厂/ }).at(-1)!);
	expect(screen.getByText("招商中")).toBeInTheDocument();
  });

	it("filters with the dedicated vendor category and shows an enrollment CTA for an empty category", async () => {
	  mockedListVendors.mockResolvedValueOnce({ items: [], page: 1, pageSize: 12, total: 0 });
	  renderVendors("/vendors?vendorCategoryId=11&province=山东");
	  expect(await screen.findByText("“传动系统厂商”暂无入驻企业，招商进行中")).toBeInTheDocument();
	  expect(screen.getAllByRole("link", { name: "申请厂商入驻" })[0]).toHaveAttribute("href", "/join");
	  expect(screen.getByRole("button", { name: /传动系统厂商/ })).toBeInTheDocument();
	  await waitFor(() => expect(mockedListVendors).toHaveBeenCalledWith(expect.objectContaining({ vendorCategoryId: "11", province: "山东" })));
	});

	it("shows the dynamic newly joined entry and applies the 90-day filter", async () => {
	  mockedListVendors.mockResolvedValueOnce({ items: [], page: 1, pageSize: 12, total: 0 });
	  renderVendors("/vendors?newlyJoined=true&sort=latest&province=山东");
	  expect(await screen.findByText("近 90 天暂无新入驻厂商")).toBeInTheDocument();
	  expect(screen.getByRole("button", { name: /新入驻厂商/ })).toBeInTheDocument();
	  expect(screen.getAllByRole("link", { name: "申请厂商入驻" })[0]).toHaveAttribute("href", "/join");
	  await waitFor(() => expect(mockedListVendors).toHaveBeenCalledWith(expect.objectContaining({ newlyJoined: true, sort: "latest", province: "山东" })));
	});
});
