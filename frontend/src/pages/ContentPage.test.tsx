import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getFriendLinks, getHome, getPage } from "../api/public";
import { ContentPage } from "./ContentPage";

vi.mock("../api/public", () => ({
  getFriendLinks: vi.fn(),
  getHome: vi.fn(),
  getPage: vi.fn(),
}));

describe("ContentPage", () => {
  beforeEach(() => {
    vi.mocked(getHome).mockResolvedValue({
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
    vi.mocked(getPage).mockResolvedValue({
      id: 1,
      slug: "links",
      title: "友情链接",
      summary: "合作伙伴和行业服务入口。",
      content: "友情链接由平台运营人员在后台维护。",
      seoKeywords: "农机行业友情链接",
      isEnabled: true,
      sortOrder: 1,
    });
    vi.mocked(getFriendLinks).mockResolvedValue([{ id: 1, name: "农机配件服务", url: "https://example.com", logo: "/uploads/logo.png", isEnabled: true, sortOrder: 1 }]);
  });

  it("renders CMS page content and friend links for the links page", async () => {
    render(
      <MemoryRouter initialEntries={["/links"]}>
        <ContentPage slug="links" />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "友情链接" })).toBeInTheDocument();
    expect(screen.getByText("合作伙伴和行业服务入口。")).toBeInTheDocument();
    expect(screen.getByText("友情链接由平台运营人员在后台维护。")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /农机配件服务/ })).toHaveAttribute("href", "https://example.com");
    expect(getPage).toHaveBeenCalledWith("links");
    expect(getFriendLinks).toHaveBeenCalled();
  });
});
