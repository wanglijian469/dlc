import { ArrowRight, BriefcaseBusiness, Building2, CheckCircle2, ChevronDown, ChevronUp, MapPin, Newspaper, Wrench, X, ZoomIn } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getVendor, getVendorPosts } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { ShowroomProductShelf } from "../components/public/ShowroomProductShelf";
import { VendorContactActions } from "../components/public/VendorContactActions";
import { PromotionTools } from "../components/public/PromotionTools";
import { ErrorState, LoadingState } from "../components/public/StateViews";
import { VendorCover } from "../components/public/VendorCover";
import type { Vendor, VendorPost } from "../types/api";
import { useSite } from "../contexts/SiteContext";
import { getMenuLabel } from "../utils/navigation";
import { vendorPath } from "../utils/vendorPath";
import { getStaticPageData } from "../utils/staticPageData";
import { publicMediaURL } from "../utils/publicMedia";
import { ImageLightbox, type LightboxImage } from "../components/public/ImageLightbox";

export function VendorDetailPage() {
  const { layout } = useSite();
  const vendorsLabel = getMenuLabel(layout.topMenus, "/vendors", "厂商资源");
  const { id = "" } = useParams();
  const staticData = getStaticPageData("vendor", id);
  const [vendor, setVendor] = useState<Vendor | null>(staticData?.vendor || null);
  const [posts, setPosts] = useState<VendorPost[]>(staticData?.posts || []);
  const [loading, setLoading] = useState(!staticData?.vendor);
  const [error, setError] = useState("");
  const [postsExpanded, setPostsExpanded] = useState(false);
  const [selectedPost, setSelectedPost] = useState<VendorPost | null>(null);
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null);
  const postDialogRef = useRef<HTMLElement>(null);
  const postDialogCloseRef = useRef<HTMLButtonElement>(null);
  const postOpenerRef = useRef<HTMLElement | null>(null);

  const requestVersion = useRef(0);
  const load = () => {
    const version = ++requestVersion.current;
    setLoading(true);
    setError("");
    getVendor(id).then(data => { if (version === requestVersion.current) setVendor(data); }).catch(() => { if (version === requestVersion.current) setError("厂商详情加载失败或资料暂未公开"); }).finally(() => { if (version === requestVersion.current) setLoading(false); });
  };
  useEffect(() => {
    load();
    return () => { requestVersion.current++; };
  }, [id]);
  useEffect(() => {
    if (!vendor?.id) return;
    let active = true;
    getVendorPosts(vendor.id).then(data => { if (active) setPosts(data); }).catch(() => { if (active) setPosts([]); });
    return () => { active = false; };
  }, [vendor?.id]);
  useEffect(() => { setLightboxIndex(null); }, [id]);

  useEffect(() => {
    if (!selectedPost) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    postDialogCloseRef.current?.focus();
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        setSelectedPost(null);
        return;
      }
      if (event.key !== "Tab") return;
      const focusable = postDialogRef.current?.querySelectorAll<HTMLElement>('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
      if (!focusable?.length) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = previousOverflow;
      postOpenerRef.current?.focus();
    };
  }, [selectedPost]);

  if (loading) return <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} title="厂商详情"><LoadingState /></PageFrame>;
  if (error || !vendor) return <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} title="厂商详情"><ErrorState text={error || "厂商不存在"} onRetry={load} /></PageFrame>;

  const region = [vendor.province, vendor.city, vendor.county].filter(Boolean).join(" · ");
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
  const visiblePosts = postsExpanded ? posts : posts.slice(0, 3);
  const lightboxImages: LightboxImage[] = (vendor.media || []).map((media) => {
    const kind = mediaKindLabel(media.kind);
    return { src: publicMediaURL(media.url), alt: media.caption || `${vendor.name}${kind}图片`, caption: media.caption ? `${kind} · ${media.caption}` : kind };
  });
  const openPost = (post: VendorPost, opener: HTMLElement) => {
    postOpenerRef.current = opener;
    setSelectedPost(post);
  };

  return (
    <PageFrame breadcrumbs={[{ label: vendorsLabel, path: "/vendors" }]} navigationCategoryIds={(vendor.vendorCategories || []).map((category) => category.id)} title={vendor.name} subtitle={region || "源头农机配件厂商"}>
      <section className="vendor-showcase-hero">
        <VendorCover variant="detail" vendor={vendor} />
        <div className="vendor-showcase-copy">
          <div className="vendor-detail-title"><div className="vendor-logo-fallback">{vendor.shortName?.slice(0, 2) || "农机"}{vendor.logo && <img alt={`${vendor.name} Logo`} src={vendor.logo} onError={(event) => { event.currentTarget.style.display = "none"; }} />}</div><div><h2>{vendor.shortName || vendor.name}</h2><div className="tag-row">{vendor.isRecommended && <span className="tag-green">推荐厂商</span>}{vendor.tags?.slice(0, 4).map((tag) => <span key={tag.id}>{tag.name}</span>)}</div></div></div>
          <p className="vendor-lead">{serviceAdvantages}</p>
          <div className="vendor-key-lines"><span><MapPin size={16} />{region || "全国供应"}</span><span><Wrench size={16} />{vendor.mainProducts || "农机配件"}</span></div>
          <div className="promotion-actions"><a className="outline-btn" href="#showroom-products">浏览本厂产品</a><PromotionTools path={vendorPath(vendor)} title={vendor.name} description={vendor.mainProducts} image={vendor.coverImage}/></div>
          <VendorContactActions vendor={vendor} path={vendorPath(vendor)} sticky/>

        </div>
      </section>

      <nav className="showroom-section-nav" aria-label="企业展厅分区"><a href="#showroom-products">主营产品</a><a href="#showroom-about">企业实力</a><a href="#showroom-gallery">企业图集</a><a href="#showroom-posts">案例动态</a></nav>
      <ShowroomProductShelf vendor={vendor}/>
      <section id="showroom-about" className="vendor-overview-grid">
        <article className="vendor-section-card vendor-about"><header><Building2 size={20} /><h2>企业概况</h2></header><p>{vendor.description || "厂商正在完善企业介绍。"}</p>{vendor.address && <p className="vendor-address"><MapPin size={16} />{vendor.address}</p>}<div className="company-stats">{[["成立年份", vendor.establishedYear], ["厂房面积", vendor.factoryArea], ["员工规模", vendor.employeeCount]].filter((item) => item[1]).map(([label, value]) => <div key={label}><strong>{value}</strong><span>{label}</span></div>)}</div></article>
        <article className="vendor-section-card"><header><CheckCircle2 size={20} /><h2>主营产品与企业实力</h2></header><div className="vendor-strength-list">{strengthRows.map(([label, value]) => <div key={label}><strong>{label}</strong><p>{value}</p></div>)}</div></article>
      </section>

      {vendor.providesProcessing && processingRows.length > 0 && <section className="vendor-section-card processing-section"><header><Wrench size={20} /><h2>加工服务能力</h2></header><div className="capability-grid">{processingRows.map(([label, value]) => <div key={label}><strong>{label}</strong><p>{value}</p></div>)}</div></section>}
      {vendor.media?.length ? <section id="showroom-gallery" className="vendor-section-card"><header><Building2 size={20} /><h2>企业图集</h2></header><div className="vendor-media-grid">{vendor.media.map((media, index) => <figure key={media.id || media.url}><button aria-label={`放大企业图片：${media.caption || mediaKindLabel(media.kind)}`} className="image-zoom-trigger" type="button" onClick={() => setLightboxIndex(index)}><img alt={media.caption || vendor.name} loading="lazy" src={publicMediaURL(media.url)} /><span aria-hidden="true" className="image-zoom-badge"><ZoomIn size={18} /></span></button><figcaption><span>{mediaKindLabel(media.kind)}</span>{media.caption}</figcaption></figure>)}</div></section> : null}
      {posts.length > 0 && <section id="showroom-posts" className="vendor-section-card vendor-posts-section"><header><Newspaper size={21} /><div><h2>企业动态与案例</h2><p>了解企业新闻、合作案例与生产动态</p></div></header><div className="vendor-post-public-list">{visiblePosts.map((post) => {
        const isCase = post.postType === "case";
        return <button className="vendor-post-public-item" key={post.id} type="button" onClick={(event) => openPost(post, event.currentTarget)}>
          <span className={`vendor-post-public-cover ${isCase ? "case" : "update"}`}>{post.coverImage ? <img alt={`${post.title}封面`} loading="lazy" src={post.coverImage} /> : isCase ? <BriefcaseBusiness aria-hidden="true" size={38} /> : <Newspaper aria-hidden="true" size={38} />}</span>
          <span className="vendor-post-public-copy"><span className={`vendor-post-public-type ${isCase ? "case" : "update"}`}>{isCase ? "客户案例" : "企业动态"}</span><strong>{post.title}</strong><span className="vendor-post-public-summary">{post.summary || post.content.slice(0, 120)}</span><span className="vendor-post-public-meta"><time>{formatPostDate(post.publishedAt)}</time><span>查看详情 <ArrowRight size={15} /></span></span></span>
        </button>;
      })}</div>{posts.length > 3 && <button className="vendor-post-toggle" type="button" onClick={() => setPostsExpanded((value) => !value)}>{postsExpanded ? <><ChevronUp size={17} />收起</> : <><ChevronDown size={17} />查看全部 {posts.length} 条</>}</button>}</section>}

      {selectedPost && <div className="vendor-post-modal-backdrop" role="presentation" onMouseDown={() => setSelectedPost(null)}><section aria-labelledby="vendor-post-dialog-title" aria-modal="true" className="vendor-post-modal" ref={postDialogRef} role="dialog" onMouseDown={(event) => event.stopPropagation()}><header><div><span className={`vendor-post-public-type ${selectedPost.postType === "case" ? "case" : "update"}`}>{selectedPost.postType === "case" ? "客户案例" : "企业动态"}</span><h2 id="vendor-post-dialog-title">{selectedPost.title}</h2><time>{formatPostDate(selectedPost.publishedAt)}</time></div><button aria-label="关闭企业动态详情" ref={postDialogCloseRef} type="button" onClick={() => setSelectedPost(null)}><X size={21} /></button></header>{selectedPost.coverImage && <img alt={`${selectedPost.title}封面`} className="vendor-post-modal-cover" src={selectedPost.coverImage} />}<div className="vendor-post-modal-content">{selectedPost.content}</div></section></div>}

      {lightboxIndex !== null && <ImageLightbox images={lightboxImages} index={lightboxIndex} onIndexChange={setLightboxIndex} onClose={() => setLightboxIndex(null)} />}
    </PageFrame>
  );
}

function mediaKindLabel(kind: string) { return ({ factory: "厂房", equipment: "设备", certificate: "证书" } as Record<string, string>)[kind] || "企业图片"; }

function formatPostDate(value?: string | null) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString("zh-CN");
}
