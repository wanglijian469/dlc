import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createResource, deleteResource, listConfigs, listResource, listResourcePage, suggestVendorSEO, updateResource } from "../../api/admin";
import { AdminResourcePage } from "./AdminResourcePage";

vi.mock("../../api/admin", () => ({
  createResource: vi.fn(),
  deleteResource: vi.fn(),
  listConfigs: vi.fn(),
  listResource: vi.fn(),
    listResourcePage: vi.fn(),
    importWorkbook: vi.fn(),
  updateConfig: vi.fn(),
  updateResource: vi.fn(),
  uploadFile: vi.fn(),
  suggestVendorSEO: vi.fn(),
}));

const mockedCreateResource = vi.mocked(createResource);
const mockedListConfigs = vi.mocked(listConfigs);
const mockedListResource = vi.mocked(listResource);
const mockedListResourcePage = vi.mocked(listResourcePage);
const mockedUpdateResource = vi.mocked(updateResource);
const mockedSuggestVendorSEO = vi.mocked(suggestVendorSEO);

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
    mockedListConfigs.mockResolvedValue([
      { id: 1, configKey: "site.meta", configValue: JSON.stringify({ siteName: "大陆农机配件" }), description: "站点品牌和页脚信息" },
      { id: 2, configKey: "home.modules", configValue: "[]", description: "首页实际展示模块" },
      { id: 3, configKey: "site.theme", configValue: "{}", description: "主题色" },
    ]);
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
    mockedSuggestVendorSEO.mockImplementation(async (vendor) => ({
      seoTitle: `${vendor.shortName || vendor.name || "厂商"}｜链条、齿轮厂家`,
      seoDescription: `${vendor.name || "该厂商"}主营链条、齿轮。查看企业资料、产品信息与联系方式。`,
      sourceFields: ["厂商名称", "主营产品"],
    }));
    vi.mocked(deleteResource).mockResolvedValue({ deleted: true });
  });

  it("groups platform configuration into focused submenus and exposes footer fields", async () => {
    renderAdmin("/admin/configs");

    expect((await screen.findAllByText("站点与页脚")).length).toBeGreaterThan(0);
    expect(screen.getByLabelText("版权年份")).toBeInTheDocument();
    expect(screen.getByLabelText("版权所有者")).toBeInTheDocument();
    expect(screen.getByLabelText("备案号")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "首页展示" }));
    expect(await screen.findByText("首页模块编排")).toBeInTheDocument();
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

  it("shows logo and cover previews with their own image fields", async () => {
    mockedListResource.mockImplementation((resource) => {
      if (resource === "vendors") return Promise.resolve([{ id: 3, name: "预览厂商", logo: "/api/media/50", logoAssetId: 50, coverImage: "/api/media/49", coverAssetId: 49 }] as never);
      return Promise.resolve([] as never);
    });
    renderAdmin("/admin/vendors");

    fireEvent.click(await screen.findByRole("button", { name: "编辑" }));

    expect(screen.getByText("Logo URL 图片预览")).toBeInTheDocument();
    expect(screen.getByText("封面 URL 图片预览")).toBeInTheDocument();
    expect(screen.getByText("Logo URL 图片预览").closest(".image-field")).not.toBeNull();
    expect(screen.getByText("封面 URL 图片预览").closest(".image-field")).not.toBeNull();
  });

  it("offers the XLSX template and bulk import on vendor and product pages", async () => {
    const { unmount } = renderAdmin("/admin/vendors");
    expect(await screen.findByRole("link", { name: "下载导入模板" })).toHaveAttribute("href", "/templates/农机配件平台_厂商产品资料采集模板.xlsx");
    expect(screen.getByText("批量导入 XLSX")).toBeInTheDocument();
    unmount();

    renderAdmin("/admin/products");
    expect(await screen.findByRole("link", { name: "下载导入模板" })).toBeInTheDocument();
    expect(screen.getByText("批量导入 XLSX")).toBeInTheDocument();
  });

  it("submits rich vendor fields with selected tag ids", async () => {
    renderAdmin("/admin/vendors");
    await openCreateEditor("新增厂商信息");

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "浙江汉丰农机有限公司" } });
    fireEvent.change(screen.getByLabelText("厂商官网 URL"), { target: { value: "https://vendor.example.com" } });
    const vendorTags = screen.getByRole("group", { name: "配件厂商标签" });
    fireEvent.click(within(vendorTags).getByRole("checkbox", { name: "源头厂商" }));
    const vendorForm = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(vendorForm).getByRole("button", { name: "创建记录" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith(
        "vendors",
        expect.objectContaining({ name: "浙江汉丰农机有限公司", websiteUrl: "https://vendor.example.com", tagIds: [1] }),
      ),
    );
  });

  it("generates vendor SEO suggestions and preserves manual overrides", async () => {
    renderAdmin("/admin/vendors");
    await openCreateEditor("新增厂商信息");

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "河北冀农农机具有限公司" } });
    fireEvent.change(screen.getByLabelText("主营产品"), { target: { value: "旋耕机链条、齿轮" } });

    await waitFor(() => expect(mockedSuggestVendorSEO).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByLabelText("SEO 标题")).toHaveValue("河北冀农农机具有限公司｜链条、齿轮厂家"));
    expect(screen.getAllByText("自动生成").length).toBe(2);
    expect(screen.getByText("建议依据：厂商名称、主营产品")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("SEO 标题"), { target: { value: "人工优化标题" } });
    expect(screen.getByText("人工设置")).toBeInTheDocument();
    const titleField = screen.getByLabelText("SEO 标题").closest("label") as HTMLLabelElement;
    fireEvent.click(within(titleField).getByRole("button", { name: "恢复自动生成" }));
    expect(screen.getByLabelText("SEO 标题")).toHaveValue("河北冀农农机具有限公司｜链条、齿轮厂家");

    const form = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(form).getByRole("button", { name: "创建记录" }));
    await waitFor(() => expect(mockedCreateResource).toHaveBeenCalledWith("vendors", expect.objectContaining({
      seoTitleManual: false,
      seoDescriptionManual: false,
      seoTitle: "河北冀农农机具有限公司｜链条、齿轮厂家",
    })));
  });

  it("publishes a vendor when the front display checkbox is selected", async () => {
    renderAdmin("/admin/vendors");
    await openCreateEditor("新增厂商信息");

    fireEvent.change(await screen.findByLabelText("厂商名称"), { target: { value: "展示测试厂商" } });
    const visibilityToggle = screen.getByLabelText("前台显示");
    expect(visibilityToggle.closest("label")).toHaveClass("admin-toggle-field");
    expect(screen.getByText("勾选并保存后发布到前台")).toBeInTheDocument();
    fireEvent.click(visibilityToggle);
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
    const processingTags = screen.getByRole("group", { name: "加工服务标签" });
    fireEvent.click(within(processingTags).getByRole("checkbox", { name: "数控车削" }));
    fireEvent.change(screen.getByLabelText("加工服务能力"), { target: { value: "数控车削、焊接加工" } });
    fireEvent.change(screen.getByLabelText("加工设备"), { target: { value: "数控车床、焊接工位" } });
    const form = screen.getByLabelText("厂商名称").closest("form") as HTMLFormElement;
    fireEvent.click(within(form).getByRole("button", { name: "创建记录" }));

    await waitFor(() =>
      expect(mockedCreateResource).toHaveBeenCalledWith(
        "vendors",
        expect.objectContaining({
          providesProcessing: true,
          tagIds: [2],
          processingServices: "数控车削、焊接加工",
          processingEquipment: "数控车床、焊接工位",
        }),
      ),
    );
  });

  it("separates vendor and processing tags into checkbox groups", async () => {
    renderAdmin("/admin/vendors");
    await openCreateEditor("新增厂商信息");

    const vendorTags = screen.getByRole("group", { name: "配件厂商标签" });
    const processingTags = screen.getByRole("group", { name: "加工服务标签" });
    expect(within(vendorTags).getByRole("checkbox", { name: "源头厂商" })).toBeInTheDocument();
    expect(within(vendorTags).queryByRole("checkbox", { name: "数控车削" })).not.toBeInTheDocument();
    expect(within(processingTags).getByRole("checkbox", { name: "数控车削" })).toBeInTheDocument();
    expect(within(processingTags).queryByRole("checkbox", { name: "源头厂商" })).not.toBeInTheDocument();
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

  it("keeps category-generated and legacy sidebar anchors out of the quick-navigation list", async () => {
    mockedListResource.mockImplementation((resource) => {
      if (resource === "menus") return Promise.resolve([
        { id: 10, name: "播种施肥配件", categoryId: 5, menuType: "sidebar", path: "/products?categoryId=5" },
        { id: 11, name: "全部分类", menuType: "mobile", path: "/products" },
      ] as never);
      if (resource === "categories") return Promise.resolve([{ id: 5, name: "播种施肥配件" }] as never);
      return Promise.resolve([] as never);
    });
    renderAdmin("/admin/menus");

    expect(await screen.findByText("分类导航已自动生成")).toBeInTheDocument();
    expect(screen.queryByText("播种施肥配件")).not.toBeInTheDocument();
    expect(screen.getByText("全部分类")).toBeInTheDocument();
    expect(screen.getByText("移动快捷入口")).toBeInTheDocument();
    expect(screen.getByText("/products")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "编辑" })).toBeInTheDocument();
  });
});
