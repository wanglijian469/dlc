import { AlertTriangle, CheckCircle2, ExternalLink, FileSearch, Image, Link2, RefreshCw, SearchCheck } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { getSEOStatus, submitBaiduURLs, type SEOStatus } from "../../api/admin";
import { AdminLayout } from "../../components/admin/AdminLayout";
import { ErrorState, LoadingState } from "../../components/public/StateViews";

const optimizationEntries = [
  { path: "/admin/vendors", title: "厂商页面", description: "补充厂商 SEO 标题、摘要、短拼音 URL、Logo 和企业介绍。" },
  { path: "/admin/products", title: "产品页面", description: "完善产品标题、详情正文、分类、主图和适配机型。" },
  { path: "/admin/categories", title: "分类落地页", description: "维护分类 URL、搜索标题和分类摘要，避免目录页内容过少。" },
  { path: "/admin/pages", title: "内容与行业文章", description: "优化正文、封面、内部链接和搜索摘要。" },
];

export function AdminSEOPage() {
  const [status, setStatus] = useState<SEOStatus | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const load = () => {
    setError("");
    setLoading(true);
    return getSEOStatus().then(setStatus).catch(() => setError("SEO 状态加载失败，请重试")).finally(() => setLoading(false));
  };
  useEffect(() => { void load(); }, []);
  const priorityIssues = useMemo(() => status ? status.issues.missingTitle + status.issues.missingDescription + status.issues.missingImage + status.issues.duplicateTitles + status.issues.duplicateDescriptions : 0, [status]);

  return <AdminLayout title="SEO 工作台">
    {!status && !error && <LoadingState />}
    {error && <ErrorState text={error} onRetry={load} />}
    {status && <>
      {message && <p className="admin-message">{message}</p>}
      <section className="seo-workbench-hero">
        <div><span className="admin-eyebrow">SEO OVERVIEW</span><h2>搜索优化任务总览</h2><p>优先处理影响搜索结果展示的内容问题，再检查抓取、重定向与 Sitemap 状态。</p></div>
        <div className={`seo-health-score ${priorityIssues ? "warning" : "healthy"}`}><strong>{priorityIssues}</strong><span>{priorityIssues ? "项内容待优化" : "核心检查已通过"}</span></div>
        <button className="outline-btn" disabled={loading} type="button" onClick={() => void load()}><RefreshCw size={16} />{loading ? "刷新中…" : "刷新检查"}</button>
      </section>

      <section className="seo-task-grid" aria-label="SEO 待办概览">
        <SEOIssueCard icon={FileSearch} title="元数据缺失" value={status.issues.missingTitle + status.issues.missingDescription} detail={`缺标题 ${status.issues.missingTitle} · 缺描述 ${status.issues.missingDescription}`} attention={status.issues.missingTitle + status.issues.missingDescription > 0} />
        <SEOIssueCard icon={Image} title="图片问题" value={status.issues.missingImage} detail="厂商 Logo、产品主图或文章封面" attention={status.issues.missingImage > 0} />
        <SEOIssueCard icon={SearchCheck} title="重复元数据" value={status.issues.duplicateTitles + status.issues.duplicateDescriptions} detail={`重复标题 ${status.issues.duplicateTitles} · 重复描述 ${status.issues.duplicateDescriptions}`} attention={status.issues.duplicateTitles + status.issues.duplicateDescriptions > 0} />
        <SEOIssueCard icon={Link2} title="访问与链接" value={status.issues.recent404 + status.issues.zeroResultSearches} detail={`近 30 日 404：${status.issues.recent404} · 零结果搜索：${status.issues.zeroResultSearches}`} attention={status.issues.recent404 + status.issues.zeroResultSearches > 0} />
      </section>

      <section className="admin-panel seo-optimization-panel">
        <div className="admin-section-title"><div><h2>开始优化内容</h2><p>选择内容类型进入对应列表，点击“编辑”即可处理 SEO 字段和正文。</p></div></div>
        <div className="seo-entry-grid">{optimizationEntries.map((entry) => <Link key={entry.path} to={entry.path}><strong>{entry.title}</strong><span>{entry.description}</span><small>进入优化 →</small></Link>)}</div>
      </section>

      <div className="seo-lower-grid">
        <section className="admin-panel seo-monitor-panel">
          <div className="admin-section-title"><div><h2>站点监测</h2><p>这些数据用于发现失效入口和搜索需求。</p></div></div>
          <dl><div><dt>近 30 日 404</dt><dd className={status.issues.recent404 ? "warning" : ""}>{status.issues.recent404}</dd></div><div><dt>零结果搜索</dt><dd className={status.issues.zeroResultSearches ? "warning" : ""}>{status.issues.zeroResultSearches}</dd></div><div><dt>301 重定向规则</dt><dd>{status.issues.redirects}</dd></div></dl>
        </section>
        <section className="admin-panel seo-sitemap-panel">
          <header><div><h2>Sitemap 与搜索提交</h2><p>只收录已发布、未到未来发布时间的规范地址。</p></div><a className="outline-btn small" href={status.sitemap.index} target="_blank" rel="noreferrer">查看 Sitemap <ExternalLink size={14} /></a></header>
          <div className="seo-sitemap-state"><span>{status.baiduSubmission.configured ? <CheckCircle2 size={18} /> : <AlertTriangle size={18} />}</span><div><strong>百度资源提交</strong><p>{status.baiduSubmission.configured ? "凭据已配置，可提交最新规范地址。" : "当前未配置，保持默认关闭。"}</p></div></div>
          <p className="seo-sitemap-sections">分片：{status.sitemap.sections.join("、")}</p>
          {status.baiduSubmission.enabled && <button className="primary-btn small" type="button" onClick={() => { setMessage("正在提交百度资源…"); submitBaiduURLs().then((result) => setMessage(`已提交 ${result.submitted} 个规范地址`)).catch(() => setMessage("百度资源提交失败，请检查凭据和网络")); }}>提交百度资源</button>}
        </section>
      </div>
    </>}
  </AdminLayout>;
}

function SEOIssueCard({ icon: Icon, title, value, detail, attention }: { icon: typeof FileSearch; title: string; value: number; detail: string; attention: boolean }) {
  return <article className={`seo-task-card ${attention ? "needs-attention" : "healthy"}`}><span className="seo-task-icon"><Icon size={20} /></span><div><span>{title}</span><strong>{value}</strong><small>{detail}</small></div>{attention ? <em>待处理</em> : <em>正常</em>}</article>;
}
