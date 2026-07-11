import { FormEvent, useEffect, useState } from "react";
import { createCMSUser, deleteCMSUser, listCMSUserPage, listResource, updateCMSUser, type CMSUser } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { Pagination } from "../../components/public/Pagination";
import type { Vendor } from "../../types/api";
import { Plus, X } from "lucide-react";
import { AdminModal } from "../../components/admin/AdminModal";

type UserForm = { username: string; password: string; role: "admin" | "vendor"; vendorId: number; isEnabled: boolean };
const emptyForm: UserForm = { username: "", password: "", role: "vendor", vendorId: 0, isEnabled: true };

export function AdminUsersPage() {
  const [users, setUsers] = useState<CMSUser[]>([]);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [form, setForm] = useState<UserForm>(emptyForm);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [message, setMessage] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const load = () => Promise.all([listCMSUserPage({ page, pageSize: 20, search: search || undefined, role: roleFilter || undefined }), listResource<Vendor & { id: number }>("vendors")]).then(([result, vendorRows]) => { setUsers(result.items); setTotal(result.total); setVendors(vendorRows); });
  useEffect(() => { void load(); }, [page, search, roleFilter]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setMessage("");
    const payload = { ...form, vendorId: form.role === "vendor" ? form.vendorId : undefined };
    const action = editingId ? updateCMSUser(editingId, payload) : createCMSUser(payload);
    action.then(() => { setMessage("账号已保存"); setEditingId(null); setForm(emptyForm); setEditorOpen(false); void load(); }).catch(() => setMessage("保存失败，请检查用户名、密码和绑定厂商"));
  };

  return (
    <AdminLayout title="CMS 账号">
      {message && <p className="admin-message">{message}</p>}
      <div className="admin-toolbar admin-panel compact"><input placeholder="搜索用户名" value={search} onChange={(e) => { setSearch(e.target.value); setPage(1); }} /><select aria-label="账号角色筛选" value={roleFilter} onChange={(e) => { setRoleFilter(e.target.value); setPage(1); }}><option value="">全部角色</option><option value="vendor">厂商用户</option><option value="admin">管理员</option></select><span>共 {total} 个账号</span><button className="primary-btn" type="button" onClick={() => { setEditingId(null); setForm(emptyForm); setEditorOpen(true); }}><Plus size={16} />新增账号</button></div>
      <div className="admin-table-panel"><table className="admin-table"><thead><tr><th>用户名</th><th>角色</th><th>绑定厂商</th><th>状态</th><th>操作</th></tr></thead><tbody>{users.map((user) => <tr key={user.id}><td data-label="用户名">{user.username}</td><td data-label="角色">{user.role === "vendor" ? "厂商用户" : "管理员"}</td><td data-label="绑定厂商">{user.vendor?.name || "—"}</td><td data-label="状态">{user.isEnabled ? "启用" : "停用"}</td><td data-label="操作"><button onClick={() => { setEditingId(user.id); setForm({ username: user.username, password: "", role: user.role, vendorId: user.vendorId || 0, isEnabled: user.isEnabled }); setEditorOpen(true); }}>编辑</button><button onClick={() => { if (window.confirm("确认删除该 CMS 账号？")) void deleteCMSUser(user.id).then(load).catch(() => setMessage("账号删除失败")); }}>删除</button></td></tr>)}</tbody></table></div>
      <Pagination onChange={setPage} page={page} pageSize={20} total={total} />
      {editorOpen && <AdminModal label={editingId ? "编辑 CMS 账号" : "新增 CMS 账号"} className="admin-user-editor" onClose={() => setEditorOpen(false)}><header><div><span>{editingId ? "编辑账号" : "新增账号"}</span><h2>CMS 账号</h2></div><button aria-label="关闭账号编辑器" type="button" onClick={() => setEditorOpen(false)}><X size={20} /></button></header><form className="admin-form admin-grouped-form" onSubmit={submit}><fieldset><legend>账号资料</legend><div className="admin-field-grid">
        <label>用户名<input required value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} /></label>
        <label>{editingId ? "新密码（留空则不修改）" : "初始密码（至少 6 位）"}<input type="password" required={!editingId} value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} /></label>
        <label>账号角色<select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value as UserForm["role"] })}><option value="vendor">厂商用户</option><option value="admin">管理员</option></select></label>
        {form.role === "vendor" && <label>绑定厂商<select required value={form.vendorId || ""} onChange={(e) => setForm({ ...form, vendorId: Number(e.target.value) })}><option value="">请选择厂商</option>{vendors.map((vendor) => <option key={vendor.id} value={vendor.id}>{vendor.name}</option>)}</select></label>}
        <label className="checkbox-field"><input type="checkbox" checked={form.isEnabled} onChange={(e) => setForm({ ...form, isEnabled: e.target.checked })} />启用账号</label>
      </div></fieldset><div className="admin-editor-actions"><button className="outline-btn" type="button" onClick={() => setEditorOpen(false)}>取消</button><button className="primary-btn" type="submit">{editingId ? "保存账号" : "创建账号"}</button></div></form></AdminModal>}
    </AdminLayout>
  );
}
