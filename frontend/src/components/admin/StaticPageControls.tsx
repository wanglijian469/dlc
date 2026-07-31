import { useCallback, useEffect, useMemo, useState } from "react";
import {
  createStaticBuildJob,
  getStaticBuildJob,
  getStaticPageResourceStatuses,
  getStaticPageSummary,
  updateStaticPageSettings,
  type StaticBuildJob,
  type StaticPageResourceStatus,
  type StaticPageSummary,
} from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";

const statusLabels: Record<string, string> = {
  unbuilt: "未生成",
  generating: "生成中",
  ready: "已生成",
  stale: "需更新",
  failed: "生成失败",
};

async function waitForJob(id: number, onProgress: (job: StaticBuildJob) => void) {
  for (;;) {
    const job = await getStaticBuildJob(id);
    onProgress(job);
    if (job.status === "completed" || job.status === "failed") return job;
    await new Promise((resolve) => window.setTimeout(resolve, 800));
  }
}

export function useStaticPageStatuses(resourceType: "vendor" | "product", ids: number[], enabled: boolean) {
  const [statuses, setStatuses] = useState<Record<number, StaticPageResourceStatus>>({});
  const [generatingId, setGeneratingId] = useState(0);
  const signature = ids.join(",");
  const reload = useCallback(async () => {
    if (!enabled || !ids.length) {
      setStatuses({});
      return;
    }
    const rows = await getStaticPageResourceStatuses(resourceType, ids);
    setStatuses(Object.fromEntries(rows.map((row) => [row.resourceId, row])));
  }, [enabled, resourceType, signature]);
  useEffect(() => { void reload().catch(() => setStatuses({})); }, [reload]);
  const generate = async (id: number) => {
    setGeneratingId(id);
    try {
      const job = await createStaticBuildJob({ scope: "single", resourceType, resourceIds: [id] });
      await waitForJob(job.id, () => undefined);
      await reload();
    } catch (error) {
      setStatuses((current) => ({
        ...current,
        [id]: {
          ...(current[id] || { resourceType, resourceId: id, canGenerate: true }),
          status: "failed",
          errorMessage: getApiErrorMessage(error, "静态页面生成失败"),
        },
      }));
    } finally {
      setGeneratingId(0);
    }
  };
  return { statuses, generatingId, generate };
}

export function StaticPageStatusCell({
  status,
  generating,
  onGenerate,
}: {
  status?: StaticPageResourceStatus;
  generating: boolean;
  onGenerate: () => void;
}) {
  if (!status) return <span className="static-page-muted">—</span>;
  const shownStatus = generating ? "generating" : status.status;
  return <div className="static-page-row-status">
    <span className={`static-page-badge ${shownStatus}`}>{statusLabels[shownStatus]}</span>
    {status.generatedAt && <small>{new Date(status.generatedAt).toLocaleString()}</small>}
    {status.errorMessage && <small className="static-page-error" title={status.errorMessage}>{status.errorMessage}</small>}
    <div>
      <button disabled={!status.canGenerate || generating} type="button" onClick={onGenerate}>
        {shownStatus === "ready" || shownStatus === "stale" || shownStatus === "failed" ? "重新生成" : "生成静态页"}
      </button>
      {status.status === "ready" && status.path && <a href={status.path} rel="noreferrer" target="_blank">访问</a>}
    </div>
  </div>;
}

export function StaticPageManager() {
  const [summary, setSummary] = useState<StaticPageSummary | null>(null);
  const [job, setJob] = useState<StaticBuildJob | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const load = useCallback(() => getStaticPageSummary().then((value) => {
    setSummary(value);
    if (value.latestJob) setJob(value.latestJob);
  }), []);
  useEffect(() => { void load().catch((error) => setMessage(getApiErrorMessage(error, "静态页面状态加载失败"))); }, [load]);
  const counts = useMemo(() => summary?.counts || { ready: 0, unbuilt: 0, stale: 0, generating: 0, failed: 0 }, [summary]);
  const setEnabled = async (enabled: boolean) => {
    setBusy(true); setMessage("");
    try {
      setSummary(await updateStaticPageSettings(enabled));
      setMessage(enabled ? "静态页面服务已启用" : "静态页面服务已停用，前台将使用动态页面");
    } catch (error) {
      setMessage(getApiErrorMessage(error, "设置保存失败"));
    } finally {
      setBusy(false);
    }
  };
  const generateAll = async () => {
    setBusy(true); setMessage("");
    try {
      const created = await createStaticBuildJob({ scope: "all" });
      setJob(created);
      const completed = await waitForJob(created.id, setJob);
      setMessage(`生成完成：成功 ${completed.succeeded}，失败 ${completed.failed}`);
      await load();
    } catch (error) {
      setMessage(getApiErrorMessage(error, "全量生成失败"));
    } finally {
      setBusy(false);
    }
  };
  if (!summary) return <section className="config-card"><p>{message || "正在加载静态页面状态…"}</p></section>;
  const progress = job?.total ? Math.round(job.processed / job.total * 100) : 0;
  return <section className="config-card static-page-manager">
    <header>
      <div><strong>静态页面生产</strong><small>仅对已发布且前台可见的厂商与产品生成；内容更新后自动回退动态页面。</small></div>
      <label className="static-page-switch"><input checked={summary.enabled} disabled={!summary.available || busy} type="checkbox" onChange={(event) => void setEnabled(event.target.checked)} />启用静态页面</label>
    </header>
    {!summary.available && <p className="static-page-warning">当前服务未配置前端生产目录或服务器未开启 STATIC_PAGES_ENABLED，暂不能启用。</p>}
    <div className="static-page-metrics">
      <article><strong>{counts.ready || 0}</strong><span>已生成</span></article>
      <article><strong>{counts.unbuilt || 0}</strong><span>待生成</span></article>
      <article><strong>{counts.stale || 0}</strong><span>需更新</span></article>
      <article><strong>{counts.failed || 0}</strong><span>失败</span></article>
    </div>
    {job && (job.status === "running" || job.status === "queued") && <div className="static-page-progress"><div><span style={{ width: `${progress}%` }} /></div><small>{job.processed}/{job.total}，成功 {job.succeeded}，失败 {job.failed}</small></div>}
    {job?.errors?.length ? <details><summary>查看失败原因（{job.errors.length}）</summary><ul>{job.errors.map((error, index) => <li key={`${index}-${error}`}>{error}</li>)}</ul></details> : null}
    <div className="static-page-actions"><button className="primary-btn" disabled={!summary.enabled || busy} type="button" onClick={() => void generateAll()}>{busy ? "正在处理…" : "生成全部厂商与产品页面"}</button><small>输出目录：{summary.outputDir || "未配置"}</small></div>
    {message && <p className="admin-message">{message}</p>}
  </section>;
}
