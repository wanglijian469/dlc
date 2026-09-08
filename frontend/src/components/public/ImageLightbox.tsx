import { ChevronLeft, ChevronRight, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";

export interface LightboxImage {
  src: string;
  alt: string;
  caption?: string;
}

export function ImageLightbox({ images, index, onIndexChange, onClose }: {
  images: LightboxImage[];
  index: number;
  onIndexChange: (index: number) => void;
  onClose: () => void;
}) {
  const closeRef = useRef<HTMLButtonElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);
  const touchStartX = useRef<number | null>(null);
  const indexRef = useRef(index);
  const imageCountRef = useRef(images.length);
  const closeHandlerRef = useRef(onClose);
  const indexChangeHandlerRef = useRef(onIndexChange);
  const [loadFailed, setLoadFailed] = useState(false);
  const current = images[index];
  const multiple = images.length > 1;
  indexRef.current = index;
  imageCountRef.current = images.length;
  closeHandlerRef.current = onClose;
  indexChangeHandlerRef.current = onIndexChange;

  const move = (offset: number) => {
    if (!multiple) return;
    onIndexChange((index + offset + images.length) % images.length);
  };

  useEffect(() => {
    previousFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    closeRef.current?.focus();
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeHandlerRef.current();
      } else if (event.key === "ArrowLeft" && imageCountRef.current > 1) {
        event.preventDefault();
        indexChangeHandlerRef.current((indexRef.current - 1 + imageCountRef.current) % imageCountRef.current);
      } else if (event.key === "ArrowRight" && imageCountRef.current > 1) {
        event.preventDefault();
        indexChangeHandlerRef.current((indexRef.current + 1) % imageCountRef.current);
      } else if (event.key === "Tab") {
        const dialog = closeRef.current?.closest<HTMLElement>('[role="dialog"]');
        const controls = dialog?.querySelectorAll<HTMLElement>('button:not([disabled]), [href], [tabindex]:not([tabindex="-1"])');
        if (!controls?.length) return;
        const first = controls[0];
        const last = controls[controls.length - 1];
        if (event.shiftKey && document.activeElement === first) {
          event.preventDefault();
          last.focus();
        } else if (!event.shiftKey && document.activeElement === last) {
          event.preventDefault();
          first.focus();
        }
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = previousOverflow;
      previousFocusRef.current?.focus();
    };
  }, []);

  useEffect(() => { setLoadFailed(false); }, [current?.src]);
  if (!current) return null;

  return <div className="image-lightbox-backdrop" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
    <section aria-label="图片预览" aria-modal="true" className="image-lightbox" role="dialog" onMouseDown={(event) => event.stopPropagation()}>
      <button aria-label="关闭图片预览" className="image-lightbox-close" ref={closeRef} type="button" onClick={onClose}><X size={24} /></button>
      {multiple && <button aria-label="上一张图片" className="image-lightbox-nav previous" type="button" onClick={() => move(-1)}><ChevronLeft size={34} /></button>}
      <figure
        onTouchStart={(event) => { touchStartX.current = event.touches[0]?.clientX ?? null; }}
        onTouchEnd={(event) => {
          if (touchStartX.current === null) return;
          const distance = (event.changedTouches[0]?.clientX ?? touchStartX.current) - touchStartX.current;
          touchStartX.current = null;
          if (Math.abs(distance) >= 40) move(distance < 0 ? 1 : -1);
        }}
      >
        {loadFailed ? <div className="image-lightbox-error" role="status">图片加载失败</div> : <img alt={current.alt} src={current.src} onError={() => setLoadFailed(true)} />}
        {(current.caption || multiple) && <figcaption>{current.caption && <span>{current.caption}</span>}{multiple && <small>{index + 1} / {images.length}</small>}</figcaption>}
      </figure>
      {multiple && <button aria-label="下一张图片" className="image-lightbox-nav next" type="button" onClick={() => move(1)}><ChevronRight size={34} /></button>}
    </section>
  </div>;
}
