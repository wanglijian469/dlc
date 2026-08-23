import { FormEvent, useEffect, useMemo, useState, type Dispatch, type SetStateAction } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { Download, FileUp, Plus, X } from "lucide-react";
import {
  createResource,
  batchSaveProductSuppliers,
  deleteResource,
	 downloadRemoteImage,
  getCaptureAISettings,
  importWorkbook,
  listConfigs,
  listResource,
  listResourcePage,
  ResourceName,
  ResourceRecord,
  saveRevision,
  suggestVendorSEO,
  submitRevision,
  updateConfig,
  updateCaptureAISettings,
  updateResource,
  uploadFile,
  type BulkImportResult,
} from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import type { Category, Menu, Product, SiteConfig, Vendor, VendorCategory, VendorOption } from "../../types/api";
import type { VendorMedia } from "../../types/api";
import { AdminModal } from "../../components/admin/AdminModal";
import { ProtectedMediaImage } from "../../components/admin/ProtectedMediaImage";
import { BlocksEditor, GalleryEditor, SpecsEditor, VendorMediaEditor } from "../../components/admin/StructuredEditors";
import { Pagination } from "../../components/public/Pagination";
import { AdminProductSuppliersEditor } from "../../components/admin/AdminProductSuppliersEditor";
import { AdminVendorProductsPanel } from "../../components/admin/AdminVendorProductsPanel";
import { NewProductVendorPicker, VendorOptionSearch } from "../../components/admin/VendorOptionSearch";
import { AdminCategoriesPage } from "./AdminCategoriesPage";
import { AdminVendorCategoriesPage } from "./AdminVendorCategoriesPage";
import { hierarchicalCategoryOptions } from "../../utils/categories";
import { getApiErrorMessage } from "../../api/client";
import { processingToggleDescription, productFieldGuidance, vendorFieldGuidance } from "../../config/formGuidance";
import { StaticPageManager, StaticPageStatusCell, useStaticPageStatuses } from "../../components/admin/StaticPageControls";
import { AccessProtectionManager } from "../../components/admin/AccessProtectionManager";

type FormValue = string | number | boolean | number[] | VendorMedia[];
type FormState = Record<string, FormValue>;

type Field = {
  key: string;
  label: string;
  type?: "text" | "number" | "checkbox" | "textarea" | "select" | "multiselect" | "checkbox-group" | "image" | "gallery" | "specs" | "blocks";
  options?: { label: string; value: string | number | boolean }[];
  refResource?: ResourceName;
  placeholder?: string;
  description?: string;
  group?: string;
  optionTagType?: "vendor" | "processing";
  maxLength?: number;
  pattern?: string;
};

const menuTypes = [
  { label: "顶部导航", value: "top" },
  { label: "侧边快捷导航", value: "sidebar" },
  { label: "辅助菜单", value: "auxiliary" },
  { label: "移动端快捷入口", value: "mobile" },
  { label: "移动端底部", value: "mobile_bottom" },
];

const reviewStatusOptions = [
  { label: "待人工复核", value: "pending" },
  { label: "已复核", value: "verified" },
  { label: "已驳回", value: "rejected" },
];

