import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getFilterOptions, getHome, listVendors } from "../api/public";
import { VendorsPage } from "./VendorsPage";

vi.mock("../api/public", () => ({
  getFilterOptions: vi.fn(),
  getHome: vi.fn(),
  listVendors: vi.fn(),
}));

const mockedGetFilterOptions = vi.mocked(getFilterOptions);
const mockedGetHome = vi.mocked(getHome);
const mockedListVendors = vi.mocked(listVendors);

describe("VendorsPage", () => {
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
      provinces: ["山东", "河北"],
      categories: [{ id: 1, name: "传动配件" }],
      serviceTags: [{ id: 1, name: "源头厂商" }],
    });
    mockedListVendors.mockResolvedValue({
      items: [{ id: 1, name: "山东沃得农机配件有限公司", province: "山东", mainProducts: "链条、齿轮", tags: [{ id: 1, name: "源头厂商" }] }],
      page: 1,
      pageSize: 12,
      total: 1,
    });
  });

  it("renders vendor directory results with filters", async () => {
    render(
      <MemoryRouter>
        <VendorsPage />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "厂商目录" })).toBeInTheDocument();
    expect(await screen.findByText("山东沃得农机配件有限公司")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "搜索" })).toBeInTheDocument();
    await waitFor(() => expect(mockedListVendors).toHaveBeenCalledWith(expect.objectContaining({ page: 1, pageSize: 12 })));
  });
});
