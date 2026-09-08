import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getVendorProfile } from "../../api/admin";
import { processingToggleDescription, vendorFieldGuidance } from "../../config/formGuidance";
import { VendorProfilePage } from "./VendorProfilePage";

import { saveWorkDraft, commitWorkDraft } from "../../api/workspace";
vi.mock("../../api/workspace", () => ({ listWorkDrafts: vi.fn().mockResolvedValue([]), saveWorkDraft: vi.fn().mockImplementation(async (key, input) => ({ id: 2, clientKey: key, ...input, version: 1 })), commitWorkDraft: vi.fn().mockResolvedValue({}) }));

vi.mock("../../api/admin", () => ({
  getVendorProfile: vi.fn(),
  logoutSession: vi.fn(),
  submitVendorProfile: vi.fn(),
  uploadFile: vi.fn(),
}));

vi.mock("../../components/admin/StructuredEditors", () => ({
  VendorMediaEditor: () => <div>企业图集编辑器</div>,
}));

vi.mock("../../components/admin/VendorProductsEditor", () => ({
  VendorProductsEditor: () => <div>产品编辑器</div>,
}));

const mockedGetVendorProfile = vi.mocked(getVendorProfile);


const guidedFields = [
  ["厂商简称", vendorFieldGuidance.shortName],
  ["主营产品", vendorFieldGuidance.mainProducts],
  ["公司简介", vendorFieldGuidance.description],
  ["服务优势", vendorFieldGuidance.serviceAdvantages],
  ["年产能", vendorFieldGuidance.annualCapacity],
  ["主要设备", vendorFieldGuidance.equipment],
  ["认证资质", vendorFieldGuidance.certifications],
  ["加工服务", vendorFieldGuidance.processingServices],
  ["加工材料 / 配件类型", vendorFieldGuidance.processingMaterials],
  ["加工设备", vendorFieldGuidance.processingEquipment],
  ["加工产能 / 交期", vendorFieldGuidance.processingCapacity],
  ["加工服务区域", vendorFieldGuidance.processingRegions],
  ["加工接单说明", vendorFieldGuidance.processingNotes],
] as const;

describe("VendorProfilePage writing guidance", () => {
  beforeEach(() => {
    localStorage.setItem("cms_role", "vendor");
    mockedGetVendorProfile.mockResolvedValue({
      vendor: { id: 7, slug: "testvendor", name: "测试厂商", publicationStatus: "draft" },
      draft: { id: 7, slug: "testvendor", name: "测试厂商" },
    } as never);
  });

  afterEach(() => {
    cleanup();
    localStorage.clear();
    vi.clearAllMocks();
  });

  it("renders all field examples and submits the processing toggle", async () => {
    render(<MemoryRouter initialEntries={["/admin/vendor-profile"]}><VendorProfilePage /></MemoryRouter>);

    await screen.findByDisplayValue("测试厂商");
		expect(screen.getByText("我的厂商网站")).toBeInTheDocument();
		expect(screen.getByText(`${window.location.origin}/v/testvendor`)).toBeInTheDocument();
		expect(screen.getByLabelText(/厂商网站地址标识/)).toHaveValue("testvendor");
    expect(screen.queryByText("产品编辑器")).not.toBeInTheDocument();
    for (const [, placeholder] of guidedFields) {
      const field = screen.getByPlaceholderText(placeholder);
      expect(field).toHaveAttribute("placeholder", placeholder);
      expect(field).toHaveClass("writing-example");
    }

    const processingToggle = screen.getByRole("checkbox", { name: /提供来图来样加工/ });
    expect(processingToggle.closest("label")).toHaveClass("admin-toggle-field", "wide-field");
    expect(screen.getByText(processingToggleDescription)).toBeInTheDocument();

    fireEvent.click(processingToggle);
    fireEvent.click(screen.getByRole("button", { name: "提交管理员审核" }));
    await waitFor(() => expect(saveWorkDraft).toHaveBeenCalledWith(expect.any(String), expect.objectContaining({ payload: expect.objectContaining({ providesProcessing: true }) })));
    await waitFor(() => expect(commitWorkDraft).toHaveBeenCalledWith(2, 1));
  });
});
