import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getVendorAnalytics } from "../../api/admin";
import { VendorAnalyticsPage } from "./VendorAnalyticsPage";

vi.mock("../../api/admin", () => ({ getVendorAnalytics: vi.fn(), logoutSession: vi.fn() }));
const mockedGetVendorAnalytics = vi.mocked(getVendorAnalytics);

const summary = {
  days: 30,
  vendor: { id: 7, name: "汉丰农机", pv: 21, uv: 9, contacts: 4, contactEvents: [{ label: "contact_phone_click", count: 4 }], trend: [{ date: "2026-07-18", pv: 7, uv: 3 }] },
  products: { pv: 16, uv: 8, trend: [{ date: "2026-07-18", pv: 5, uv: 2 }], items: [{ productId: 3, name: "收割机链条", path: "/products/chain", pv: 16, uv: 8 }] },
} as const;

describe("VendorAnalyticsPage", () => {
  beforeEach(() => { localStorage.setItem("cms_role", "vendor"); mockedGetVendorAnalytics.mockResolvedValue(summary as never); });
  afterEach(() => { cleanup(); localStorage.clear(); vi.clearAllMocks(); });

  it("shows scoped metrics, ranking and shared-product notice", async () => {
    render(<MemoryRouter><VendorAnalyticsPage /></MemoryRouter>);
    expect(await screen.findByText("汉丰农机访问概览")).toBeInTheDocument();
    expect(screen.getByText("收割机链条")).toBeInTheDocument();
    expect(screen.getByText(/不等同于本厂独立曝光/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "近 7 天" }));
    await waitFor(() => expect(mockedGetVendorAnalytics).toHaveBeenLastCalledWith(7));
  });

  it("renders an actionable error state", async () => {
    mockedGetVendorAnalytics.mockRejectedValueOnce(new Error("failed"));
    render(<MemoryRouter><VendorAnalyticsPage /></MemoryRouter>);
    expect(await screen.findByRole("alert")).toHaveTextContent("访问数据加载失败");
    fireEvent.click(screen.getByRole("button", { name: "重新加载" }));
    await waitFor(() => expect(mockedGetVendorAnalytics).toHaveBeenCalledTimes(2));
  });
});
