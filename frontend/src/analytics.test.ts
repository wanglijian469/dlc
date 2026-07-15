import { afterEach, describe, expect, it, vi } from "vitest";
import { publicClient } from "./api/client";
import { trackAnalytics, trackRoute } from "./analytics";

describe("public analytics", () => {
  afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); });

  it("reports public route visits through the existing HTTP client", () => {
    const post = vi.spyOn(publicClient, "post").mockResolvedValue({} as never);
    trackRoute("/products", "?categoryId=8");
    expect(post).toHaveBeenCalledWith("/api/analytics/events", { eventType: "page_view", path: "/products", contentType: "category", contentId: 8 });
  });

  it("does not report administrative routes", () => {
    const post = vi.spyOn(publicClient, "post").mockResolvedValue({} as never);
    trackRoute("/admin/dashboard");
    trackAnalytics({ eventType: "page_view", path: "/admin/users" });
    expect(post).not.toHaveBeenCalled();
  });
});
