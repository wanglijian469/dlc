import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { checkOwnProductDuplicate, createOwnProduct } from "../../api/admin";
import { VendorProductsEditor } from "./VendorProductsEditor";

vi.mock("../../api/admin", () => ({
  checkOwnProductDuplicate: vi.fn().mockResolvedValue({ exact: false, similar: [] }),
  createOwnProduct: vi.fn().mockResolvedValue({}),
  deleteOwnProduct: vi.fn(), listOwnProducts: vi.fn().mockResolvedValue([]), updateOwnProduct: vi.fn(), updateOwnProductSubmission: vi.fn(), uploadFile: vi.fn(), withdrawOwnProductSubmission: vi.fn(),
}));
vi.mock("../../api/public", () => ({ getFilterOptions: vi.fn().mockResolvedValue({ categories: [{ id: 1, name: "农机配件", parentId: 0, isEnabled: true }, { id: 2, name: "传动配件", parentId: 1, isEnabled: true }] }) }));

describe("VendorProductsEditor", () => {
  afterEach(() => { cleanup(); document.body.style.overflow = ""; vi.clearAllMocks(); });

  it("only exposes self-service product entry and submits vendor-owned fields", async () => {
    render(<VendorProductsEditor />);
    await screen.findByText("尚未录入本厂产品，可点击“添加本厂产品”开始填写。");
    expect(screen.queryByText(/平台已有产品|关联平台产品|产品候选/)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "添加本厂产品" }));
    expect(screen.getByRole("dialog", { name: "添加本厂产品" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "农机配件" })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "传动配件" })).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("本厂产品名称"), { target: { value: "液压翻转犁" } });
    fireEvent.change(screen.getByLabelText("本厂型号"), { target: { value: "1LF-260" } });
    fireEvent.change(screen.getByLabelText("产品大类"), { target: { value: "1" } });
    fireEvent.change(screen.getByLabelText("供货能力"), { target: { value: "月供 100 台" } });
    fireEvent.click(screen.getByRole("button", { name: "提交审核" }));
    await waitFor(() => expect(createOwnProduct).toHaveBeenCalledWith(expect.objectContaining({ vendorProductName: "液压翻转犁", vendorModel: "1LF-260", supplyAbility: "月供 100 台" })));
  });

  it("shows same-vendor similarity without exposing the platform catalog", async () => {
    vi.mocked(checkOwnProductDuplicate).mockResolvedValueOnce({ exact: false, similar: [{ id: 7, recordType: "supplier", name: "液压翻转犁", model: "1LF-360", status: "approved" }] });
    render(<VendorProductsEditor />);
    await screen.findByText(/尚未录入本厂产品/);
    fireEvent.click(screen.getByRole("button", { name: "添加本厂产品" }));
    fireEvent.change(screen.getByLabelText("本厂产品名称"), { target: { value: "液压翻转犁配件" } });
    await screen.findByText(/本厂已有相似产品/);
    expect(screen.queryByText(/平台已有产品|关联平台产品|产品候选/)).not.toBeInTheDocument();
  });
});
