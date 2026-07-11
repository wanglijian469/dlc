import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getHome, getProduct, listProducts } from "../api/public";
import { ProductDetailPage } from "./ProductDetailPage";

vi.mock("../api/public", () => ({
  getHome: vi.fn(),
  getProduct: vi.fn(),
  listProducts: vi.fn(),
}));

const mockedGetHome = vi.mocked(getHome);
const mockedGetProduct = vi.mocked(getProduct);
const mockedListProducts = vi.mocked(listProducts);

describe("ProductDetailPage", () => {
  beforeEach(() => {
    mockedGetHome.mockResolvedValue({
      topMenus: [],
      sidebarMenus: [],
      auxiliaryMenus: [],
      mobileMenus: [],
      mobileBottomMenus: [],
      banner: { title: "" },
      recommendedVendors: [],
      moreVendors: [],
      stats: [],
      safeguards: [],
      join: { text: "", buttonText: "", path: "/" },
    });
    mockedGetProduct.mockResolvedValue({
      id: 7,
      name: "液压油泵总成",
      categoryId: 5,
      vendorId: 3,
      compatibleModels: "农机液压系统",
      description: "压力稳定",
      detailContent: "适配多种农机液压系统，可按样品定制。",
      gallery: ["/uploads/pump-1.jpg"],
      specs: [{ name: "质保", value: "12个月" }],
      priceNote: "面议 / 批量报价",
      inquiryText: "联系供应商",
      inquiryPath: "/vendors/3",
      category: { id: 5, name: "液压系统配件" },
      vendor: { id: 3, name: "江苏东成农机配件有限公司", province: "江苏" },
    });
    mockedListProducts.mockResolvedValue({ items: [], page: 1, pageSize: 4, total: 0 });
  });

  it("renders CMS-managed product detail content", async () => {
    render(
      <MemoryRouter initialEntries={["/products/7"]}>
        <Routes><Route element={<ProductDetailPage />} path="/products/:id" /></Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "液压油泵总成", level: 1 })).toBeInTheDocument();
    expect(screen.getByText("适配多种农机液压系统，可按样品定制。")).toBeInTheDocument();
    expect(screen.getByText(/质保/)).toBeInTheDocument();
    expect(screen.getByText(/12个月/)).toBeInTheDocument();
    expect(screen.getByText("面议 / 批量报价")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "联系供应商" })).toHaveAttribute("href", "/vendors/3");
    expect(mockedGetProduct).toHaveBeenCalledWith("7");
  });
});
