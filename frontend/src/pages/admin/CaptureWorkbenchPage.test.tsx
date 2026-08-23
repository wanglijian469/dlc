import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";
import { CaptureWorkbenchPage } from "./CaptureWorkbenchPage";
import { listCapturePackages } from "../../api/admin";

vi.mock("../../api/admin", () => ({
  listCapturePackages: vi.fn(), createCapturePackage: vi.fn(), getCapturePackage: vi.fn(), deleteCapturePackage: vi.fn(), uploadCaptureDocuments: vi.fn(),
  deleteCaptureDocument: vi.fn(), updateCaptureDocument: vi.fn(), recognizeCapturePackage: vi.fn(), updateCaptureDraft: vi.fn(), commitCapturePackage: vi.fn(), createVendorInvitation: vi.fn(),
}));
vi.mock("../../components/admin/AdminLayout", () => ({ AdminLayout: ({ children }: { children: ReactNode }) => <div>{children}</div> }));
vi.mock("../../components/admin/VendorOptionSearch", () => ({ VendorOptionSearch: () => <div>厂商搜索</div> }));

describe("CaptureWorkbenchPage", () => {
  beforeEach(() => { localStorage.setItem("cms_role", "admin"); vi.mocked(listCapturePackages).mockResolvedValue([]); });
  it("offers the one-vendor package and mobile capture workflow", async () => {
    render(<MemoryRouter><CaptureWorkbenchPage /></MemoryRouter>);
    expect(screen.getByText("新建厂商资料包")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("例如：2026 展会－某某机械")).toBeInTheDocument();
    await waitFor(() => expect(listCapturePackages).toHaveBeenCalled());
    expect(screen.getByText(/一家厂商一个资料包/)).toBeInTheDocument();
  });
});
