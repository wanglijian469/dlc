import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getFilterOptions, getHome, listProducts } from "../api/public";
import { ProductsPage } from "./ProductsPage";

vi.mock("../api/public", () => ({
  getFilterOptions: vi.fn(),
  getHome: vi.fn(),
  listProducts: vi.fn(),
}));

const mockedGetFilterOptions = vi.mocked(getFilterOptions);
const mockedGetHome = vi.mocked(getHome);
const mockedListProducts = vi.mocked(listProducts);

describe("ProductsPage", () => {
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
    mockedGetFilterOptions.mockResolvedValue({
      provinces: ["山东"],
      categories: [{ id: 2, name: "传动配件" }],
      serviceTags: [],
    });
    mockedListProducts.mockResolvedValue({
      items: [
        {
          id: 9,
          name: "变速箱齿轮总成",
          categoryId: 2,
          vendorId: 1,
          vendor: { id: 1, name: "河北金瑞农机制造有限公司" },
          isHot: true,
        },
      ],
      page: 1,
      pageSize: 12,
      total: 1,
    });
  });

  it("renders product list and uses industry cover fallback instead of Parts placeholder", async () => {
    const { container } = render(
      <MemoryRouter>
        <ProductsPage />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "配件产品" })).toBeInTheDocument();
    expect(await screen.findByText("变速箱齿轮总成")).toBeInTheDocument();
    expect(screen.getByText("河北金瑞农机制造有限公司")).toBeInTheDocument();

    const cover = container.querySelector(".product-image") as HTMLElement;
    expect(cover.style.backgroundImage).not.toContain("Parts");
    expect(cover).toHaveClass("industry-cover-transmission");
    await waitFor(() => expect(mockedListProducts).toHaveBeenCalledWith(expect.objectContaining({ page: 1, pageSize: 12 })));
  });
});
