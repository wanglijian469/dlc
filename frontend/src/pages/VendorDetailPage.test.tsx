import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getHome, getVendor, getVendorContact, listProducts } from "../api/public";
import { VendorDetailPage } from "./VendorDetailPage";

vi.mock("../api/public", () => ({
  getHome: vi.fn(),
  getVendor: vi.fn(),
  getVendorContact: vi.fn(),
  listProducts: vi.fn(),
}));

const mockedGetHome = vi.mocked(getHome);
const mockedGetVendor = vi.mocked(getVendor);
const mockedGetVendorContact = vi.mocked(getVendorContact);
const mockedListProducts = vi.mocked(listProducts);

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
      serviceModels: "收割机、拖拉机、播种机",
      serviceAdvantages: "源头工厂、支持定制、交付稳定",
      description: "专注农机液压件生产与配套服务。",
      websiteUrl: "https://vendor.example.com",
      phone: "400-800-0008",
      contactName: "王经理",
      isVerified: true,
      establishedYear: "2012 年",
      factoryArea: "12000 平方米",
      employeeCount: "80 人",
      annualCapacity: "年产液压件 20 万套",
      equipment: "数控车床、自动焊接线、液压测试台",
      certifications: "ISO9001 质量管理体系",
      afterSalesService: "质保 12 个月，提供技术选型支持",
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
    expect(screen.getByText("质保 12 个月，提供技术选型支持")).toBeInTheDocument();
    expect(screen.getByText("加工服务能力")).toBeInTheDocument();
    expect(screen.getByText(/数控车削、焊接加工/)).toBeInTheDocument();
    expect(screen.getByText(/数控车床、焊接工位/)).toBeInTheDocument();
    expect(screen.getByText("液压油缸总成")).toBeInTheDocument();
    const breadcrumbs = screen.getByRole("navigation", { name: "面包屑" });
    expect(breadcrumbs.querySelector('a[href="/vendors"]')).toHaveTextContent("厂商目录");
    await waitFor(() => expect(mockedListProducts).toHaveBeenCalledWith({ vendorId: "8", pageSize: 6 }));
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

	it("loads the branded vendor site route", async () => {
		mockedGetVendor.mockResolvedValue({ id: 8, slug: "hanfeng-parts", name: "测试厂商" });
		renderDetail("/v/hanfeng-parts");
		await screen.findByRole("heading", { name: "测试厂商", level: 1 });
		expect(mockedGetVendor).toHaveBeenCalledWith("hanfeng-parts");
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
    renderDetail();
    const buttons = await screen.findAllByRole("button", { name: /登录查看完整?联系方式/ });
    fireEvent.click(buttons[0]);
    expect(await screen.findByRole("link", { name: "13812345678" })).toHaveAttribute("href", "tel:13812345678");
    expect(screen.getByText("hanfeng-parts")).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "测试厂商 微信二维码" })).toHaveAttribute("src", "/api/vendors/8/contact-qr");
    expect(screen.queryByText(/今日还可查看/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /复制电话|复制微信/ })).not.toBeInTheDocument();
    expect(mockedGetVendorContact).toHaveBeenCalledWith(8);
  });

  it("shows public phone, WeChat and QR code directly without copy actions", async () => {
    mockedGetVendor.mockResolvedValue({
      id: 10,
      name: "公开联系厂商",
      phone: "0319-5666294",
      phonePublic: true,
      wechat: "public-wechat",
      wechatQrCode: "/api/media/88",
      wechatPublic: true,
    });
    renderDetail("/v/public-vendor");
    expect(await screen.findByRole("link", { name: "0319-5666294" })).toHaveAttribute("href", "tel:0319-5666294");
    expect(screen.getByText("public-wechat")).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "公开联系厂商 微信二维码" })).toHaveAttribute("src", "/api/media/88");
    expect(screen.queryByRole("button", { name: /复制电话|复制微信|电话联系/ })).not.toBeInTheDocument();
  });

});
