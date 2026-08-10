import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getHome, getProduct, getProductSuppliers, listProducts } from "../api/public";
import { ProductDetailPage } from "./ProductDetailPage";

vi.mock("../api/public", () => ({
  getHome: vi.fn(),
  getProduct: vi.fn(),
  getProductSuppliers: vi.fn(),
  listProducts: vi.fn(),
}));

const mockedGetHome = vi.mocked(getHome);
const mockedGetProduct = vi.mocked(getProduct);
const mockedGetProductSuppliers = vi.mocked(getProductSuppliers);
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
      supplierCount: 1,
      supplierRegions: ["江苏"],
      compatibleModels: "农机液压系统",
      description: "压力稳定",
      detailContent: "适配多种农机液压系统，可按样品定制。",
      gallery: ["/uploads/pump-1.jpg"],
      specs: [{ name: "质保", value: "12个月" }],
      category: { id: 5, name: "液压系统配件" },
    });
    mockedGetProductSuppliers.mockResolvedValue([{ id: 9, productId: 7, vendorId: 3, status: "approved", priceNote: "面议 / 批量报价", inquiryText: "联系该厂商", vendor: { id: 3, name: "江苏东成农机配件有限公司", province: "江苏" } }]);
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
    expect(screen.getByRole("link", { name: "联系该厂商" })).toHaveAttribute("href", "/v/3");
    expect(screen.getByRole("link", { name: "查看供应商（1）" })).toHaveAttribute("href", "#suppliers");
    const mainImage = await screen.findByRole("img", { name: "液压油泵总成" });
    expect(mainImage).toHaveAttribute("src", "/uploads/pump-1.jpg");
    expect(mainImage.closest(".product-main-media")?.querySelector(".industry-cover-main-icon")).toBeInTheDocument();
    mainImage.dispatchEvent(new Event("error", { bubbles: true }));
    expect(mainImage).toHaveStyle({ display: "none" });
    const breadcrumbs = screen.getByRole("navigation", { name: "面包屑" });
    expect(breadcrumbs.querySelector('a[href="/products"]')).toHaveTextContent("配件货源");
    expect(breadcrumbs.querySelector('a[href="/products?categoryId=5"]')).toHaveTextContent("液压系统配件");
    expect(mockedGetProduct).toHaveBeenCalledWith("7");
  });

  it("uses static-page preload data without requesting detail APIs", async () => {
    vi.clearAllMocks();
    const node = document.createElement("script");
    node.id = "static-page-data";
    node.type = "application/json";
    node.textContent = JSON.stringify({
      kind: "product",
      slug: "static-pump",
      layout: {},
      product: {
        id: 12,
        slug: "static-pump",
        name: "静态液压泵",
        categoryId: 5,
        specs: [{ name: "压力", value: "20MPa" }],
      },
      suppliers: [],
      related: [],
    });
    document.body.appendChild(node);

    render(
      <MemoryRouter initialEntries={["/products/static-pump"]}>
        <Routes><Route element={<ProductDetailPage />} path="/products/:id" /></Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "静态液压泵", level: 1 })).toBeInTheDocument();
    expect(mockedGetProduct).not.toHaveBeenCalled();
    expect(mockedGetProductSuppliers).not.toHaveBeenCalled();
    expect(mockedListProducts).not.toHaveBeenCalled();
    node.remove();
  });
});
