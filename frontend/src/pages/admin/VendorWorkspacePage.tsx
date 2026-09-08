import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Camera, PackagePlus, Store, TrendingUp, Bell } from "lucide-react";
import { getWorkspace, type WorkspaceSummary } from "../../api/workspace";
import { getVendorAnalytics, type VendorAnalyticsSummary } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { PromotionTools } from "../../components/public/PromotionTools";
export function VendorWorkspacePage() {
 const [data, setData] = useState<WorkspaceSummary | null>(null), [stats, setStats] = useState<VendorAnalyticsSummary | null>(null), [error, setError] = useState("");
 const load = () => { setError(""); getWorkspace().then(setData).catch(() => setError("工作台加载失败，请重试")); getVendorAnalytics(30).then(setStats).catch(() => setStats(null)); };
 useEffect(load, []);
 return <AdminLayout title="厂商工作台">
  {error && <p role="alert">{error} <button onClick={load}>重试</button></p>}
  {!data ? <p role="status">正在加载工作台…</p> : <>
   <section className="workspace-welcome"><div><span>让客户看见您的产品</span><h2>{data.vendor.shortName || data.vendor.name}</h2><p>完善展厅、及时更新供货信息，再把本厂页面分享给客户。</p></div><Link className="primary-btn" to="/admin/vendor-products?new=1"><PackagePlus size={20}/>发布产品</Link></section>
   <nav className="workspace-shortcuts" aria-label="厂商快捷操作"><Link to="/admin/vendor-products?new=1"><PackagePlus/>录入产品<small>拍照上传，三步提交</small></Link><Link to="/admin/capture"><Camera/>拍照识别<small>名片、彩页转草稿</small></Link><Link to="/admin/vendor-profile"><Store/>完善企业展厅<small>展示主营与企业实力</small></Link><Link to="/admin/vendor-analytics"><TrendingUp/>推广效果<small>了解访问和联系操作</small></Link></nav>
   <section className="admin-panel"><h2>待办事项</h2><div className="workspace-tasks">{[["未完成草稿", data.draftCount, "/admin/vendor-products?status=draft"], ["待审核产品", data.counts.pending || 0, "/admin/vendor-products?status=pending"], ["需修改产品", data.counts.rejected || 0, "/admin/vendor-products?status=rejected"], ["待更新价格", data.counts.expired || 0, "/admin/vendor-products?status=expired"]].map(([label, count, path]) => <Link key={label} to={String(path)}><strong>{count}</strong><span>{label}</span></Link>)}</div>{!!data.profileDraftCount && <p><Link to="/admin/vendor-profile">继续完善已保存的企业资料草稿 →</Link></p>}<Link className="workspace-notifications" to="/account/notifications"><Bell size={17}/>站内消息 · 未读 {data.unreadCount} 条</Link>{data.profileStatus === "rejected" && <p role="status">企业资料需修改：{data.profileReviewNote} <Link to="/admin/vendor-profile">去修改</Link></p>}{data.profileStatus === "pending" && <p>企业资料审核中；当前公开资料保持不变。</p>}</section>
   <section className="admin-panel"><h2>展厅完善建议</h2><div className="workspace-checklist">{!data.vendor.logo && <Link to="/admin/vendor-profile">补充企业 Logo，让客户更容易识别 →</Link>}{!data.vendor.mainProducts && <Link to="/admin/vendor-profile">填写主营产品，让客户了解您能供应什么 →</Link>}{!data.vendor.phone && !data.vendor.wechat && <Link to="/admin/vendor-profile">完善联系方式，并确认公开范围 →</Link>}{data.counts.missingImage > 0 && <Link to="/admin/vendor-products">为 {data.counts.missingImage} 个产品补充本厂图片 →</Link>}<Link to="/admin/vendor-products">检查产品图片、型号和供货信息 →</Link></div></section>
   <section className="admin-panel"><h2>最近 30 天推广摘要</h2>{stats?.showroom ? <div className="workspace-tasks"><div><strong>{stats.showroom.uv}</strong><span>展厅与本厂产品访客</span></div><div><strong>{stats.showroom.contactUV}</strong><span>有联系操作的访客</span></div><div><strong>{(stats.showroom.conversionRate * 100).toFixed(1)}%</strong><span>联系操作转化率</span></div></div> : <p>暂无推广摘要，可前往效果看板重试。</p>}<p>联系操作不代表接通、添加微信或成交。<Link to="/admin/vendor-analytics">查看完整数据 →</Link></p>{data.vendor.slug && data.vendor.isVisible && data.vendor.publicationStatus === "published" ? <div className="promotion-actions"><Link className="outline-btn" to={"/v/" + data.vendor.slug}>查看企业展厅</Link><PromotionTools path={"/v/" + data.vendor.slug} title={data.vendor.name} description={data.vendor.mainProducts} image={data.vendor.coverImage}/></div> : <p>企业审核发布后可生成推广链接和海报。</p>}</section>
  </>}
 </AdminLayout>;
}
