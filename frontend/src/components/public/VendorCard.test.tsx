import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import { VendorCard } from "./VendorCard";

describe("VendorCard", () => {
  afterEach(() => cleanup());
  it("keeps one cover in the visual and places the logo beside the company name", () => {
    const { container } = render(<MemoryRouter><VendorCard vendor={{ id: 1, name: "测试厂商", coverImage: "/cover.jpg", logo: "/logo.png" }} /></MemoryRouter>);
    expect(container.querySelectorAll(".vendor-card-visual img")).toHaveLength(1);
    expect(screen.getByAltText("测试厂商 Logo").parentElement).toHaveClass("vendor-card-header");
    expect(screen.getByRole("heading", { name: "测试厂商" })).toHaveAttribute("title", "测试厂商");
  });

  it("replaces a failed cover with the same placeholder used for missing covers and retries a new URL", () => {
    const { container, rerender } = render(<MemoryRouter><VendorCard vendor={{ id: 1, name: "测试厂商", coverImage: "/bad.jpg" }} /></MemoryRouter>);
    fireEvent.error(screen.getByAltText("测试厂商 企业展示"));
    expect(container.querySelector(".vendor-card-visual")).toHaveTextContent("测试厂商 · 企业展厅");
    rerender(<MemoryRouter><VendorCard vendor={{ id: 1, name: "测试厂商" }} /></MemoryRouter>);
    expect(container.querySelector(".vendor-card-visual")).toHaveTextContent("测试厂商 · 企业展厅");
    expect(container.querySelector(".vendor-card-logo")).not.toBeInTheDocument();
    rerender(<MemoryRouter><VendorCard vendor={{ id: 1, name: "测试厂商", coverImage: "/new.jpg" }} /></MemoryRouter>);
    expect(screen.getByAltText("测试厂商 企业展示")).toHaveAttribute("src", "/new.jpg");
  });

  it("removes a broken logo without removing the company name or cover", () => {
    render(<MemoryRouter><VendorCard compact vendor={{ id: 1, name: "测试厂商", coverImage: "/cover.jpg", logo: "/bad.png" }} /></MemoryRouter>);
    fireEvent.error(screen.getByAltText("测试厂商 Logo"));
    expect(screen.queryByAltText("测试厂商 Logo")).not.toBeInTheDocument();
    expect(screen.getByAltText("测试厂商 企业展示")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "测试厂商" })).toBeInTheDocument();
  });

  it("renders an information-first card without an empty image area", () => {
    const { container } = render(<MemoryRouter><VendorCard vendor={{ id: 1, name: "山东测试农机配件有限公司", province: "山东", city: "潍坊", mainProducts: "变速箱、链条、齿轮", serviceAdvantages: "设备先进，品控严格", isVerified: true, websiteUrl: "https://example.com", tags: [{ id: 1, name: "源头厂商" }] }} /></MemoryRouter>);
    expect(container.querySelector(".vendor-cover")).not.toBeInTheDocument();
    expect(screen.getByText("山东测试农机配件有限公司")).toBeInTheDocument();
    expect(screen.queryByText("平台认证")).not.toBeInTheDocument();
    expect(screen.getByText(/山东 · 潍坊/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "进入展厅" })).toHaveAttribute("href", "/v/1");
    expect(screen.getByRole("link", { name: "查看产品" })).toHaveAttribute("href", "/v/1#showroom-products");
    expect(screen.getByRole("link", { name: /访问官网/ })).toHaveAttribute("href", "https://example.com");
  });

  it("shows a disabled website state when no official site is available", () => {
    render(<MemoryRouter><VendorCard directory vendor={{ id: 2, name: "河北测试农机配件有限公司" }} /></MemoryRouter>);
    expect(screen.queryByText("暂无官网")).not.toBeInTheDocument();
    expect(screen.getByRole("link", {name:"进入展厅"})).toBeInTheDocument();
  });
});
