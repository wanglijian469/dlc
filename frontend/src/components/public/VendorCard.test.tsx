import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { VendorCard } from "./VendorCard";

describe("VendorCard", () => {
  it("renders an information-first vendor card without cover templates", () => {
    const { container } = render(
      <MemoryRouter>
        <VendorCard
          vendor={{
            id: 1,
            name: "山东测试农机配件有限公司",
            logo: "https://img.example.com/logo.png",
            coverImage: "https://dummyimage.com/600x320/eaf3ff/1f2a3d&text=Factory+01",
            province: "山东",
            city: "潍坊",
            mainProducts: "变速箱、链条、齿轮、轴承、液压件",
            serviceAdvantages: "设备先进，品控严格，交期稳定",
            isVerified: true,
            tags: [{ id: 1, name: "源头厂商" }],
          }}
        />
      </MemoryRouter>,
    );

    expect(container.querySelector(".vendor-cover")).not.toBeInTheDocument();
    expect(container.querySelector(".vendor-cover-index")).not.toBeInTheDocument();
    expect(container.textContent).not.toContain("Factory");
    expect(container.textContent).not.toContain("01");

    expect(screen.getByRole("img", { name: "山东测试农机配件有限公司" })).toHaveClass("vendor-logo");
    expect(screen.getByText("山东测试农机配件有限公司")).toBeInTheDocument();
    expect(screen.getByText("平台认证")).toHaveClass("tag-blue");
    expect(screen.getByText(/地区：山东 · 潍坊/)).toBeInTheDocument();
    expect(screen.getByText(/主营：变速箱、链条、齿轮、轴承、液压件/)).toBeInTheDocument();
    expect(screen.getByText(/优势：设备先进，品控严格，交期稳定/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看厂商" })).toHaveAttribute("href", "/vendors/1");
    expect(screen.getByRole("link", { name: "查看产品" })).toHaveAttribute("href", "/products?vendorId=1");
  });

  it("shows the inquiry action for directory cards", () => {
    render(
      <MemoryRouter>
        <VendorCard directory vendor={{ id: 2, name: "河北测试农机配件有限公司" }} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "在线询价" })).toHaveAttribute("href", "/vendors/2");
  });
});
