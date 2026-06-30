import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getVendor, listProducts } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { ProductCard } from "../components/public/ProductCard";
import { ErrorState, LoadingState } from "../components/public/StateViews";
import type { Product, Vendor } from "../types/api";

const logoFallback = "https://dummyimage.com/120x80/ffffff/0b5fea&text=DL";
const websiteLead = "平台可为源头厂商搭建独立展示网站，提升询盘转化";

type DetailItem = {
  label: string;
  value?: string;
};

export function VendorDetailPage() {
  const { id = "" } = useParams();
  const [vendor, setVendor] = useState<Vendor | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = () => {
    setLoading(true);
    setError("");
    Promise.all([getVendor(id), listProducts({ vendorId: id, pageSize: 8 })])
      .then(([vendorData, productData]) => {
        setVendor(vendorData);
        setProducts(productData.items);
      })
      .catch(() => setError("厂商详情加载失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [id]);

  if (loading) return <PageFrame title="厂商详情"><LoadingState /></PageFrame>;
  if (error || !vendor) return <PageFrame title="厂商详情"><ErrorState text={error || "厂商不存在"} onRetry={load} /></PageFrame>;

  const region = [vendor.province, vendor.city, vendor.county].filter(Boolean).join(" · ");
  const overviewItems: DetailItem[] = [
    { label: "成立年份", value: vendor.establishedYear },
    { label: "厂房面积", value: vendor.factoryArea },
    { label: "员工规模", value: vendor.employeeCount },
  ];

  return (
    <PageFrame title={vendor.name} subtitle={region || "源头农机配件厂商"}>
      <section className="vendor-profile-hero">
        <div className="vendor-profile-main">
          <div className="vendor-profile-heading">
            <img
              alt={vendor.name}
              src={vendor.logo || logoFallback}
              onError={(event) => {
                event.currentTarget.src = logoFallback;
              }}
            />
            <div>
              <h2>{vendor.shortName || "企业档案"}</h2>
              <div className="tag-row">
                {vendor.isVerified && <span className="tag-blue">平台认证</span>}
                {vendor.isRecommended && <span className="tag-green">推荐厂商</span>}
                {vendor.tags?.map((tag) => (
                  <span className="tag-green" key={tag.id}>
                    {tag.name}
                  </span>
                ))}
              </div>
            </div>
          </div>
          <div className="vendor-profile-lines">
            <p>地区：{region || "全国供应"}</p>
            <p>主营：{vendor.mainProducts || "农机配件"}</p>
            {vendor.serviceModels && <p>适配机型：{vendor.serviceModels}</p>}
            {vendor.contactName && <p>联系人：{vendor.contactName}</p>}
            {vendor.phone && <p>联系电话：{vendor.phone}</p>}
            {vendor.address && <p>地址：{vendor.address}</p>}
          </div>
        </div>
        <aside className="vendor-website-card">
          <span>厂商独立官网</span>
          <h3>{vendor.websiteUrl ? "已有官网展示入口" : "开通独立官网获取更多询盘"}</h3>
          <p>{websiteLead}</p>
          {vendor.websiteUrl ? (
            <a className="primary-btn vendor-website-main" href={vendor.websiteUrl} rel="noreferrer" target="_blank">
              进入厂商官网
            </a>
          ) : (
            <Link className="primary-btn vendor-website-main" to="/join">
              申请开通独立官网
            </Link>
          )}
        </aside>
      </section>

      <section className="vendor-profile-grid">
        <DetailCard items={overviewItems} title="公司实力" />
        <TextCard title="公司简介" value={vendor.description} />
        <TextCard title="主营产品" value={vendor.mainProducts} />
        <TextCard label="年产能" title="生产能力" value={vendor.annualCapacity} />
        <TextCard label="主要设备" title="设备能力" value={vendor.equipment} />
        <TextCard label="认证资质" title="质量认证" value={vendor.certifications} />
        <TextCard label="质检能力" title="质检能力" value={vendor.qualityControl} />
        <TextCard title="服务优势" value={vendor.serviceAdvantages} />
        <TextCard title="供货范围" value={vendor.supplyRegions} />
        <TextCard title="合作说明" value={vendor.cooperationTerms} />
        <TextCard label="售后服务" title="售后服务" value={vendor.afterSalesService} />
        <DetailCard
          title="联系方式"
          items={[
            { label: "联系人", value: vendor.contactName },
            { label: "电话", value: vendor.phone },
            { label: "微信", value: vendor.wechat },
            { label: "地址", value: vendor.address || region },
          ]}
        />
      </section>

      <section className="section-block">
        <div className="section-title">
          <h2>关联产品</h2>
          <Link to={`/products?vendorId=${vendor.id}`}>查看全部产品</Link>
        </div>
        <div className="product-grid">
          {products.map((product) => (
            <ProductCard compact key={product.id} product={{ ...product, vendor }} />
          ))}
        </div>
      </section>
    </PageFrame>
  );
}

function TextCard({ title, value, label }: { title: string; value?: string; label?: string }) {
  if (!value) return null;
  return (
    <article className="vendor-profile-card">
      <h3>{title}</h3>
      <p>{label ? `${label}：` : ""}{value}</p>
    </article>
  );
}

function DetailCard({ title, items }: { title: string; items: DetailItem[] }) {
  const visible = items.filter((item) => item.value);
  if (!visible.length) return null;
  return (
    <article className="vendor-profile-card">
      <h3>{title}</h3>
      <div className="vendor-detail-list">
        {visible.map((item) => (
          <p key={item.label}>{item.label}：{item.value}</p>
        ))}
      </div>
    </article>
  );
}
