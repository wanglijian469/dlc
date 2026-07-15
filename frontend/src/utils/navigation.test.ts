import { describe, expect, it } from "vitest";
import { getMenuLabel, getSearchPath, getVendorEntryTarget } from "./navigation";

describe("navigation helpers", () => {
  it("creates encoded search paths", () => {
    expect(getSearchPath(" 液压油泵 ")).toBe("/search?keyword=%E6%B6%B2%E5%8E%8B%E6%B2%B9%E6%B3%B5");
  });

  it("does not create a search path for empty keywords", () => {
    expect(getSearchPath("   ")).toBe(null);
  });

  it("uses the label configured for a public menu path", () => {
    expect(getMenuLabel([{ id: 1, name: "配件货源", path: "/products" }], "/products", "配件产品")).toBe("配件货源");
    expect(getMenuLabel([], "/vendors", "厂商目录")).toBe("厂商目录");
  });

  it("prefers vendor website when present", () => {
    expect(getVendorEntryTarget({ id: 7, websiteUrl: "https://example.com" })).toEqual({
      type: "external",
      href: "https://example.com",
    });
  });

  it("falls back to internal vendor detail", () => {
    expect(getVendorEntryTarget({ id: 7, websiteUrl: "" })).toEqual({
      type: "internal",
      href: "/vendors/7",
    });
  });
});
