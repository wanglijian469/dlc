export function Pagination({ page, pageSize, total, onChange }: { page: number; pageSize: number; total: number; onChange: (page: number) => void }) {
  const pages = Math.max(1, Math.ceil(total / pageSize));
  if (pages <= 1) return null;
  return <nav aria-label="分页" className="pagination"><button disabled={page <= 1} type="button" onClick={() => onChange(page - 1)}>上一页</button><span>第 {page} / {pages} 页</span><button disabled={page >= pages} type="button" onClick={() => onChange(page + 1)}>下一页</button></nav>;
}
