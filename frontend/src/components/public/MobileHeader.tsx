import { Menu, Search } from "lucide-react";
import { Link } from "react-router-dom";
import type { SiteMeta } from "../../types/api";

export function MobileHeader({ siteMeta }: { siteMeta?: SiteMeta }) {
  return (
    <header className="mobile-header">
      <Menu size={22} />
      <Link className="brand mobile-brand" to="/">
        <span className="brand-mark">{siteMeta?.mobileBrandMark || siteMeta?.brandMark || "农"}</span>
        <span>{siteMeta?.mobileBrandName || siteMeta?.siteName || "大陆农机配件"}</span>
      </Link>
      <Link aria-label="搜索" to="/search">
        <Search size={22} />
      </Link>
    </header>
  );
}
