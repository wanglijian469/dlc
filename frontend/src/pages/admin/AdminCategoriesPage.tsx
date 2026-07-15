import { type FormEvent, useEffect, useMemo, useState } from "react";
import { Plus, X } from "lucide-react";
import { createResource, deleteResource, listResource, updateResource } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { AdminModal } from "../../components/admin/AdminModal";
import type { Category } from "../../types/api";

type LevelFilter = "all" | "root" | "child";
type CategoryForm = {
  name: string;
  level: "root" | "child";
  parentId: number;
  icon: string;
  seoTitle: string;
  seoDescription: string;
  sortOrder: number;
  isEnabled: boolean;
};

const emptyForm: CategoryForm = { name: "", level: "root", parentId: 0, icon: "", seoTitle: "", seoDescription: "", sortOrder: 0, isEnabled: true };

export function AdminCategoriesPage() {
  const [rows, setRows] = useState<Category[]>([]);
  const [filter, setFilter] = useState<LevelFilter>("all");
  const [message, setMessage] = useState("");
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<CategoryForm>(emptyForm);
  const [editorOpen, setEditorOpen] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = () => listResource<Category & { id: number }>("categories").then(setRows).catch((error) => setMessage(getApiErrorMessage(error, "分类列表加载失败")));
  useEffect(() => { void load(); }, []);

  const roots = useMemo(() => sortCategories(rows.filter((row) => !row.parentId)), [rows]);
  const rootNames = useMemo(() => new Map(roots.map((row) => [row.id, row.name])), [roots]);
  const displayedRows = useMemo(() => {
    if (filter === "root") return roots;
    if (filter === "child") return sortCategories(rows.filter((row) => Boolean(row.parentId)));
    const children = new Map<number, Category[]>();
    rows.filter((row) => Boolean(row.parentId)).forEach((row) => children.set(row.parentId || 0, [...(children.get(row.parentId || 0) || []), row]));
    return roots.flatMap((root) => [root, ...sortCategories(children.get(root.id) || [])]);
  }, [filter, roots, rows]);

  const openCreate = (level: "root" | "child", parentId = 0) => {
    setEditingId(null);
    setForm({ ...emptyForm, level, parentId: level === "child" ? parentId : 0 });
    setMessage("");
    setEditorOpen(true);
  };
  const openEdit = (category: Category) => {
    setEditingId(category.id);
    setForm({
      name: category.name,
      level: category.parentId ? "child" : "root",
      parentId: category.parentId || 0,
      icon: category.icon || "",
      seoTitle: category.seoTitle || "",
      seoDescription: category.seoDescription || "",
      sortOrder: category.sortOrder || 0,
      isEnabled: category.isEnabled !== false,
    });
    setMessage("");
    setEditorOpen(true);
  };
  const submit = (event: FormEvent) => {
    event.preventDefault();
    if (form.level === "child" && !form.parentId) {
      setMessage("二级分类必须选择一级分类");
      return;
    }
    setSaving(true);
    setMessage("");
    const payload: Partial<Category> = { ...form, parentId: form.level === "root" ? 0 : form.parentId };
    delete (payload as Partial<Category> & { level?: string }).level;
    const request = editingId ? updateResource("categories", editingId, payload) : createResource("categories", payload);
    void request.then(() => {
      setEditorOpen(false);
      setMessage(editingId ? "分类修改成功" : "分类创建成功，刷新前台后菜单将自动更新");
      void load();
    }).catch((error) => setMessage(getApiErrorMessage(error, "分类保存失败"))).finally(() => setSaving(false));
  };

  return (
    <AdminLayout title="配件分类">
      {message && !editorOpen && <p className="admin-message">{message}</p>}
      <div className="admin-toolbar admin-panel compact category-toolbar">
        <select aria-label="分类级别筛选" value={filter} onChange={(event) => setFilter(event.target.value as LevelFilter)}>
          <option value="all">全部分类</option>
          <option value="root">一级分类</option>
          <option value="child">二级分类</option>
        </select>
        <div className="category-admin-actions">
          <button className="outline-btn" type="button" onClick={() => openCreate("child")}><Plus size={16} />新增二级分类</button>
          <button className="primary-btn" type="button" onClick={() => openCreate("root")}><Plus size={16} />新增一级分类</button>
        </div>
      </div>
      <div className="admin-table-panel">
        <table className="admin-table category-tree-table">
          <thead><tr><th>分类名称</th><th>级别</th><th>父级分类</th><th>图标</th><th>排序</th><th>启用状态</th><th>操作</th></tr></thead>
          <tbody>
            {displayedRows.map((category) => {
              const isChild = Boolean(category.parentId);
              return <tr className={isChild ? "category-child-row" : ""} key={category.id}>
                <td data-label="分类名称"><span className="category-tree-name">{isChild && <span aria-hidden="true">└</span>}{category.name}</span></td>
                <td data-label="级别"><span className={`category-level-badge ${isChild ? "child" : "root"}`}>{isChild ? "二级分类" : "一级分类"}</span></td>
                <td data-label="父级分类">{isChild ? rootNames.get(category.parentId || 0) || "父级已删除" : "—"}</td>
                <td data-label="图标">{category.icon || "—"}</td>
                <td data-label="排序">{category.sortOrder || 0}</td>
                <td data-label="启用状态">{category.isEnabled === false ? "已停用" : "已启用"}</td>
                <td data-label="操作"><div className="category-row-actions">{!isChild && <button type="button" onClick={() => openCreate("child", category.id)}>新增子分类</button>}<button type="button" onClick={() => openEdit(category)}>编辑</button><button type="button" onClick={() => { if (!window.confirm(`确认删除分类“${category.name}”吗？`)) return; void deleteResource("categories", category.id).then(() => { setMessage("分类删除成功"); void load(); }).catch((error) => setMessage(getApiErrorMessage(error, "分类删除失败"))); }}>删除</button></div></td>
              </tr>;
            })}
            {!displayedRows.length && <tr><td colSpan={7}>暂无符合条件的分类</td></tr>}
          </tbody>
        </table>
      </div>
      {editorOpen && <AdminModal label={`${editingId ? "编辑" : "新增"}配件分类`} onClose={() => setEditorOpen(false)}>
        <header><div><span>{editingId ? "编辑分类" : "创建分类"}</span><h2>{form.level === "root" ? "一级分类" : "二级分类"}</h2></div><button aria-label="关闭分类编辑器" type="button" onClick={() => setEditorOpen(false)}><X size={20} /></button></header>
        {message && <p className="admin-message category-editor-message">{message}</p>}
        <form className="admin-form admin-grouped-form" onSubmit={submit}><fieldset><legend>分类信息</legend><div className="admin-field-grid">
          <label>分类级别<select aria-label="分类级别" value={form.level} onChange={(event) => setForm({ ...form, level: event.target.value as CategoryForm["level"], parentId: event.target.value === "root" ? 0 : form.parentId })}><option value="root">一级分类</option><option value="child">二级分类</option></select></label>
          {form.level === "child" && <label>所属一级分类<select aria-label="所属一级分类" required value={form.parentId || ""} onChange={(event) => setForm({ ...form, parentId: Number(event.target.value) })}><option value="">请选择一级分类</option>{roots.filter((root) => root.id !== editingId).map((root) => <option key={root.id} value={root.id}>{root.name}</option>)}</select></label>}
          <label>分类名称<input required value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} /></label>
          <label>图标<input placeholder="图标名称或图标地址" value={form.icon} onChange={(event) => setForm({ ...form, icon: event.target.value })} /></label>
          <label>SEO 标题<input value={form.seoTitle} onChange={(event) => setForm({ ...form, seoTitle: event.target.value })} /></label>
          <label className="field-wide">SEO 摘要<textarea value={form.seoDescription} onChange={(event) => setForm({ ...form, seoDescription: event.target.value })} /></label>
          <label>排序<input type="number" value={form.sortOrder} onChange={(event) => setForm({ ...form, sortOrder: Number(event.target.value) })} /></label>
          <label className="admin-toggle-field"><input aria-label="启用分类" checked={form.isEnabled} type="checkbox" onChange={(event) => setForm({ ...form, isEnabled: event.target.checked })} /><span><strong>启用分类</strong><small>停用一级分类会同时从前台隐藏其全部二级分类</small></span></label>
        </div></fieldset><div className="admin-editor-actions"><button className="outline-btn" type="button" onClick={() => setEditorOpen(false)}>取消</button><button className="primary-btn" disabled={saving} type="submit">{saving ? "保存中…" : "保存分类"}</button></div></form>
      </AdminModal>}
    </AdminLayout>
  );
}

function sortCategories(rows: Category[]) {
  return [...rows].sort((left, right) => (left.sortOrder || 0) - (right.sortOrder || 0) || left.id - right.id);
}
