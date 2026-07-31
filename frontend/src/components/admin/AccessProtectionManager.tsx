import { useEffect, useState } from "react";
import {
  createProtectionBlock,
  createWatermarkBuildJob,
  getProtectionStatus,
  getWatermarkBuildJob,
  listProtectionBlocks,
  listProtectionEvents,
  releaseProtectionBlock,
  updateProtectionConfig,
  type ProtectionConfig,
  type ProtectionStatus,
  type ScrapeClientBlock,
  type ScrapeRiskEvent,
  type WatermarkBuildJob,
} from "../../api/admin";
import { getApiErrorMessage } from "../../api/client";

export function AccessProtectionManager() {
  const [status, setStatus] = useState<ProtectionStatus | null>(null);
  const [config, setConfig] = useState<ProtectionConfig | null>(null);
  const [events, setEvents] = useState<ScrapeRiskEvent[]>([]);
  const [blocks, setBlocks] = useState<ScrapeClientBlock[]>([]);
  const [message, setMessage] = useState("");
  const [saving, setSaving] = useState(false);
  const [watermarkJob, setWatermarkJob] = useState<WatermarkBuildJob | null>(null);

  const load = () => Promise.all([getProtectionStatus(), listProtectionEvents(), listProtectionBlocks()]).then(([nextStatus, eventPage, blockPage]) => {
    setStatus(nextStatus);
    setConfig(nextStatus.config);
    setWatermarkJob(nextStatus.latestWatermarkJob || null);
    setEvents(eventPage.items);
    setBlocks(blockPage.items);
  });
  useEffect(() => { void load().catch((error) => setMessage(getApiErrorMessage(error, "访问保护状态加载失败"))); }, []);
  useEffect(() => {
    if (!watermarkJob || !["queued", "running"].includes(watermarkJob.status)) return;
    const timer = window.setInterval(() => {
      void getWatermarkBuildJob(watermarkJob.id).then(setWatermarkJob).catch(() => undefined);
    }, 1500);
    return () => window.clearInterval(timer);
  }, [watermarkJob?.id, watermarkJob?.status]);
  const save = () => {
    if (!config) return;
    setSaving(true); setMessage("");
    updateProtectionConfig(config).then((value) => {
      setConfig(value);
      setMessage(value.auditOnly ? "配置已保存，当前仅记录风险，不执行自动封禁" : "配置已保存，自动封禁已经启用");
      return load();
    }).catch((error) => setMessage(getApiErrorMessage(error, "访问保护配置保存失败"))).finally(() => setSaving(false));
  };
  const setNumber = (key: keyof ProtectionConfig, value: string) => setConfig((current) => current ? { ...current, [key]: Number(value) } : current);
  const setLines = (key: "blockedAiAgents" | "allowCidrs", value: string) => setConfig((current) => current ? { ...current, [key]: value.split("\n").map((item) => item.trim()).filter(Boolean) } : current);
  if (!status || !config) return <section className="config-card"><p>{message || "正在加载访问保护配置…"}</p></section>;
  return <div className="access-protection-manager">
    <section className="config-card">
      <header><div><strong>访问与采集防护</strong><small>保留搜索引擎收录，识别批量详情采集；首次上线建议保持“仅记录”。</small></div><button className="primary-btn small" disabled={saving} type="button" onClick={save}>{saving ? "保存中…" : "保存配置"}</button></header>
      <div className="static-page-metrics"><article><strong>{status.events24h}</strong><span>24小时风险事件</span></article><article><strong>{status.activeBlocks}</strong><span>当前封禁</span></article><article><strong>{status.contactViewsToday}</strong><span>今日联系方式查看</span></article><article><strong>{config.distinctResourceLimit}</strong><span>详情访问阈值</span></article></div>
      <div className="config-field-grid protection-config-grid">
        <label className="checkbox-field"><input checked={config.enabled} type="checkbox" onChange={(event) => setConfig({ ...config, enabled: event.target.checked })} />启用访问行为识别</label>
        <label className="checkbox-field"><input checked={config.auditOnly} type="checkbox" onChange={(event) => setConfig({ ...config, auditOnly: event.target.checked })} />仅记录，不自动封禁</label>
        <label>统计窗口（分钟）<input min={1} type="number" value={config.windowMinutes} onChange={(event) => setNumber("windowMinutes", event.target.value)} /></label>
        <label>不同详情阈值<input min={10} type="number" value={config.distinctResourceLimit} onChange={(event) => setNumber("distinctResourceLimit", event.target.value)} /></label>
        <label>首次封禁（小时）<input min={1} type="number" value={config.blockHours} onChange={(event) => setNumber("blockHours", event.target.value)} /></label>
        <label>升级触发次数<input min={2} type="number" value={config.escalationStrikes} onChange={(event) => setNumber("escalationStrikes", event.target.value)} /></label>
        <label>升级封禁（小时）<input min={1} type="number" value={config.escalatedBlockHours} onChange={(event) => setNumber("escalatedBlockHours", event.target.value)} /></label>
        <label>水印透明度（%）<input max={90} min={5} type="number" value={config.watermarkOpacity} onChange={(event) => setNumber("watermarkOpacity", event.target.value)} /></label>
        <label className="checkbox-field"><input checked={config.watermarkEnabled} type="checkbox" onChange={(event) => setConfig({ ...config, watermarkEnabled: event.target.checked })} />公开展示图启用水印</label>
        <label>水印文字<input value={config.watermarkText} onChange={(event) => setConfig({ ...config, watermarkText: event.target.value })} /></label>
        <label className="field-wide">禁止的 AI User-Agent（每行一个）<textarea value={config.blockedAiAgents.join("\n")} onChange={(event) => setLines("blockedAiAgents", event.target.value)} /></label>
        <label className="field-wide">IP/CIDR 白名单（每行一个）<textarea placeholder="例如：203.0.113.10 或 203.0.113.0/24" value={config.allowCidrs.join("\n")} onChange={(event) => setLines("allowCidrs", event.target.value)} /></label>
      </div>
      {message && <p className="admin-message">{message}</p>}
    </section>
    <section className="config-card">
      <header><div><strong>公开图片水印</strong><small>产品、规格、厂房和设备图片使用公开水印副本；Logo、证书及后台预览保留原图。</small></div><button className="primary-btn small" disabled={!!watermarkJob && ["queued", "running"].includes(watermarkJob.status)} type="button" onClick={() => void createWatermarkBuildJob().then(setWatermarkJob).catch((error) => setMessage(getApiErrorMessage(error, "水印任务创建失败")))}>批量生成历史水印</button></header>
      {watermarkJob
        ? <div className="static-page-progress"><strong>{watermarkJob.processed} / {watermarkJob.total}</strong><div><span style={{ width: `${watermarkJob.total ? Math.round(watermarkJob.processed * 100 / watermarkJob.total) : 0}%` }} /></div><small>成功 {watermarkJob.succeeded}，跳过 {watermarkJob.skipped}，失败 {watermarkJob.failed}{watermarkJob.error ? `；${watermarkJob.error}` : ""}</small></div>
        : <p>尚未在本次会话执行历史图片批量任务；新公开图片会在首次访问时自动生成水印副本。</p>}
    </section>
    <section className="config-card"><header><div><strong>最近风险事件</strong><small>IP仅显示截断网段，客户端标识仅保存哈希。</small></div></header><ProtectionEvents rows={events} onBlock={(clientKeyHash) => void createProtectionBlock(clientKeyHash).then(load).catch((error) => setMessage(getApiErrorMessage(error, "手动封禁失败")))} /></section>
    <section className="config-card"><header><div><strong>封禁记录</strong><small>可手动提前解除；自动封禁到期后自行恢复。</small></div></header><ProtectionBlocks rows={blocks} onRelease={(id) => void releaseProtectionBlock(id).then(load).catch((error) => setMessage(getApiErrorMessage(error, "解除封禁失败")))} /></section>
  </div>;
}

