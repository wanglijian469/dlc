import { useEffect, useRef } from "react";
import { useLocation, useNavigationType } from "react-router-dom";

// Only public list position is retained, never contact details or draft payloads.
const positions = new Map<string, number>();
export function useListScrollRestoration(ready: boolean) {
  const location = useLocation();
  const navigation = useNavigationType();
  const key = location.pathname + location.search;
  const restored = useRef("");
  useEffect(() => {
    const remember = () => positions.set(key, window.scrollY);
    window.addEventListener("scroll", remember, { passive: true });
    return () => window.removeEventListener("scroll", remember);
  }, [key]);
  useEffect(() => {
    if (!ready || restored.current === location.key) return;
    restored.current = location.key;
    if (navigation === "POP" && positions.has(key)) {
      const top = positions.get(key)!;
      const frame = requestAnimationFrame(() => window.scrollTo({ top, behavior: "instant" }));
      return () => cancelAnimationFrame(frame);
    }
  }, [ready, key, location.key, navigation]);
}
