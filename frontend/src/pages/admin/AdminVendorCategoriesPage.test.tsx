import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createResource, deleteResource, listResource, updateResource } from "../../api/admin";
import { AdminVendorCategoriesPage } from "./AdminVendorCategoriesPage";

vi.mock("../../api/admin", () => ({ createResource: vi.fn(), deleteResource: vi.fn(), listResource: vi.fn(), updateResource: vi.fn() }));

describe("AdminVendorCategoriesPage", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(listResource).mockResolvedValue([{ id: 1, name: "传动配件", isEnabled: true }, { id: 2, name: "变速箱齿轮", parentId: 1, isEnabled: true }] as never);
		vi.mocked(createResource).mockResolvedValue({ id: 3, name: "液压件" } as never);
		vi.mocked(updateResource).mockResolvedValue({ id: 1, name: "传动配件" } as never);
		vi.mocked(deleteResource).mockResolvedValue({ deleted: true });
	});

	it("renders the two-level tree and creates a child category", async () => {
		render(<MemoryRouter><AdminVendorCategoriesPage /></MemoryRouter>);
		expect(await screen.findByText("变速箱齿轮")).toBeInTheDocument();
		fireEvent.click(screen.getByRole("button", { name: "新增子分类" }));
		fireEvent.change(screen.getByLabelText("厂商分类名称"), { target: { value: "链条链轮" } });
		fireEvent.click(screen.getByRole("button", { name: "保存分类" }));
		await waitFor(() => expect(createResource).toHaveBeenCalledWith("vendor-categories", expect.objectContaining({ name: "链条链轮", parentId: 1 })));
	});
});
