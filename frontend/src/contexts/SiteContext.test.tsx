import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getLayoutConfig, getVendorCategories } from "../api/public";
import { getStaticPageData } from "../utils/staticPageData";
import { SiteProvider, useSite } from "./SiteContext";

vi.mock("../api/public", () => ({ getLayoutConfig: vi.fn(), getVendorCategories: vi.fn() }));
vi.mock("../utils/staticPageData", () => ({ getStaticPageData: vi.fn() }));

const staticLayout = {
  siteMeta: { siteName: "静态站点", brandMark: "农", submitVendorText: "厂商入驻", adminLoginText: "厂商登录", mobileBrandName: "静态站点", mobileBrandMark: "农" },
  theme: { primaryColor: "#1559c7", accentColor: "#0d8b6f" },
  topMenus: [], sidebarMenus: [], auxiliaryMenus: [], mobileMenus: [], mobileBottomMenus: [], version: "static",
};

function Probe() {
  const site = useSite();
  return <div>{site.layout.siteMeta.siteName}|{site.layout.topMenus.map((menu) => menu.name).join(",")}|{site.vendorCategories.map((category) => category.name).join(",")}</div>;
}

describe("SiteProvider", () => {
  beforeEach(() => {
	vi.mocked(getStaticPageData).mockReturnValue({ kind: "vendor", slug: "static-vendor", layout: staticLayout, vendor: { id: 1, name: "静态厂商" } });
	vi.mocked(getLayoutConfig).mockResolvedValue(staticLayout);
	vi.mocked(getVendorCategories).mockResolvedValue([{ id: 9, name: "最新厂商分类", vendorCount: 1 }]);
  });

  it("refreshes vendor categories even when the page uses a static layout payload", async () => {
	render(<SiteProvider><Probe /></SiteProvider>);
	expect(screen.getByText(/静态站点/)).toBeInTheDocument();
	expect(screen.getByText(/首页,厂商资源,配件货源,加工服务,供求信息/)).toBeInTheDocument();
	expect(await screen.findByText(/最新厂商分类/)).toBeInTheDocument();
	expect(getVendorCategories).toHaveBeenCalledTimes(1);
	await waitFor(() => expect(getLayoutConfig).not.toHaveBeenCalled());
  });
});
