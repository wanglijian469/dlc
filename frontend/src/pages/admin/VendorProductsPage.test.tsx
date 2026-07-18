import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { VendorProductsPage } from "./VendorProductsPage";

vi.mock("../../components/admin/VendorProductsEditor", () => ({ VendorProductsEditor: () => <div>独立产品资料编辑器</div> }));

describe("VendorProductsPage", () => {
  beforeEach(() => localStorage.setItem("cms_role", "vendor"));
  afterEach(() => { cleanup(); localStorage.clear(); });

  it("hosts the existing product editor on its own workspace page", () => {
    render(<MemoryRouter initialEntries={["/admin/vendor-products"]}><VendorProductsPage /></MemoryRouter>);
    expect(screen.getByRole("heading", { name: "我的产品资料" })).toBeInTheDocument();
    expect(screen.getByText("独立产品资料编辑器")).toBeInTheDocument();
  });
});
