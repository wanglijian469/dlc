import { Building2, CheckCircle2, ExternalLink, MapPin, Phone, Wrench } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getVendor, getVendorContact, listProducts } from "../api/public";
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
  const vendorsLabel = getMenuLabel(layout.topMenus, "/vendors", "厂商目录");
  const { id = "" } = useParams();
  const staticData = getStaticPageData("vendor", id);
  const [vendor, setVendor] = useState<Vendor | null>(staticData?.vendor || null);
  const [products, setProducts] = useState<Product[]>(staticData?.products || []);
  const [loading, setLoading] = useState(!staticData?.vendor);
  const [error, setError] = useState("");
  const [productsError, setProductsError] = useState(false);
  const [contactMessage, setContactMessage] = useState("");
  const [revealingContact, setRevealingContact] = useState(false);

  const load = () => {
    setLoading(true); setError("");
    setProductsError(false);
    getVendor(id).then(setVendor).catch(() => setError("厂商详情加载失败或资料暂未公开")).finally(() => setLoading(false));
    listProducts({ vendorId: id, pageSize: 6 }).then((result) => setProducts(result.items)).catch(() => { setProducts([]); setProductsError(true); });
  };
  useEffect(() => {
    if (staticData?.vendor) return;
    load();
  }, [id, staticData]);

  if (loading) return <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} title="厂商详情"><LoadingState /></PageFrame>;
  if (error || !vendor) return <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} title="厂商详情"><ErrorState text={error || "厂商不存在"} onRetry={load} /></PageFrame>;

  const region = [vendor.province, vendor.city, vendor.county].filter(Boolean).join(" · ");
  const phoneMasked = Boolean(vendor.phone?.includes("*"));
  const protectedContactAvailable = Boolean(
    (!vendor.phonePublic && vendor.phoneAvailable)
    || (!vendor.wechatPublic && vendor.wechatAvailable)
    || (!vendor.contactNamePublic && vendor.contactNameAvailable),
  );
  const revealContact = () => {
    setRevealingContact(true);
    setContactMessage("");
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
  const capabilityRows = [
    ["年产能", vendor.annualCapacity], ["主要设备", vendor.equipment], ["售后服务", vendor.afterSalesService],
  ].filter((row) => row[1]);
  const processingRows = [
    ["加工能力", vendor.processingServices], ["材料 / 类型", vendor.processingMaterials], ["加工设备", vendor.processingEquipment], ["产能 / 交期", vendor.processingCapacity], ["服务区域", vendor.processingRegions], ["接单说明", vendor.processingNotes],
  ].filter((row) => row[1]);

  return (
    <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} navigationCategoryIds={(vendor.vendorCategories || []).map((category) => category.id)} title={vendor.name} subtitle={region || "源头农机配件厂商"}>
      <section className="vendor-showcase-hero">
        <VendorCover variant="detail" vendor={vendor} />
        <div className="vendor-showcase-copy">
          <div className="vendor-detail-title"><div className="vendor-logo-fallback">{vendor.shortName?.slice(0, 2) || "农机"}{vendor.logo && <img alt={`${vendor.name} Logo`} src={vendor.logo} onError={(event) => { event.currentTarget.style.display = "none"; }} />}</div><div><h2>{vendor.shortName || vendor.name}</h2><div className="tag-row">{vendor.isVerified && <span className="tag-blue">平台认证</span>}{vendor.isRecommended && <span className="tag-green">推荐厂商</span>}{vendor.tags?.slice(0, 4).map((tag) => <span key={tag.id}>{tag.name}</span>)}</div></div></div>
          <p className="vendor-lead">{vendor.serviceAdvantages || vendor.description || "专注农机配件生产与供应"}</p>
          <div className="vendor-key-lines"><span><MapPin size={16} />{region || "全国供应"}</span><span><Wrench size={16} />{vendor.mainProducts || "农机配件"}</span></div>
          <div className="vendor-direct-contact" aria-label="联系方式">
            <div className="vendor-direct-contact-copy">
              <strong><Phone size={17} />联系方式</strong>
              <dl>
                {vendor.contactName && <div><dt>联系人</dt><dd>{vendor.contactName}</dd></div>}
                {vendor.phone && <div><dt>电话</dt><dd>{phoneMasked ? <span>{vendor.phone}</span> : <a href={`tel:${vendor.phone}`} onClick={() => trackAnalytics({ eventType: "contact_phone_click", path: vendorPath(vendor), contentType: "vendor", contentId: vendor.id })}>{vendor.phone}</a>}</dd></div>}
                {vendor.wechat && <div><dt>微信</dt><dd>{vendor.wechat}</dd></div>}
                {vendor.address && <div><dt>地址</dt><dd>{vendor.address}</dd></div>}
              </dl>
              <div className="vendor-contact-links">
                {protectedContactAvailable && <button className="primary-btn small" disabled={revealingContact} type="button" onClick={revealContact}><Phone size={15} />{revealingContact ? "正在获取…" : "登录查看完整联系方式"}</button>}
                {vendor.websiteUrl && <a className="outline-btn small" href={vendor.websiteUrl} rel="noreferrer" target="_blank" onClick={() => trackAnalytics({ eventType: "vendor_website_click", path: vendorPath(vendor), contentType: "vendor", contentId: vendor.id })}><ExternalLink size={15} />访问官网</a>}
              </div>
              {contactMessage && <span className="contact-access-message" role="status">{contactMessage}{contactMessage.startsWith("请先登录") && <>，<Link to="/account/login">前往登录</Link></>}</span>}
            </div>
            {vendor.wechatQrCode && <figure className="vendor-wechat-qr"><img alt={`${vendor.shortName || vendor.name} 微信二维码`} src={vendor.wechatQrCode} /><figcaption>微信扫码联系</figcaption></figure>}
          </div>
        </div>
      </section>

      <section className="vendor-overview-grid">
        <article className="vendor-section-card vendor-about"><header><Building2 size={20} /><h2>企业概况</h2></header><p>{vendor.description || "厂商正在完善企业介绍。"}</p><div className="company-stats">{[["成立年份", vendor.establishedYear], ["厂房面积", vendor.factoryArea], ["员工规模", vendor.employeeCount]].filter((item) => item[1]).map(([label, value]) => <div key={label}><strong>{value}</strong><span>{label}</span></div>)}</div></article>
        <article className="vendor-section-card"><header><Wrench size={20} /><h2>主营与适配</h2></header><dl><div><dt>主营产品</dt><dd>{vendor.mainProducts || "农机配件"}</dd></div>{vendor.serviceModels && <div><dt>适配机型</dt><dd>{vendor.serviceModels}</dd></div>}</dl></article>
      </section>

      {capabilityRows.length > 0 && <section className="vendor-section-card"><header><CheckCircle2 size={20} /><h2>生产能力与服务保障</h2></header><div className="capability-grid">{capabilityRows.map(([label, value]) => <div key={label}><strong>{label}</strong><p>{value}</p></div>)}</div></section>}
      {vendor.providesProcessing && processingRows.length > 0 && <section className="vendor-section-card processing-section"><header><Wrench size={20} /><h2>加工服务能力</h2></header><div className="capability-grid">{processingRows.map(([label, value]) => <div key={label}><strong>{label}</strong><p>{value}</p></div>)}</div></section>}

      {vendor.media?.length ? <section className="vendor-section-card"><header><Building2 size={20} /><h2>企业图集</h2></header><div className="vendor-media-grid">{vendor.media.map((media) => <figure key={media.id || media.url}><img alt={media.caption || vendor.name} loading="lazy" src={media.url} /><figcaption><span>{mediaKindLabel(media.kind)}</span>{media.caption}</figcaption></figure>)}</div></section> : null}

      {vendor.certifications && <section className="vendor-section-card"><header><CheckCircle2 size={20} /><h2>资质与认证</h2></header><p>{vendor.certifications}</p></section>}

      {productsError && <section className="inline-notice" role="status">关联产品暂时加载失败，企业主体资料不受影响。</section>}
      {products.length > 0 && <section className="section-block"><div className="section-title"><h2>关联产品</h2><Link to={`/products?vendorId=${vendor.id}`}>查看全部</Link></div><div className="product-grid related-products">{products.map((product) => <ProductCard compact key={product.id} product={{ ...product, vendor }} />)}</div></section>}

    </PageFrame>
  );
}

function mediaKindLabel(kind: string) { return ({ factory: "厂房", equipment: "设备", certificate: "证书" } as Record<string, string>)[kind] || "企业图片"; }
