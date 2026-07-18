import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { getProcessingFilterOptions, listProcessingVendors, type VendorListParams } from "../api/public";
import { PageFrame } from "../components/public/PageFrame";
import { EmptyState, ErrorState, LoadingState } from "../components/public/StateViews";
import type { FilterOptions, PageResult, Tag, Vendor } from "../types/api";
import { SlidersHorizontal, X } from "lucide-react";
import { MobileDirectorySearch } from "../components/public/MobileDirectorySearch";

const pageSize = 12;

export function ProcessingServicesPage() {
  const [params, setParams] = useSearchParams();
  const [keyword, setKeyword] = useState(params.get("keyword") || "");
  const [province, setProvince] = useState(params.get("province") || "");
  const [tagId, setTagId] = useState(params.get("tagId") || "");
  const [sort, setSort] = useState<"recommended" | "latest">((params.get("sort") as "recommended" | "latest") || "recommended");
  const [filters, setFilters] = useState<FilterOptions>({ provinces: [], categories: [], serviceTags: [] });
  const [result, setResult] = useState<PageResult<Vendor> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filterOpen, setFilterOpen] = useState(false);

  const query = useMemo<VendorListParams>(
    () => ({
      keyword: params.get("keyword") || undefined,
      province: params.get("province") || undefined,
      tagId: params.get("tagId") || undefined,
      sort: ((params.get("sort") as "recommended" | "latest") || "recommended"),
      page: Number(params.get("page") || 1),
      pageSize,
    }),
    [params],
  );

  const load = () => {
    setLoading(true);
    setError("");
    Promise.all([getProcessingFilterOptions(), listProcessingVendors(query)])
      .then(([filterOptions, vendors]) => {
        setFilters(filterOptions);
        setResult(vendors);
      })
      .catch(() => setError("加工服务加载失败，请稍后重试"))
      .finally(() => setLoading(false));
  };

  useEffect(load, [query]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const next = new URLSearchParams();
    if (keyword.trim()) next.set("keyword", keyword.trim());
    if (province) next.set("province", province);
    if (tagId) next.set("tagId", tagId);
    if (sort !== "recommended") next.set("sort", sort);
    next.set("page", "1");
    setParams(next);
  };

  const total = result?.total || 0;
  const currentPage = result?.page || 1;
  const canLoadMore = result ? currentPage * result.pageSize < result.total : false;
  const activeFilters = [keyword.trim(), province, tagId ? filters.serviceTags.find((tag) => String(tag.id) === tagId)?.name : ""].filter(Boolean) as string[];
  const clearFilters = () => { setKeyword(""); setProvince(""); setTagId(""); setSort("recommended"); setParams(new URLSearchParams()); };

  return (
    <PageFrame title="加工服务" subtitle="展示提供加工服务的厂商能力，帮助厂商承接匹配的加工订单机会">
      <MobileDirectorySearch value={keyword} placeholder="搜索厂商、加工能力、材料、设备" onChange={setKeyword} onSubmit={submit} />
      <button aria-expanded={filterOpen} className="mobile-filter-toggle" type="button" onClick={() => setFilterOpen(!filterOpen)}><SlidersHorizontal size={17} />筛选加工服务{activeFilters.length > 0 && <span>{activeFilters.length}</span>}</button>
      <form className={`filter-bar ${filterOpen ? "open" : ""}`} onSubmit={submit}>
        <input className="filter-keyword" value={keyword} placeholder="搜索厂商、加工能力、材料、设备" onChange={(event) => setKeyword(event.target.value)} />
        <select value={province} onChange={(event) => setProvince(event.target.value)}>
          <option value="">全部地区</option>
          {filters.provinces.map((item) => (
            <option key={item} value={item}>
              {item}
            </option>
          ))}
        </select>
        <select value={tagId} onChange={(event) => setTagId(event.target.value)}>
          <option value="">全部加工服务</option>
          {filters.serviceTags.map((tag) => (
            <option key={tag.id} value={tag.id}>
              {tag.name}
            </option>
          ))}
        </select>
        <select value={sort} onChange={(event) => setSort(event.target.value as "recommended" | "latest")}>
          <option value="recommended">推荐优先</option>
          <option value="latest">最新入驻</option>
        </select>
        <button className="primary-btn" type="submit">
          搜索
        </button>
      </form>
      {activeFilters.length > 0 && <div className="active-filter-row">{activeFilters.map((item) => <span key={item}>{item}</span>)}<button type="button" onClick={clearFilters}><X size={14} />清除筛选</button></div>}
      {loading && <LoadingState />}
      {error && <ErrorState text={error} onRetry={load} />}
      {!loading && !error && result && (
        <>
          <div className="list-summary">共找到 {total} 家加工服务厂商</div>
          {result.items.length ? (
            <div className="vendor-grid directory-grid">
              {result.items.map((vendor) => (
                <ProcessingVendorCard key={vendor.id} vendor={vendor} />
              ))}
            </div>
          ) : (
            <EmptyState text="暂无符合条件的加工服务厂商，请调整筛选条件" />
          )}
          {canLoadMore && (
            <button
              className="outline-btn load-more"
              type="button"
              onClick={() => {
                const next = new URLSearchParams(params);
                next.set("page", String(currentPage + 1));
                setParams(next);
              }}
            >
              加载更多
            </button>
          )}
        </>
      )}
    </PageFrame>
  );
}

function ProcessingVendorCard({ vendor }: { vendor: Vendor }) {
  const region = [vendor.province, vendor.city].filter(Boolean).join(" / ") || "全国服务";
  const processingTags = processingOnly(vendor.tags);
  return (
    <article className="vendor-card processing-card">
      <div className="vendor-card-header">
        <div className="vendor-title-block">
          <h3>{vendor.name}</h3>
          <div className="tag-row">
            {vendor.isVerified && <span className="tag-blue">平台认证</span>}
            {processingTags.map((tag) => (
              <span className="tag-green" key={tag.id}>
                {tag.name}
              </span>
            ))}
          </div>
        </div>
      </div>
      <div className="vendor-body">
        <p className="vendor-line">地区：{region}</p>
        {vendor.processingServices && <p className="vendor-line">加工能力：{vendor.processingServices}</p>}
        {vendor.processingMaterials && <p className="vendor-line">材料/类型：{vendor.processingMaterials}</p>}
        {vendor.processingEquipment && <p className="vendor-line">设备：{vendor.processingEquipment}</p>}
        {vendor.processingCapacity && <p className="vendor-line">产能/交期：{vendor.processingCapacity}</p>}
      </div>
      <div className="card-actions">
        <Link className="primary-btn small" to={`/vendors/${vendor.slug || vendor.id}`}>
          查看厂商
        </Link>
        <Link className="outline-btn small" to={`/vendors/${vendor.slug || vendor.id}`}>
          对接加工
        </Link>
      </div>
    </article>
  );
}

function processingOnly(tags: Tag[] = []) {
  return tags.filter((tag) => !tag.tagType || tag.tagType === "processing").slice(0, 4);
}
