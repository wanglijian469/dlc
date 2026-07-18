import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { VendorProductsEditor } from "./VendorProductsEditor";
import { productFieldGuidance } from "../../config/formGuidance";

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
    cleanup();
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

  it("shows product-writing guidance in the new product drawer", async () => {
    render(<VendorProductsEditor />);
    await waitFor(() => expect(screen.getByText("暂无产品供应信息，可先搜索平台产品并建立供应关联。")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "提交新产品" }));

    expect(screen.getByText("补充产品资料和标准配图，提交后由管理员审核。")).toBeInTheDocument();
    const description = screen.getByLabelText("发布产品说明");
    const detail = screen.getByLabelText("产品详细说明");
    expect(description).toHaveAttribute("placeholder", productFieldGuidance.description);
    expect(detail).toHaveAttribute("placeholder", productFieldGuidance.detailContent);
    expect(description).toHaveClass("writing-example");

    fireEvent.change(description, { target: { value: "适用于联合收割机传动系统。" } });
    expect(description).toHaveValue("适用于联合收割机传动系统。");
  });
});
