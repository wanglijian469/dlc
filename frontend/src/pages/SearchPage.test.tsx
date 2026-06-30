import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getHome, search } from "../api/public";
import { SearchPage } from "./SearchPage";

vi.mock("../api/public", () => ({
  getHome: vi.fn(),
  search: vi.fn(),
}));

const mockedGetHome = vi.mocked(getHome);
const mockedSearch = vi.mocked(search);

describe("SearchPage", () => {
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
    mockedSearch.mockResolvedValue({
      vendors: {
        items: [{ id: 1, name: "山东测试农机配件有限公司", province: "山东", city: "潍坊", mainProducts: "齿轮、链条", isVerified: true }],
        page: 1,
        pageSize: 10,
        total: 1,
      },
      products: {
        items: [{ id: 2, name: "齿轮箱总成", categoryId: 2, compatibleModels: "收割机", vendor: { id: 1, name: "山东测试农机配件有限公司" } }],
        page: 1,
        pageSize: 10,
        total: 1,
      },
      categories: {
        items: [{ id: 3, name: "传动配件" }],
        page: 1,
        pageSize: 10,
        total: 1,
      },
    });
  });

  it("renders search results as unified cards", async () => {
    const { container } = render(
      <MemoryRouter initialEntries={["/search?keyword=齿轮"]}>
        <SearchPage />
      </MemoryRouter>,
    );

    expect(await screen.findByText("山东测试农机配件有限公司")).toBeInTheDocument();
    expect(await screen.findByText("齿轮箱总成")).toBeInTheDocument();
    expect(container.querySelectorAll(".search-card").length).toBeGreaterThanOrEqual(3);
    expect(container.querySelector(".vendor-cover")).not.toBeInTheDocument();
    expect(container.textContent).not.toContain("Factory");
    expect(container.textContent).not.toContain("Parts");
  });
});
