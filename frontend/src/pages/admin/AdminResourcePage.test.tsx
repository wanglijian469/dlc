import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createResource, deleteResource, listResource, updateResource } from "../../api/admin";
import { AdminResourcePage } from "./AdminResourcePage";

vi.mock("../../api/admin", () => ({
  createResource: vi.fn(),
  deleteResource: vi.fn(),
  listConfigs: vi.fn(),
  listResource: vi.fn(),
  updateConfig: vi.fn(),
  updateResource: vi.fn(),
  uploadFile: vi.fn(),
}));

const mockedCreateResource = vi.mocked(createResource);
const mockedListResource = vi.mocked(listResource);
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

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "浙江汉丰农机有限公司" } });
    fireEvent.change(screen.getByLabelText("厂商官网 URL"), { target: { value: "https://vendor.example.com" } });
    const tagSelect = screen.getByLabelText("厂商标签") as HTMLSelectElement;
    const tagOption = within(tagSelect).getByRole("option", { name: "源头厂商" }) as HTMLOptionElement;
    tagOption.selected = true;
    fireEvent.change(tagSelect);
    const vendorForm = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(vendorForm).getByRole("button", { name: "新增" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith(
        "vendors",
        expect.objectContaining({ name: "浙江汉丰农机有限公司", websiteUrl: "https://vendor.example.com", tagIds: [1] }),
      ),
    );
  });

  it("submits processing service fields for vendors", async () => {
    renderAdmin("/admin/vendors");

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "山东精工加工有限公司" } });
    fireEvent.click(screen.getByLabelText("是否提供加工服务"));
    fireEvent.change(screen.getByLabelText("加工服务能力"), { target: { value: "数控车削、焊接加工" } });
    fireEvent.change(screen.getByLabelText("加工设备"), { target: { value: "数控车床、焊接工位" } });
    const form = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(form).getByRole("button", { name: "新增" }));

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

  it("submits public source and review fields for vendors", async () => {
    renderAdmin("/admin/vendors");

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "河北冀农农机具有限公司" } });
    fireEvent.change(screen.getByLabelText("公开信息来源 URL"), { target: { value: "https://www.hbjinong.com/" } });
    fireEvent.change(screen.getByLabelText("采集备注"), { target: { value: "公开官网首页采集，人工复核前不标记平台认证。" } });
    fireEvent.change(screen.getByLabelText("复核状态"), { target: { value: "pending" } });
    const form = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(form).getByRole("button", { name: "新增" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith(
        "vendors",
        expect.objectContaining({
          name: "河北冀农农机具有限公司",
          sourceUrl: "https://www.hbjinong.com/",
          sourceNote: "公开官网首页采集，人工复核前不标记平台认证。",
          reviewStatus: "pending",
        }),
      ),
    );
  });

  it("supports processing tag type in tag forms", async () => {
    renderAdmin("/admin/tags");

    const tagTypeSelect = await screen.findByLabelText("标签类型");
    expect(within(tagTypeSelect).getByRole("option", { name: "加工服务" })).toHaveValue("processing");
  });

  it("uses category and vendor selects for products", async () => {
    renderAdmin("/admin/products");

    fireEvent.change(await screen.findByLabelText("产品名称"), { target: { value: "液压油泵总成" } });
    fireEvent.change(screen.getByLabelText("所属分类"), { target: { value: "5" } });
    fireEvent.change(screen.getByLabelText("所属厂商"), { target: { value: "3" } });
    const form = screen.getByLabelText("产品名称").closest("form") as HTMLFormElement;
    expect(within(form).getByRole("option", { name: "液压系统配件" })).toBeInTheDocument();
    expect(within(form).getByRole("option", { name: "江苏东成农机配件有限公司" })).toBeInTheDocument();
    fireEvent.click(within(form).getByRole("button", { name: "新增" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith("products", expect.objectContaining({ name: "液压油泵总成", categoryId: 5, vendorId: 3 })),
    );
  });
});
