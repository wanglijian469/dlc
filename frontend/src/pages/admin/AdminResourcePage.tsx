import { FormEvent, useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
  createResource,
  deleteResource,
  listConfigs,
  listResource,
  ResourceName,
  ResourceRecord,
  updateConfig,
  updateResource,
  uploadFile,
} from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import type { SiteConfig } from "../../types/api";

type FormValue = string | number | boolean | number[];
type FormState = Record<string, FormValue>;

type Field = {
  key: string;
  label: string;
  type?: "text" | "number" | "checkbox" | "textarea" | "select" | "multiselect" | "image";
  options?: { label: string; value: string | number | boolean }[];
  refResource?: ResourceName;
  placeholder?: string;
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
      { key: "qualityControl", label: "质检能力", type: "textarea" },
      { key: "supplyRegions", label: "供货区域", type: "textarea" },
      { key: "cooperationTerms", label: "合作方式", type: "textarea" },
      { key: "afterSalesService", label: "售后服务", type: "textarea" },
      { key: "sourceUrl", label: "公开信息来源 URL" },
      { key: "sourceNote", label: "采集备注", type: "textarea" },
      { key: "reviewStatus", label: "复核状态", type: "select", options: reviewStatusOptions },
      { key: "websiteUrl", label: "厂商官网 URL" },
      { key: "phone", label: "联系电话" },
      { key: "wechat", label: "微信" },
      { key: "contactName", label: "联系人" },
      { key: "tagIds", label: "厂商标签", type: "multiselect", refResource: "tags" },
      { key: "isRecommended", label: "推荐", type: "checkbox" },
      { key: "isVerified", label: "认证", type: "checkbox" },
      { key: "isVisible", label: "前台显示", type: "checkbox" },
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
      { key: "image", label: "图片 URL", type: "image" },
      { key: "categoryId", label: "所属分类", type: "select", refResource: "categories" },
      { key: "vendorId", label: "所属厂商", type: "select", refResource: "vendors" },
      { key: "compatibleModels", label: "适配机型" },
      { key: "description", label: "列表描述", type: "textarea" },
      { key: "detailContent", label: "详情正文", type: "textarea" },
      { key: "galleryRaw", label: "图库 JSON", type: "textarea", placeholder: '["/uploads/a.jpg"]' },
      { key: "specsRaw", label: "规格参数 JSON", type: "textarea", placeholder: '[{"name":"质保","value":"12个月"}]' },
      { key: "priceNote", label: "价格说明" },
      { key: "inquiryText", label: "询价按钮文案" },
      { key: "inquiryPath", label: "询价跳转路径" },
      { key: "isHot", label: "热门", type: "checkbox" },
      { key: "isRecommended", label: "推荐", type: "checkbox" },
      { key: "status", label: "状态", type: "select", options: [{ label: "上架", value: 1 }, { label: "下架", value: 2 }] },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
  banners: {
    title: "Banner 管理",
    fields: [
      { key: "title", label: "标题" },
      { key: "subtitle", label: "副标题", type: "textarea" },
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
  const [message, setMessage] = useState("");

  const load = () => listResource<ResourceRecord>(name).then(setRows);
  useEffect(() => {
    setMessage("");
    setForm(defaultForm(schema.fields));
    setEditingId(null);
    void load();
    const resources = Array.from(new Set(schema.fields.map((field) => field.refResource).filter(Boolean))) as ResourceName[];
    if (resources.length) {
      void Promise.all(resources.map((resourceName) => listResource<ResourceRecord>(resourceName).then((items) => [resourceName, items] as const))).then((entries) => setRefs(Object.fromEntries(entries)));
    } else {
      setRefs({});
    }
  }, [name]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");
    const payload = payloadFromForm(schema.fields, form) as Partial<ResourceRecord>;
    const action = editingId ? updateResource<ResourceRecord>(name, editingId, payload) : createResource<ResourceRecord>(name, payload);
    action
      .then(() => {
        setForm(defaultForm(schema.fields));
        setEditingId(null);
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
      <form className="admin-form" onSubmit={submit}>
        {schema.fields.map((field) => (
          <label key={field.key}>
            {field.label}
            <FieldInput field={field} form={form} refs={refs} setForm={setForm} />
          </label>
        ))}
        {(name === "vendors" || name === "products" || name === "banners" || name === "friend-links") && <ImagePreview form={form} />}
        <button className="primary-btn" type="submit">
          {editingId ? "保存修改" : "新增"}
        </button>
        {editingId && (
          <button className="outline-btn" type="button" onClick={() => { setEditingId(null); setForm(defaultForm(schema.fields)); }}>
            取消编辑
          </button>
        )}
      </form>
      <div className="admin-toolbar admin-panel compact">
        <input value={keyword} placeholder="搜索当前列表" onChange={(event) => setKeyword(event.target.value)} />
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
          }}
        />
      </div>
    </AdminLayout>
  );
}

function FieldInput({ field, form, refs, setForm }: { field: Field; form: FormState; refs: Partial<Record<ResourceName, ResourceRecord[]>>; setForm: (form: FormState) => void }) {
  if (field.type === "textarea") {
    return <textarea placeholder={field.placeholder} value={String(form[field.key] ?? "")} onChange={(event) => setForm({ ...form, [field.key]: event.target.value })} />;
  }
  if (field.type === "checkbox") {
    return <input checked={Boolean(form[field.key])} type="checkbox" onChange={(event) => setForm({ ...form, [field.key]: event.target.checked })} />;
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
            uploadFile(file).then((result) => setForm({ ...form, [field.key]: result.url }));
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
  return (refs[field.refResource] || []).map((item) => ({ label: String((item as unknown as { name?: string; title?: string }).name || (item as unknown as { title?: string }).title || item.id), value: item.id }));
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
              <td key={key}>{formatCell((row as unknown as Record<string, unknown>)[key])}</td>
            ))}
            <td>
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
        {rows.map((row) => (
          <label key={row.configKey}>
            <strong>{configLabel(row.configKey)}</strong>
            <small>{row.description}</small>
            <textarea
              defaultValue={row.configValue}
              onBlur={(event) =>
                updateConfig(row.configKey, { ...row, configValue: event.target.value })
                  .then(() => setMessage("配置已保存"))
                  .catch(() => setMessage("配置保存失败，请检查 JSON 格式"))
              }
            />
          </label>
        ))}
      </div>
    </AdminLayout>
  );
}

function ImagePreview({ form }: { form: FormState }) {
  const src = String(form.logo || form.coverImage || form.image || form.backgroundImage || "");
  if (!src) return null;
  return (
    <div className="image-preview">
      <span>图片预览</span>
      <img alt="" src={src} />
    </div>
  );
}

function defaultForm(fields: Field[]) {
  return Object.fromEntries(fields.map((field) => [field.key, field.type === "checkbox" ? false : field.type === "number" ? 0 : field.type === "multiselect" ? [] : ""]));
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
