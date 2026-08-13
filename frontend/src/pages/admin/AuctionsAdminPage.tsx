import { useEffect, useState } from "react";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { listAdminAuctions, updateAdminAuctionStatus } from "../../api/admin";
import type { ProcurementAuction } from "../../types/api";

export function AuctionsAdminPage() {
  const [items, setItems] = useState<ProcurementAuction[]>([]);
  const load = () => listAdminAuctions().then(setItems);
  useEffect(() => { void load(); }, []);
  const update = (id: number, status: "suspended" | "cancelled" | "unawarded") => { const reason = window.prompt("请填写监管处理原因") || ""; if (reason) void updateAdminAuctionStatus(id, status, reason).then(load); };
  return <AdminLayout title="采购竞价监管"><div className="admin-panel vendor-post-list">{items.map((item) => <article key={item.id}><div><span>{item.status}</span><h3>{item.title}</h3><p>{item.quantity} {item.unit} · 截止 {new Date(item.endAt).toLocaleString()}</p></div><div>{item.status === "open" && <button className="outline-btn small" type="button" onClick={() => update(item.id, "suspended")}>暂停</button>}{!["awarded", "cancelled"].includes(item.status) && <button className="outline-btn small" type="button" onClick={() => update(item.id, "cancelled")}>异常取消</button>}</div></article>)}</div></AdminLayout>;
}
