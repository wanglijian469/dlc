import { Search } from "lucide-react";
import type { FormEvent } from "react";

export function MobileDirectorySearch({
  value,
  placeholder,
  onChange,
  onSubmit,
}: {
  value: string;
  placeholder: string;
  onChange: (value: string) => void;
  onSubmit: (event: FormEvent) => void;
}) {
  return (
    <form className="mobile-directory-search" onSubmit={onSubmit}>
      <Search aria-hidden="true" size={18} />
      <input aria-label="搜索关键词" value={value} placeholder={placeholder} onChange={(event) => onChange(event.target.value)} />
      <button aria-label="快速搜索" type="submit">搜索</button>
    </form>
  );
}
