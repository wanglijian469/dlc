import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getFilterOptions, listProducts } from "../api/public";
import { ProductsPage } from "./ProductsPage";

vi.mock("../api/public", () => ({
  getFilterOptions: vi.fn(),
  listProducts: vi.fn(),
}));

const mockedGetFilterOptions = vi.mocked(getFilterOptions);
const mockedListProducts = vi.mocked(listProducts);

describe("ProductsPage", () => {
  beforeEach(() => {
    mockedGetFilterOptions.mockResolvedValue({
      provinces: ["山东"],
      categories: [{ id: 2, name: "液压系统" }],
      serviceTags: [],
    });
    mockedListProducts.mockResolvedValue({
      items: [{ id: 9, name: "液压油泵总成", categoryId: 2, vendorId: 1, vendor: { id: 1, name: "河北金瑞农机制造有限公司" }, isHot: true }],
      page: 1,
      pageSize: 12,
      total: 1,
    });
  });

  it("renders product list and linked vendor", async () => {
    render(
      <MemoryRouter>
        <ProductsPage />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "配件产品" })).toBeInTheDocument();
    expect(await screen.findByText("液压油泵总成")).toBeInTheDocument();
    expect(screen.getByText("河北金瑞农机制造有限公司")).toBeInTheDocument();
    await waitFor(() => expect(mockedListProducts).toHaveBeenCalledWith(expect.objectContaining({ page: 1, pageSize: 12 })));
  });
});
