export function getSearchPath(keyword: string): string | null {
  const value = keyword.trim();
  if (!value) return null;
  return `/search?keyword=${encodeURIComponent(value)}`;
}

export function getVendorEntryTarget(vendor: { id: number; websiteUrl?: string | null }) {
  const websiteUrl = vendor.websiteUrl?.trim();
  if (websiteUrl) {
    return { type: "external" as const, href: websiteUrl };
  }
  return { type: "internal" as const, href: `/vendors/${vendor.id}` };
}
