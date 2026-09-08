import { useEffect, useRef, useState } from "react";
import { uploadFile } from "../../api/admin";
import { ProtectedMediaImage } from "./ProtectedMediaImage";
import { getApiErrorMessage } from "../../api/client";
type Photos = { image?: string; galleryRaw?: string };
export function ProductPhotosEditor({ value, onChange, onBusy }: { value: Photos; onChange: (patch: Photos) => void; onBusy: (busy: boolean) => void }) {
 const latest = useRef(value); latest.current = value;
 const [jobs, setJobs] = useState<Array<{ key: number; file: File; progress: number; error: string }>>([]);
 const pending = useRef(0), counter = useRef(0);
 useEffect(() => { onBusy(jobs.length > 0); }, [jobs.length, onBusy]);
 const gallery = (raw?: string): string[] => { try { return JSON.parse(raw || "[]"); } catch { return []; } };
 const send = async (job: {key: number;file: File}) => {
  pending.current++; onBusy(true);
  setJobs(rows => rows.map(row => row.key === job.key ? { ...row, error: "", progress: 0 } : row));
  try {
   const result = await uploadFile(job.file, {}, progress => setJobs(rows => rows.map(row => row.key === job.key ? { ...row, progress } : row)));
   const current = latest.current;
   const patch = current.image ? { galleryRaw: JSON.stringify([...gallery(current.galleryRaw), result.url]) } : { image: result.url };
   latest.current = { ...current, ...patch }; onChange(patch);
   setJobs(rows => rows.filter(row => row.key !== job.key));
  } catch (e) { setJobs(rows => rows.map(row => row.key === job.key ? { ...row, error: getApiErrorMessage(e, "上传失败，请重试") } : row)); }
  finally { pending.current--; }
 };
 const add = (files: FileList | null) => { if (!files) return; const tasks = Array.from(files).map(file => ({ key: ++counter.current, file, progress: 0, error: "" })); setJobs(rows => [...rows, ...tasks]); tasks.forEach(job => void send(job)); };
 const urls = [value.image, ...gallery(value.galleryRaw)].filter(Boolean) as string[];
 const reorder = (rows: string[]) => onChange({ image: rows[0] || "", galleryRaw: JSON.stringify(rows.slice(1)) });
 return <section className="product-photos-editor"><header><strong>产品图片</strong><span>第一张为主图；仅审核通过后公开</span></header><div className="promotion-actions"><label className="primary-btn upload-button">拍照<input aria-label="拍摄产品图片" type="file" accept="image/jpeg,image/png,image/webp" capture="environment" onChange={e => { add(e.target.files); e.target.value = ""; }}/></label><label className="outline-btn upload-button">从相册选图<input aria-label="选择产品图片" type="file" multiple accept="image/jpeg,image/png,image/webp" onChange={e => { add(e.target.files); e.target.value = ""; }}/></label></div><div className="draft-photo-grid">{urls.map((url, i) => <article key={url + i}><ProtectedMediaImage src={url} assetId={Number(url.match(/\/api\/media\/(\d+)/)?.[1]) || undefined} alt={i === 0 ? "产品主图" : "产品图集"}/><span>{i === 0 ? "主图" : "图集"}</span><div><button type="button" disabled={i === 0} onClick={() => reorder([url, ...urls.filter((_, j) => i !== j)])}>设为主图</button><button type="button" onClick={() => reorder(urls.filter((_, j) => i !== j))}>删除</button></div></article>)}</div>{jobs.map(job => <div key={job.key} role="status">{job.file.name}：{job.error ? <>{job.error} <button type="button" onClick={() => void send(job)}>重试此图</button><button type="button" onClick={() => setJobs(rows => rows.filter(row => row.key !== job.key))}>移除</button></> : <progress aria-label={job.file.name + "上传进度"} value={job.progress} max={100}/>}</div>)}</section>;
}
