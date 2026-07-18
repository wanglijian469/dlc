import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getVendorProfile, submitVendorProfile } from "../../api/admin";
import { processingToggleDescription, vendorFieldGuidance } from "../../config/formGuidance";
import { VendorProfilePage } from "./VendorProfilePage";

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
const mockedSubmitVendorProfile = vi.mocked(submitVendorProfile);

const guidedFields = [
  ["厂商简称", vendorFieldGuidance.shortName],
  ["主营产品", vendorFieldGuidance.mainProducts],
  ["适配机型", vendorFieldGuidance.serviceModels],
  ["公司简介", vendorFieldGuidance.description],
  ["服务优势", vendorFieldGuidance.serviceAdvantages],
  ["年产能", vendorFieldGuidance.annualCapacity],
  ["主要设备", vendorFieldGuidance.equipment],
  ["认证资质", vendorFieldGuidance.certifications],
  ["售后服务", vendorFieldGuidance.afterSalesService],
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
      vendor: { id: 7, name: "测试厂商", publicationStatus: "draft" },
      draft: { id: 7, name: "测试厂商" },
    } as never);
    mockedSubmitVendorProfile.mockResolvedValue({} as never);
  });

  afterEach(() => {
    cleanup();
    localStorage.clear();
    vi.clearAllMocks();
  });

  it("renders all field examples and submits the processing toggle", async () => {
    render(<MemoryRouter initialEntries={["/admin/vendor-profile"]}><VendorProfilePage /></MemoryRouter>);

    await screen.findByDisplayValue("测试厂商");
    expect(screen.queryByText("产品编辑器")).not.toBeInTheDocument();
    for (const [label, placeholder] of guidedFields) {
      const field = screen.getByLabelText(label);
      expect(field).toHaveAttribute("placeholder", placeholder);
      expect(field).toHaveClass("writing-example");
    }

    const processingToggle = screen.getByRole("checkbox", { name: /提供来图来样加工/ });
    expect(processingToggle.closest("label")).toHaveClass("admin-toggle-field", "wide-field");
    expect(screen.getByText(processingToggleDescription)).toBeInTheDocument();

    fireEvent.click(processingToggle);
    fireEvent.click(screen.getByRole("button", { name: "提交管理员审核" }));
    await waitFor(() => expect(mockedSubmitVendorProfile).toHaveBeenCalledWith(expect.objectContaining({ providesProcessing: true })));
  });
});
