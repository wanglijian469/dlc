import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getHome, getVendor, listProducts } from "../api/public";
import { VendorDetailPage } from "./VendorDetailPage";

vi.mock("../api/public", () => ({
  getHome: vi.fn(),
  getVendor: vi.fn(),
  listProducts: vi.fn(),
}));

const mockedGetHome = vi.mocked(getHome);
const mockedGetVendor = vi.mocked(getVendor);
const mockedListProducts = vi.mocked(listProducts);

function renderDetail(path = "/vendors/8") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route element={<VendorDetailPage />} path="/vendors/:id" />
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

  it("renders a rich B2B vendor profile with a prominent website entry", async () => {
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
      qualityControl: "来料检验、压力测试、出厂抽检",
      supplyRegions: "华东、华北、东北农机维修市场",
      cooperationTerms: "支持来图定制，常规件 7 天交付",
      afterSalesService: "质保 12 个月，提供技术选型支持",
      tags: [{ id: 1, name: "源头厂商" }],
    });

    renderDetail();

    expect(await screen.findByText("浙江汉丰农机有限公司")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "进入厂商官网" })).toHaveAttribute("href", "https://vendor.example.com");
    expect(screen.getByRole("link", { name: "进入厂商官网" })).toHaveAttribute("target", "_blank");
    expect(screen.getByText("平台可为源头厂商搭建独立展示网站，提升询盘转化")).toBeInTheDocument();
    expect(screen.getByText(/成立年份：2012 年/)).toBeInTheDocument();
    expect(screen.getByText(/厂房面积：12000 平方米/)).toBeInTheDocument();
    expect(screen.getByText(/年产能：年产液压件 20 万套/)).toBeInTheDocument();
    expect(screen.getByText(/主要设备：数控车床、自动焊接线、液压测试台/)).toBeInTheDocument();
    expect(screen.getByText(/认证资质：ISO9001 质量管理体系/)).toBeInTheDocument();
    expect(screen.getByText(/售后服务：质保 12 个月，提供技术选型支持/)).toBeInTheDocument();
    expect(screen.getByText("液压油缸总成")).toBeInTheDocument();
    await waitFor(() => expect(mockedListProducts).toHaveBeenCalledWith({ vendorId: "8", pageSize: 8 }));
  });

  it("shows an independent website application entry when the vendor has no website", async () => {
    mockedGetVendor.mockResolvedValue({
      id: 9,
      name: "河北力捷机械有限公司",
      province: "河北",
      city: "邢台",
      mainProducts: "齿轮、轴承、传动轴",
    });

    renderDetail("/vendors/9");

    expect(await screen.findByRole("heading", { name: "河北力捷机械有限公司" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "申请开通独立官网" })).toHaveAttribute("href", "/join");
    expect(screen.queryByText("生产能力")).not.toBeInTheDocument();
  });
});
