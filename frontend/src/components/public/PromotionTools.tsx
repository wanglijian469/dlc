import { useState } from "react";
import { Share2 } from "lucide-react";
import { Modal } from "./Modal";
import { publicMediaURL } from "../../utils/publicMedia";

export function promotionURL(path: string, source: "share" | "qr") {
 const url = new URL(path, window.location.origin);
 if (url.origin !== window.location.origin || !/^\/v\/[^/]+(?:\/products\/\d+)?$/.test(url.pathname)) throw new Error("仅支持已发布的本厂页面");
 url.search = ""; url.hash = ""; url.searchParams.set("from", source);
 return url.toString();
}
async function loadImage(src: string) {
 return new Promise<HTMLImageElement>((resolve, reject) => { const img = new Image(); img.crossOrigin = "anonymous"; img.onload = () => resolve(img); img.onerror = reject; img.src = src; });
}
export function PromotionTools({ path, title, description = "", image = "" }: { path: string; title: string; description?: string; image?: string }) {
 const [open, setOpen] = useState(false), [qr, setQr] = useState(""), [poster, setPoster] = useState(""), [message, setMessage] = useState(""), [busy, setBusy] = useState(false);
 const show = async () => {
  setOpen(true); setMessage(""); setBusy(true);
  try { const { default: QRCode } = await import("qrcode"); setQr(await QRCode.toDataURL(promotionURL(path, "qr"), { width: 320, margin: 4, errorCorrectionLevel: "M" })); }
  catch { setMessage("二维码生成失败，可先复制链接"); } finally { setBusy(false); }
 };
 const copy = async () => { try { await navigator.clipboard.writeText(promotionURL(path, "share")); setMessage("推广链接已复制"); } catch { setMessage("复制失败，请长按下方链接复制"); } };
 const share = async () => {
  if (!navigator.share) { await copy(); return; }
  try { await navigator.share({ title, url: promotionURL(path, "share") }); }
  catch (e) { if ((e as Error).name !== "AbortError") await copy(); }
 };
 const createPoster = async () => {
  setBusy(true); setMessage("");
  try {
   const canvas = document.createElement("canvas"); canvas.width = 750; canvas.height = 1080;
   const ctx = canvas.getContext("2d"); if (!ctx) throw new Error();
   ctx.fillStyle = "#f2f6fb"; ctx.fillRect(0, 0, 750, 1080);
   ctx.fillStyle = "#1554bd"; ctx.fillRect(0, 0, 750, 140);
   ctx.fillStyle = "#fff"; ctx.font = "bold 32px sans-serif"; ctx.fillText("大陆农机 · 企业展厅", 38, 82);
   ctx.fillStyle = "#132a47"; ctx.font = "bold 30px sans-serif";
   const wrap = (text: string, y: number, maxLines: number) => {
    let line = "", row = 0;
    for (const c of text) { if (ctx.measureText(line + c).width > 666) { ctx.fillText(line, 42, y + row * 43); line = ""; if (++row >= maxLines) return; } line += c; }
    if (line) ctx.fillText(line, 42, y + row * 43);
   };
   wrap(title, 198, 2);
   ctx.fillStyle = "#fff"; ctx.fillRect(38, 282, 674, 350);
   // Read only publicly served, same-origin imagery. Private previews are never exported.
   let drawn = false;
   if (image) {
    const url = new URL(publicMediaURL(image), window.location.origin);
    if (url.origin === window.location.origin && (/^\/api\/media\/\d+$/.test(url.pathname) || /^\/(images|uploads)\//.test(url.pathname))) {
     try { const photo = await loadImage(url.href); const scale = Math.min(654 / photo.width, 330 / photo.height); const w = photo.width * scale, h = photo.height * scale; ctx.drawImage(photo, (750 - w) / 2, 292 + (330 - h) / 2, w, h); drawn = true; } catch { /* Keep a truthful placeholder. */ }
    }
   }
   if (!drawn) { ctx.fillStyle = "#526880"; ctx.font = "24px sans-serif"; ctx.fillText("扫码查看企业与产品资料", 214, 465); }
   ctx.fillStyle = "#526880"; ctx.font = "24px sans-serif"; wrap(description || "了解本厂产品、型号与供货信息", 685, 2);
   if (!qr) throw new Error(); const code = await loadImage(qr); ctx.drawImage(code, 42, 792, 220, 220);
   ctx.fillStyle = "#132a47"; ctx.font = "bold 28px sans-serif"; ctx.fillText("扫码了解 · 直接联系厂商", 286, 862);
   ctx.font = "22px sans-serif"; ctx.fillText("产品供应情况以厂商确认为准", 286, 909);
   setPoster(canvas.toDataURL("image/png")); setMessage("海报已生成，可保存或长按图片保存");
  } catch { setMessage("海报生成失败，请重试或使用二维码"); } finally { setBusy(false); }
 };
 return <><button type="button" className="outline-btn small" onClick={() => void show()}><Share2 size={16}/>分享推广</button>{open && <Modal title="分享推广" onClose={() => setOpen(false)}><p>{title}</p><div className="promotion-actions"><button className="primary-btn" onClick={() => void copy()}>复制链接</button><button className="outline-btn" onClick={() => void share()}>系统分享</button><button className="outline-btn" disabled={busy || !qr} onClick={() => void createPoster()}>生成海报</button></div><input aria-label="推广链接" readOnly value={promotionURL(path, "share")} onFocus={e => e.target.select()}/>{busy && <p role="status">正在生成…</p>}{message && <p role="status">{message}</p>}{qr && <figure className="promotion-qr"><img src={qr} alt="推广二维码"/><figcaption>扫码访问本厂页面</figcaption></figure>}{poster && <><img className="promotion-poster" src={poster} alt={title + "推广海报"}/><a className="primary-btn" href={poster} download="企业推广海报.png">保存推广海报</a></>}<small>海报不包含隐藏联系方式；电话点击不代表实际接通。</small></Modal>}</>;
}
