import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { PublicHeader } from "./PublicHeader";

const meta = { siteName: "大陆农机供应链", brandMark: "农", brandLogo: "/api/media/88", submitVendorText: "提交厂商", adminLoginText: "后台登录", mobileBrandName: "大陆农机供应链", mobileBrandMark: "农" };

describe("PublicHeader", () => {
  afterEach(cleanup);
  beforeEach(() => localStorage.clear());

  it("renders an uploaded brand logo before the site name", () => {
    const { container } = render(<MemoryRouter><PublicHeader menus={[]} siteMeta={meta} /></MemoryRouter>);
    expect(container.querySelector(".brand-logo")).toHaveAttribute("src", "/api/media/88");
    expect(screen.getByText("大陆农机供应链")).toBeInTheDocument();
  });

  it("restores all required public entries when the supplied menu is empty", () => {
    render(<MemoryRouter><PublicHeader menus={[]} siteMeta={meta} /></MemoryRouter>);
    const navigation = screen.getByRole("navigation");
    expect(Array.from(navigation.querySelectorAll("a")).map((link) => link.textContent)).toEqual([
      "首页", "厂商资源", "配件货源", "加工服务", "供求信息",
    ]);
    expect(screen.getByRole("link", { name: "供求信息" })).toHaveAttribute("href", "/purchase");
  });

  it("shows the signed-in vendor and supports logout", () => {
    localStorage.setItem("cms_authenticated", "true");
    localStorage.setItem("cms_role", "vendor");
    localStorage.setItem("cms_username", "vendor-a");
    render(<MemoryRouter><PublicHeader menus={[]} siteMeta={meta} /></MemoryRouter>);
    expect(screen.getByText("厂商工作台")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "退出" }));
    expect(localStorage.getItem("cms_authenticated")).toBeNull();
    expect(screen.getByText("厂商登录")).toBeInTheDocument();
  });
});
