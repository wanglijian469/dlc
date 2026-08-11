import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { listMarketPosts } from "../api/market";
import { MarketPostsPage } from "./MarketPostsPage";

vi.mock("../api/market", () => ({ listMarketPosts: vi.fn() }));
vi.mock("../components/public/PageFrame", () => ({ PageFrame: ({ children }: { children: React.ReactNode }) => <main>{children}</main> }));

describe("MarketPostsPage", () => {
  beforeEach(() => vi.mocked(listMarketPosts).mockResolvedValue({ items: [{ id: 1, type: "demand", title: "求购收割机链条", description: "长期采购收割机链条和相关配件", publisherName: "采购商", status: "published", images: [], expiresAt: "2026-09-01T00:00:00+08:00", createdAt: "2026-08-01T00:00:00+08:00", updatedAt: "2026-08-01T00:00:00+08:00" }], page: 1, pageSize: 18, total: 1 }));
  it("renders public market data without contact fields", async () => {
    render(<MemoryRouter><MarketPostsPage /></MemoryRouter>);
    expect(await screen.findByText("求购收割机链条")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /发布信息/ })).toHaveAttribute("href", "/publish");
    expect(screen.queryByText(/138/)).not.toBeInTheDocument();
  });
});
