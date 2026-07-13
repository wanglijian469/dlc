import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { VendorProductsEditor } from "./VendorProductsEditor";

vi.mock("../../api/admin", () => ({
  createOwnProduct: vi.fn(),
  deleteOwnProduct: vi.fn(),
  linkOwnProduct: vi.fn(),
  listOwnProducts: vi.fn().mockResolvedValue([]),
  searchVendorProductCatalog: vi.fn().mockResolvedValue([]),
  updateOwnProduct: vi.fn(),
  uploadFile: vi.fn(),
}));

vi.mock("../../api/public", () => ({
  getFilterOptions: vi.fn().mockResolvedValue({ categories: [] }),
}));

describe("VendorProductsEditor", () => {
  afterEach(() => {
    document.body.style.overflow = "";
  });

  it("opens product creation in a modal drawer and restores page scrolling when closed", async () => {
    render(<VendorProductsEditor />);
    await waitFor(() => expect(screen.getByText("暂无产品供应信息，可先搜索平台产品并建立供应关联。")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "提交新产品" }));

    expect(screen.getByRole("dialog", { name: "提交新产品候选" })).toBeInTheDocument();
    expect(document.body.style.overflow).toBe("hidden");

    fireEvent.click(screen.getAllByRole("button", { name: "关闭产品编辑抽屉" })[1]);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(document.body.style.overflow).toBe("");
  });
});
