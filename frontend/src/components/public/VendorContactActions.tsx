import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import type { Vendor } from "../../types/api";
import { getVendorContact, getVendorContactQRCode } from "../../api/public";
import { trackAnalytics } from "../../analytics";
import { Modal } from "./Modal";

export function VendorContactActions({ vendor, path, sticky = false }: { vendor: Vendor; path: string; sticky?: boolean }) {
 const location = useLocation();
 const [contact, setContact] = useState({ phone: "", wechat: "", wechatQrCodeUrl: "" });
 const [requestedContact, setRequestedContact] = useState("phone");
 const [open, setOpen] = useState(false), [loading, setLoading] = useState(false), [message, setMessage] = useState(""), [login, setLogin] = useState(false), [qr, setQr] = useState("");
 const phone = contact.phone || (vendor.phonePublic && !vendor.phone?.includes("*") ? vendor.phone : "");
 const wechat = contact.wechat || (vendor.wechatPublic ? vendor.wechat : "");
 const qrURL = contact.wechatQrCodeUrl || (vendor.wechatPublic ? vendor.wechatQrCode : "");
 const event = (eventType: "contact_phone_click" | "contact_wechat_copy" | "contact_wechat_qr_view" | "vendor_website_click") => trackAnalytics({ eventType, path });
 useEffect(() => { setContact({ phone: "", wechat: "", wechatQrCodeUrl: "" }); setOpen(false); setLogin(false); setMessage(""); }, [vendor.id, path]);
 const reveal = async (wechatRequested = false) => {
  setRequestedContact(wechatRequested ? "wechat" : "phone"); setLoading(true); setMessage(""); setLogin(false);
  try { const data = await getVendorContact(vendor.id); setContact({ phone: data.phone || "", wechat: data.wechat || "", wechatQrCodeUrl: data.wechatQrCodeUrl || "" }); if (wechatRequested) setOpen(true); }
  catch (e) { const status = (e as {response?: {status?: number}}).response?.status; setLogin(status === 401); setMessage(status === 401 ? "登录后可查看完整联系方式" : status === 429 ? "访问过于频繁，请稍后重试" : "联系方式加载失败，请重试"); }
  finally { setLoading(false); }
 };
 useEffect(() => {
  if (new URLSearchParams(location.search).get("contact") && localStorage.getItem("cms_authenticated") === "true") void reveal(new URLSearchParams(location.search).get("contact") === "wechat");
 }, [vendor.id]);
 useEffect(() => {
  let disposed = false, objectURL = ""; setQr("");
  if (open && qrURL) {
   if (qrURL.includes("/contact-qr")) getVendorContactQRCode(vendor.id).then(blob => { if (!disposed) { objectURL = URL.createObjectURL(blob); setQr(objectURL); } }).catch(() => { if (!disposed) setMessage("二维码加载失败，请关闭后重试"); });
   else setQr(qrURL);
  }
  return () => { disposed = true; if (objectURL) URL.revokeObjectURL(objectURL); };
 }, [open, qrURL, vendor.id]);
 const copyWechat = async () => { try { await navigator.clipboard.writeText(wechat || ""); event("contact_wechat_copy"); setMessage("微信号已复制"); } catch { setMessage("复制失败，请长按微信号复制"); } };
 const returnParams = new URLSearchParams(location.search); returnParams.set("contact", requestedContact);
 const returnTo = location.pathname + "?" + returnParams.toString();
 return <div className={sticky ? "showroom-contact sticky-contact" : "showroom-contact"} aria-label="联系厂商">
  <div className="showroom-contact-buttons">
   {phone ? <a className="primary-btn" href={"tel:" + phone} onClick={() => event("contact_phone_click")}>电话联系：{phone}</a> : (vendor.phoneAvailable || vendor.phone) ? <button className="primary-btn" disabled={loading} onClick={() => void reveal()}>查看联系电话</button> : <span>联系电话待完善</span>}
   {(wechat || qrURL || vendor.wechatAvailable) && <button className="outline-btn" disabled={loading} onClick={() => { if (wechat || qrURL) setOpen(true); else void reveal(true); }}>微信联系</button>}
   {vendor.websiteUrl && <a className="outline-btn contact-website" href={vendor.websiteUrl} target="_blank" rel="noreferrer" onClick={() => event("vendor_website_click")}>访问官网</a>}
  </div>
  {!open && message && <p role="status">{message} {login && <Link to={"/account/login?returnTo=" + encodeURIComponent(returnTo)}>前往登录</Link>}</p>}
  {open && <Modal title="微信联系方式" onClose={() => setOpen(false)}><p>{vendor.name}</p>{wechat && <p>微信号：<strong>{wechat}</strong> <button className="outline-btn" onClick={() => void copyWechat()}>复制微信号</button></p>}{qr && <img className="contact-qr" src={qr} alt={vendor.name + " 微信二维码"} onLoad={() => event("contact_wechat_qr_view")} onError={() => { setQr(""); setMessage("二维码加载失败，请重试"); }}/>}<p role="status">{message}</p></Modal>}
 </div>;
}
