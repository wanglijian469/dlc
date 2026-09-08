import { useEffect, useRef, type ReactNode } from "react";
export function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: ReactNode }) {
 const panel = useRef<HTMLElement>(null); const close = useRef<HTMLButtonElement>(null);
 const callback = useRef(onClose); callback.current = onClose;
 useEffect(() => {
  const previous = document.activeElement as HTMLElement; const overflow = document.body.style.overflow;
  document.body.style.overflow = "hidden"; close.current?.focus();
  const keys = (e: KeyboardEvent) => {
   if (e.key === "Escape") { e.preventDefault(); callback.current(); }
   if (e.key !== "Tab") return;
   const buttons = Array.from(panel.current?.querySelectorAll<HTMLElement>('button:not(:disabled), a[href], input:not(:disabled), select:not(:disabled)') || []);
   const first = buttons[0], last = buttons[buttons.length - 1];
   if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last?.focus(); }
   else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first?.focus(); }
  };
  document.addEventListener("keydown", keys);
  return () => { document.body.style.overflow = overflow; document.removeEventListener("keydown", keys); previous?.focus(); };
 }, []);
 return <div className="promotion-modal-backdrop" onMouseDown={e => { if (e.target === e.currentTarget) onClose(); }}><section className="promotion-modal" role="dialog" aria-modal="true" aria-label={title} ref={panel}><header><h2>{title}</h2><button type="button" aria-label={"关闭" + title} ref={close} onClick={onClose}>×</button></header>{children}</section></div>;
}
