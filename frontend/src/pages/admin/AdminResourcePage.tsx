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

type Field = { key: string; label: string; type?: "text" | "number" | "checkbox" | "textarea" };

const schemas: Record<ResourceName, { title: string; fields: Field[] }> = {
  menus: {
    title: "Menus",
    fields: [
      { key: "name", label: "Name" },
      { key: "parentId", label: "Parent ID", type: "number" },
      { key: "icon", label: "Icon" },
      { key: "menuType", label: "Type" },
      { key: "path", label: "Path" },
      { key: "sortOrder", label: "Sort", type: "number" },
      { key: "isEnabled", label: "Enabled", type: "checkbox" },
      { key: "isDefaultOpen", label: "Default open", type: "checkbox" },
    ],
  },
  vendors: {
    title: "Vendors",
    fields: [
      { key: "name", label: "Name" },
      { key: "logo", label: "Logo URL" },
      { key: "coverImage", label: "Cover URL" },
      { key: "province", label: "Province" },
      { key: "city", label: "City" },
      { key: "mainProducts", label: "Main products", type: "textarea" },
      { key: "serviceAdvantages", label: "Advantages", type: "textarea" },
      { key: "websiteUrl", label: "Website" },
      { key: "isRecommended", label: "Recommended", type: "checkbox" },
      { key: "isVisible", label: "Visible", type: "checkbox" },
      { key: "sortOrder", label: "Sort", type: "number" },
    ],
  },
  tags: {
    title: "Tags",
    fields: [
      { key: "name", label: "Name" },
      { key: "tagType", label: "Type" },
      { key: "color", label: "Color" },
      { key: "sortOrder", label: "Sort", type: "number" },
    ],
  },
  categories: {
    title: "Categories",
    fields: [
      { key: "name", label: "Name" },
      { key: "parentId", label: "Parent ID", type: "number" },
      { key: "icon", label: "Icon" },
      { key: "sortOrder", label: "Sort", type: "number" },
      { key: "isEnabled", label: "Enabled", type: "checkbox" },
    ],
  },
  products: {
    title: "Products",
    fields: [
      { key: "name", label: "Name" },
      { key: "image", label: "Image URL" },
      { key: "categoryId", label: "Category ID", type: "number" },
      { key: "vendorId", label: "Vendor ID", type: "number" },
      { key: "compatibleModels", label: "Models" },
      { key: "description", label: "Description", type: "textarea" },
      { key: "isHot", label: "Hot", type: "checkbox" },
      { key: "isRecommended", label: "Recommended", type: "checkbox" },
      { key: "status", label: "Status", type: "number" },
      { key: "sortOrder", label: "Sort", type: "number" },
    ],
  },
  banners: {
    title: "Banners",
    fields: [
      { key: "title", label: "Title" },
      { key: "subtitle", label: "Subtitle", type: "textarea" },
      { key: "backgroundImage", label: "Background URL" },
      { key: "searchPlaceholder", label: "Search placeholder" },
      { key: "hotKeywordsRaw", label: "Hot keywords" },
      { key: "isEnabled", label: "Enabled", type: "checkbox" },
      { key: "sortOrder", label: "Sort", type: "number" },
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

  const load = () => listResource<ResourceRecord>(name).then(setRows);
  useEffect(() => {
    void load();
  }, [name]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const action = editingId ? updateResource<ResourceRecord>(name, editingId, form) : createResource<ResourceRecord>(name, form);
    action.then(() => {
      setForm({});
      setEditingId(null);
      void load();
    });
  };

  return (
    <AdminLayout title={schema.title}>
      <form className="admin-form" onSubmit={submit}>
        {schema.fields.map((field) => (
          <label key={field.key}>
            {field.label}
            {field.type === "textarea" ? (
              <textarea value={String(form[field.key] ?? "")} onChange={(event) => setForm({ ...form, [field.key]: event.target.value })} />
            ) : field.type === "checkbox" ? (
              <input checked={Boolean(form[field.key])} type="checkbox" onChange={(event) => setForm({ ...form, [field.key]: event.target.checked })} />
            ) : (
              <input
                value={String(form[field.key] ?? "")}
                type={field.type || "text"}
                onChange={(event) => setForm({ ...form, [field.key]: field.type === "number" ? Number(event.target.value) : event.target.value })}
              />
            )}
          </label>
        ))}
        <button className="primary-btn" type="submit">
          {editingId ? "Save" : "Create"}
        </button>
      </form>
      <ResourceTable
        rows={rows}
        onDelete={(id) => deleteResource(name, id).then(() => void load())}
        onEdit={(row) => {
          setEditingId(row.id);
          setForm(row as unknown as Record<string, string | number | boolean>);
        }}
      />
    </AdminLayout>
  );
}

function ResourceTable({ rows, onEdit, onDelete }: { rows: ResourceRecord[]; onEdit: (row: ResourceRecord) => void; onDelete: (id: number) => void }) {
  const keys = useMemo(() => Object.keys(rows[0] || {}).filter((key) => ["id", "name", "title", "sortOrder", "isEnabled", "isVisible"].includes(key)), [rows]);
  return (
    <table className="admin-table">
      <thead>
        <tr>
          {keys.map((key) => (
            <th key={key}>{key}</th>
          ))}
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => (
          <tr key={row.id}>
            {keys.map((key) => (
              <td key={key}>{String((row as unknown as Record<string, unknown>)[key] ?? "")}</td>
            ))}
            <td>
              <button type="button" onClick={() => onEdit(row)}>Edit</button>
              <button type="button" onClick={() => onDelete(row.id)}>Delete</button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function ConfigPage() {
  const [rows, setRows] = useState<SiteConfig[]>([]);
  useEffect(() => {
    void listConfigs().then(setRows);
  }, []);
  return (
    <AdminLayout title="Configs">
      <div className="config-list">
        {rows.map((row) => (
          <label key={row.configKey}>
            {row.configKey}
            <textarea defaultValue={row.configValue} onBlur={(event) => updateConfig(row.configKey, { ...row, configValue: event.target.value })} />
          </label>
        ))}
      </div>
    </AdminLayout>
  );
}
