import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { search } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import type { SearchPayload } from "../types/api";

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
            {result.vendors.items.map((vendor) => (
              <Link className="result-row" key={vendor.id} to={`/vendors/${vendor.id}`}>
                <strong>{vendor.name}</strong>
                <span>{vendor.province} {vendor.mainProducts}</span>
              </Link>
            ))}
          </section>
          <section className="search-section">
            <h2>产品结果</h2>
            {result.products.items.map((product) => (
              <div className="result-row" key={product.id}>
                <strong>{product.name}</strong>
                <span>{product.compatibleModels || product.category?.name}</span>
              </div>
            ))}
          </section>
          <section className="search-section">
            <h2>分类结果</h2>
            {result.categories.items.map((category) => (
              <Link className="result-row" key={category.id} to={`/products?categoryId=${category.id}`}>
                {category.name}
              </Link>
            ))}
          </section>
        </>
      )}
    </PageFrame>
  );
}
