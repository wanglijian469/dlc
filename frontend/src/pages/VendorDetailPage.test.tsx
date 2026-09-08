import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getHome, getVendor, getVendorContact, getVendorContactQRCode, getVendorPosts } from "../api/public";
import { VendorDetailPage } from "./VendorDetailPage";

import { getShowroomProducts } from "../api/workspace";
vi.mock("../api/workspace", () => ({ getShowroomProducts: vi.fn() }));
vi.mock("../analytics", () => ({ trackAnalytics: vi.fn() }));

vi.mock("../api/public", () => ({
  getHome: vi.fn(),
  getFilterOptions: vi.fn().mockResolvedValue({ categories: [] }),
  getVendor: vi.fn(),
  getVendorContact: vi.fn(),
  getVendorContactQRCode: vi.fn(),
  getVendorPosts: vi.fn(),
  listProducts: vi.fn(),
}));

const mockedGetHome = vi.mocked(getHome);
const mockedGetVendor = vi.mocked(getVendor);
const mockedGetVendorContact = vi.mocked(getVendorContact);
const mockedGetVendorContactQRCode = vi.mocked(getVendorContactQRCode);
const mockedGetVendorPosts = vi.mocked(getVendorPosts);
const mockedListProducts = vi.mocked(getShowroomProducts);

function renderDetail(path = "/vendors/8") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route element={<VendorDetailPage />} path="/vendors/:id" />
		<Route element={<VendorDetailPage />} path="/v/:id" />
      </Routes>
    </MemoryRouter>,
  );
}

