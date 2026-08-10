import { type FormEvent, useEffect, useMemo, useState } from "react";
import { Plus, X } from "lucide-react";
import { createResource, deleteResource, listResource, updateResource } from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { AdminModal } from "../../components/admin/AdminModal";
import type { VendorCategory } from "../../types/api";

type CategoryForm = { name: string; level: "root" | "child"; parentId: number; icon: string; sortOrder: number; isEnabled: boolean };
const emptyForm: CategoryForm = { name: "", level: "root", parentId: 0, icon: "", sortOrder: 0, isEnabled: true };

export function AdminVendorCategoriesPage() {
	const [rows, setRows] = useState<VendorCategory[]>([]);
	const [message, setMessage] = useState("");
	const [editingId, setEditingId] = useState<number | null>(null);
	const [form, setForm] = useState<CategoryForm>(emptyForm);
	const [open, setOpen] = useState(false);
	const [saving, setSaving] = useState(false);
	const load = () => listResource<VendorCategory & { id: number }>("vendor-categories").then(setRows).catch((error) => setMessage(getApiErrorMessage(error, "厂商分类加载失败")));
	useEffect(() => { void load(); }, []);
	const roots = useMemo(() => sortRows(rows.filter((row) => !row.parentId)), [rows]);
	const displayed = useMemo(() => roots.flatMap((root) => [root, ...sortRows(rows.filter((row) => row.parentId === root.id))]), [roots, rows]);
	const rootNames = useMemo(() => new Map(roots.map((row) => [row.id, row.name])), [roots]);
	const create = (level: CategoryForm["level"], parentId = 0) => { setEditingId(null); setForm({ ...emptyForm, level, parentId }); setMessage(""); setOpen(true); };
	const edit = (category: VendorCategory) => { setEditingId(category.id); setForm({ name: category.name, level: category.parentId ? "child" : "root", parentId: category.parentId || 0, icon: category.icon || "", sortOrder: category.sortOrder || 0, isEnabled: category.isEnabled !== false }); setMessage(""); setOpen(true); };
	const submit = (event: FormEvent) => {
		event.preventDefault();
		if (form.level === "child" && !form.parentId) { setMessage("二级分类必须选择一级分类"); return; }
		setSaving(true);
		const payload = { name: form.name, parentId: form.level === "root" ? 0 : form.parentId, icon: form.icon, sortOrder: form.sortOrder, isEnabled: form.isEnabled };
		const request = editingId ? updateResource("vendor-categories", editingId, payload) : createResource("vendor-categories", payload);
		void request.then(() => { setOpen(false); setMessage(editingId ? "厂商分类修改成功" : "厂商分类创建成功"); void load(); }).catch((error) => setMessage(getApiErrorMessage(error, "厂商分类保存失败"))).finally(() => setSaving(false));
	};
	return <AdminLayout title="厂商分类">
		{message && !open && <p className="admin-message">{message}</p>}
		<div className="admin-toolbar admin-panel compact category-toolbar"><p>独立维护厂商展示位置；停用分类只会从前台隐藏导航，不会删除已有归属。</p><div className="category-admin-actions"><button className="outline-btn" type="button" onClick={() => create("child")}><Plus size={16} />新增二级分类</button><button className="primary-btn" type="button" onClick={() => create("root")}><Plus size={16} />新增一级分类</button></div></div>
		<div className="admin-table-panel"><table className="admin-table category-tree-table"><thead><tr><th>分类名称</th><th>级别</th><th>父级分类</th><th>图标</th><th>排序</th><th>状态</th><th>操作</th></tr></thead><tbody>
			{displayed.map((category) => { const child = Boolean(category.parentId); return <tr className={child ? "category-child-row" : ""} key={category.id}><td><span className="category-tree-name">{child && <span aria-hidden="true">└</span>}{category.name}</span></td><td>{child ? "二级分类" : "一级分类"}</td><td>{child ? rootNames.get(category.parentId || 0) || "父级已删除" : "—"}</td><td>{category.icon || "—"}</td><td>{category.sortOrder || 0}</td><td>{category.isEnabled === false ? "已停用" : "已启用"}</td><td><div className="category-row-actions">{!child && <button type="button" onClick={() => create("child", category.id)}>新增子分类</button>}<button type="button" onClick={() => edit(category)}>编辑</button><button type="button" onClick={() => { if (!window.confirm(`确认删除厂商分类“${category.name}”吗？`)) return; void deleteResource("vendor-categories", category.id).then(() => { setMessage("厂商分类删除成功"); void load(); }).catch((error) => setMessage(getApiErrorMessage(error, "厂商分类删除失败"))); }}>删除</button></div></td></tr>; })}
			{!displayed.length && <tr><td colSpan={7}>暂无厂商分类</td></tr>}
		</tbody></table></div>
		{open && <AdminModal label={`${editingId ? "编辑" : "新增"}厂商分类`} onClose={() => setOpen(false)}><header><div><span>{editingId ? "编辑分类" : "创建分类"}</span><h2>{form.level === "root" ? "一级分类" : "二级分类"}</h2></div><button aria-label="关闭厂商分类编辑器" type="button" onClick={() => setOpen(false)}><X size={20} /></button></header>{message && <p className="admin-message category-editor-message">{message}</p>}<form className="admin-form admin-grouped-form" onSubmit={submit}><fieldset><legend>分类信息</legend><div className="admin-field-grid"><label>分类级别<select aria-label="厂商分类级别" value={form.level} onChange={(event) => setForm({ ...form, level: event.target.value as CategoryForm["level"], parentId: event.target.value === "root" ? 0 : form.parentId })}><option value="root">一级分类</option><option value="child">二级分类</option></select></label>{form.level === "child" && <label>所属一级分类<select aria-label="所属一级厂商分类" required value={form.parentId || ""} onChange={(event) => setForm({ ...form, parentId: Number(event.target.value) })}><option value="">请选择一级分类</option>{roots.filter((root) => root.id !== editingId).map((root) => <option key={root.id} value={root.id}>{root.name}</option>)}</select></label>}<label>分类名称<input aria-label="厂商分类名称" required value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} /></label><label>图标<input value={form.icon} onChange={(event) => setForm({ ...form, icon: event.target.value })} /></label><label>排序<input type="number" value={form.sortOrder} onChange={(event) => setForm({ ...form, sortOrder: Number(event.target.value) })} /></label><label className="admin-toggle-field"><input aria-label="启用厂商分类" checked={form.isEnabled} type="checkbox" onChange={(event) => setForm({ ...form, isEnabled: event.target.checked })} /><span><strong>启用分类</strong><small>停用一级分类会同时隐藏其二级导航</small></span></label></div></fieldset><div className="admin-editor-actions"><button className="outline-btn" type="button" onClick={() => setOpen(false)}>取消</button><button className="primary-btn" disabled={saving} type="submit">{saving ? "保存中…" : "保存分类"}</button></div></form></AdminModal>}
	</AdminLayout>;
}

function sortRows(rows: VendorCategory[]) { return [...rows].sort((left, right) => (left.sortOrder || 0) - (right.sortOrder || 0) || left.id - right.id); }
