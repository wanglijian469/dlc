import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deleteProductSupplier, listProductSuppliers, listVendorOptions } from "../../api/admin";
import { AdminProductSuppliersEditor } from "./AdminProductSuppliersEditor";

vi.mock("../../api/admin", () => ({
  deleteProductSupplier: vi.fn(),
  listProductSuppliers: vi.fn(),
  listVendorOptions: vi.fn(),
  saveProductSupplier: vi.fn(),
}));

const mockedDelete = vi.mocked(deleteProductSupplier);
const mockedListSuppliers = vi.mocked(listProductSuppliers);
const mockedListVendorOptions = vi.mocked(listVendorOptions);

describe("AdminProductSuppliersEditor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal("confirm", vi.fn(() => true));
    mockedListSuppliers.mockResolvedValueOnce([{ id: 18, productId: 5, vendorId: 2, status: "approved", vendor: { id: 2, name: "测试供应商" } }] as never).mockResolvedValue([] as never);
    mockedListVendorOptions.mockResolvedValue({ items: [], page: 1, pageSize: 20, total: 0 });
    mockedDelete.mockResolvedValue({ deleted: true });
  });

  afterEach(() => vi.unstubAllGlobals());

  it("removes the supplier row after confirming deletion", async () => {
    render(<AdminProductSuppliersEditor productId={5} />);
    expect(await screen.findByText("测试供应商")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "删除供应关系" }));

    await waitFor(() => expect(mockedDelete).toHaveBeenCalledWith(5, 18));
    expect(await screen.findByText("暂无关联厂商。")).toBeInTheDocument();
  });
});
