import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { listProductSubmissionMatches, listProductSubmissions, reviewProductSubmission } from "../../api/admin";
import { ProductReviewsPage } from "./ProductReviewsPage";

vi.mock("../../api/admin", () => ({ listProductSubmissions: vi.fn(), listProductSubmissionMatches: vi.fn(), reviewProductSubmission: vi.fn() }));
vi.mock("../../api/public", () => ({ getFilterOptions: vi.fn().mockResolvedValue({ categories: [{ id: 1, name: "农机配件", parentId: 0 }] }) }));
const mockedList = vi.mocked(listProductSubmissions);
const mockedMatches = vi.mocked(listProductSubmissionMatches);
const mockedReview = vi.mocked(reviewProductSubmission);

describe("ProductReviewsPage", () => {
  afterEach(() => { cleanup(); vi.clearAllMocks(); });

  it("requires an explicit human create-or-link resolution", async () => {
    mockedList.mockResolvedValue({ items: [{ id: 11, vendorId: 3, submissionType: "new_product", status: "pending", productDraft: { name: "液压翻转犁", categoryId: 1 }, supplierDraft: { vendorProductName: "冀丰翻转犁", vendorModel: "1LF-260" }, createdAt: "", updatedAt: "" }], page: 1, pageSize: 10, total: 1 });
    mockedMatches.mockResolvedValue([{ product: { id: 7, name: "液压翻转犁 1LF-260", category: { id: 1, name: "农机配件" } }, score: 85, reasons: ["型号匹配", "分类相符"], vendorAlreadyLinked: false }]);
    mockedReview.mockResolvedValue({} as never);
    render(<MemoryRouter><ProductReviewsPage /></MemoryRouter>);
    expect(await screen.findByDisplayValue("冀丰翻转犁")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /关联已有产品/ }));
    const candidate = await screen.findByRole("button", { name: /液压翻转犁 1LF-260/ });
    fireEvent.click(candidate);
    fireEvent.click(screen.getByRole("button", { name: "通过" }));
    await waitFor(() => expect(mockedReview).toHaveBeenCalledWith(11, expect.objectContaining({ status: "approved", resolution: "link_product", targetProductId: 7 })));
  });
});
