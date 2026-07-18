import { publicClient } from "./api/client";

export type AnalyticsEventType = "page_view" | "search_submit" | "search_zero_results" | "not_found" | "join_cta_click" | "vendor_register_success" | "contact_phone_click" | "contact_wechat_copy" | "vendor_website_click";

type AnalyticsPayload = { eventType: AnalyticsEventType; path: string; contentType?: "category" | "vendor" | "product" | "article"; contentId?: number };

const publicPaths = ["/", "/vendors", "/products", "/service", "/join", "/about", "/purchase", "/links", "/guides", "/search"];

export function trackAnalytics(payload: AnalyticsPayload) {
  if (payload.eventType !== "not_found" && !isPublicPath(payload.path)) return;
  // Analytics is deliberately best-effort. Use the application's Axios/XHR
  // transport instead of browser fetch so older 360 kernels cannot break page
  // rendering when they do not implement fetch or keepalive.
  try {
    void publicClient.post("/api/analytics/events", payload).catch(() => undefined);
  } catch {
    // An unavailable Promise/XHR implementation must only disable analytics.
  }
}

export function trackRoute(pathname: string, search = "") {
  if (!isPublicPath(pathname)) return;
  const payload: AnalyticsPayload = { eventType: "page_view", path: pathname };
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
  return publicPaths.indexOf(path) >= 0 || /^\/(vendors|products)\/[^/]+$/.test(path) || /^\/products\/category\/[^/]+$/.test(path) || /^\/guides\/[^/]+$/.test(path);
}
