import { useEffect, useState } from "react";
import { listOperationLogs, type OperationLog } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { Pagination } from "../../components/public/Pagination";

export function AdminLogsPage() {
  const [rows, setRows] = useState<OperationLog[]>([]); const [page, setPage] = useState(1); const [total, setTotal] = useState(0); const [search, setSearch] = useState(""); const [error, setError] = useState("");
  const load = () => { setError(""); return listOperationLogs({ page, pageSize: 30, search: search || undefined }).then((result) => { setRows(result.items); setTotal(result.total); }).catch(() => setError("操作日志加载失败，请重试")); };
  useEffect(() => { const timer = window.setTimeout(() => void load(), 300); return () => window.clearTimeout(timer); }, [page, search]);
  return <AdminLayout title="操作日志"><div className="admin-toolbar admin-panel compact"><input aria-label="搜索操作日志" placeholder="搜索账号、动作或资源" value={search} onChange={(event) => { setSearch(event.target.value); setPage(1); }} /><span>共 {total} 条</span></div>{error && <p className="admin-message">{error} <button onClick={() => void load()} type="button">重试</button></p>}<div className="admin-table-panel"><table className="admin-table"><thead><tr><th>时间</th><th>账号</th><th>动作</th><th>资源</th><th>记录 ID</th></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td data-label="时间">{new Date(row.createdAt).toLocaleString()}</td><td data-label="账号">{row.username}</td><td data-label="动作">{row.action}</td><td data-label="资源">{row.resource}</td><td data-label="记录 ID">{row.recordId || "—"}</td></tr>)}</tbody></table></div><Pagination onChange={setPage} page={page} pageSize={30} total={total} /></AdminLayout>;
}
