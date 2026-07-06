import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getHome, getProcessingFilterOptions, listProcessingVendors } from "../api/public";
import { ProcessingServicesPage } from "./ProcessingServicesPage";

vi.mock("../api/public", () => ({
  getHome: vi.fn(),
  getProcessingFilterOptions: vi.fn(),
  listProcessingVendors: vi.fn(),
}));

const mockedGetHome = vi.mocked(getHome);
const mockedGetProcessingFilterOptions = vi.mocked(getProcessingFilterOptions);
const mockedListProcessingVendors = vi.mocked(listProcessingVendors);

describe("ProcessingServicesPage", () => {
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
    mockedGetProcessingFilterOptions.mockResolvedValue({
      provinces: ["山东"],
      categories: [],
      serviceTags: [{ id: 21, name: "数控车削", tagType: "processing" }],
    });
    mockedListProcessingVendors.mockResolvedValue({
      items: [
        {
          id: 3,
          name: "山东精工加工有限公司",
          province: "山东",
          providesProcessing: true,
          processingServices: "数控车削、来图来样加工",
          processingEquipment: "数控车床、加工中心",
          tags: [{ id: 21, name: "数控车削", tagType: "processing" }],
        },
      ],
      page: 1,
      pageSize: 12,
      total: 1,
    });
  });

  it("renders processing vendors with processing filters", async () => {
    render(
      <MemoryRouter>
        <ProcessingServicesPage />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "加工服务" })).toBeInTheDocument();
    expect(screen.getByText("山东精工加工有限公司")).toBeInTheDocument();
    expect(screen.getByText(/数控车削、来图来样加工/)).toBeInTheDocument();
    expect(screen.getByText(/数控车床、加工中心/)).toBeInTheDocument();
    await waitFor(() => expect(mockedListProcessingVendors).toHaveBeenCalledWith(expect.objectContaining({ page: 1, pageSize: 12 })));
  });
});
