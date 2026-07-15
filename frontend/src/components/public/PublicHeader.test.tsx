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

  it("shows the signed-in member and supports logout", () => {
    localStorage.setItem("admin_token", "token");
    localStorage.setItem("cms_role", "user");
    localStorage.setItem("cms_username", "member-a");
    render(<MemoryRouter><PublicHeader menus={[]} siteMeta={meta} /></MemoryRouter>);
    expect(screen.getByText("member-a")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "退出" }));
    expect(localStorage.getItem("admin_token")).toBeNull();
    expect(screen.getByText("后台登录")).toBeInTheDocument();
  });
});
