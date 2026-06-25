import { Menu, Search } from "lucide-react";

export function MobileHeader() {
  return (
    <header className="mobile-header">
      <Menu size={22} />
      <div className="brand mobile-brand">
        <span className="brand-mark">农</span>
        <span>大陆农机配件</span>
      </div>
      <Search size={22} />
    </header>
  );
}