describe("VendorDetailPage", () => {
  beforeEach(() => {
    mockedGetVendorPosts.mockResolvedValue([]);
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: vi.fn(() => "blob:vendor-qr") });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: vi.fn() });
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
    mockedListProducts.mockResolvedValue({
      items: [
        {
          id: 9,
          name: "液压油缸总成",
          compatibleModels: "联合收割机、拖拉机",
          vendor: { id: 8, name: "浙江汉丰农机有限公司" },
        },
      ],
      page: 1,
      pageSize: 8,
      total: 1,
    });
  });

  afterEach(() => cleanup());

  it("renders a rich B2B vendor profile with lightweight contact actions", async () => {
    mockedGetVendor.mockResolvedValue({
      id: 8,
      name: "浙江汉丰农机有限公司",
      shortName: "汉丰农机",
      logo: "https://img.example.com/logo.png",
      province: "浙江",
      city: "宁波",
      address: "浙江宁波农机产业园",
      mainProducts: "液压油缸、液压油泵、高压油管",
      serviceAdvantages: "源头工厂、支持定制、交付稳定",
      description: "专注农机液压件生产与配套服务。",
      websiteUrl: "https://vendor.example.com",
      phone: "400-800-0008", phonePublic: true,
      contactName: "王经理",
      isVerified: true,
      establishedYear: "2012 年",
      factoryArea: "12000 平方米",
      employeeCount: "80 人",
      annualCapacity: "年产液压件 20 万套",
      equipment: "数控车床、自动焊接线、液压测试台",
      certifications: "ISO9001 质量管理体系",
      providesProcessing: true,
      processingServices: "数控车削、焊接加工",
      processingMaterials: "钢件、轴套、齿轮坯",
      processingEquipment: "数控车床、焊接工位",
      processingCapacity: "支持小批量试制和批量代工",
      processingRegions: "全国发货",
      processingNotes: "来图来样均可",
      tags: [{ id: 1, name: "源头厂商" }],
    });

    renderDetail();

    expect(await screen.findByRole("heading", { name: "浙江汉丰农机有限公司", level: 1 })).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: "访问官网" })[0]).toHaveAttribute("href", "https://vendor.example.com");
    expect(screen.getAllByRole("link", { name: "访问官网" })[0]).toHaveAttribute("target", "_blank");
    expect(screen.getByText("2012 年")).toBeInTheDocument();
    expect(screen.getByText("12000 平方米")).toBeInTheDocument();
    expect(screen.getByText("年产液压件 20 万套")).toBeInTheDocument();
    expect(screen.getByText("数控车床、自动焊接线、液压测试台")).toBeInTheDocument();
    expect(screen.getByText("ISO9001 质量管理体系")).toBeInTheDocument();
    expect(screen.getByText("加工服务能力")).toBeInTheDocument();
    expect(screen.getByText(/数控车削、焊接加工/)).toBeInTheDocument();
    expect(screen.getByText(/数控车床、焊接工位/)).toBeInTheDocument();
    expect(await screen.findByText("液压油缸总成")).toBeInTheDocument();
    const breadcrumbs = screen.getByRole("navigation", { name: "面包屑" });
    expect(breadcrumbs.querySelector('a[href="/vendors"]')).toHaveTextContent("厂商资源");
    await waitFor(() => expect(mockedListProducts).toHaveBeenCalledWith(8, expect.objectContaining({ pageSize: 12 })));
  });

  it("does not invent a website or inquiry entry when contact details are missing", async () => {
    mockedGetVendor.mockResolvedValue({
      id: 9,
      name: "河北力捷机械有限公司",
      province: "河北",
      city: "邢台",
      mainProducts: "齿轮、轴承、传动轴",
    });

    renderDetail("/vendors/9");

    expect(await screen.findByRole("heading", { name: "河北力捷机械有限公司", level: 1 })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "访问官网" })).not.toBeInTheDocument();
    expect(screen.queryByText("在线询价")).not.toBeInTheDocument();
    expect(screen.queryByText("生产能力")).not.toBeInTheDocument();
  });

  it("renders enterprise gallery images in the complete-image grid", async () => {
    mockedGetVendor.mockResolvedValue({
      id: 11,
      name: "图集测试厂商",
      media: [{ id: 1, kind: "factory", caption: "竖版厂房", url: "/api/media/11" }, { id: 2, kind: "equipment", caption: "加工中心", url: "/api/media/12" }],
    });
    const { container } = renderDetail("/vendors/11");
    expect(await screen.findByRole("heading", { name: "企业图集" })).toBeInTheDocument();
    expect(container.querySelector(".vendor-media-grid img")?.getAttribute("src")).toMatch(/^\/api\/media\/11(?:\?|$)/);
    fireEvent.click(screen.getByRole("button", { name: "放大企业图片：竖版厂房" }));
    const dialog = screen.getByRole("dialog", { name: "图片预览" });
    expect(dialog).toHaveTextContent("厂房 · 竖版厂房");
    expect(within(dialog).getByRole("img", { name: "竖版厂房" }).getAttribute("src")).toMatch(/^\/api\/media\/11(?:\?|$)/);
    fireEvent.click(screen.getByRole("button", { name: "下一张图片" }));
    expect(within(dialog).getByRole("img", { name: "加工中心" })).toBeInTheDocument();
  });

	it("loads the branded vendor site route", async () => {
		mockedGetVendor.mockResolvedValue({ id: 8, slug: "hanfeng", name: "测试厂商" });
		renderDetail("/v/hanfeng");
		await screen.findByRole("heading", { name: "测试厂商", level: 1 });
		expect(mockedGetVendor).toHaveBeenCalledWith("hanfeng");
	});

  it("reveals protected contact only after an authenticated contact request", async () => {
    mockedGetVendor.mockResolvedValue({
      id: 8,
      name: "测试厂商",
      phone: "138******78",
      phonePublic: false,
      phoneAvailable: true,
    });
    mockedGetVendorContact.mockResolvedValue({
      vendorId: 8,
      phone: "13812345678",
      wechat: "hanfeng-parts",
      wechatQrCodeUrl: "/api/vendors/8/contact-qr",
    });
    mockedGetVendorContactQRCode.mockResolvedValue(new Blob(["qr"], { type: "image/png" }));
    renderDetail();
    const buttons = await screen.findAllByRole("button", { name: "查看联系电话" });
    fireEvent.click(buttons[0]);
    expect(await screen.findByRole("link", { name: /13812345678/ })).toHaveAttribute("href", "tel:13812345678");
    fireEvent.click(screen.getByRole("button", { name: "微信联系" }));
    expect(screen.getByText("hanfeng-parts")).toBeInTheDocument();
    expect(await screen.findByRole("img", { name: "测试厂商 微信二维码" })).toHaveAttribute("src", "blob:vendor-qr");
    expect(screen.queryByText(/今日还可查看/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /复制微信/ })).toBeInTheDocument();
    expect(mockedGetVendorContact).toHaveBeenCalledWith(8);
  });

  it("shows public phone and WeChat with explicit successful-copy action", async () => {
    mockedGetVendor.mockResolvedValue({
      id: 10,
      name: "公开联系厂商",
      phone: "0319-5666294",
      phonePublic: true,
      wechat: "public-wechat",
      wechatQrCode: "/api/media/88",
      wechatPublic: true,
    });
    renderDetail("/v/publicvendor");
    expect(await screen.findByRole("link", { name: /0319-5666294/ })).toHaveAttribute("href", "tel:0319-5666294");
    fireEvent.click(screen.getByRole("button", { name: "微信联系" }));
    expect(screen.getByText("public-wechat")).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "公开联系厂商 微信二维码" })).toHaveAttribute("src", "/api/media/88");
    expect(screen.getByRole("button", { name: /复制微信/ })).toBeInTheDocument();
  });

  it("shows vendor posts as a compact list, expands it and opens the full article", async () => {
    mockedGetVendor.mockResolvedValue({ id: 12, name: "动态测试厂商" });
    mockedGetVendorPosts.mockResolvedValue(Array.from({ length: 4 }, (_, index) => ({
      id: index + 1,
      vendorId: 12,
      postType: index === 1 ? "case" as const : "update" as const,
      title: `动态${index + 1}`,
      summary: `摘要${index + 1}`,
      content: `完整正文${index + 1}\n第二行`,
      status: "approved" as const,
      publishedAt: `2026-08-${13 - index}T08:00:00Z`,
      createdAt: "2026-08-01T08:00:00Z",
      updatedAt: "2026-08-01T08:00:00Z",
    })));

    const { container } = renderDetail("/vendors/12");

    expect(await screen.findByRole("heading", { name: "企业动态与案例" })).toBeInTheDocument();
    expect(container.querySelectorAll(".vendor-post-public-item")).toHaveLength(3);
    expect(container.querySelector(".vendor-post-public-cover.update svg")).toBeInTheDocument();
    expect(screen.queryByText("动态4")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "查看全部 4 条" }));
    expect(container.querySelectorAll(".vendor-post-public-item")).toHaveLength(4);
    fireEvent.click(screen.getByRole("button", { name: /动态1/ }));

    const dialog = screen.getByRole("dialog", { name: "动态1" });
    expect(dialog).toHaveTextContent("完整正文1");
    expect(document.body).toHaveStyle({ overflow: "hidden" });
    expect(screen.getByRole("button", { name: "关闭企业动态详情" })).toHaveFocus();
    fireEvent.keyDown(document, { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(document.body.style.overflow).toBe("");
  });

});
