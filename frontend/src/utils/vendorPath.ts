import type { Vendor } from "../types/api";

/** The public, shareable address for a vendor's branded site. */
export function vendorPath(vendor: Pick<Vendor, "id" | "slug">) {
  return `/v/${vendor.slug || vendor.id}`;
}
