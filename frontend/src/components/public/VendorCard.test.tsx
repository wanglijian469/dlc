import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import { VendorCard } from "./VendorCard";

describe("VendorCard", () => {
  afterEach(() => cleanup());

  it("renders an information-first card without an empty image area", () => {
    const { container } = render(<MemoryRouter><VendorCard vendor={{ id: 1, name: "山东测试农机配件有限公司", province: "山东", city: "潍坊", mainProducts: "变速箱、链条、齿轮", serviceAdvantages: "设备先进，品控严格", isVerified: true, websiteUrl: "https://example.com", tags: [{ id: 1, name: "源头厂商" }] }} /></MemoryRouter>);
    expect(container.querySelector(".vendor-cover")).not.toBeInTheDocument();
    expect(screen.getByText("山东测试农机配件有限公司")).toBeInTheDocument();
    expect(screen.getByText("平台认证")).toHaveClass("tag-blue");
    expect(screen.getByText(/山东 · 潍坊/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看详情" })).toHaveAttribute("href", "/vendors/1");
    expect(screen.getByRole("link", { name: "查看产品" })).toHaveAttribute("href", "/products?vendorId=1");
    expect(screen.getByRole("link", { name: /访问官网/ })).toHaveAttribute("href", "https://example.com");
  });

  it("shows a disabled website state when no official site is available", () => {
    render(<MemoryRouter><VendorCard directory vendor={{ id: 2, name: "河北测试农机配件有限公司" }} /></MemoryRouter>);
    expect(screen.getByText("暂无官网")).toHaveAttribute("aria-disabled", "true");
  });
});
