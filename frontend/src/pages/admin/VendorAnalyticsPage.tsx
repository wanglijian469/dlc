import { ExternalLink, Eye, MousePointerClick, PackageSearch, UsersRound } from "lucide-react";
import { useEffect, useState } from "react";
import { getVendorAnalytics, type VendorAnalyticsSummary } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";

const contactLabels: Record<string, string> = {
  contact_phone_click: "电话点击",
  contact_wechat_copy: "微信复制",
  vendor_website_click: "官网访问",
};

export function VendorAnalyticsPage() {
  const [days, setDays] = useState<7 | 30 | 90>(30);
  const [data, setData] = useState<VendorAnalyticsSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    getVendorAnalytics(days).then((result) => { if (active) setData(result); }).catch(() => {
      if (active) { setData(null); setError("访问数据加载失败，请稍后重试"); }
    }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [days, reloadKey]);

  return <AdminLayout title="访问数据">
    <section className="admin-panel analytics-panel vendor-analytics-panel">
      <div className="admin-section-title">
        <div><h2>{data?.vendor.name || "厂商"}访问概览</h2><span>查看公开厂商页与已审核供应产品的匿名访问汇总。</span></div>
        <div className="analytics-range" aria-label="统计时间范围">{([7, 30, 90] as const).map((value) => <button aria-pressed={days === value} className={days === value ? "active" : ""} key={value} onClick={() => setDays(value)} type="button">近 {value} 天</button>)}</div>
      </div>
      {error && <div className="analytics-error"><p className="admin-message" role="alert">{error}</p><button className="outline-btn small" onClick={() => setReloadKey((value) => value + 1)} type="button">重新加载</button></div>}
      {loading && <p className="structured-empty" role="status">正在加载访问数据…</p>}
      {!loading && data && <>
        <div className="admin-metric-grid analytics-metrics vendor-analytics-metrics">
          <article><Eye size={20} /><span>厂商页 PV<strong>{data.vendor.pv}</strong></span></article>
          <article><UsersRound size={20} /><span>厂商页 UV<strong>{data.vendor.uv}</strong></span></article>
          <article><MousePointerClick size={20} /><span>联系操作<strong>{data.vendor.contacts}</strong></span></article>
          <article><PackageSearch size={20} /><span>产品页 PV / UV<strong>{data.products.pv} / {data.products.uv}</strong></span></article>
        </div>
        <div className="analytics-grid vendor-analytics-grid">
          <TrendCard title="厂商页访问趋势" trend={data.vendor.trend} />
          <TrendCard title="产品页访问趋势" trend={data.products.trend} />
          <section className="analytics-card"><h3>联系操作</h3>{data.vendor.contactEvents.length ? <ul className="vendor-contact-events">{data.vendor.contactEvents.map((item) => <li key={item.label}><span>{contactLabels[item.label] || item.label}</span><strong>{item.count}</strong></li>)}</ul> : <p>当前时间范围内暂无联系操作</p>}</section>
          <section className="analytics-card vendor-product-ranking"><h3>产品访问排行</h3>{data.products.items.length ? <div className="admin-table-wrap"><table className="admin-table"><thead><tr><th>产品名称</th><th>PV</th><th>UV</th><th>公开页面</th></tr></thead><tbody>{data.products.items.map((item) => <tr key={item.productId}><td data-label="产品名称">{item.name}</td><td data-label="PV">{item.pv}</td><td data-label="UV">{item.uv}</td><td data-label="公开页面"><a className="vendor-public-link" href={item.path} rel="noreferrer" target="_blank">查看 <ExternalLink size={14} /></a></td></tr>)}</tbody></table></div> : <p>暂无已审核的供应产品或访问数据</p>}</section>
        </div>
        <p className="vendor-analytics-note">说明：共享产品可能由多家厂商供应，产品访问量代表该产品页面的整体热度，不等同于本厂独立曝光。</p>
      </>}
    </section>
  </AdminLayout>;
}

function TrendCard({ title, trend }: { title: string; trend: Array<{ date: string; pv: number; uv: number }> }) {
  const max = Math.max(...trend.map((row) => row.pv), 1);
  return <section className="analytics-card"><h3>{title}</h3>{trend.length ? <div className="analytics-trend">{trend.map((row) => <div key={row.date} title={`${row.date}：${row.pv} PV，${row.uv} UV`}><span style={{ height: `${Math.max(5, row.pv / max * 100)}%` }} /><small>{row.date.slice(5)}</small></div>)}</div> : <p>当前时间范围内暂无访问数据</p>}</section>;
}
