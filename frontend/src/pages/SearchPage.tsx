import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { search } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import type { Category, Product, SearchPayload, Vendor } from "../types/api";

function regionOf(vendor?: Vendor) {
  return [vendor?.province, vendor?.city].filter(Boolean).join(" · ");
}

function VendorSearchCard({ vendor }: { vendor: Vendor }) {
  return (
    <Link className="search-card vendor-search-card" to={`/vendors/${vendor.id}`}>
      <div className="search-card-head">
        <strong>{vendor.name}</strong>
        {vendor.isVerified && <span className="tag-blue">平台认证</span>}
      </div>
      <p>地区：{regionOf(vendor) || "全国供应"}</p>
      <p>主营：{vendor.mainProducts || "农机配件"}</p>
    </Link>
  );
}

function ProductSearchCard({ product }: { product: Product }) {
  const vendor = product.vendor;
  return (
    <div className="search-card product-search-card">
      <div className="search-card-head">
        <strong>{product.name}</strong>
        {product.isHot && <span className="tag-orange">热销</span>}
      </div>
      <p>适配机型：{product.compatibleModels || "通用农机配件"}</p>
      <p>分类：{product.category?.name || "农机配件"}</p>
      <p>供应商：{vendor?.name || "平台供应商"}</p>
      <div className="card-actions">
        {vendor?.id ? (
          <Link className="primary-btn small" to={`/vendors/${vendor.id}`}>
            联系供应商
          </Link>
        ) : (
          <Link className="primary-btn small" to="/vendors">
            联系供应商
          </Link>
        )}
      </div>
    </div>
  );
}

function CategorySearchCard({ category }: { category: Category }) {
  return (
    <Link className="search-card category-search-card" to={`/products?categoryId=${category.id}`}>
      <div className="search-card-head">
        <strong>{category.name}</strong>
        <span className="tag-blue">分类入口</span>
      </div>
      <p>进入该分类查看相关配件产品和供应商</p>
    </Link>
  );
}

export function SearchPage() {
  const [params] = useSearchParams();
  const keyword = params.get("keyword") || "";
  const [result, setResult] = useState<SearchPayload | null>(null);
  const [loading, setLoading] = useState(Boolean(keyword));
  const [error, setError] = useState("");

  const load = () => {
    if (!keyword) return;
    setLoading(true);
    setError("");
    search(keyword)
      .then(setResult)
      .catch(() => setError("搜索失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [keyword]);

  const total = (result?.vendors.total || 0) + (result?.products.total || 0) + (result?.categories.total || 0);

  return (
    <PageFrame title={`搜索：${keyword || "请输入关键词"}`} subtitle="同时检索厂商、配件产品和分类">
      {!keyword && <EmptyState text="请输入配件名称、农机型号或厂商名称进行搜索" />}
      {loading && <LoadingState text="正在搜索..." />}
      {error && <ErrorState text={error} onRetry={load} />}
      {!loading && !error && result && (
        <>
          {total === 0 && <EmptyState text="没有找到匹配结果，请换个关键词试试" />}
          <section className="search-section">
            <h2>厂商结果</h2>
            <div className="search-card-list">
              {result.vendors.items.map((vendor) => (
                <VendorSearchCard key={vendor.id} vendor={vendor} />
              ))}
            </div>
          </section>
          <section className="search-section">
            <h2>产品结果</h2>
            <div className="search-card-list">
              {result.products.items.map((product) => (
                <ProductSearchCard key={product.id} product={product} />
              ))}
            </div>
          </section>
          <section className="search-section">
            <h2>分类结果</h2>
            <div className="search-card-list">
              {result.categories.items.map((category) => (
                <CategorySearchCard category={category} key={category.id} />
              ))}
            </div>
          </section>
        </>
      )}
    </PageFrame>
  );
}
