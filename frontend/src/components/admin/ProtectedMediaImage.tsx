import { useEffect, useState } from "react";
import { API_BASE_URL } from "../../api/client";

export function ProtectedMediaImage({ assetId, src, alt, className }: { assetId?: number; src: string; alt: string; className?: string }) {
  const [resolved, setResolved] = useState(assetId ? "" : src);
  useEffect(() => {
    if (!assetId) { setResolved(src); return; }
    const controller = new AbortController(); let objectURL = "";
    fetch(`${API_BASE_URL}/api/admin/media/${assetId}`, { credentials: "include", signal: controller.signal }).then((response) => { if (!response.ok) throw new Error(); return response.blob(); }).then((blob) => { objectURL = URL.createObjectURL(blob); setResolved(objectURL); }).catch(() => setResolved(""));
    return () => { controller.abort(); if (objectURL) URL.revokeObjectURL(objectURL); };
  }, [assetId, src]);
  return resolved ? <img alt={alt} className={className} src={resolved} /> : <span className="media-preview-unavailable">图片预览不可用</span>;
}
