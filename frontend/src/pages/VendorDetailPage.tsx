import { Building2, CheckCircle2, ExternalLink, MapPin, MessageCircle, Phone, Wrench, X } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getVendor, getVendorContact, getVendorContactQRCode, listProducts } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { ProductCard } from "../components/public/ProductCard";
import { ErrorState, LoadingState } from "../components/public/StateViews";
import { VendorCover } from "../components/public/VendorCover";
import type { Product, Vendor } from "../types/api";
import { useSite } from "../contexts/SiteContext";
import { getMenuLabel } from "../utils/navigation";
import { vendorPath } from "../utils/vendorPath";
import { trackAnalytics } from "../analytics";
import { getStaticPageData } from "../utils/staticPageData";

export function VendorDetailPage() {
  const { layout } = useSite();
  const vendorsLabel = getMenuLabel(layout.topMenus, "/vendors", "厂商资源");
  const { id = "" } = useParams();
  const staticData = getStaticPageData("vendor", id);
  const [vendor, setVendor] = useState<Vendor | null>(staticData?.vendor || null);
  const [products, setProducts] = useState<Product[]>(staticData?.products || []);
  const [loading, setLoading] = useState(!staticData?.vendor);
  const [error, setError] = useState("");
  const [productsError, setProductsError] = useState(false);
  const [contactMessage, setContactMessage] = useState("");
  const [revealingContact, setRevealingContact] = useState(false);
  const [wechatOpen, setWechatOpen] = useState(false);
  const [qrSource, setQrSource] = useState("");
  const [qrError, setQrError] = useState("");

  const load = () => {
    setLoading(true);
    setError("");
    setProductsError(false);
    getVendor(id).then(setVendor).catch(() => setError("厂商详情加载失败或资料暂未公开")).finally(() => setLoading(false));
    listProducts({ vendorId: id, pageSize: 6 }).then((result) => setProducts(result.items)).catch(() => { setProducts([]); setProductsError(true); });
  };
  useEffect(() => {
    if (staticData?.vendor) return;
    load();
  }, [id, staticData]);

  useEffect(() => {
    if (!wechatOpen || !vendor?.wechatQrCode) {
      setQrSource("");
      setQrError("");
      return;
    }
    if (!vendor.wechatQrCode.includes("/contact-qr")) {
      setQrSource(vendor.wechatQrCode);
      setQrError("");
      return;
    }
    let objectURL = "";
    let cancelled = false;
    setQrSource("");
    setQrError("");
    getVendorContactQRCode(vendor.id)
      .then((blob) => {
        if (cancelled) return;
        objectURL = URL.createObjectURL(blob);
        setQrSource(objectURL);
      })
      .catch(() => { if (!cancelled) setQrError("二维码加载失败，请重试"); });
    return () => {
      cancelled = true;
      if (objectURL) URL.revokeObjectURL(objectURL);
    };
  }, [wechatOpen, vendor?.id, vendor?.wechatQrCode]);

  if (loading) return <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} title="厂商详情"><LoadingState /></PageFrame>;
  if (error || !vendor) return <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} title="厂商详情"><ErrorState text={error || "厂商不存在"} onRetry={load} /></PageFrame>;

  const region = [vendor.province, vendor.city, vendor.county].filter(Boolean).join(" · ");
  const phoneMasked = Boolean(vendor.phone?.includes("*"));
  const protectedContactAvailable = Boolean(
    (!vendor.phonePublic && vendor.phoneAvailable)
    || (!vendor.wechatPublic && vendor.wechatAvailable)
    || (!vendor.contactNamePublic && vendor.contactNameAvailable),
  );
  const wechatAvailable = Boolean(vendor.wechat || vendor.wechatQrCode || vendor.wechatAvailable);
  const revealContact = (openWechat = false) => {
    setRevealingContact(true);
    setContactMessage("");
    if (openWechat) setWechatOpen(true);
    getVendorContact(vendor.id).then((contact) => {
      setVendor((current) => current ? {
        ...current,
        phone: contact.phone,
        wechat: contact.wechat,
        wechatQrCode: contact.wechatQrCodeUrl,
        contactName: contact.contactName,
        phoneAvailable: false,
        wechatAvailable: false,
        contactNameAvailable: false,
      } : current);
    }).catch((requestError: { response?: { status?: number } }) => {
      setContactMessage(requestError.response?.status === 429 ? "访问过于频繁，请稍后再试" : "请先登录厂商账号后查看完整联系方式");
    }).finally(() => setRevealingContact(false));
  };
  const strengthRows = [
    ["主营产品", vendor.mainProducts || "农机配件"],
    ["年产能", vendor.annualCapacity],
    ["主要设备", vendor.equipment],
    ["资质认证", vendor.certifications],
  ].filter((row) => row[1]);
  const processingRows = [
    ["加工能力", vendor.processingServices], ["材料 / 类型", vendor.processingMaterials], ["加工设备", vendor.processingEquipment], ["产能 / 交期", vendor.processingCapacity], ["服务区域", vendor.processingRegions], ["接单说明", vendor.processingNotes],
  ].filter((row) => row[1]);
  const serviceAdvantages = Array.from(vendor.serviceAdvantages || vendor.description || "专注农机配件生产与供应").slice(0, 80).join("");

  return (
    <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} navigationCategoryIds={(vendor.vendorCategories || []).map((category) => category.id)} title={vendor.name} subtitle={region || "源头农机配件厂商"}>
      <section className="vendor-showcase-hero">
        <VendorCover variant="detail" vendor={vendor} />
        <div className="vendor-showcase-copy">
          <div className="vendor-detail-title"><div className="vendor-logo-fallback">{vendor.shortName?.slice(0, 2) || "农机"}{vendor.logo && <img alt={`${vendor.name} Logo`} src={vendor.logo} onError={(event) => { event.currentTarget.style.display = "none"; }} />}</div><div><h2>{vendor.shortName || vendor.name}</h2><div className="tag-row">{vendor.isVerified && <span className="tag-blue">平台认证</span>}{vendor.isRecommended && <span className="tag-green">推荐厂商</span>}{vendor.tags?.slice(0, 4).map((tag) => <span key={tag.id}>{tag.name}</span>)}</div></div></div>
          <p className="vendor-lead">{serviceAdvantages}</p>
          <div className="vendor-key-lines"><span><MapPin size={16} />{region || "全国供应"}</span><span><Wrench size={16} />{vendor.mainProducts || "农机配件"}</span></div>
          <div className="vendor-contact-bar" aria-label="联系方式">
            {vendor.phone && !phoneMasked
              ? <a className="primary-btn small" href={`tel:${vendor.phone}`} onClick={() => trackAnalytics({ eventType: "contact_phone_click", path: vendorPath(vendor), contentType: "vendor", contentId: vendor.id })}><Phone size={15} />{vendor.phone}</a>
              : protectedContactAvailable && <button className="primary-btn small" disabled={revealingContact} type="button" onClick={() => revealContact()}><Phone size={15} />{revealingContact ? "正在获取…" : "查看联系方式"}</button>}
            {vendor.websiteUrl && <a className="outline-btn small" href={vendor.websiteUrl} rel="noreferrer" target="_blank" onClick={() => trackAnalytics({ eventType: "vendor_website_click", path: vendorPath(vendor), contentType: "vendor", contentId: vendor.id })}><ExternalLink size={15} />访问官网</a>}
            {wechatAvailable && <button className="outline-btn small" disabled={revealingContact} type="button" onClick={() => vendor.wechatPublic || vendor.wechat || vendor.wechatQrCode ? setWechatOpen(true) : revealContact(true)}><MessageCircle size={15} />微信联系</button>}
          </div>
          {contactMessage && <span className="contact-access-message" role="status">{contactMessage}{contactMessage.startsWith("请先登录") && <>，<Link to="/account/login">前往登录</Link></>}</span>}
        </div>
      </section>

      <section className="vendor-overview-grid">
        <article className="vendor-section-card vendor-about"><header><Building2 size={20} /><h2>企业概况</h2></header><p>{vendor.description || "厂商正在完善企业介绍。"}</p>{vendor.address && <p className="vendor-address"><MapPin size={16} />{vendor.address}</p>}<div className="company-stats">{[["成立年份", vendor.establishedYear], ["厂房面积", vendor.factoryArea], ["员工规模", vendor.employeeCount]].filter((item) => item[1]).map(([label, value]) => <div key={label}><strong>{value}</strong><span>{label}</span></div>)}</div></article>
        <article className="vendor-section-card"><header><CheckCircle2 size={20} /><h2>主营产品与企业实力</h2></header><div className="vendor-strength-list">{strengthRows.map(([label, value]) => <div key={label}><strong>{label}</strong><p>{value}</p></div>)}</div></article>
      </section>

      {vendor.providesProcessing && processingRows.length > 0 && <section className="vendor-section-card processing-section"><header><Wrench size={20} /><h2>加工服务能力</h2></header><div className="capability-grid">{processingRows.map(([label, value]) => <div key={label}><strong>{label}</strong><p>{value}</p></div>)}</div></section>}
      {vendor.media?.length ? <section className="vendor-section-card"><header><Building2 size={20} /><h2>企业图集</h2></header><div className="vendor-media-grid">{vendor.media.map((media) => <figure key={media.id || media.url}><img alt={media.caption || vendor.name} loading="lazy" src={media.url} /><figcaption><span>{mediaKindLabel(media.kind)}</span>{media.caption}</figcaption></figure>)}</div></section> : null}
      {productsError && <section className="inline-notice" role="status">关联产品暂时加载失败，企业主体资料不受影响。</section>}
      {products.length > 0 && <section className="section-block"><div className="section-title"><h2>关联产品</h2><Link to={`/products?vendorId=${vendor.id}`}>查看全部</Link></div><div className="product-grid related-products">{products.map((product) => <ProductCard compact key={product.id} product={{ ...product, vendor }} />)}</div></section>}

      {wechatOpen && <div className="vendor-contact-modal-backdrop" role="presentation" onMouseDown={() => setWechatOpen(false)}><section aria-label="微信联系方式" aria-modal="true" className="vendor-contact-modal" role="dialog" onMouseDown={(event) => event.stopPropagation()}><header><div><small>{vendor.shortName || vendor.name}</small><h2>微信联系方式</h2></div><button aria-label="关闭微信联系方式" type="button" onClick={() => setWechatOpen(false)}><X size={20} /></button></header>{vendor.contactName && <p>联系人：{vendor.contactName}</p>}{vendor.wechat && <p>微信号：<strong>{vendor.wechat}</strong></p>}{vendor.wechatQrCode && <figure className="vendor-wechat-qr">{qrSource ? <img alt={`${vendor.shortName || vendor.name} 微信二维码`} src={qrSource} onError={() => { setQrSource(""); setQrError("二维码加载失败，请重试"); }} /> : qrError ? <div className="qr-load-error"><span>{qrError}</span><button className="outline-btn small" type="button" onClick={() => { setQrError(""); setWechatOpen(false); window.setTimeout(() => setWechatOpen(true)); }}>重新加载</button></div> : <div className="qr-loading">二维码加载中…</div>}<figcaption>微信扫码联系</figcaption></figure>}{!vendor.wechat && !vendor.wechatQrCode && <p className="contact-access-message">{contactMessage || "微信联系方式暂未公开"}</p>}</section></div>}
    </PageFrame>
  );
}

function mediaKindLabel(kind: string) { return ({ factory: "厂房", equipment: "设备", certificate: "证书" } as Record<string, string>)[kind] || "企业图片"; }
