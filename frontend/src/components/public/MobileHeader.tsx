import { Menu, Search } from "lucide-react";
import { Link } from "react-router-dom";

export function MobileHeader() {
  return (
    <header className="mobile-header">
      <Menu size={22} />
      <Link className="brand mobile-brand" to="/">
        <span className="brand-mark">农</span>
        <span>大陆农机配件</span>
      </Link>
      <Link aria-label="搜索" to="/search">
        <Search size={22} />
      </Link>
    </header>
  );
}