const schemas: Record<ResourceName, { title: string; fields: Field[] }> = {
  menus: {
    title: "页面与快捷导航",
    fields: [
      { key: "name", label: "菜单名称" },
      { key: "parentId", label: "父级菜单", type: "select", refResource: "menus" },
      { key: "icon", label: "图标" },
      { key: "menuType", label: "菜单类型", type: "select", options: menuTypes },
      { key: "path", label: "跳转路径" },
      { key: "sortOrder", label: "排序", type: "number" },
      { key: "isEnabled", label: "启用", type: "checkbox" },
      { key: "isTop", label: "置顶", type: "checkbox" },
      { key: "isDefaultOpen", label: "默认展开", type: "checkbox" },
    ],
  },
  vendors: {
    title: "厂商信息",
    fields: [
      { key: "name", label: "厂商名称" },
      { key: "shortName", label: "简称", placeholder: vendorFieldGuidance.shortName },
      { key: "province", label: "省份" },
      { key: "city", label: "城市" },
      { key: "county", label: "区县" },
      { key: "address", label: "详细地址" },
      { key: "description", label: "公司介绍", type: "textarea", placeholder: vendorFieldGuidance.description },
      { key: "serviceAdvantages", label: "服务优势", type: "textarea", placeholder: vendorFieldGuidance.serviceAdvantages, maxLength: 80, group: "展示信息" },
      { key: "mainProducts", label: "主营产品", type: "textarea", placeholder: vendorFieldGuidance.mainProducts, group: "展示信息" },
      { key: "logo", label: "Logo URL", type: "image", group: "展示信息" },
      { key: "coverImage", label: "封面 URL", type: "image", group: "展示信息" },
      { key: "seoTitle", label: "SEO 标题" },
      { key: "seoDescription", label: "SEO 摘要", type: "textarea" },
      { key: "seoTitleManual", label: "SEO 标题人工设置", type: "checkbox", group: "SEO 优化建议" },
      { key: "seoDescriptionManual", label: "SEO 摘要人工设置", type: "checkbox", group: "SEO 优化建议" },
      { key: "establishedYear", label: "成立年份" },
      { key: "factoryArea", label: "厂房面积" },
      { key: "employeeCount", label: "员工规模" },
      { key: "annualCapacity", label: "年产能", type: "textarea", placeholder: vendorFieldGuidance.annualCapacity },
      { key: "equipment", label: "主要设备", type: "textarea", placeholder: vendorFieldGuidance.equipment },
      { key: "certifications", label: "认证资质", type: "textarea", placeholder: vendorFieldGuidance.certifications },
      { key: "reviewStatus", label: "复核状态", type: "select", options: reviewStatusOptions },
      { key: "websiteUrl", label: "厂商官网 URL" },
      { key: "phone", label: "联系电话" },
      { key: "phonePublic", label: "电话公开展示", type: "checkbox", description: "公开后会直接出现在前台和静态页面中，可能被搜索引擎和第三方采集。" },
      { key: "wechat", label: "微信" },
      { key: "wechatQrCode", label: "微信二维码", type: "image", description: "请上传清晰的正方形 PNG、JPG 或 WebP 图片；公开范围跟随“微信公开展示”。" },
      { key: "wechatPublic", label: "微信公开展示", type: "checkbox", description: "公开后会直接出现在前台和静态页面中，可能被搜索引擎和第三方采集。" },
      { key: "contactName", label: "联系人" },
      { key: "contactNamePublic", label: "联系人公开展示", type: "checkbox", description: "公开后会直接出现在前台和静态页面中，可能被搜索引擎和第三方采集。" },
      { key: "tagIds", label: "配件厂商标签", type: "checkbox-group", refResource: "tags", optionTagType: "vendor", group: "展示信息", description: "用于标识厂商经营与展示属性，可多选。" },
	  { key: "vendorCategoryIds", label: "厂商分类", type: "checkbox-group", refResource: "vendor-categories", group: "展示信息", description: "用于确定企业在厂商资源导航中的展示位置，可多选。" },
      { key: "isRecommended", label: "推荐厂商", type: "checkbox", description: "在推荐厂商区域优先展示" },
      { key: "isVerified", label: "平台认证", type: "checkbox", description: "在前台显示认证标识" },
      { key: "isVisible", label: "前台显示", type: "checkbox", description: "勾选并保存后发布到前台" },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
  tags: {
    title: "厂商标签",
    fields: [
      { key: "name", label: "标签名称" },
      { key: "tagType", label: "标签类型", type: "select", options: [{ label: "厂商", value: "vendor" }, { label: "产品", value: "product" }] },
      { key: "color", label: "颜色" },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
  categories: {
    title: "配件分类",
    fields: [
      { key: "name", label: "分类名称" },
      { key: "parentId", label: "父级分类", type: "select", refResource: "categories" },
      { key: "icon", label: "图标" },
      { key: "sortOrder", label: "排序", type: "number" },
      { key: "isEnabled", label: "启用", type: "checkbox" },
    ],
  },
	"vendor-categories": { title: "厂商分类", fields: [] },
  products: {
    title: "配件产品",
    fields: [
      { key: "name", label: "产品名称" },
      { key: "image", label: "产品主图", type: "image" },
      { key: "categoryId", label: "所属分类", type: "select", refResource: "categories" },
      { key: "compatibleModels", label: "适配机型" },
      { key: "description", label: "产品说明", type: "textarea", placeholder: productFieldGuidance.description },
      { key: "detailContent", label: "产品详细说明", type: "textarea", placeholder: productFieldGuidance.detailContent },
      { key: "seoTitle", label: "SEO 标题" },
      { key: "seoDescription", label: "SEO 摘要", type: "textarea" },
      { key: "galleryRaw", label: "产品图库", type: "gallery" },
      { key: "specsRaw", label: "规格参数", type: "specs" },
      { key: "priceNote", label: "价格说明" },
      { key: "inquiryText", label: "询价按钮文案" },
      { key: "inquiryPath", label: "询价跳转路径" },
      { key: "isHot", label: "热门", type: "checkbox" },
      { key: "isRecommended", label: "推荐", type: "checkbox" },
      { key: "publicationStatus", label: "目录发布状态", type: "select", options: [{ label: "已发布", value: "published" }, { label: "草稿", value: "draft" }, { label: "已隐藏", value: "hidden" }] },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
  banners: {
    title: "Banner 管理",
    fields: [
      { key: "title", label: "标题" },
      { key: "backgroundImage", label: "背景图 URL", type: "image" },
      { key: "searchPlaceholder", label: "搜索占位文案" },
      { key: "hotKeywordsRaw", label: "热门关键词", placeholder: "用英文逗号分隔" },
      { key: "isEnabled", label: "启用", type: "checkbox" },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
  pages: {
    title: "页面与行业文章",
    fields: [
      { key: "pageType", label: "内容类型", type: "select", options: [{ label: "平台页面", value: "page" }, { label: "行业文章", value: "article" }] },
      { key: "slug", label: "页面标识" },
      { key: "title", label: "页面标题" },
      { key: "summary", label: "摘要", type: "textarea" },
      { key: "coverImage", label: "文章封面", type: "image" },
      { key: "authorName", label: "作者/来源" },
      { key: "publishedAt", label: "发布时间（ISO 日期）", placeholder: "2026-07-15T00:00:00+08:00" },
      { key: "content", label: "正文", type: "textarea" },
      { key: "blocksRaw", label: "结构化内容", type: "blocks" },
      { key: "seoKeywords", label: "SEO 关键词" },
      { key: "seoTitle", label: "SEO 标题" },
      { key: "seoDescription", label: "SEO 摘要", type: "textarea" },
      { key: "relatedCategoryId", label: "关联配件分类", type: "select", refResource: "categories" },
      { key: "relatedProductId", label: "关联产品", type: "select", refResource: "products" },
      { key: "relatedVendorId", label: "关联厂商", type: "select", refResource: "vendors" },
      { key: "isEnabled", label: "启用", type: "checkbox" },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
  "friend-links": {
    title: "友情链接",
    fields: [
      { key: "name", label: "链接名称" },
      { key: "url", label: "链接 URL" },
      { key: "logo", label: "Logo URL", type: "image" },
      { key: "isEnabled", label: "启用", type: "checkbox" },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
};

const processingVendorFields: Field[] = [
  { key: "providesProcessing", label: "是否提供加工服务", type: "checkbox", description: processingToggleDescription },
  { key: "tagIds", label: "加工服务标签", type: "checkbox-group", refResource: "tags", optionTagType: "processing", group: "加工能力", description: "用于标识可承接的加工方式，可多选。" },
  { key: "processingServices", label: "加工服务能力", type: "textarea", placeholder: vendorFieldGuidance.processingServices },
  { key: "processingMaterials", label: "可加工材料 / 配件类型", type: "textarea", placeholder: vendorFieldGuidance.processingMaterials },
  { key: "processingEquipment", label: "加工设备", type: "textarea", placeholder: vendorFieldGuidance.processingEquipment },
  { key: "processingCapacity", label: "产能 / 交期", type: "textarea", placeholder: vendorFieldGuidance.processingCapacity },
  { key: "processingRegions", label: "加工服务区域", type: "textarea", placeholder: vendorFieldGuidance.processingRegions },
  { key: "processingNotes", label: "接单说明", type: "textarea", placeholder: vendorFieldGuidance.processingNotes },
];

extendProcessingSchemas();
extendSEOSchemas();
extendProductGuidance();

function extendProductGuidance() {
	const structuredFields = new Set(["image", "categoryId", "galleryRaw", "specsRaw"]);
	for (const field of schemas.products.fields) {
		const guidance = productFieldGuidance[field.key as keyof typeof productFieldGuidance];
		if (!guidance) continue;
		field.placeholder = guidance;
		if (structuredFields.has(field.key)) field.description = guidance;
	}
}

function extendSEOSchemas() {
  (["vendors", "products", "categories"] as ResourceName[]).forEach((resource) => {
    const fields = schemas[resource].fields;
    if (!fields.some((field) => field.key === "slug")) {
      const nameIndex = fields.findIndex((field) => field.key === "name");
		fields.splice(nameIndex + 1, 0, resource === "vendors"
			? { key: "slug", label: "厂商网站地址标识", description: "3–16 位小写字母或数字；留空时按厂商品牌自动生成，修改后保留 301 历史重定向", maxLength: 16, pattern: "[a-z0-9]{3,16}" }
			: { key: "slug", label: "短拼音标识", description: "发布后修改会自动保留 301 历史重定向" });
    }
  });
  const categoryFields = schemas.categories.fields;
  if (!categoryFields.some((field) => field.key === "seoTitle")) {
    categoryFields.push({ key: "seoTitle", label: "SEO 标题" }, { key: "seoDescription", label: "SEO 摘要", type: "textarea" });
  }
}

function extendProcessingSchemas() {
  const vendorFields = schemas.vendors.fields;
  if (!vendorFields.some((field) => field.key === "providesProcessing")) {
    const certificationsIndex = vendorFields.findIndex((field) => field.key === "certifications");
    vendorFields.splice(certificationsIndex >= 0 ? certificationsIndex + 1 : vendorFields.length, 0, ...processingVendorFields);
  }
  const tagTypeField = schemas.tags.fields.find((field) => field.key === "tagType");
  if (tagTypeField?.options && !tagTypeField.options.some((option) => option.value === "processing")) {
    tagTypeField.options.push({ label: "加工服务", value: "processing" });
  }
}

export function AdminResourcePage({ resourceName }: { resourceName?: ResourceName }) {
  const { resource: routeResource = "menus" } = useParams();
  const resource = resourceName || routeResource;
  if (resource === "configs") return <ConfigPage />;
  if (resource === "categories") return <AdminCategoriesPage />;
	if (resource === "vendor-categories") return <AdminVendorCategoriesPage />;
  const name = (schemas[resource as ResourceName] ? resource : "menus") as ResourceName;
  const schema = schemas[name];
  const isEditor = localStorage.getItem("cms_role") === "editor";
  const [rows, setRows] = useState<ResourceRecord[]>([]);
  const [refs, setRefs] = useState<Partial<Record<ResourceName, ResourceRecord[]>>>({});
  const [form, setForm] = useState<FormState>({});
  const [editingId, setEditingId] = useState<number | null>(null);
  const [keyword, setKeyword] = useState("");
	const [statusFilter, setStatusFilter] = useState("");
	const [provinceFilter, setProvinceFilter] = useState("");
  const [categoryFilter, setCategoryFilter] = useState(0);
  const [associationFilter, setAssociationFilter] = useState<"" | "linked" | "unlinked">("");
  const [vendorFilter, setVendorFilter] = useState<VendorOption | null>(null);
  const [selectedProductIds, setSelectedProductIds] = useState<Set<number>>(new Set());
  const [batchEditorOpen, setBatchEditorOpen] = useState(false);
  const [batchVendor, setBatchVendor] = useState<VendorOption | null>(null);
  const [batchSaving, setBatchSaving] = useState(false);
  const [batchMessage, setBatchMessage] = useState("");
  const [newProductVendors, setNewProductVendors] = useState<VendorOption[]>([]);
  const [message, setMessage] = useState("");
  const [importResult, setImportResult] = useState<BulkImportResult | null>(null);
  const [importing, setImporting] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const serverPaged = name === "vendors" || name === "products";
	const debouncedKeyword = useDebouncedValue(keyword, 300);

  const load = () => {
    if (serverPaged) {
	  const filters = name === "vendors"
        ? { publicationStatus: statusFilter || undefined, province: provinceFilter || undefined }
        : {
            publicationStatus: statusFilter || undefined,
            categoryId: categoryFilter || undefined,
            vendorId: vendorFilter?.id,
            associationStatus: associationFilter || undefined,
          };
      return listResourcePage<ResourceRecord>(name, { page, pageSize: 20, search: debouncedKeyword.trim() || undefined, ...filters }).then((result) => { setRows(result.items); setTotal(result.total); });
    }
    return listResource<ResourceRecord>(name).then((items) => { setRows(items); setTotal(items.length); });
  };
  useEffect(() => {
    setMessage("");
    setForm(defaultForm(schema.fields));
    setEditingId(null);
    setEditorOpen(false);
    setNewProductVendors([]);
    setSelectedProductIds(new Set());
    setVendorFilter(null);
    setCategoryFilter(0);
    setAssociationFilter("");
    setPage(1);
    void load();
    const resources = Array.from(new Set(schema.fields.map((field) => field.refResource).filter(Boolean))) as ResourceName[];
		if (name === "menus" && !resources.includes("categories")) resources.push("categories");
    if (resources.length) {
      void Promise.all(resources.map((resourceName) => listResource<ResourceRecord>(resourceName).then((items) => [resourceName, items] as const))).then((entries) => setRefs(Object.fromEntries(entries)));
    } else {
      setRefs({});
    }
  }, [name]);

  useEffect(() => {
    if (serverPaged) void load();
  }, [page, debouncedKeyword, statusFilter, provinceFilter, categoryFilter, associationFilter, vendorFilter?.id]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");
    const payload = payloadFromForm(schema.fields, form) as Partial<ResourceRecord> & { media?: VendorMedia[] };
    if (name === "pages") {
      const pagePayload = payload as Record<string, unknown>;
      if (!pagePayload.pageType) pagePayload.pageType = "page";
      for (const key of ["publishedAt", "relatedCategoryId", "relatedProductId", "relatedVendorId"]) {
        if (!pagePayload[key]) delete pagePayload[key];
      }
    }
    if (name === "vendors") {
      if (Array.from(String(form.serviceAdvantages || "")).length > 80) {
        setMessage("服务优势不能超过 80 个字符，请缩短后再保存");
        return;
      }
      payload.media = (form.media as VendorMedia[] | undefined) || [];
      (payload as { publicationStatus?: "published" | "hidden" }).publicationStatus = Boolean(form.isVisible) ? "published" : "hidden";
    }
    if (name === "products" && !editingId) {
      (payload as Record<string, unknown>).vendorIds = newProductVendors.map((vendor) => vendor.id);
      const publishing = (payload as Record<string, unknown>).publicationStatus === "published";
      const hasPublishableVendor = newProductVendors.some((vendor) => vendor.publicationStatus === "published" && vendor.isVisible);
      if (publishing && !hasPublishableVendor) {
        setMessage("产品发布前必须关联至少一家前台已发布的厂商");
        return;
      }
    }
    const action = async () => {
      if (isEditor) {
        if (!editingId) throw new Error("编辑角色只能提交现有内容的修订");
        const resourceType = name === "pages"
          ? ((payload as Record<string, unknown>).pageType === "article" ? "article" : "page")
          : name.slice(0, -1) as "vendor" | "product";
        const revision = await saveRevision({
          resourceType,
          resourceId: editingId,
          baseVersion: Number(form.contentVersion || 1),
          snapshot: payload,
        });
        await submitRevision(revision.id);
        return;
      }
      await (editingId ? updateResource<ResourceRecord>(name, editingId, payload) : createResource<ResourceRecord>(name, payload));
    };
    void action()
      .then(() => {
        setForm(defaultForm(schema.fields));
        setEditingId(null);
        setEditorOpen(false);
        setNewProductVendors([]);
        setMessage(isEditor ? "修订已保存并送审" : "保存成功");
        void load();
      })
      .catch((error) => setMessage(getApiErrorMessage(error, "保存失败，请检查字段")));
  };

  const filteredRows = useMemo(() => {
    const value = keyword.trim().toLowerCase();
    const visibleRows = name === "menus"
      ? rows.filter((row) => {
          const menu = row as Menu;
          return menu.menuType !== "sidebar" && !menu.categoryId;
        })
      : rows;
    if (!value) return visibleRows;
    return visibleRows.filter((row) => JSON.stringify(row).toLowerCase().includes(value));
  }, [keyword, rows]);

  return (
    <AdminLayout title={schema.title}>
      {message && !editorOpen && !batchEditorOpen && <p className="admin-message">{message}</p>}
      {name === "menus" && <section className="menu-management-note"><strong>分类导航已自动生成</strong><span>一级、二级配件分类请在“配件分类”维护；此处仅维护页面入口、非分类快捷项和“全部分类”等移动端入口。</span></section>}
      <div className="admin-toolbar admin-panel compact">
        <input value={keyword} placeholder="搜索当前列表" onChange={(event) => { setKeyword(event.target.value); setPage(1); if (name === "products") setSelectedProductIds(new Set()); }} />
		{name === "vendors" && <><select aria-label="发布状态筛选" value={statusFilter} onChange={(event) => { setStatusFilter(event.target.value); setPage(1); }}><option value="">全部发布状态</option><option value="published">已发布</option><option value="draft">草稿</option><option value="hidden">已隐藏</option></select><input aria-label="地区筛选" placeholder="输入省份" value={provinceFilter} onChange={(event) => { setProvinceFilter(event.target.value); setPage(1); }} /></>}
		{name === "products" && <>
          <select aria-label="产品发布状态筛选" value={statusFilter} onChange={(event) => { setStatusFilter(event.target.value); setPage(1); setSelectedProductIds(new Set()); }}><option value="">全部发布状态</option><option value="published">已发布</option><option value="draft">草稿</option><option value="hidden">已隐藏</option></select>
          <select aria-label="产品分类筛选" value={categoryFilter || ""} onChange={(event) => { setCategoryFilter(Number(event.target.value)); setPage(1); setSelectedProductIds(new Set()); }}><option value="">全部分类</option>{hierarchicalCategoryOptions((refs.categories || []) as Category[]).map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}</select>
          <select aria-label="厂商关联状态筛选" value={associationFilter} onChange={(event) => { setAssociationFilter(event.target.value as "" | "linked" | "unlinked"); setPage(1); setSelectedProductIds(new Set()); }}><option value="">全部关联状态</option><option value="linked">已关联厂商</option><option value="unlinked">未关联厂商</option></select>
          <div className="product-vendor-filter">{vendorFilter ? <button className="filter-chip" type="button" onClick={() => { setVendorFilter(null); setPage(1); setSelectedProductIds(new Set()); }}>{vendorFilter.name}<X size={14} /></button> : <VendorOptionSearch label="按厂商筛选" onSelect={(vendor) => { setVendorFilter(vendor); setPage(1); setSelectedProductIds(new Set()); }} />}</div>
        </>}
        {!isEditor && name === "products" && selectedProductIds.size > 0 && <button className="outline-btn batch-associate-button" type="button" onClick={() => { setBatchVendor(null); setBatchMessage(""); setBatchEditorOpen(true); }}>批量关联厂商（{selectedProductIds.size}）</button>}
        {!isEditor && (name === "vendors" || name === "products") && <div className="admin-import-actions"><a className="outline-btn" download href="/templates/农机配件平台_厂商产品资料采集模板.xlsx"><Download size={16} />下载导入模板</a><label className={`outline-btn admin-import-button ${importing ? "disabled" : ""}`}><FileUp size={16} />{importing ? "正在导入…" : "批量导入 XLSX"}<input accept=".xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" disabled={importing} type="file" onChange={(event) => { const file = event.target.files?.[0]; event.target.value = ""; if (!file) return; setImporting(true); setImportResult(null); setMessage(""); void importWorkbook(name, file).then((result) => { setImportResult(result); if (result.imported) { setMessage("批量导入成功"); void load(); } }).catch((error) => setMessage(error?.response?.data?.message || "批量导入失败，请检查文件格式")).finally(() => setImporting(false)); }} /></label></div>}
        {!isEditor && <button className="primary-btn" type="button" onClick={() => { setEditingId(null); setNewProductVendors([]); setForm({ ...defaultForm(schema.fields), ...(name === "vendors" ? { media: [] } : {}), ...(name === "products" ? { publicationStatus: "draft" } : {}) }); setEditorOpen(true); }}><Plus size={16} />新增{schema.title}</button>}
      </div>
      {importResult && <section className={`admin-import-result ${importResult.imported ? "success" : "error"}`}><strong>{importResult.imported ? "导入完成" : "表格校验未通过，未写入数据"}</strong><span>读取 {importResult.totalRows} 行 · 新增 {importResult.created} 条 · 更新 {importResult.updated} 条{name === "products" ? ` · 新增供应关系 ${importResult.relationsCreated} 条 · 更新供应关系 ${importResult.relationsUpdated} 条` : ""}</span>{importResult.issues.length > 0 && <ul>{importResult.issues.slice(0, 30).map((issue, index) => <li key={`${issue.sheet}-${issue.row}-${index}`}>{issue.sheet} 第 {issue.row} 行：{issue.message}</li>)}</ul>}{(importResult.warnings || []).length > 0 && <div className="admin-import-warnings"><strong>以下图片未能本地化，已保留原始链接：</strong><ul>{(importResult.warnings || []).slice(0, 30).map((warning, index) => <li key={`${warning.sheet}-${warning.row}-${index}`}>{warning.sheet} 第 {warning.row} 行：{warning.message}</li>)}</ul></div>}</section>}
      <div className="admin-table-panel">
        {name === "products" ? <ProductResourceTable
          rows={filteredRows as Product[]}
          selectionEnabled={!isEditor}
          selectedIds={selectedProductIds}
          onSelectionChange={setSelectedProductIds}
          onEdit={(row) => {
            setEditingId(row.id);
            setNewProductVendors([]);
            setForm({ ...defaultForm(schema.fields), ...(row as unknown as FormState) });
            setEditorOpen(true);
          }}
          onDelete={isEditor ? undefined : (id) => {
            if (!window.confirm("确认删除这条产品记录？")) return;
            void deleteResource(name, id).then(() => { setMessage("删除成功"); setSelectedProductIds((current) => { const next = new Set(current); next.delete(id); return next; }); void load(); }).catch(() => setMessage("删除失败"));
          }}
        /> : <ResourceTable
		  resource={name}
          rows={filteredRows}
          onDelete={isEditor ? undefined : (id) => {
            if (!window.confirm("确认删除这条记录？")) return;
            deleteResource(name, id)
              .then(() => {
                setMessage("删除成功");
                void load();
              })
              .catch(() => setMessage("删除失败"));
          }}
          onEdit={(row) => {
            setEditingId(row.id);
            setForm({ ...defaultForm(schema.fields), ...(row as unknown as FormState) });
            setEditorOpen(true);
          }}
        />}
        {serverPaged && <Pagination onChange={setPage} page={page} pageSize={20} total={total} />}
      </div>
      {editorOpen && <AdminModal label={`${editingId ? "编辑" : "新增"}${schema.title}`} onClose={() => setEditorOpen(false)}><header><div><span>{editingId ? "编辑记录" : "新增记录"}</span><h2>{schema.title}</h2></div><button aria-label="关闭编辑器" type="button" onClick={() => setEditorOpen(false)}><X size={20} /></button></header>{message && <p className="admin-message admin-editor-message">{message}</p>}<form className="admin-form admin-grouped-form" onSubmit={submit}>
        {groupFields(name, schema.fields).map((group) => name === "vendors" && group.title === "SEO 优化建议" ? <fieldset className="vendor-seo-workbench" key={group.title}><legend>{group.title}</legend><VendorSEOEditor form={form} setForm={setForm} /></fieldset> : <fieldset key={group.title}><legend>{group.title}</legend><div className="admin-field-grid">{group.fields.map((field) => {
          const fieldWide = ["textarea", "gallery", "specs", "blocks", "checkbox-group"].includes(field.type || "");
          const structuredEditor = ["gallery", "specs", "blocks"].includes(field.type || "");
          if (field.type === "checkbox-group") return <div className={`admin-checkbox-group-field ${fieldWide ? "field-wide" : ""}`} key={`${field.key}-${field.optionTagType || "all"}`}><strong>{field.label}</strong>{field.description && <small>{field.description}</small>}<FieldInput field={field} form={form} refs={refs} setForm={setForm} /></div>;
          if (field.type === "checkbox" && field.description) return <label className="admin-toggle-field" key={field.key}><FieldInput field={field} form={form} refs={refs} setForm={setForm} /><span><strong>{field.label}</strong><small>{field.description}</small></span></label>;
          if (structuredEditor) return <div className={`admin-guided-field ${fieldWide ? "field-wide" : ""}`} key={field.key}><FieldInput field={field} form={form} refs={refs} setForm={setForm} />{field.description && <small className="field-guidance">{field.description}</small>}</div>;
          if (field.description) return <div className={`admin-guided-field ${fieldWide ? "field-wide" : ""}`} key={field.key}><label>{field.label}<FieldInput field={field} form={form} refs={refs} setForm={setForm} /></label><small className="field-guidance">{field.description}</small></div>;
          return <label className={fieldWide ? "field-wide" : ""} key={field.key}>{field.label}<FieldInput field={field} form={form} refs={refs} setForm={setForm} /></label>;
        })}</div></fieldset>)}
        {name === "vendors" && <fieldset><legend>企业图集</legend><VendorMediaEditor value={(form.media as VendorMedia[] | undefined) || []} onChange={(media) => setForm({ ...form, media })} /></fieldset>}
        {!isEditor && name === "products" && !editingId && <NewProductVendorPicker value={newProductVendors} onChange={setNewProductVendors} />}
        {(["vendors", "products", "categories", "pages"] as ResourceName[]).includes(name) && <SEOPublishChecklist resource={name} form={form} />}
        <div className="admin-editor-actions"><button className="outline-btn" type="button" onClick={() => setEditorOpen(false)}>取消</button><button className="primary-btn" type="submit">{editingId ? "保存修改" : "创建记录"}</button></div>
      </form>{!isEditor && name === "products" && editingId && <AdminProductSuppliersEditor productId={editingId} onChanged={() => void load()} />}{!isEditor && name === "vendors" && editingId && <AdminVendorProductsPanel vendorId={editingId} />}</AdminModal>}
      {batchEditorOpen && <AdminModal label="批量关联厂商" onClose={() => setBatchEditorOpen(false)}><header><div><span>批量操作</span><h2>为 {selectedProductIds.size} 个产品关联厂商</h2></div><button aria-label="关闭批量关联" type="button" onClick={() => setBatchEditorOpen(false)}><X size={20} /></button></header><section className="batch-association-panel"><p>本操作只追加一家厂商，不会替换产品已有的厂商关系。重复关系会自动跳过。</p>{batchMessage && <p className="admin-message">{batchMessage}</p>}{batchVendor ? <div className="batch-vendor-summary"><strong>{batchVendor.name}</strong><span>{[batchVendor.province, batchVendor.city, batchVendor.mainProducts].filter(Boolean).join(" · ")}</span><button type="button" onClick={() => setBatchVendor(null)}>重新选择</button></div> : <VendorOptionSearch label="搜索批量关联厂商" onSelect={setBatchVendor} />}<div className="admin-editor-actions batch-actions"><button className="outline-btn" type="button" onClick={() => setBatchEditorOpen(false)}>取消</button><button className="primary-btn" disabled={!batchVendor || batchSaving} type="button" onClick={() => { if (!batchVendor) return; setBatchSaving(true); setBatchMessage(""); void batchSaveProductSuppliers(Array.from(selectedProductIds), batchVendor.id).then((result) => { setMessage(`批量关联完成：新增 ${result.created} 条，已有 ${result.existing} 条`); setSelectedProductIds(new Set()); setBatchEditorOpen(false); void load(); }).catch((error) => setBatchMessage(getApiErrorMessage(error, "批量关联失败"))).finally(() => setBatchSaving(false)); }}>{batchSaving ? "正在关联…" : `确认关联 ${selectedProductIds.size} 个产品`}</button></div></section></AdminModal>}
    </AdminLayout>
  );
}

function SEOPublishChecklist({ resource, form }: { resource: ResourceName; form: FormState }) {
  const title = String(form.seoTitle || form.title || form.name || "").trim();
  const description = String(form.seoDescription || form.summary || form.description || "").trim();
  const slug = String(form.slug || "").trim();
  const image = String(form.coverImage || form.image || form.logo || "").trim();
  const body = String(form.content || form.detailContent || form.description || "");
  const issues = [
    !slug && "缺少短拼音 URL 标识",
    !title && "缺少标题",
    title.length > 32 && "标题建议控制在 32 个汉字以内",
    !description && "缺少搜索摘要",
    description.length > 120 && "摘要建议控制在 120 个汉字以内",
    resource !== "categories" && !image && "缺少分享图片",
    resource === "pages" && body && !/\/(products|vendors|guides)/.test(body) && "正文没有内部链接",
  ].filter(Boolean) as string[];
  const score = Math.max(0, 100 - issues.length * 14);
  return <fieldset className="seo-publish-checklist"><legend>SEO 发布检查</legend><div className="seo-score"><strong>{score}</strong><span>/ 100</span></div><div className="seo-serp-preview"><small>搜索结果预览</small><h3>{title || "请填写 SEO 标题"}</h3><span>https://example.com/{resource}/{slug || "slug"}</span><p>{description || "请填写能够准确概括页面内容的摘要。"}</p></div>{issues.length ? <ul>{issues.map((issue) => <li key={issue}>{issue}</li>)}</ul> : <p className="seo-check-passed">发布前 SEO 检查已通过</p>}</fieldset>;
}

type VendorSuggestion = Awaited<ReturnType<typeof suggestVendorSEO>>;

function VendorSEOEditor({ form, setForm }: { form: FormState; setForm: Dispatch<SetStateAction<FormState>> }) {
  const [suggestion, setSuggestion] = useState<VendorSuggestion | null>(null);
  const [loading, setLoading] = useState(false);
  const sourceSignature = JSON.stringify([
    form.name, form.shortName, form.province, form.city, form.mainProducts,
    form.description, form.providesProcessing, form.processingServices,
  ]);
  const titleManual = Boolean(form.seoTitleManual);
  const descriptionManual = Boolean(form.seoDescriptionManual);

  useEffect(() => {
    if (!String(form.name || "").trim()) {
      setSuggestion(null);
      return;
    }
    let cancelled = false;
    const timer = window.setTimeout(() => {
      setLoading(true);
      void suggestVendorSEO(form as unknown as Partial<Vendor>)
        .then((result) => {
          if (cancelled) return;
          setSuggestion(result);
          setForm((current) => ({
			...current,
			seoTitle: Boolean(current.seoTitleManual) ? current.seoTitle : result.seoTitle,
			seoDescription: Boolean(current.seoDescriptionManual) ? current.seoDescription : result.seoDescription,
		  }));
        })
        .finally(() => { if (!cancelled) setLoading(false); });
    }, 320);
    return () => { cancelled = true; window.clearTimeout(timer); };
  }, [sourceSignature, titleManual, descriptionManual]);

  const title = String(form.seoTitle || "");
  const description = String(form.seoDescription || "");
	const vendorSiteAddress = `${typeof window === "undefined" ? "https://example.com" : window.location.origin}/v/${String(form.slug || "vendor-slug")}`;
  const restore = (field: "title" | "description") => {
    if (field === "title") setForm((current) => ({ ...current, seoTitleManual: false, seoTitle: suggestion?.seoTitle || String(current.seoTitle || "") }));
    else setForm((current) => ({ ...current, seoDescriptionManual: false, seoDescription: suggestion?.seoDescription || String(current.seoDescription || "") }));
  };

  return <div className="vendor-seo-editor">
    <p className="vendor-seo-intro">根据厂商名称、地区、主营产品与加工能力生成。自动模式会随资料更新，人工设置后不再覆盖。</p>
		<div className="vendor-site-address"><strong>厂商独立站地址</strong><code>{vendorSiteAddress}</code><button type="button" onClick={() => void navigator.clipboard.writeText(vendorSiteAddress)}>复制地址</button><a href={vendorSiteAddress} rel="noreferrer" target="_blank">预览网站</a></div>
    <div className="vendor-seo-fields">
      <label>
        <span className="vendor-seo-field-head"><strong>SEO 标题</strong><em className={titleManual ? "manual" : "auto"}>{titleManual ? "人工设置" : "自动生成"}</em></span>
        <input aria-label="SEO 标题" value={title} onChange={(event) => setForm({ ...form, seoTitle: event.target.value, seoTitleManual: true })} />
        <small><span className={Array.from(title).length > 32 ? "over-limit" : ""}>{Array.from(title).length} / 32 字</span>{titleManual && <button type="button" onClick={() => restore("title")}>恢复自动生成</button>}</small>
      </label>
      <label className="field-wide">
        <span className="vendor-seo-field-head"><strong>SEO 摘要</strong><em className={descriptionManual ? "manual" : "auto"}>{descriptionManual ? "人工设置" : "自动生成"}</em></span>
        <textarea aria-label="SEO 摘要" value={description} onChange={(event) => setForm({ ...form, seoDescription: event.target.value, seoDescriptionManual: true })} />
        <small><span className={Array.from(description).length > 120 ? "over-limit" : ""}>{Array.from(description).length} / 120 字</span>{descriptionManual && <button type="button" onClick={() => restore("description")}>恢复自动生成</button>}</small>
      </label>
    </div>
    <div className="vendor-seo-preview"><small>{loading ? "正在更新建议…" : `建议依据：${suggestion?.sourceFields.join("、") || "填写厂商资料后自动生成"}`}</small><h3>{title || "等待生成 SEO 标题"}</h3><span>{vendorSiteAddress}</span><p>{description || "等待生成 SEO 摘要"}</p></div>
  </div>;
}

function useDebouncedValue<T>(value: T, delay: number) {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => { const timer = window.setTimeout(() => setDebounced(value), delay); return () => window.clearTimeout(timer); }, [value, delay]);
  return debounced;
}

function FieldInput({ field, form, refs, setForm }: { field: Field; form: FormState; refs: Partial<Record<ResourceName, ResourceRecord[]>>; setForm: (form: FormState) => void }) {
  if (field.type === "gallery") return <GalleryEditor value={String(form[field.key] || "")} onChange={(value) => setForm({ ...form, [field.key]: value })} />;
  if (field.type === "specs") return <SpecsEditor value={String(form[field.key] || "")} onChange={(value) => setForm({ ...form, [field.key]: value })} />;
  if (field.type === "blocks") return <BlocksEditor value={String(form[field.key] || "")} onChange={(value) => setForm({ ...form, [field.key]: value })} />;
  if (field.type === "textarea") {
    const value = String(form[field.key] ?? "");
    return <div className="limited-field"><textarea className={field.placeholder ? "writing-example" : undefined} placeholder={field.placeholder} value={value} onChange={(event) => setForm({ ...form, [field.key]: event.target.value })} />{field.maxLength && <small className={Array.from(value).length > field.maxLength ? "character-count over-limit" : "character-count"}>{Array.from(value).length} / {field.maxLength} 字</small>}</div>;
  }
  if (field.type === "checkbox") {
    return <input aria-label={field.label} checked={Boolean(form[field.key])} type="checkbox" onChange={(event) => setForm({ ...form, [field.key]: event.target.checked })} />;
  }
  if (field.type === "select") {
    const options = optionsFor(field, refs);
    return (
      <select value={String(form[field.key] ?? "")} onChange={(event) => setForm({ ...form, [field.key]: numericSelect(field) ? Number(event.target.value) : event.target.value })}>
        <option value="">请选择</option>
        {options.map((option) => (
          <option key={String(option.value)} value={String(option.value)}>
            {option.label}
          </option>
        ))}
      </select>
    );
  }
  if (field.type === "multiselect") {
    const selected = Array.isArray(form[field.key]) ? (form[field.key] as number[]).map(String) : [];
    return (
      <select multiple value={selected} onChange={(event) => setForm({ ...form, [field.key]: Array.from(event.target.selectedOptions).map((option) => Number(option.value)) })}>
        {optionsFor(field, refs).map((option) => (
          <option key={String(option.value)} value={String(option.value)}>
            {option.label}
          </option>
        ))}
      </select>
    );
  }
  if (field.type === "checkbox-group") {
    const selected = Array.isArray(form[field.key]) ? (form[field.key] as number[]) : [];
    const options = optionsFor(field, refs);
    if (!options.length) return <p className="tag-checkbox-empty">暂无可选标签，请先在“厂商标签”中配置。</p>;
    return (
      <div aria-label={field.label} className="tag-checkbox-grid" role="group">
        {options.map((option) => {
          const id = Number(option.value);
          return (
            <label className="tag-checkbox-option" key={String(option.value)}>
              <input
                checked={selected.includes(id)}
                type="checkbox"
                onChange={(event) => setForm({
                  ...form,
                  [field.key]: event.target.checked ? Array.from(new Set([...selected, id])) : selected.filter((value) => value !== id),
                })}
              />
              <span>{option.label}</span>
            </label>
          );
        })}
      </div>
    );
  }
  if (field.type === "image") return <ImageField field={field} form={form} setForm={setForm} />;
  if (field.key === "slug") {
    const value = String(form[field.key] ?? "");
    return <div className="limited-field"><input value={value} type="text" maxLength={16} pattern="[a-z0-9]{3,16}" placeholder="例如 jinong" onChange={(event) => setForm({ ...form, [field.key]: event.target.value.toLowerCase().replace(/[^a-z0-9]/g, "") })} /><small className="character-count">公开地址：/v/{value || "自动生成"} · 剩余 {Math.max(0, 16 - value.length)} 位</small></div>;
  }
  return (
    <input
      value={String(form[field.key] ?? "")}
      type={field.type || "text"}
      className={field.placeholder ? "writing-example" : undefined}
      placeholder={field.placeholder}
      maxLength={field.maxLength}
      pattern={field.pattern}
      onChange={(event) => setForm({ ...form, [field.key]: field.type === "number" ? Number(event.target.value) : event.target.value })}
    />
  );
}

function ImageField({ field, form, setForm }: { field: Field; form: FormState; setForm: (form: FormState) => void }) {
  const [downloading, setDownloading] = useState(false);
  const [downloadError, setDownloadError] = useState("");
  const value = String(form[field.key] ?? "");
  const assetKey = field.key === "logo" ? "logoAssetId" : field.key === "coverImage" ? "coverAssetId" : field.key === "wechatQrCode" ? "wechatQrCodeAssetId" : undefined;
  const mediaMatch = value.match(/^\/api\/media\/(\d+)$/);
  const canDownload = /^https?:\/\//i.test(value);

  const useLocalCopy = () => {
    if (!canDownload || downloading) return;
    setDownloading(true);
    setDownloadError("");
    void downloadRemoteImage(value)
      .then((result) => setForm({ ...form, [field.key]: result.url, ...(assetKey ? { [assetKey]: result.assetId } : {}) }))
      .catch((error) => setDownloadError(error?.response?.data?.message || "下载失败，请确认图片地址可公开访问"))
      .finally(() => setDownloading(false));
  };

  return (
    <div className="image-field">
      <input value={value} type="text" placeholder={field.placeholder} onChange={(event) => { setDownloadError(""); setForm({ ...form, [field.key]: event.target.value }); }} />
      <div className="image-field-actions">
        <input
          aria-label={`${field.label} 上传`}
          type="file"
          accept="image/*"
          onChange={(event) => {
            const file = event.target.files?.[0];
            if (!file) return;
            uploadFile(file).then((result) => setForm({ ...form, [field.key]: result.url, ...(assetKey ? { [assetKey]: result.assetId } : {}) }));
          }}
        />
        <button className="outline-btn small" disabled={!canDownload || downloading} onClick={useLocalCopy} type="button">{downloading ? "正在下载…" : "下载到本地"}</button>
        {value && <button className="outline-btn small danger" onClick={() => setForm({ ...form, [field.key]: "", ...(assetKey ? { [assetKey]: "" } : {}) })} type="button">删除图片</button>}
      </div>
      {downloadError && <small className="image-download-error">{downloadError}</small>}
      {value && <div className="image-field-preview"><span>{field.label} 图片预览</span><ProtectedMediaImage alt={`${field.label} 图片预览`} assetId={assetKey ? Number(form[assetKey] || 0) || (mediaMatch ? Number(mediaMatch[1]) : undefined) : mediaMatch ? Number(mediaMatch[1]) : undefined} src={value} /></div>}
    </div>
  );
}

function optionsFor(field: Field, refs: Partial<Record<ResourceName, ResourceRecord[]>>) {
  if (field.options) return field.options;
  if (!field.refResource) return [];
	if (field.refResource === "categories") return hierarchicalCategoryOptions((refs.categories || []) as Category[]);
	if (field.refResource === "vendor-categories") return hierarchicalVendorCategoryOptions((refs["vendor-categories"] || []) as VendorCategory[]);
	if (field.refResource === "menus") return (refs.menus || []).map((item) => {
		const menu = item as Menu;
		const label = menu.categoryId ? `分类快捷项：${menu.name}` : menu.name;
		return { label, value: item.id };
	});
  return (refs[field.refResource] || [])
    .filter((item) => !field.optionTagType || (item as unknown as { tagType?: string }).tagType === field.optionTagType)
    .map((item) => ({ label: String((item as unknown as { name?: string; title?: string }).name || (item as unknown as { title?: string }).title || item.id), value: item.id }));
}

function numericSelect(field: Field) {
  return Boolean(field.refResource || field.key.endsWith("Id") || field.key === "status" || field.key === "parentId");
}

function ResourceTable({ resource, rows, onEdit, onDelete, categoryManagedMenuIDs }: { resource: ResourceName; rows: ResourceRecord[]; onEdit: (row: ResourceRecord) => void; onDelete?: (id: number) => void; categoryManagedMenuIDs?: Set<number> }) {
  const keys = useMemo(() => Object.keys(rows[0] || {}).filter((key) => ["id", "name", "title", "slug", "province", "menuType", "path", "publicationStatus", "sortOrder", "isEnabled", "isVisible", "isRecommended"].includes(key)), [rows]);
	const showSource = categoryManagedMenuIDs !== undefined;
  const showStaticPages = resource === "vendors" && localStorage.getItem("cms_role") === "admin";
  const staticPages = useStaticPageStatuses("vendor", rows.map((row) => row.id), showStaticPages);
  return (
    <table className="admin-table">
      <thead>
        <tr>
          {keys.map((key) => (
            <th key={key}>{columnLabel(key, resource)}</th>
          ))}
			{showSource && <th>来源</th>}
          {showStaticPages && <th>静态页面</th>}
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => (
          <tr key={row.id}>
            {keys.map((key) => (
              <td data-label={columnLabel(key, resource)} key={key}>{formatCell((row as unknown as Record<string, unknown>)[key], key)}</td>
            ))}
			{showSource && <td data-label="来源">{categoryManagedMenuIDs.has(row.id) ? <span className="category-derived-badge">分类生成</span> : "人工维护"}</td>}
            {showStaticPages && <td data-label="静态页面"><StaticPageStatusCell generating={staticPages.generatingId === row.id} status={staticPages.statuses[row.id]} onGenerate={() => void staticPages.generate(row.id)} /></td>}
            <td data-label="操作">
				{categoryManagedMenuIDs?.has(row.id) ? <span className="category-derived-note">请到“配件分类”维护</span> : <><button type="button" onClick={() => onEdit(row)}>编辑</button>{onDelete && <button type="button" onClick={() => onDelete(row.id)}>删除</button>}</>}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function ProductResourceTable({
  rows,
  selectionEnabled,
  selectedIds,
  onSelectionChange,
  onEdit,
  onDelete,
}: {
  rows: Product[];
  selectionEnabled: boolean;
  selectedIds: Set<number>;
  onSelectionChange: Dispatch<SetStateAction<Set<number>>>;
  onEdit: (row: Product) => void;
  onDelete?: (id: number) => void;
}) {
  const showStaticPages = localStorage.getItem("cms_role") === "admin";
  const staticPages = useStaticPageStatuses("product", rows.map((row) => row.id), showStaticPages);
  const currentIDs = rows.map((row) => row.id);
  const allCurrentSelected = currentIDs.length > 0 && currentIDs.every((id) => selectedIds.has(id));
  const toggleCurrentPage = () => onSelectionChange((current) => {
    const next = new Set(current);
    if (allCurrentSelected) currentIDs.forEach((id) => next.delete(id));
    else currentIDs.forEach((id) => next.add(id));
    return next;
  });
  return (
    <table className="admin-table product-management-table">
      <thead>
        <tr>
          {selectionEnabled && <th><input aria-label="选择当前页产品" checked={allCurrentSelected} type="checkbox" onChange={toggleCurrentPage} /></th>}
          <th>产品</th>
          <th>分类</th>
          <th>发布状态</th>
          <th>关联厂商</th>
          {showStaticPages && <th>静态页面</th>}
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => {
          const associated = row.associatedVendors || [];
          const extra = Math.max(0, (row.associationCount || 0) - associated.length);
          return (
            <tr key={row.id}>
              {selectionEnabled && <td data-label="选择"><input aria-label={`选择产品 ${row.name}`} checked={selectedIds.has(row.id)} type="checkbox" onChange={() => onSelectionChange((current) => { const next = new Set(current); if (next.has(row.id)) next.delete(row.id); else next.add(row.id); return next; })} /></td>}
              <td data-label="产品"><strong>{row.name}</strong><small className="product-row-slug">{row.slug || `ID ${row.id}`}</small></td>
              <td data-label="分类">{row.category?.name || "未分类"}</td>
              <td data-label="发布状态"><span className={`publication-badge ${row.publicationStatus || "draft"}`}>{formatCell(row.publicationStatus, "publicationStatus")}</span></td>
              <td data-label="关联厂商">
                {associated.length ? <div className="associated-vendor-summary">{associated.map((vendor) => <span key={vendor.id}>{vendor.name}</span>)}{extra > 0 && <em>另有 {extra} 家</em>}<small>共 {row.associationCount || associated.length} 家</small></div> : <span className="unlinked-product-badge">未关联厂商</span>}
              </td>
              {showStaticPages && <td data-label="静态页面"><StaticPageStatusCell generating={staticPages.generatingId === row.id} status={staticPages.statuses[row.id]} onGenerate={() => void staticPages.generate(row.id)} /></td>}
              <td data-label="操作"><button type="button" onClick={() => onEdit(row)}>编辑</button>{onDelete && <button type="button" onClick={() => onDelete(row.id)}>删除</button>}</td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}

type ConfigSection = "site" | "home" | "theme" | "static" | "security" | "capture";

const configSections: Array<{ key: ConfigSection; label: string; description: string; keys: string[] }> = [
  { key: "site", label: "站点与页脚", description: "品牌、页脚版权与备案信息", keys: ["site.meta"] },
  { key: "home", label: "首页展示", description: "管理当前前台实际展示的厂商与加工服务模块", keys: ["home.modules"] },
  { key: "theme", label: "主题样式", description: "前台主色与强调色", keys: ["site.theme"] },
  { key: "static", label: "静态化与缓存", description: "手动生成厂商与产品公开静态页，管理更新与失败状态", keys: [] },
  { key: "security", label: "访问与采集防护", description: "管理行为识别、封禁、AI 爬虫规则、联系方式额度和公开图片水印", keys: [] },
  { key: "capture", label: "智能采集云服务", description: "配置腾讯云 OCR 与混元视觉识别，不在页面回显密钥", keys: [] },
];

function ConfigPage() {
  const [rows, setRows] = useState<SiteConfig[]>([]);
  const [message, setMessage] = useState("");
  const [searchParams, setSearchParams] = useSearchParams();
  const requestedSection = searchParams.get("section") as ConfigSection | null;
  const activeSection = configSections.some((section) => section.key === requestedSection) ? requestedSection as ConfigSection : "site";
  const section = configSections.find((item) => item.key === activeSection) || configSections[0];
  const visibleRows = rows.filter((row) => section.keys.includes(row.configKey));
  useEffect(() => {
    void listConfigs().then(setRows);
  }, []);
  return (
    <AdminLayout title="平台配置">
      {message && <p className="admin-message">{message}</p>}
      <section className="config-section-header">
        <div><strong>{section.label}</strong><span>{section.description}</span></div>
        <nav aria-label="平台配置子菜单" className="config-section-tabs">
          {configSections.map((item) => <button className={item.key === activeSection ? "active" : ""} key={item.key} type="button" onClick={() => setSearchParams({ section: item.key })}>{item.label}</button>)}
        </nav>
      </section>
      <div className="config-list">
        {activeSection === "static"
          ? <StaticPageManager />
          : activeSection === "security"
            ? <AccessProtectionManager />
            : activeSection === "capture"
              ? <CaptureAISettingsManager onMessage={setMessage} />
            : visibleRows.map((row) => row.configKey === "home.modules" || row.configKey === "site.theme" ? <AdvancedConfigEditor key={row.configKey} row={row} onMessage={setMessage} /> : <ReadableConfigEditor key={row.configKey} row={row} onMessage={setMessage} />)}
      </div>
    </AdminLayout>
  );
}

function CaptureAISettingsManager({ onMessage }: { onMessage: (message: string) => void }) {
  const [settings, setSettings] = useState({ enabled: false, secretIdConfigured: false, secretKeyConfigured: false, tokenHubKeyConfigured: false, region: "ap-guangzhou", baseUrl: "https://tokenhub.tencentmaas.com/v1", visionModel: "hunyuan-t1-vision-20250916", environmentOverrides: [] as string[] });
  const [secretId, setSecretId] = useState("");
  const [secretKey, setSecretKey] = useState("");
  const [tokenHubKey, setTokenHubKey] = useState("");
  const [clearCredentials, setClearCredentials] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  useEffect(() => {
    void getCaptureAISettings().then(setSettings).catch((error) => onMessage(getApiErrorMessage(error, "智能采集配置加载失败"))).finally(() => setLoading(false));
  }, [onMessage]);
  const save = () => {
    setSaving(true);
    void updateCaptureAISettings({ enabled: settings.enabled, secretId, secretKey, tokenHubKey, region: settings.region, baseUrl: settings.baseUrl, visionModel: settings.visionModel, clearCredentials })
      .then((next) => {
        setSettings(next);
        setSecretId("");
        setSecretKey("");
        setTokenHubKey("");
        setClearCredentials(false);
        onMessage("智能采集云服务配置已保存并即时生效");
      })
      .catch((error) => onMessage(getApiErrorMessage(error, "智能采集配置保存失败")))
      .finally(() => setSaving(false));
  };
  if (loading) return <section className="config-card"><p>正在读取智能采集配置…</p></section>;
  return <section className="config-card capture-ai-settings-card">
    <header><div><strong>腾讯云 OCR 与 TokenHub 视觉</strong><small>OCR 与视觉服务使用不同密钥；所有密钥均加密保存且不会回显。</small></div><button className="primary-btn small" disabled={saving} type="button" onClick={save}>{saving ? "正在保存…" : "保存并生效"}</button></header>
    {settings.environmentOverrides.length > 0 && <p className="capture-ai-override-note">以下环境变量优先于页面配置：{settings.environmentOverrides.join("、")}</p>}
    <div className="config-field-grid">
      <label className="checkbox-field field-wide"><input checked={settings.enabled} type="checkbox" onChange={(event) => setSettings({ ...settings, enabled: event.target.checked })} />启用智能采集识别</label>
      <label>腾讯云 SecretId<input aria-label="腾讯云 SecretId" autoComplete="off" placeholder={settings.secretIdConfigured ? "已配置；留空表示不修改" : "请输入 SecretId"} type="password" value={secretId} onChange={(event) => { setSecretId(event.target.value); setClearCredentials(false); }} /><small>{settings.secretIdConfigured ? "当前已配置" : "当前未配置"}</small></label>
      <label>腾讯云 SecretKey<input aria-label="腾讯云 SecretKey" autoComplete="new-password" placeholder={settings.secretKeyConfigured ? "已配置；留空表示不修改" : "请输入 SecretKey"} type="password" value={secretKey} onChange={(event) => { setSecretKey(event.target.value); setClearCredentials(false); }} /><small>{settings.secretKeyConfigured ? "当前已配置" : "当前未配置"}</small></label>
      <label className="field-wide">TokenHub API Key<input aria-label="TokenHub API Key" autoComplete="new-password" placeholder={settings.tokenHubKeyConfigured ? "已配置；留空表示不修改" : "请输入 TokenHub API Key"} type="password" value={tokenHubKey} onChange={(event) => { setTokenHubKey(event.target.value); setClearCredentials(false); }} /><small>{settings.tokenHubKeyConfigured ? "当前已配置" : "当前未配置；需在 TokenHub 控制台创建"}</small></label>
      <ConfigInput label="腾讯云地域" value={settings.region} onChange={(region) => setSettings({ ...settings, region })} />
      <label>TokenHub 接入地址<select aria-label="TokenHub 接入地址" value={settings.baseUrl} onChange={(event) => setSettings({ ...settings, baseUrl: event.target.value })}><option value="https://tokenhub.tencentmaas.com/v1">广州（中国大陆）</option><option value="https://tokenhub-intl.tencentmaas.com/v1">新加坡（全球）</option></select></label>
      <ConfigInput label="视觉模型" value={settings.visionModel} onChange={(visionModel) => setSettings({ ...settings, visionModel })} />
      {(settings.secretIdConfigured || settings.secretKeyConfigured || settings.tokenHubKeyConfigured) && <label className="checkbox-field field-wide capture-ai-clear"><input checked={clearCredentials} type="checkbox" onChange={(event) => setClearCredentials(event.target.checked)} />清除后台已保存的 OCR 与 TokenHub 密钥</label>}
    </div>
    <p className="config-hint">保存时不会记录密钥、OCR 原文或联系方式。环境变量已设置的字段始终优先。</p>
  </section>;
}

function AdvancedConfigEditor({ row, onMessage }: { row: SiteConfig; onMessage: (message: string) => void }) {
  const initial = () => { try { return JSON.parse(row.configValue) as unknown; } catch { return row.configKey === "home.modules" ? [] : {}; } };
  const [value, setValue] = useState<unknown>(initial);
  const save = () => updateConfig(row.configKey, { ...row, configValue: JSON.stringify(value) }).then(() => onMessage("配置已保存")).catch(() => onMessage("配置保存失败，请检查字段"));
  if (row.configKey === "site.theme") {
    const theme = value as Record<string, unknown>;
    return <section className="config-card"><header><div><strong>站点主题</strong><small>统一控制前台产业主色与可信强调色。</small></div><button className="primary-btn small" type="button" onClick={save}>保存主题</button></header><div className="config-field-grid"><ConfigInput label="产业主色" value={theme.primaryColor} onChange={(next) => setValue({ ...theme, primaryColor: next })} /><ConfigInput label="可信强调色" value={theme.accentColor} onChange={(next) => setValue({ ...theme, accentColor: next })} /></div></section>;
  }
  const rows = Array.isArray(value) ? value as Array<Record<string, unknown>> : [];
  const update = (index: number, key: string, next: unknown) => setValue(rows.map((item, current) => current === index ? { ...item, [key]: next } : item));
  const names: Record<string, string> = { recommendedVendors: "优质厂商", processingServices: "加工服务", moreVendors: "更多厂商" };
  return <section className="config-card home-modules-editor"><header><div><strong>首页模块编排</strong><small>按排序值控制展示顺序，支持标题、图片、数量与开关。</small></div><button className="primary-btn small" type="button" onClick={save}>保存模块</button></header><div className="module-editor-list">{rows.map((item, index) => <article key={String(item.type)}><div className="module-editor-head"><strong>{names[String(item.type)] || String(item.type)}</strong><label className="checkbox-field"><input checked={Boolean(item.visible)} type="checkbox" onChange={(event) => update(index, "visible", event.target.checked)} />显示</label></div><div className="config-field-grid"><ConfigInput label="标题" value={item.title} onChange={(next) => update(index, "title", next)} /><ConfigInput label="副标题" value={item.subtitle} onChange={(next) => update(index, "subtitle", next)} /><ConfigInput label="跳转路径" value={item.path} onChange={(next) => update(index, "path", next)} /><ConfigInput label="背景图片" value={item.image} onChange={(next) => update(index, "image", next)} /><ConfigInput label="展示数量" type="number" value={item.limit} onChange={(next) => update(index, "limit", Number(next))} /><ConfigInput label="排序" type="number" value={item.sortOrder} onChange={(next) => update(index, "sortOrder", Number(next))} /></div></article>)}</div></section>;
}

function ReadableConfigEditor({ row, onMessage }: { row: SiteConfig; onMessage: (message: string) => void }) {
  const parse = () => { try { return JSON.parse(row.configValue) as Record<string, unknown> | unknown[]; } catch { return {}; } };
  const [value, setValue] = useState<Record<string, unknown> | unknown[]>(parse);
  const objectValue = Array.isArray(value) ? {} : value;
  const update = (key: string, next: unknown) => setValue({ ...objectValue, [key]: next });
  const save = () => updateConfig(row.configKey, { ...row, configValue: JSON.stringify(value) }).then(() => onMessage("配置已保存")).catch(() => onMessage("配置保存失败，请检查字段"));
  if (row.configKey === "home.stats") return <section className="config-card"><header><div><strong>首页真实统计</strong><small>{row.description}</small></div></header><p>厂商、产品、加工服务厂商和覆盖省份均从当前数据库实时聚合，无需手工填写。</p></section>;
  return <section className="config-card"><header><div><strong>{configLabel(row.configKey)}</strong><small>{row.description}</small></div><button className="primary-btn small" type="button" onClick={save}>保存配置</button></header><div className="config-field-grid">
    {row.configKey === "site.meta" && <><ConfigInput label="站点名称" value={objectValue.siteName} onChange={(next) => update("siteName", next)} /><BrandLogoConfigInput value={objectValue.brandLogo} onChange={(next) => update("brandLogo", next)} /><ConfigInput label="品牌标识文字（未上传 Logo 时显示）" value={objectValue.brandMark} onChange={(next) => update("brandMark", next)} /><p className="config-hint">前台入驻入口统一显示为“厂商入驻”。</p><ConfigInput label="后台入口" value={objectValue.adminLoginText} onChange={(next) => update("adminLoginText", next)} /><ConfigInput label="版权年份" value={objectValue.copyrightYear || String(new Date().getFullYear())} onChange={(next) => update("copyrightYear", next)} /><ConfigInput label="版权所有者" value={objectValue.copyrightOwner || objectValue.siteName} onChange={(next) => update("copyrightOwner", next)} /><ConfigInput label="备案号" value={objectValue.filingNumber || "待运营方配置"} onChange={(next) => update("filingNumber", next)} /><ConfigInput label="正式站点 URL" value={objectValue.siteUrl} onChange={(next) => update("siteUrl", next)} /><ConfigInput label="默认 SEO 标题" value={objectValue.defaultSeoTitle} onChange={(next) => update("defaultSeoTitle", next)} /><ConfigInput label="默认 SEO 摘要" value={objectValue.defaultSeoDescription} onChange={(next) => update("defaultSeoDescription", next)} /><ConfigInput label="百度站长验证码" value={objectValue.baiduVerification} onChange={(next) => update("baiduVerification", next)} /><ConfigInput label="Google 验证码" value={objectValue.googleVerification} onChange={(next) => update("googleVerification", next)} /></>}
    {row.configKey === "home.sections" && <><ConfigInput label="推荐区标题" value={objectValue.recommendedTitle} onChange={(next) => update("recommendedTitle", next)} /><ConfigInput label="更多区标题" value={objectValue.moreTitle} onChange={(next) => update("moreTitle", next)} /><ConfigInput label="推荐数量" type="number" value={objectValue.recommendedLimit} onChange={(next) => update("recommendedLimit", Number(next))} /><ConfigInput label="更多数量" type="number" value={objectValue.moreLimit} onChange={(next) => update("moreLimit", Number(next))} /><label className="checkbox-field"><input checked={Boolean(objectValue.showRecommended)} type="checkbox" onChange={(event) => update("showRecommended", event.target.checked)} />显示推荐厂商</label><label className="checkbox-field"><input checked={Boolean(objectValue.showMore)} type="checkbox" onChange={(event) => update("showMore", event.target.checked)} />显示更多厂商</label></>}
    {row.configKey === "home.join" && <><ConfigInput label="引导文案" value={objectValue.text} onChange={(next) => update("text", next)} /><ConfigInput label="按钮文案" value={objectValue.buttonText} onChange={(next) => update("buttonText", next)} /><ConfigInput label="跳转路径" value={objectValue.path} onChange={(next) => update("path", next)} /></>}
    {row.configKey === "home.safeguards" && <label className="field-wide">保障文案（每行一条）<textarea value={(Array.isArray(value) ? value : []).join("\n")} onChange={(event) => setValue(event.target.value.split("\n").map((item) => item.trim()).filter(Boolean))} /></label>}
  </div></section>;
}

function ConfigInput({ label, value, type = "text", onChange }: { label: string; value: unknown; type?: string; onChange: (value: string) => void }) {
  return <label>{label}<input type={type} value={String(value ?? "")} onChange={(event) => onChange(event.target.value)} /></label>;
}

function BrandLogoConfigInput({ value, onChange }: { value: unknown; onChange: (value: string) => void }) {
  const [downloading, setDownloading] = useState(false);
  const [error, setError] = useState("");
  const url = String(value ?? "");
  const remoteURL = /^https?:\/\//i.test(url);
  return <label className="field-wide brand-logo-config">品牌 Logo<small>上传透明背景的图标效果最佳；会显示在站点名称左侧。</small><input value={url} placeholder="上传后自动填写图片地址" type="text" onChange={(event) => { setError(""); onChange(event.target.value); }} /><div className="image-field-actions"><input aria-label="品牌 Logo 上传" type="file" accept="image/*" onChange={(event) => { const file = event.target.files?.[0]; if (!file) return; void uploadFile(file).then((result) => onChange(result.url)).catch(() => setError("图片上传失败，请重试")); }} /><button className="outline-btn small" disabled={!remoteURL || downloading} type="button" onClick={() => { if (!remoteURL || downloading) return; setDownloading(true); setError(""); void downloadRemoteImage(url).then((result) => onChange(result.url)).catch((requestError) => setError(requestError?.response?.data?.message || "下载失败，请确认图片地址可公开访问")).finally(() => setDownloading(false)); }}>{downloading ? "正在下载…" : "下载到本地"}</button></div>{error && <small className="image-download-error">{error}</small>}{url && <div className="brand-logo-config-preview"><img alt="品牌 Logo 预览" src={url} /><span>前台会在站点名称左侧展示此 Logo。</span></div>}</label>;
}

function defaultForm(fields: Field[]) {
  return Object.fromEntries(fields.map((field) => [field.key, field.type === "checkbox" ? false : field.type === "number" ? 0 : field.type === "multiselect" || field.type === "checkbox-group" ? [] : ""]));
}

function groupFields(resource: ResourceName, fields: Field[]) {
  const groupTitle = (key: string) => {
    if (resource === "vendors") {
		if (["name", "shortName", "slug", "province", "city", "county", "address", "description"].includes(key)) return "基础资料";
		if (["logo", "coverImage", "mainProducts", "serviceAdvantages", "tagIds", "vendorCategoryIds"].includes(key)) return "展示信息";
      if (["establishedYear", "factoryArea", "employeeCount", "annualCapacity", "equipment", "certifications"].includes(key)) return "生产与服务";
      if (key.startsWith("processing") || key === "providesProcessing") return "加工能力";
      if (["websiteUrl", "phone", "phonePublic", "wechat", "wechatQrCode", "wechatPublic", "contactName", "contactNamePublic"].includes(key)) return "联系方式";
      if (["seoTitle", "seoDescription", "seoTitleManual", "seoDescriptionManual"].includes(key)) return "SEO 优化建议";
      return "平台状态";
    }
    if (resource === "products") {
      if (["name", "categoryId", "compatibleModels", "description", "priceNote", "publicationStatus"].includes(key)) return "基础信息";
      if (["image", "galleryRaw"].includes(key)) return "展示图片";
      if (["detailContent", "specsRaw"].includes(key)) return "详情与规格";
      return "发布设置";
    }
    if (resource === "pages") return key === "blocksRaw" ? "内容区块" : "页面信息";
    return "记录信息";
  };
  const groups: Array<{ title: string; fields: Field[] }> = [];
  fields.forEach((field) => { const title = field.group || groupTitle(field.key); let group = groups.find((item) => item.title === title); if (!group) { group = { title, fields: [] }; groups.push(group); } group.fields.push(field); });
  return groups;
}

function payloadFromForm(fields: Field[], form: FormState) {
  return Object.fromEntries(fields.map((field) => [field.key, form[field.key]]));
}

function hierarchicalVendorCategoryOptions(rows: VendorCategory[]) {
	const roots = rows.filter((row) => !row.parentId).sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0) || a.id - b.id);
	return roots.flatMap((root) => [
		{ label: root.name, value: root.id },
		...rows.filter((row) => row.parentId === root.id).sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0) || a.id - b.id).map((child) => ({ label: `└ ${child.name}`, value: child.id })),
	]);
}

function columnLabel(key: string, resource?: ResourceName) {
	if (key === "slug" && resource === "vendors") return "厂商网站地址标识";
  return ({ id: "ID", name: "名称", title: "标题", slug: "URL 标识", province: "地区", menuType: "展示位置", path: "访问路径", publicationStatus: "发布状态", sortOrder: "排序", isEnabled: "启用", isVisible: "显示", isRecommended: "推荐" } as Record<string, string>)[key] || key;
}

function formatCell(value: unknown, key = "") {
  if (typeof value === "boolean") return value ? "是" : "否";
  if (key === "menuType") return ({ top: "顶部导航", auxiliary: "辅助入口", mobile: "移动快捷入口", mobile_bottom: "移动底部" } as Record<string, string>)[String(value)] || String(value || "");
  if (key === "publicationStatus") return ({ published: "已发布", draft: "草稿", hidden: "已隐藏", archived: "已归档", scheduled: "定时发布" } as Record<string, string>)[String(value)] || String(value || "");
  return String(value ?? "");
}

function configLabel(key: string) {
  return ({ "site.meta": "站点与页脚", "home.modules": "首页展示模块" } as Record<string, string>)[key] || key;
}
