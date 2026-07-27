import { useEffect, useMemo, useState } from "react";
import { Search, X } from "lucide-react";
import { listVendorOptions } from "../../api/admin";
import type { VendorOption } from "../../types/api";

type VendorOptionSearchProps = {
  excludedIds?: number[];
  label?: string;
  onSelect: (vendor: VendorOption) => void;
};

export function VendorOptionSearch({ excludedIds = [], label = "搜索厂商", onSelect }: VendorOptionSearchProps) {
  const [keyword, setKeyword] = useState("");
  const [debouncedKeyword, setDebouncedKeyword] = useState("");
  const [options, setOptions] = useState<VendorOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [open, setOpen] = useState(false);
  const excluded = useMemo(() => new Set(excludedIds), [excludedIds]);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedKeyword(keyword.trim()), 300);
    return () => window.clearTimeout(timer);
  }, [keyword]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    void listVendorOptions({ search: debouncedKeyword || undefined })
      .then((result) => {
        if (active) setOptions(result.items);
      })
      .catch(() => {
        if (active) setError("厂商搜索失败，请稍后重试");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [debouncedKeyword]);

  const available = options.filter((option) => !excluded.has(option.id));
  return (
    <div className="vendor-option-search">
      <label>
        <Search size={16} />
        <input
          aria-label={label}
          placeholder="输入厂商名称、简称、地区或主营产品"
          value={keyword}
          onFocus={() => setOpen(true)}
          onBlur={() => window.setTimeout(() => setOpen(false), 100)}
          onChange={(event) => {
            setKeyword(event.target.value);
            setOpen(true);
          }}
        />
      </label>
      {open && <div className="vendor-option-results" role="listbox" aria-label={`${label}结果`}>
        {available.map((vendor) => (
          <button key={vendor.id} type="button" onMouseDown={(event) => event.preventDefault()} onClick={() => { onSelect(vendor); setOpen(false); }}>
            <span>
              <strong>{vendor.name}</strong>
              <small>{[vendor.province, vendor.city, vendor.mainProducts].filter(Boolean).join(" · ") || "未填写地区和主营产品"}</small>
            </span>
            <em className={vendor.publicationStatus === "published" && vendor.isVisible ? "published" : "inactive"}>
              {vendor.publicationStatus === "published" && vendor.isVisible ? "前台已发布" : "未发布"}
            </em>
          </button>
        ))}
        {!loading && !error && !available.length && <p>没有可选厂商</p>}
        {loading && <p>正在搜索厂商…</p>}
        {error && <p className="vendor-option-error">{error}</p>}
      </div>}
    </div>
  );
}

export function NewProductVendorPicker({ value, onChange }: { value: VendorOption[]; onChange: (vendors: VendorOption[]) => void }) {
  return (
    <fieldset className="product-vendor-picker">
      <legend>关联厂商</legend>
      <p>产品可关联多家厂商；发布时至少需要一家“前台已发布”的厂商。</p>
      {value.length > 0 && (
        <div className="selected-vendor-list">
          {value.map((vendor) => (
            <article key={vendor.id}>
              <span>
                <strong>{vendor.name}</strong>
                <small>{[vendor.province, vendor.city].filter(Boolean).join(" · ") || "地区未填写"}</small>
              </span>
              <em className={vendor.publicationStatus === "published" && vendor.isVisible ? "published" : "inactive"}>
                {vendor.publicationStatus === "published" && vendor.isVisible ? "前台已发布" : "未发布"}
              </em>
              <button aria-label={`移除厂商 ${vendor.name}`} type="button" onClick={() => onChange(value.filter((item) => item.id !== vendor.id))}>
                <X size={15} />
              </button>
            </article>
          ))}
        </div>
      )}
      <VendorOptionSearch excludedIds={value.map((vendor) => vendor.id)} onSelect={(vendor) => onChange([...value, vendor])} />
    </fieldset>
  );
}
