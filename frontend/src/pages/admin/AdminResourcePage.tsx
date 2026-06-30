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
} from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import type { SiteConfig } from "../../types/api";

type Field = {
  key: string;
  label: string;
  type?: "text" | "number" | "checkbox" | "textarea" | "select";
  options?: { label: string; value: string | number | boolean }[];
  placeholder?: string;
};

const menuTypes = [
  { label: "顶部导航", value: "top" },
  { label: "左侧菜单", value: "sidebar" },
  { label: "辅助菜单", value: "auxiliary" },
  { label: "移动端宫格", value: "mobile" },
];

const schemas: Record<ResourceName, { title: string; fields: Field[] }> = {
  menus: {
    title: "导航菜单",
    fields: [
      { key: "name", label: "菜单名称" },
      { key: "parentId", label: "父级菜单 ID", type: "number" },
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
      { key: "logo", label: "Logo URL" },
      { key: "coverImage", label: "封面 URL" },
      { key: "province", label: "省份" },
      { key: "city", label: "城市" },
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
      { key: "websiteUrl", label: "厂商官网 URL" },
      { key: "phone", label: "联系电话" },
      { key: "wechat", label: "微信" },
      { key: "contactName", label: "联系人" },
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
      { key: "parentId", label: "父级分类 ID", type: "number" },
      { key: "icon", label: "图标" },
      { key: "sortOrder", label: "排序", type: "number" },
      { key: "isEnabled", label: "启用", type: "checkbox" },
    ],
  },
  products: {
    title: "配件产品",
    fields: [
      { key: "name", label: "产品名称" },
      { key: "image", label: "图片 URL" },
      { key: "categoryId", label: "分类 ID", type: "number" },
      { key: "vendorId", label: "厂商 ID", type: "number" },
      { key: "compatibleModels", label: "适配机型" },
      { key: "description", label: "描述", type: "textarea" },
      { key: "isHot", label: "热门", type: "checkbox" },
      { key: "isRecommended", label: "推荐", type: "checkbox" },
      { key: "status", label: "状态", type: "number" },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
  banners: {
    title: "Banner 管理",
    fields: [
      { key: "title", label: "标题" },
      { key: "subtitle", label: "副标题", type: "textarea" },
      { key: "backgroundImage", label: "背景图 URL" },
      { key: "searchPlaceholder", label: "搜索占位文案" },
      { key: "hotKeywordsRaw", label: "热门关键词", placeholder: "用英文逗号分隔" },
      { key: "isEnabled", label: "启用", type: "checkbox" },
      { key: "sortOrder", label: "排序", type: "number" },
    ],
  },
};

export function AdminResourcePage() {
  const { resource = "menus" } = useParams();
  if (resource === "configs") return <ConfigPage />;
  const name = (schemas[resource as ResourceName] ? resource : "menus") as ResourceName;
  const schema = schemas[name];
  const [rows, setRows] = useState<ResourceRecord[]>([]);
  const [form, setForm] = useState<Record<string, string | number | boolean>>({});
  const [editingId, setEditingId] = useState<number | null>(null);
  const [keyword, setKeyword] = useState("");
  const [message, setMessage] = useState("");

  const load = () => listResource<ResourceRecord>(name).then(setRows);
  useEffect(() => {
    setMessage("");
    setForm(defaultForm(schema.fields));
    setEditingId(null);
    void load();
  }, [name]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");
    const action = editingId ? updateResource<ResourceRecord>(name, editingId, form) : createResource<ResourceRecord>(name, form);
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
            <FieldInput field={field} form={form} setForm={setForm} />
          </label>
        ))}
        {(name === "vendors" || name === "products" || name === "banners") && <ImagePreview form={form} />}
        <button className="primary-btn" type="submit">
          {editingId ? "保存修改" : "新增"}
        </button>
        {editingId && (
          <button className="outline-btn" type="button" onClick={() => { setEditingId(null); setForm(defaultForm(schema.fields)); }}>
            取消编辑
          </button>
        )}
      </form>
      <div className="admin-toolbar">
        <input value={keyword} placeholder="搜索当前列表" onChange={(event) => setKeyword(event.target.value)} />
      </div>
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
          setForm(row as unknown as Record<string, string | number | boolean>);
        }}
      />
    </AdminLayout>
  );
}

function FieldInput({ field, form, setForm }: { field: Field; form: Record<string, string | number | boolean>; setForm: (form: Record<string, string | number | boolean>) => void }) {
  if (field.type === "textarea") {
    return <textarea value={String(form[field.key] ?? "")} onChange={(event) => setForm({ ...form, [field.key]: event.target.value })} />;
  }
  if (field.type === "checkbox") {
    return <input checked={Boolean(form[field.key])} type="checkbox" onChange={(event) => setForm({ ...form, [field.key]: event.target.checked })} />;
  }
  if (field.type === "select") {
    return (
      <select value={String(form[field.key] ?? "")} onChange={(event) => setForm({ ...form, [field.key]: event.target.value })}>
        <option value="">请选择</option>
        {field.options?.map((option) => (
          <option key={String(option.value)} value={String(option.value)}>
            {option.label}
          </option>
        ))}
      </select>
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

function ResourceTable({ rows, onEdit, onDelete }: { rows: ResourceRecord[]; onEdit: (row: ResourceRecord) => void; onDelete: (id: number) => void }) {
  const keys = useMemo(() => Object.keys(rows[0] || {}).filter((key) => ["id", "name", "title", "province", "sortOrder", "isEnabled", "isVisible", "isRecommended"].includes(key)), [rows]);
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
                  .catch(() => setMessage("配置保存失败"))
              }
            />
          </label>
        ))}
      </div>
    </AdminLayout>
  );
}

function ImagePreview({ form }: { form: Record<string, string | number | boolean> }) {
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
  return Object.fromEntries(fields.map((field) => [field.key, field.type === "checkbox" ? false : field.type === "number" ? 0 : ""]));
}

function columnLabel(key: string) {
  return ({ id: "ID", name: "名称", title: "标题", province: "地区", sortOrder: "排序", isEnabled: "启用", isVisible: "显示", isRecommended: "推荐" } as Record<string, string>)[key] || key;
}

function formatCell(value: unknown) {
  if (typeof value === "boolean") return value ? "是" : "否";
  return String(value ?? "");
}

function configLabel(key: string) {
  return ({ "home.stats": "首页统计", "home.safeguards": "底部保障文案", "home.join": "入驻引导" } as Record<string, string>)[key] || key;
}
