import { FormEvent, useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Plus, X } from "lucide-react";
import {
  createResource,
  deleteResource,
  listConfigs,
  listResource,
  listResourcePage,
  ResourceName,
  ResourceRecord,
  updateConfig,
  updateResource,
  uploadFile,
} from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import type { SiteConfig } from "../../types/api";
import type { VendorMedia } from "../../types/api";
import { AdminModal } from "../../components/admin/AdminModal";
import { ProtectedMediaImage } from "../../components/admin/ProtectedMediaImage";
import { BlocksEditor, GalleryEditor, SpecsEditor, VendorMediaEditor } from "../../components/admin/StructuredEditors";
import { Pagination } from "../../components/public/Pagination";
import { AdminProductSuppliersEditor } from "../../components/admin/AdminProductSuppliersEditor";
import { AdminVendorProductsPanel } from "../../components/admin/AdminVendorProductsPanel";

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
};

const menuTypes = [
  { label: "顶部导航", value: "top" },
  { label: "左侧菜单", value: "sidebar" },
  { label: "辅助菜单", value: "auxiliary" },
  { label: "移动端宫格", value: "mobile" },
  { label: "移动端底部", value: "mobile_bottom" },
];

const reviewStatusOptions = [
  { label: "待人工复核", value: "pending" },
  { label: "已复核", value: "verified" },
  { label: "已驳回", value: "rejected" },
];

