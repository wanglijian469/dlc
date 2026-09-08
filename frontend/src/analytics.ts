import { publicClient } from "./api/client";

export type AnalyticsEventType = "page_view" | "search_submit" | "search_zero_results" | "not_found" | "join_cta_click" | "vendor_register_success" | "contact_phone_click" | "contact_wechat_copy" | "contact_wechat_qr_view" | "vendor_website_click";

type AnalyticsPayload = { eventType: AnalyticsEventType; path: string; contentType?: "category" | "vendor" | "product" | "article" | "supplier"; contentId?: number; source?: string };

let previousPath = "";
let campaign: {owner: string; source: string} | undefined;
let routeSource = "unknown";
const publicPaths = ["/", "/vendors", "/products", "/service", "/join", "/about", "/purchase", "/links", "/guides", "/search"];

export function trackAnalytics(payload: AnalyticsPayload) {
  if (payload.eventType !== "not_found" && !isPublicPath(payload.path)) return;
  // Analytics is deliberately best-effort. Use the application's Axios/XHR
  // transport instead of browser fetch so older 360 kernels cannot break page
  // rendering when they do not implement fetch or keepalive.
  try {
    const source = new URLSearchParams(window.location.search).get("from");
    let site = false;
    try { site = !!document.referrer && new URL(document.referrer).origin === window.location.origin; } catch { /* Unknown source. */ }
    void publicClient.post("/api/analytics/events", { ...payload, source: payload.source || (source === "share" || source === "qr" ? source : site ? "site" : routeSource) }).catch(() => undefined);
  } catch {
    // An unavailable Promise/XHR implementation must only disable analytics.
  }
}

export function trackRoute(pathname: string, search = "") {
  if (!isPublicPath(pathname)) return;
  const source = new URLSearchParams(search).get("from");
  const owner = pathname.match(/^\/v\/[^/]+/)?.[0] || pathname;
  if (source === "share" || source === "qr") campaign = {owner, source};
  else if (campaign?.owner !== owner) campaign = undefined;
  routeSource = campaign?.source || (previousPath && previousPath !== pathname ? "site" : "unknown");
  previousPath = pathname;
  const payload: AnalyticsPayload = { eventType: "page_view", path: pathname, source: routeSource };
  if (pathname === "/products") {
    const categoryMatch = /(?:^|[?&])categoryId=(\d+)(?:&|$)/.exec(search);
    const category = categoryMatch ? Number(categoryMatch[1]) : 0;
    if (category > 0 && category % 1 === 0) {
      payload.contentType = "category";
      payload.contentId = category;
    }
  }
  trackAnalytics(payload);
}

function isPublicPath(path: string) {
  return /^\/v\/[^/]+(?:\/products\/\d+)?$/.test(path) || publicPaths.indexOf(path) >= 0 || /^\/(vendors|products)\/[^/]+$/.test(path) || /^\/products\/category\/[^/]+$/.test(path) || /^\/guides\/[^/]+$/.test(path);
}
