import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createResource, deleteResource, listResource, listResourcePage, updateResource } from "../../api/admin";
import { AdminResourcePage } from "./AdminResourcePage";

vi.mock("../../api/admin", () => ({
  createResource: vi.fn(),
  deleteResource: vi.fn(),
  listConfigs: vi.fn(),
  listResource: vi.fn(),
  listResourcePage: vi.fn(),
  updateConfig: vi.fn(),
  updateResource: vi.fn(),
  uploadFile: vi.fn(),
}));

const mockedCreateResource = vi.mocked(createResource);
const mockedListResource = vi.mocked(listResource);
const mockedListResourcePage = vi.mocked(listResourcePage);
const mockedUpdateResource = vi.mocked(updateResource);

function renderAdmin(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route element={<AdminResourcePage />} path="/admin/:resource" />
      </Routes>
    </MemoryRouter>,
  );
}

async function openCreateEditor(name: string) {
  fireEvent.click(await screen.findByRole("button", { name }));
}

describe("AdminResourcePage CMS forms", () => {
  afterEach(() => cleanup());

  beforeEach(() => {
    mockedListResource.mockImplementation((resource) => {
      if (resource === "tags") return Promise.resolve([{ id: 1, name: "源头厂商", tagType: "vendor" }, { id: 2, name: "数控车削", tagType: "processing" }] as never);
      if (resource === "categories") return Promise.resolve([{ id: 5, name: "液压系统配件" }] as never);
      if (resource === "vendors") {
        return Promise.resolve([
          { id: 3, name: "江苏东成农机配件有限公司", tagIds: [1], tags: [{ id: 1, name: "源头厂商" }] },
        ] as never);
      }
      return Promise.resolve([] as never);
    });
    mockedListResourcePage.mockImplementation(async (resource) => {
      const items = await mockedListResource(resource) as never[];
      return { items, page: 1, pageSize: 20, total: items.length };
    });
    mockedCreateResource.mockResolvedValue({ id: 1, name: "测试记录" });
    mockedUpdateResource.mockResolvedValue({ id: 1, name: "测试记录" });
    vi.mocked(deleteResource).mockResolvedValue({ deleted: true });
  });

  it("does not submit preloaded tag objects when editing vendors", async () => {
    renderAdmin("/admin/vendors");

    fireEvent.click(await screen.findByRole("button", { name: "编辑" }));
    const vendorForm = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(vendorForm).getByRole("button", { name: "保存修改" }));

    await waitFor(() => expect(mockedUpdateResource).toHaveBeenCalled());
    const payload = mockedUpdateResource.mock.calls[0][2] as Record<string, unknown>;
    expect(payload.tagIds).toEqual([1]);
    expect(payload).not.toHaveProperty("tags");
  });

  it("submits rich vendor fields with selected tag ids", async () => {
    renderAdmin("/admin/vendors");
    await openCreateEditor("新增厂商信息");

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "浙江汉丰农机有限公司" } });
    fireEvent.change(screen.getByLabelText("厂商官网 URL"), { target: { value: "https://vendor.example.com" } });
    const tagSelect = screen.getByLabelText("厂商标签") as HTMLSelectElement;
    const tagOption = within(tagSelect).getByRole("option", { name: "源头厂商" }) as HTMLOptionElement;
    tagOption.selected = true;
    fireEvent.change(tagSelect);
    const vendorForm = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(vendorForm).getByRole("button", { name: "创建记录" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith(
        "vendors",
        expect.objectContaining({ name: "浙江汉丰农机有限公司", websiteUrl: "https://vendor.example.com", tagIds: [1] }),
      ),
    );
  });

  it("publishes a vendor when the front display checkbox is selected", async () => {
    renderAdmin("/admin/vendors");
    await openCreateEditor("新增厂商信息");

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "展示测试厂商" } });
    fireEvent.click(screen.getByLabelText("前台显示（勾选即发布）"));
    const vendorForm = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(vendorForm).getByRole("button", { name: "创建记录" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith(
        "vendors",
        expect.objectContaining({ name: "展示测试厂商", isVisible: true, publicationStatus: "published" }),
      ),
    );
  });

  it("submits processing service fields for vendors", async () => {
    renderAdmin("/admin/vendors");
    await openCreateEditor("新增厂商信息");

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "山东精工加工有限公司" } });
    fireEvent.click(screen.getByLabelText("是否提供加工服务"));
    fireEvent.change(screen.getByLabelText("加工服务能力"), { target: { value: "数控车削、焊接加工" } });
    fireEvent.change(screen.getByLabelText("加工设备"), { target: { value: "数控车床、焊接工位" } });
    const form = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(form).getByRole("button", { name: "创建记录" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith(
        "vendors",
        expect.objectContaining({
          providesProcessing: true,
          processingServices: "数控车削、焊接加工",
          processingEquipment: "数控车床、焊接工位",
        }),
      ),
    );
  });

  it("supports processing tag type in tag forms", async () => {
    renderAdmin("/admin/tags");
    await openCreateEditor("新增厂商标签");

    const tagTypeSelect = await screen.findByLabelText("标签类型");
    expect(within(tagTypeSelect).getByRole("option", { name: "加工服务" })).toHaveValue("processing");
  });

  it("edits the shared product catalog without assigning a single vendor", async () => {
    renderAdmin("/admin/products");
    await openCreateEditor("新增配件产品");

    fireEvent.change(await screen.findByLabelText("产品名称"), { target: { value: "液压油泵总成" } });
    fireEvent.change(screen.getByLabelText("所属分类"), { target: { value: "5" } });
    fireEvent.change(screen.getByLabelText("目录发布状态"), { target: { value: "published" } });
    const form = screen.getByLabelText("产品名称").closest("form") as HTMLFormElement;
    expect(within(form).getByRole("option", { name: "液压系统配件" })).toBeInTheDocument();
    expect(within(form).queryByLabelText("所属厂商")).not.toBeInTheDocument();
    fireEvent.click(within(form).getByRole("button", { name: "创建记录" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith("products", expect.objectContaining({ name: "液压油泵总成", categoryId: 5, publicationStatus: "published" })),
    );
  });
});
