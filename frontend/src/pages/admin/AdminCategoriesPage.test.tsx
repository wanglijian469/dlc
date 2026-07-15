import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createResource, deleteResource, listResource, updateResource } from "../../api/admin";
import { AdminCategoriesPage } from "./AdminCategoriesPage";

vi.mock("../../api/admin", () => ({
  createResource: vi.fn(),
  deleteResource: vi.fn(),
  listResource: vi.fn(),
  updateResource: vi.fn(),
}));

const categories = [
  { id: 10, name: "播种施肥配件", parentId: 0, sortOrder: 10, isEnabled: true },
  { id: 11, name: "旋耕机/耕作机械", parentId: 10, sortOrder: 5, isEnabled: true },
];

describe("AdminCategoriesPage", () => {
  afterEach(cleanup);
  beforeEach(() => {
    vi.mocked(listResource).mockResolvedValue(categories as never);
    vi.mocked(createResource).mockResolvedValue(categories[1] as never);
    vi.mocked(updateResource).mockResolvedValue(categories[1] as never);
    vi.mocked(deleteResource).mockResolvedValue({ deleted: true });
  });

  it("renders a two-level tree and creates a child from its parent row", async () => {
    render(<MemoryRouter><AdminCategoriesPage /></MemoryRouter>);

    expect(await screen.findByText("旋耕机/耕作机械")).toBeInTheDocument();
    expect(screen.getAllByText("二级分类").length).toBeGreaterThan(0);
    fireEvent.click(screen.getByRole("button", { name: "新增子分类" }));

    const dialog = screen.getByRole("dialog", { name: "新增配件分类" });
    expect(within(dialog).getByLabelText("分类级别")).toHaveValue("child");
    expect(within(dialog).getByLabelText("所属一级分类")).toHaveValue("10");
    fireEvent.change(within(dialog).getByLabelText("分类名称"), { target: { value: "播种盘" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "保存分类" }));

    await waitFor(() => expect(createResource).toHaveBeenCalledWith("categories", expect.objectContaining({ name: "播种盘", parentId: 10 })));
  });

  it("shows backend validation messages directly", async () => {
    vi.mocked(createResource).mockRejectedValue({ response: { data: { message: "同一父级下已存在同名分类" } } });
    render(<MemoryRouter><AdminCategoriesPage /></MemoryRouter>);
    expect((await screen.findAllByText("播种施肥配件")).length).toBeGreaterThan(0);
    fireEvent.click(screen.getByRole("button", { name: "新增一级分类" }));
    fireEvent.change(screen.getByLabelText("分类名称"), { target: { value: "播种施肥配件" } });
    fireEvent.click(screen.getByRole("button", { name: "保存分类" }));
    expect(await screen.findByText("同一父级下已存在同名分类")).toBeInTheDocument();
  });
});
