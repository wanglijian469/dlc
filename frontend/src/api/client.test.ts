import { describe, expect, it } from "vitest";
import { API_BASE_URL, adminClient, publicClient } from "./client";

describe("api client base url", () => {
  it("uses same-origin API routes by default for production packages", () => {
    expect(API_BASE_URL).toBe("");
    expect(publicClient.defaults.baseURL).toBe("");
    expect(adminClient.defaults.baseURL).toBe("");
  });
});