const schemas: Record<ResourceName, { title: string; fields: Field[] }> = {
  menus: {
    title: "导航菜单",
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
      { key: "shortName", label: "简称" },
      { key: "logo", label: "Logo URL", type: "image" },
      { key: "coverImage", label: "封面 URL", type: "image" },
      { key: "province", label: "省份" },
      { key: "city", label: "城市" },
      { key: "county", label: "区县" },
      { key: "address", label: "详细地址" },
      { key: "mainProducts", label: "主营产品", type: "textarea" },
      { key: "serviceModels", label: "适配机型", type: "textarea" },
      { key: "serviceAdvantages", label: "服务优势", type: "textarea" },
      { key: "description", label: "公司介绍", type: "textarea" },
      { key: "establishedYear", label: "成立年份" },
      { key: "factoryArea", label: "厂房面积" },
      { key: "employeeCount", label: "员工规模" },
      { key: "annualCapacity", label: "年产能", type: "textarea" },
      { key: "equipment", label: "主要设备", type: "textarea" },
      { key: "certifications", label: "认证资质", type: "textarea" },
      { key: "afterSalesService", label: "售后服务", type: "textarea" },
      { key: "reviewStatus", label: "复核状态", type: "select", options: reviewStatusOptions },
      { key: "websiteUrl", label: "厂商官网 URL" },
      { key: "phone", label: "联系电话" },
      { key: "wechat", label: "微信" },
      { key: "contactName", label: "联系人" },
      { key: "tagIds", label: "配件厂商标签", type: "checkbox-group", refResource: "tags", optionTagType: "vendor", group: "展示信息", description: "用于标识厂商经营与展示属性，可多选。" },
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
  products: {
    title: "配件产品",
    fields: [
      { key: "name", label: "产品名称" },
      { key: "image", label: "产品主图", type: "image" },
      { key: "categoryId", label: "所属分类", type: "select", refResource: "categories" },
      { key: "compatibleModels", label: "适配机型" },
      { key: "description", label: "列表描述", type: "textarea" },
      { key: "detailContent", label: "详情正文", type: "textarea" },
      { key: "galleryRaw", label: "产品图库", type: "gallery" },
      { key: "specsRaw", label: "规格参数", type: "specs" },
      { key: "priceNote", label: "价格说明" },
      { key: "inquiryText", label: "询价按钮文案" },
      { key: "inquiryPath", label: "询价跳转路径" },
      { key: "isHot", label: "热门", type: "checkbox" },
      { key: "isRecommended", label: "推荐", type: "checkbox" },
      { key: "status", label: "状态", type: "select", options: [{ label: "上架", value: 1 }, { label: "下架", value: 2 }] },
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
    title: "内容页面",
    fields: [
      { key: "slug", label: "页面标识" },
      { key: "title", label: "页面标题" },
      { key: "summary", label: "摘要", type: "textarea" },
      { key: "content", label: "正文", type: "textarea" },
      { key: "blocksRaw", label: "结构化内容", type: "blocks" },
      { key: "seoKeywords", label: "SEO 关键词" },
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
  { key: "providesProcessing", label: "是否提供加工服务", type: "checkbox" },
  { key: "tagIds", label: "加工服务标签", type: "checkbox-group", refResource: "tags", optionTagType: "processing", group: "加工能力", description: "用于标识可承接的加工方式，可多选。" },
  { key: "processingServices", label: "加工服务能力", type: "textarea" },
  { key: "processingMaterials", label: "可加工材料 / 配件类型", type: "textarea" },
  { key: "processingEquipment", label: "加工设备", type: "textarea" },
  { key: "processingCapacity", label: "产能 / 交期", type: "textarea" },
  { key: "processingRegions", label: "加工服务区域", type: "textarea" },
  { key: "processingNotes", label: "接单说明", type: "textarea" },
];

extendProcessingSchemas();

function extendProcessingSchemas() {
  const vendorFields = schemas.vendors.fields;
  if (!vendorFields.some((field) => field.key === "providesProcessing")) {
    const afterSalesIndex = vendorFields.findIndex((field) => field.key === "afterSalesService");
    vendorFields.splice(afterSalesIndex >= 0 ? afterSalesIndex + 1 : vendorFields.length, 0, ...processingVendorFields);
  }
  const tagTypeField = schemas.tags.fields.find((field) => field.key === "tagType");
  if (tagTypeField?.options && !tagTypeField.options.some((option) => option.value === "processing")) {
    tagTypeField.options.push({ label: "加工服务", value: "processing" });
  }
}

export function AdminResourcePage() {
  const { resource = "menus" } = useParams();
  if (resource === "configs") return <ConfigPage />;
  const name = (schemas[resource as ResourceName] ? resource : "menus") as ResourceName;
  const schema = schemas[name];
  const [rows, setRows] = useState<ResourceRecord[]>([]);
  const [refs, setRefs] = useState<Partial<Record<ResourceName, ResourceRecord[]>>>({});
  const [form, setForm] = useState<FormState>({});
  const [editingId, setEditingId] = useState<number | null>(null);
  const [keyword, setKeyword] = useState("");
	const [statusFilter, setStatusFilter] = useState("");
	const [provinceFilter, setProvinceFilter] = useState("");
  const [message, setMessage] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const serverPaged = name === "vendors" || name === "products";
	const debouncedKeyword = useDebouncedValue(keyword, 300);

  const load = () => {
    if (serverPaged) {
	  const filters = name === "vendors" ? { publicationStatus: statusFilter || undefined, province: provinceFilter || undefined } : { status: statusFilter || undefined };
      return listResourcePage<ResourceRecord>(name, { page, pageSize: 20, search: debouncedKeyword.trim() || undefined, ...filters }).then((result) => { setRows(result.items); setTotal(result.total); });
    }
    return listResource<ResourceRecord>(name).then((items) => { setRows(items); setTotal(items.length); });
  };
  useEffect(() => {
    setMessage("");
    setForm(defaultForm(schema.fields));
    setEditingId(null);
    setEditorOpen(false);
    setPage(1);
    void load();
    const resources = Array.from(new Set(schema.fields.map((field) => field.refResource).filter(Boolean))) as ResourceName[];
    if (resources.length) {
      void Promise.all(resources.map((resourceName) => listResource<ResourceRecord>(resourceName).then((items) => [resourceName, items] as const))).then((entries) => setRefs(Object.fromEntries(entries)));
    } else {
      setRefs({});
    }
  }, [name]);

  useEffect(() => {
    if (serverPaged) void load();
  }, [page, debouncedKeyword, statusFilter, provinceFilter]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");
    const payload = payloadFromForm(schema.fields, form) as Partial<ResourceRecord> & { media?: VendorMedia[] };
    if (name === "vendors") {
      payload.media = (form.media as VendorMedia[] | undefined) || [];
      (payload as { publicationStatus?: "published" | "hidden" }).publicationStatus = Boolean(form.isVisible) ? "published" : "hidden";
    }
    const action = editingId ? updateResource<ResourceRecord>(name, editingId, payload) : createResource<ResourceRecord>(name, payload);
    action
      .then(() => {
        setForm(defaultForm(schema.fields));
        setEditingId(null);
        setEditorOpen(false);
        setMessage("保存成功");
        void load();
      })
      .catch(() => setMessage("保存失败，请检查字段"));
  };

  const filteredRows = useMemo(() => {
    const value = keyword.trim().toLowerCase();
    if (!value) return rows;
    return rows.filter((row) => JSON.stringify(row).toLowerCase().includes(value));
  }, [keyword, rows]);

  return (
    <AdminLayout title={schema.title}>
      {message && <p className="admin-message">{message}</p>}
      <div className="admin-toolbar admin-panel compact">
        <input value={keyword} placeholder="搜索当前列表" onChange={(event) => setKeyword(event.target.value)} />
		{name === "vendors" && <><select aria-label="发布状态筛选" value={statusFilter} onChange={(event) => { setStatusFilter(event.target.value); setPage(1); }}><option value="">全部发布状态</option><option value="published">已发布</option><option value="draft">草稿</option><option value="hidden">已隐藏</option></select><input aria-label="地区筛选" placeholder="输入省份" value={provinceFilter} onChange={(event) => { setProvinceFilter(event.target.value); setPage(1); }} /></>}
		{name === "products" && <select aria-label="产品状态筛选" value={statusFilter} onChange={(event) => { setStatusFilter(event.target.value); setPage(1); }}><option value="">全部产品状态</option><option value="1">已上架</option><option value="2">已下架</option></select>}
        <button className="primary-btn" type="button" onClick={() => { setEditingId(null); setForm({ ...defaultForm(schema.fields), ...(name === "vendors" ? { media: [] } : {}) }); setEditorOpen(true); }}><Plus size={16} />新增{schema.title}</button>
      </div>
      <div className="admin-table-panel">
        <ResourceTable
          rows={filteredRows}
          onDelete={(id) => {
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
        />
        {serverPaged && <Pagination onChange={setPage} page={page} pageSize={20} total={total} />}
      </div>
      {editorOpen && <AdminModal label={`${editingId ? "编辑" : "新增"}${schema.title}`} onClose={() => setEditorOpen(false)}><header><div><span>{editingId ? "编辑记录" : "新增记录"}</span><h2>{schema.title}</h2></div><button aria-label="关闭编辑器" type="button" onClick={() => setEditorOpen(false)}><X size={20} /></button></header><form className="admin-form admin-grouped-form" onSubmit={submit}>
        {groupFields(name, schema.fields).map((group) => <fieldset key={group.title}><legend>{group.title}</legend><div className="admin-field-grid">{group.fields.map((field) => {
          const fieldWide = ["textarea", "gallery", "specs", "blocks", "checkbox-group"].includes(field.type || "");
          if (field.type === "checkbox-group") return <div className={`admin-checkbox-group-field ${fieldWide ? "field-wide" : ""}`} key={`${field.key}-${field.optionTagType || "all"}`}><strong>{field.label}</strong>{field.description && <small>{field.description}</small>}<FieldInput field={field} form={form} refs={refs} setForm={setForm} /></div>;
          if (field.type === "checkbox" && field.description) return <label className="admin-toggle-field" key={field.key}><FieldInput field={field} form={form} refs={refs} setForm={setForm} /><span><strong>{field.label}</strong><small>{field.description}</small></span></label>;
          return <label className={fieldWide ? "field-wide" : ""} key={field.key}>{field.label}<FieldInput field={field} form={form} refs={refs} setForm={setForm} /></label>;
        })}</div></fieldset>)}
        {name === "vendors" && <fieldset><legend>企业图集</legend><VendorMediaEditor value={(form.media as VendorMedia[] | undefined) || []} onChange={(media) => setForm({ ...form, media })} /></fieldset>}
        {(name === "vendors" || name === "products" || name === "banners" || name === "friend-links") && <ImagePreview form={form} />}
        <div className="admin-editor-actions"><button className="outline-btn" type="button" onClick={() => setEditorOpen(false)}>取消</button><button className="primary-btn" type="submit">{editingId ? "保存修改" : "创建记录"}</button></div>
      </form>{name === "products" && editingId && <AdminProductSuppliersEditor productId={editingId} />}{name === "vendors" && editingId && <AdminVendorProductsPanel vendorId={editingId} />}</AdminModal>}
    </AdminLayout>
  );
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
    return <textarea placeholder={field.placeholder} value={String(form[field.key] ?? "")} onChange={(event) => setForm({ ...form, [field.key]: event.target.value })} />;
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
  if (field.type === "image") {
    return (
      <div className="image-field">
        <input value={String(form[field.key] ?? "")} type="text" placeholder={field.placeholder} onChange={(event) => setForm({ ...form, [field.key]: event.target.value })} />
        <input
          aria-label={`${field.label} 上传`}
          type="file"
          accept="image/*"
          onChange={(event) => {
            const file = event.target.files?.[0];
            if (!file) return;
            const assetKey = field.key === "logo" ? "logoAssetId" : field.key === "coverImage" ? "coverAssetId" : undefined;
            uploadFile(file).then((result) => setForm({ ...form, [field.key]: result.url, ...(assetKey ? { [assetKey]: result.assetId } : {}) }));
          }}
        />
      </div>
    );
  }
  return (
    <input
      value={String(form[field.key] ?? "")}
      type={field.type || "text"}
      placeholder={field.placeholder}
      onChange={(event) => setForm({ ...form, [field.key]: field.type === "number" ? Number(event.target.value) : event.target.value })}
    />
  );
}

function optionsFor(field: Field, refs: Partial<Record<ResourceName, ResourceRecord[]>>) {
  if (field.options) return field.options;
  if (!field.refResource) return [];
  return (refs[field.refResource] || [])
    .filter((item) => !field.optionTagType || (item as unknown as { tagType?: string }).tagType === field.optionTagType)
    .map((item) => ({ label: String((item as unknown as { name?: string; title?: string }).name || (item as unknown as { title?: string }).title || item.id), value: item.id }));
}

function numericSelect(field: Field) {
  return Boolean(field.refResource || field.key.endsWith("Id") || field.key === "status" || field.key === "parentId");
}

function ResourceTable({ rows, onEdit, onDelete }: { rows: ResourceRecord[]; onEdit: (row: ResourceRecord) => void; onDelete: (id: number) => void }) {
  const keys = useMemo(() => Object.keys(rows[0] || {}).filter((key) => ["id", "name", "title", "slug", "province", "sortOrder", "isEnabled", "isVisible", "isRecommended"].includes(key)), [rows]);
  return (
    <table className="admin-table">
      <thead>
        <tr>
          {keys.map((key) => (
            <th key={key}>{columnLabel(key)}</th>
          ))}
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => (
          <tr key={row.id}>
            {keys.map((key) => (
              <td data-label={columnLabel(key)} key={key}>{formatCell((row as unknown as Record<string, unknown>)[key])}</td>
            ))}
            <td data-label="操作">
              <button type="button" onClick={() => onEdit(row)}>编辑</button>
              <button type="button" onClick={() => onDelete(row.id)}>删除</button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function ConfigPage() {
  const [rows, setRows] = useState<SiteConfig[]>([]);
  const [message, setMessage] = useState("");
  useEffect(() => {
    void listConfigs().then(setRows);
  }, []);
  return (
    <AdminLayout title="平台配置">
      {message && <p className="admin-message">{message}</p>}
      <div className="config-list">
        {rows.map((row) => row.configKey === "home.modules" || row.configKey === "site.theme" ? <AdvancedConfigEditor key={row.configKey} row={row} onMessage={setMessage} /> : <ReadableConfigEditor key={row.configKey} row={row} onMessage={setMessage} />)}
      </div>
    </AdminLayout>
  );
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
  const names: Record<string, string> = { categories: "热门品类", recommendedVendors: "优质厂商", featuredProducts: "热门产品", processingServices: "加工服务", moreVendors: "更多厂商", safeguards: "平台保障", join: "厂商入驻" };
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
    {row.configKey === "site.meta" && <><ConfigInput label="站点名称" value={objectValue.siteName} onChange={(next) => update("siteName", next)} /><ConfigInput label="品牌标识" value={objectValue.brandMark} onChange={(next) => update("brandMark", next)} /><ConfigInput label="提交厂商入口" value={objectValue.submitVendorText} onChange={(next) => update("submitVendorText", next)} /><ConfigInput label="后台入口" value={objectValue.adminLoginText} onChange={(next) => update("adminLoginText", next)} /></>}
    {row.configKey === "home.sections" && <><ConfigInput label="推荐区标题" value={objectValue.recommendedTitle} onChange={(next) => update("recommendedTitle", next)} /><ConfigInput label="更多区标题" value={objectValue.moreTitle} onChange={(next) => update("moreTitle", next)} /><ConfigInput label="推荐数量" type="number" value={objectValue.recommendedLimit} onChange={(next) => update("recommendedLimit", Number(next))} /><ConfigInput label="更多数量" type="number" value={objectValue.moreLimit} onChange={(next) => update("moreLimit", Number(next))} /><label className="checkbox-field"><input checked={Boolean(objectValue.showRecommended)} type="checkbox" onChange={(event) => update("showRecommended", event.target.checked)} />显示推荐厂商</label><label className="checkbox-field"><input checked={Boolean(objectValue.showMore)} type="checkbox" onChange={(event) => update("showMore", event.target.checked)} />显示更多厂商</label></>}
    {row.configKey === "home.join" && <><ConfigInput label="引导文案" value={objectValue.text} onChange={(next) => update("text", next)} /><ConfigInput label="按钮文案" value={objectValue.buttonText} onChange={(next) => update("buttonText", next)} /><ConfigInput label="跳转路径" value={objectValue.path} onChange={(next) => update("path", next)} /></>}
    {row.configKey === "home.safeguards" && <label className="field-wide">保障文案（每行一条）<textarea value={(Array.isArray(value) ? value : []).join("\n")} onChange={(event) => setValue(event.target.value.split("\n").map((item) => item.trim()).filter(Boolean))} /></label>}
  </div></section>;
}

function ConfigInput({ label, value, type = "text", onChange }: { label: string; value: unknown; type?: string; onChange: (value: string) => void }) {
  return <label>{label}<input type={type} value={String(value ?? "")} onChange={(event) => onChange(event.target.value)} /></label>;
}

function ImagePreview({ form }: { form: FormState }) {
  const src = String(form.logo || form.coverImage || form.image || form.backgroundImage || "");
  if (!src) return null;
  const match = src.match(/^\/api\/media\/(\d+)$/);
  return (
    <div className="image-preview">
      <span>图片预览</span>
      <ProtectedMediaImage alt="图片预览" assetId={match ? Number(match[1]) : undefined} src={src} />
    </div>
  );
}

function defaultForm(fields: Field[]) {
  return Object.fromEntries(fields.map((field) => [field.key, field.type === "checkbox" ? false : field.type === "number" ? 0 : field.type === "multiselect" || field.type === "checkbox-group" ? [] : ""]));
}

function groupFields(resource: ResourceName, fields: Field[]) {
  const groupTitle = (key: string) => {
    if (resource === "vendors") {
      if (["name", "shortName", "province", "city", "county", "address", "description"].includes(key)) return "基础资料";
      if (["logo", "coverImage", "mainProducts", "serviceModels", "serviceAdvantages", "tagIds"].includes(key)) return "展示信息";
      if (["establishedYear", "factoryArea", "employeeCount", "annualCapacity", "equipment", "certifications", "afterSalesService"].includes(key)) return "生产与服务";
      if (key.startsWith("processing") || key === "providesProcessing") return "加工能力";
      if (["websiteUrl", "phone", "wechat", "contactName"].includes(key)) return "联系方式";
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

function columnLabel(key: string) {
  return ({ id: "ID", name: "名称", title: "标题", slug: "标识", province: "地区", sortOrder: "排序", isEnabled: "启用", isVisible: "显示", isRecommended: "推荐" } as Record<string, string>)[key] || key;
}

function formatCell(value: unknown) {
  if (typeof value === "boolean") return value ? "是" : "否";
  return String(value ?? "");
}

function configLabel(key: string) {
  return ({ "site.meta": "站点基础信息", "home.sections": "首页模块配置", "home.stats": "首页统计", "home.safeguards": "底部保障文案", "home.join": "入驻引导" } as Record<string, string>)[key] || key;
}
