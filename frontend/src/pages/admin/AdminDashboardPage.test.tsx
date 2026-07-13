import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { AdminDashboardPage } from "./AdminDashboardPage";

describe("AdminDashboardPage", () => {
  it("renders CMS quick entry cards for common admin tasks", () => {
    render(
      <MemoryRouter>
        <AdminDashboardPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "快捷入口" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "厂商信息" })).toHaveAttribute("href", "/admin/vendors");
    expect(screen.getByRole("link", { name: "资料审核" })).toHaveAttribute("href", "/admin/vendor-reviews");
    expect(screen.getByRole("link", { name: "产品审核" })).toHaveAttribute("href", "/admin/product-reviews");
    expect(screen.getByRole("link", { name: "配件产品" })).toHaveAttribute("href", "/admin/products");
  });
});
