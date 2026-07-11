import { type ReactNode, useEffect, useRef } from "react";

export function AdminModal({ label, onClose, children, className = "" }: { label: string; onClose: () => void; children: ReactNode; className?: string }) {
  const panelRef = useRef<HTMLElement>(null);
  const previousFocus = useRef<HTMLElement | null>(null);
	const closeRef = useRef(onClose);
	closeRef.current = onClose;
  useEffect(() => {
    previousFocus.current = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const panel = panelRef.current;
    window.setTimeout(() => (panel?.querySelector<HTMLElement>("input, select, textarea, button") || panel)?.focus(), 0);
    const keydown = (event: KeyboardEvent) => {
	  if (event.key === "Escape") { event.preventDefault(); closeRef.current(); return; }
      if (event.key !== "Tab" || !panel) return;
      const focusable = Array.from(panel.querySelectorAll<HTMLElement>('button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), a[href], [tabindex]:not([tabindex="-1"])'));
      if (!focusable.length) { event.preventDefault(); panel.focus(); return; }
      const first = focusable[0]; const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus(); }
      else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
    };
    document.addEventListener("keydown", keydown);
    return () => { document.removeEventListener("keydown", keydown); document.body.style.overflow = previousOverflow; previousFocus.current?.focus(); };
	// The modal stays mounted while fields change; installing this once avoids
	// resetting focus on every parent render.
  }, []);
  return <div className="admin-editor-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}><section aria-label={label} aria-modal="true" className={`admin-editor-panel ${className}`} ref={panelRef} role="dialog" tabIndex={-1}>{children}</section></div>;
}
