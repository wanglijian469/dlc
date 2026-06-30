import { FormEvent, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { getFilterOptions, listProducts, type ProductListParams } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { ProductCard } from "../components/public/ProductCard";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import type { FilterOptions, PageResult, Product } from "../types/api";

const pageSize = 12;

export function ProductsPage() {
  const [params, setParams] = useSearchParams();
  const [keyword, setKeyword] = useState(params.get("keyword") || "");
  const [categoryId, setCategoryId] = useState(params.get("categoryId") || params.get("category") || "");
  const [onlyHot, setOnlyHot] = useState(params.get("hot") === "true");
  const [filters, setFilters] = useState<FilterOptions>({ provinces: [], categories: [], serviceTags: [] });
  const [result, setResult] = useState<PageResult<Product> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const query = useMemo<ProductListParams>(
    () => ({
      keyword: params.get("keyword") || undefined,
      categoryId: params.get("categoryId") || params.get("category") || undefined,
      vendorId: params.get("vendorId") || undefined,
      hot: params.get("hot") === "true" || undefined,
      page: Number(params.get("page") || 1),
      pageSize,
    }),
    [params],
  );

  const load = () => {
    setLoading(true);
    setError("");
    Promise.all([getFilterOptions(), listProducts(query)])
      .then(([filterOptions, products]) => {
        setFilters(filterOptions);
        setResult(products);
      })
      .catch(() => setError("产品列表加载失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [query]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const next = new URLSearchParams();
    if (keyword.trim()) next.set("keyword", keyword.trim());
    if (categoryId) next.set("categoryId", categoryId);
    if (onlyHot) next.set("hot", "true");
    next.set("page", "1");
    setParams(next);
  };

  const total = result?.total || 0;

  return (
    <PageFrame title="配件产品" subtitle="按分类、关键词和热门标识查找农机配件产品">
      <form className="filter-bar" onSubmit={submit}>
        <input value={keyword} placeholder="搜索配件名称、适配机型" onChange={(event) => setKeyword(event.target.value)} />
        <select value={categoryId} onChange={(event) => setCategoryId(event.target.value)}>
          <option value="">全部分类</option>
          {filters.categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </select>
        <label className="inline-check">
          <input checked={onlyHot} type="checkbox" onChange={(event) => setOnlyHot(event.target.checked)} />
          只看热门
        </label>
        <button className="primary-btn" type="submit">
          搜索
        </button>
      </form>
      {loading && <LoadingState />}
      {error && <ErrorState text={error} onRetry={load} />}
      {!loading && !error && result && (
        <>
          <div className="list-summary">共找到 {total} 个产品</div>
          {result.items.length ? (
            <div className="product-grid">
              {result.items.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          ) : (
            <EmptyState text="暂无符合条件的产品，请调整筛选条件" />
          )}
        </>
      )}
    </PageFrame>
  );
}