function ProtectionEvents({ rows, onBlock }: { rows: ScrapeRiskEvent[]; onBlock: (clientKeyHash: string) => void }) {
  if (!rows.length) return <p>暂无风险事件。</p>;
  return <div className="admin-table-scroll"><table className="admin-table"><thead><tr><th>时间</th><th>IP网段</th><th>页面</th><th>处理</th><th>原因</th><th>操作</th></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td>{new Date(row.createdAt).toLocaleString()}</td><td>{row.ipPrefix || "—"}</td><td>{row.path}</td><td><span className={`static-page-badge ${row.action === "blocked" ? "failed" : "stale"}`}>{row.action === "blocked" ? "已封禁" : "仅记录"}</span></td><td>{row.reason}</td><td><button type="button" onClick={() => onBlock(row.clientKeyHash)}>封禁24小时</button></td></tr>)}</tbody></table></div>;
}

function ProtectionBlocks({ rows, onRelease }: { rows: ScrapeClientBlock[]; onRelease: (id: number) => void }) {
  if (!rows.length) return <p>暂无封禁记录。</p>;
  return <div className="admin-table-scroll"><table className="admin-table"><thead><tr><th>IP网段</th><th>触发次数</th><th>截止时间</th><th>状态</th><th>操作</th></tr></thead><tbody>{rows.map((row) => { const active = !row.releasedAt && new Date(row.blockedUntil).getTime() > Date.now(); return <tr key={row.id}><td>{row.ipPrefix || "—"}</td><td>{row.strikes}</td><td>{new Date(row.blockedUntil).toLocaleString()}</td><td>{active ? "生效中" : "已结束"}</td><td>{active && <button type="button" onClick={() => onRelease(row.id)}>解除封禁</button>}</td></tr>; })}</tbody></table></div>;
}
