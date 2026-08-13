import { describe, expect, it } from "vitest";
import { publicMediaURL } from "./publicMedia";

describe("publicMediaURL", () => {
  it("adds the current watermark version to local media", () => {
    expect(publicMediaURL("/api/media/97")).toBe("/api/media/97?wm=7");
    expect(publicMediaURL("/api/media/97?size=large#photo")).toBe("/api/media/97?size=large&wm=7#photo");
  });

  it("replaces an old watermark version and leaves unrelated URLs alone", () => {
    expect(publicMediaURL("/api/media/97?wm=4")).toBe("/api/media/97?wm=7");
    expect(publicMediaURL("/uploads/product.jpg")).toBe("/uploads/product.jpg");
  });
});
