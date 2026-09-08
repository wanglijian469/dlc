import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { checkOwnProductDuplicate, listOwnProducts, updateOwnProductPrice } from "../../api/admin";
import { VendorProductsEditor } from "./VendorProductsEditor";

import { saveWorkDraft, commitWorkDraft } from "../../api/workspace";
vi.mock("../../api/workspace", () => ({ listWorkDrafts: vi.fn().mockResolvedValue([]), saveWorkDraft: vi.fn().mockImplementation(async (key, input) => ({ id: 1, clientKey: key, ...input, version: input.version + 1, updatedAt: new Date().toISOString() })), commitWorkDraft: vi.fn().mockResolvedValue({}), deleteWorkDraft: vi.fn(), updateShowroomOrder: vi.fn() }));

vi.mock("../../api/admin", () => ({
  checkOwnProductDuplicate: vi.fn().mockResolvedValue({ exact: false, similar: [] }),
  createOwnProduct: vi.fn().mockResolvedValue({}),
  deleteOwnProduct: vi.fn(), listOwnProducts: vi.fn().mockResolvedValue([]), updateOwnProduct: vi.fn().mockResolvedValue({}), updateOwnProductPrice: vi.fn().mockResolvedValue({}), updateOwnProductSubmission: vi.fn(), uploadFile: vi.fn(), withdrawOwnProductSubmission: vi.fn(),
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
    expect(screen.getByLabelText("计价单位")).toHaveValue("件");
    expect(screen.getByLabelText("起订量")).toHaveValue("1");
    expect(screen.getByLabelText("库存 / 可供应量")).toHaveValue("100");
    expect(screen.getByLabelText("交期")).toHaveValue("7天内");
    expect(screen.getByLabelText("运费说明")).toHaveValue("按实际运费结算");
    expect(screen.getByLabelText("供货能力")).toHaveValue("按订单生产");
    for (const label of ["计价单位", "起订量", "库存 / 可供应量", "交期", "运费说明", "供货能力"]) {
      expect(screen.getByLabelText(label).tagName).toBe("SELECT");
    }
    fireEvent.change(screen.getByLabelText("本厂产品名称"), { target: { value: "液压翻转犁" } });
    fireEvent.change(screen.getByLabelText("本厂型号"), { target: { value: "1LF-260" } });
    fireEvent.change(screen.getByLabelText("产品大类"), { target: { value: "1" } });
    fireEvent.change(screen.getByLabelText("供货能力"), { target: { value: "支持批量供货" } });
    fireEvent.click(screen.getByRole("button", { name: "下一步" }));
    fireEvent.click(screen.getByRole("button", { name: "下一步" }));
    fireEvent.click(screen.getByLabelText("我已确认产品资料及供货信息真实准确"));
    fireEvent.click(screen.getByRole("button", { name: "确认并提交审核" }));
    await waitFor(() => expect(saveWorkDraft).toHaveBeenCalledWith(expect.any(String), expect.objectContaining({ payload: expect.objectContaining({ vendorProductName: "液压翻转犁", vendorModel: "1LF-260", priceUnit: "件", minOrderQuantity: 1, availableQuantity: 100, leadTime: "7天内", freightNote: "按实际运费结算", supplyAbility: "支持批量供货" }) })));
    await waitFor(() => expect(commitWorkDraft).toHaveBeenCalledWith(1, expect.any(Number)));
  });

  it("preserves legacy values when editing without allowing new free text", async () => {
    vi.mocked(listOwnProducts).mockResolvedValueOnce([{
      id: 4, supplierId: 4, recordType: "supplier", vendorProductName: "液压油缸", categoryId: 1, status: "approved",
      priceUnit: "桶", minOrderQuantity: 3, availableQuantity: 0, leadTime: "10个工作日", freightNote: "专线到付", supplyAbility: "月供100台",
    } as never]);
    render(<VendorProductsEditor />);
    await screen.findByText("液压油缸");
    fireEvent.click(screen.getByRole("button", { name: "编辑本厂产品" }));
    expect(screen.getByLabelText("计价单位")).toHaveValue("桶");
    expect(screen.getByRole("option", { name: "当前值：桶", hidden: true })).toBeInTheDocument();
    expect(screen.getByLabelText("起订量")).toHaveValue("3");
    expect(screen.getByRole("option", { name: "当前值：3", hidden: true })).toBeInTheDocument();
    expect(screen.getByLabelText("库存 / 可供应量")).toHaveValue("0");
    expect(screen.getByLabelText("交期")).toHaveValue("10个工作日");
    expect(screen.getByLabelText("运费说明")).toHaveValue("专线到付");
    expect(screen.getByLabelText("供货能力")).toHaveValue("月供100台");
    fireEvent.click(screen.getByRole("button", { name: "下一步" }));
    fireEvent.click(screen.getByRole("button", { name: "下一步" }));
    fireEvent.click(screen.getByLabelText("我已确认产品资料及供货信息真实准确"));
    fireEvent.click(screen.getByRole("button", { name: "确认并提交审核" }));
    await waitFor(() => expect(saveWorkDraft).toHaveBeenCalledWith(expect.any(String), expect.objectContaining({ targetType: "supplier", targetId: 4, payload: expect.objectContaining({ priceUnit: "桶", minOrderQuantity: 3, availableQuantity: 0, supplyAbility: "月供100台" }) })));
  });

  it("uses the same presets for realtime updates and submits supply ability", async () => {
    vi.mocked(listOwnProducts).mockResolvedValueOnce([{
      id: 8, supplierId: 8, recordType: "supplier", vendorProductName: "传动轴", categoryId: 1, status: "approved", unitPriceCents: 10000,
      priceUnit: "件", minOrderQuantity: 1, availableQuantity: 0, leadTime: "现货", freightNote: "物流到付", supplyAbility: "现货供应", priceVersion: 2,
    } as never]);
    render(<VendorProductsEditor />);
    await screen.findByText("传动轴");
    fireEvent.click(screen.getByRole("button", { name: "即时更新价格库存" }));
    for (const label of ["计价单位", "起订量", "库存 / 可供应量", "交期", "运费说明", "供货能力"]) {
      expect(screen.getByLabelText(label).tagName).toBe("SELECT");
    }
    expect(screen.getByLabelText("库存 / 可供应量")).toHaveValue("0");
    fireEvent.change(screen.getByLabelText("供货能力"), { target: { value: "支持小批量供货" } });
    fireEvent.click(screen.getByRole("button", { name: "立即更新" }));
    await waitFor(() => expect(updateOwnProductPrice).toHaveBeenCalledWith(8, expect.objectContaining({ availableQuantity: 0, supplyAbility: "支持小批量供货", expectedVersion: 2 })));
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
