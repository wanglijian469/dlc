import { BarChart3, Eye, MapPinned, MousePointerClick, UsersRound } from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";
import { getAnalyticsSummary, type AnalyticsSummary } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";

const eventLabels: Record<string, string> = {
  search_submit: "提交搜索",
  join_cta_click: "点击厂商入驻",
  vendor_register_success: "注册成功",
  contact_phone_click: "点击电话",
  contact_wechat_copy: "复制微信",
  vendor_website_click: "访问厂商官网",
};

export function AdminAnalyticsPage() {
  const [days, setDays] = useState<7 | 30 | 90>(30);
  const [data, setData] = useState<AnalyticsSummary | null>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    setError("");
    getAnalyticsSummary(days).then(setData).catch(() => setError("访问分析加载失败，请稍后重试"));
  }, [days]);
  const trend = data?.trend || [];
  const provinces = data?.provinces || [];
  const contents = data?.contents || [];
  const conversionEvents = data?.conversionEvents || [];
  const maxTrend = Math.max(...trend.map((row) => row.pv), 1);
  return <AdminLayout title="访问分析">
    <section className="admin-panel analytics-panel">
      <div className="admin-section-title"><div><h2>公开站点访问概览</h2><span>仅统计匿名访问；地区精度为省级，明细保留 90 天。</span></div><div className="analytics-range">{([7, 30, 90] as const).map((value) => <button className={days === value ? "active" : ""} key={value} onClick={() => setDays(value)} type="button">近 {value} 天</button>)}</div></div>
      {error && <p className="admin-message" role="alert">{error}</p>}
      {!data && !error && <p className="structured-empty">正在加载访问数据…</p>}
      {data && <>
        <div className="admin-metric-grid analytics-metrics">
          <article><Eye size={20} /><span>浏览量 PV<strong>{data.pv}</strong></span></article>
          <article><UsersRound size={20} /><span>访客数 UV<strong>{data.uv}</strong></span></article>
          <article><MousePointerClick size={20} /><span>转化操作<strong>{data.conversions}</strong></span></article>
          <article><BarChart3 size={20} /><span>转化率<strong>{(data.conversionRate * 100).toFixed(1)}%</strong></span></article>
        </div>
        <div className="analytics-grid">
          <section className="analytics-card"><h3>访问趋势</h3>{trend.length ? <div className="analytics-trend">{trend.map((row) => <div key={row.date} title={`${row.date}：${row.pv} PV，${row.uv} UV`}><span style={{ height: `${Math.max(5, row.pv / maxTrend * 100)}%` }} /><small>{row.date.slice(5)}</small></div>)}</div> : <p>暂无访问数据</p>}</section>
          <AnalyticsList icon={<MapPinned size={18} />} title="访问省份" rows={provinces} />
          <AnalyticsList title="访问内容排行" rows={contents.map((row) => ({ label: contentLabel(row.path, row.contentType, row.contentId), count: row.count }))} />
          <AnalyticsList title="转化动作" rows={conversionEvents.map((row) => ({ label: eventLabels[row.label] || row.label, count: row.count }))} />
        </div>
      </>}
    </section>
  </AdminLayout>;
}

function AnalyticsList({ title, rows, icon }: { title: string; rows: Array<{ label: string; count: number }>; icon?: ReactNode }) {
  const max = Math.max(...rows.map((row) => row.count), 1);
  return <section className="analytics-card"><h3>{icon}{title}</h3>{rows.length ? <ol className="analytics-list">{rows.map((row) => <li key={row.label}><span>{row.label}</span><i><b style={{ width: `${row.count / max * 100}%` }} /></i><strong>{row.count}</strong></li>)}</ol> : <p>暂无数据</p>}</section>;
}

function contentLabel(path: string, type: string, id: number) {
  if (type === "vendor") return `厂商资源 #${id}`;
  if (type === "product") return `配件货源 #${id}`;
  if (type === "category") return `配件分类 #${id}`;
  if (type === "article") return `行业文章 #${id}`;
  return path;
}
