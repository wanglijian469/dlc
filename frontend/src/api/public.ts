import { publicClient } from "./client";
import type { ContentPageRecord, FilterOptions, FriendLink, HomePayload, LayoutConfig, PageResult, Product, ProductSupplier, SearchPayload, SiteMeta, Vendor, VendorCategory } from "../types/api";
import { API_BASE_URL } from "./client";

export function getHome() {
  return publicClient.get<never, HomePayload>("/api/home");
}

export function getSiteMeta() {
  return publicClient.get<never, SiteMeta>("/api/site-meta");
}

export function getLayoutConfig() {
  return publicClient.get<never, LayoutConfig>("/api/layout-config");
}

export function getPage(slug: string) {
  return publicClient.get<never, ContentPageRecord>(`/api/pages/${slug}`);
}

export function listArticles(page = 1, pageSize = 12, keyword = "") {
  return publicClient.get<never, PageResult<ContentPageRecord>>("/api/articles", { params: { page, pageSize, keyword: keyword || undefined } });
}

export function getArticle(slug: string) {
  return publicClient.get<never, ContentPageRecord>(`/api/articles/${slug}`);
}

export function getFriendLinks() {
  return publicClient.get<never, FriendLink[]>("/api/friend-links");
}

export function getVendor(slug: string) {
  const path = /^\d+$/.test(slug) ? `/api/vendors/${slug}` : `/api/vendors/slug/${encodeURIComponent(slug)}`;
  return publicClient.get<never, Vendor>(path);
}

export interface VendorContact {
  vendorId: number;
  phone?: string;
  wechat?: string;
  wechatQrCodeUrl?: string;
  contactName?: string;
}

export function getVendorContact(vendorId: number) {
  return publicClient.get<never, VendorContact>(`/api/vendors/${vendorId}/contact`);
}

export async function getVendorContactQRCode(vendorId: number) {
  const response = await fetch(`${API_BASE_URL}/api/vendors/${vendorId}/contact-qr`, { credentials: "include" });
  if (!response.ok) throw new Error("二维码加载失败");
  return response.blob();
}

export function getProduct(slug: string) {
  const path = /^\d+$/.test(slug) ? `/api/products/${slug}` : `/api/products/slug/${encodeURIComponent(slug)}`;
  return publicClient.get<never, Product>(path);
}

export function getProductSuppliers(id: string) {
  return publicClient.get<never, ProductSupplier[]>(`/api/products/${id}/suppliers`);
}

export interface VendorListParams {
  keyword?: string;
  province?: string;
  tagId?: number | string;
  categoryId?: number | string;
	vendorCategoryId?: number | string;
	newlyJoined?: boolean;
  sort?: "recommended" | "latest";
  page?: number;
  pageSize?: number;
}

export interface ProductListParams {
  keyword?: string;
  categoryId?: number | string;
  categorySlug?: string;
  vendorId?: number | string;
  hot?: boolean;
  recommended?: boolean;
  sort?: "recommended" | "latest";
  page?: number;
  pageSize?: number;
}

export function listVendors(params: VendorListParams = {}) {
  return publicClient.get<never, PageResult<Vendor>>("/api/vendors", { params });
}

export function getVendorCategories() {
	return publicClient.get<never, VendorCategory[]>("/api/vendor-categories");
}

export function listProcessingVendors(params: VendorListParams = {}) {
  return publicClient.get<never, PageResult<Vendor>>("/api/processing-vendors", { params });
}

export function listProducts(params: ProductListParams = {}) {
  return publicClient.get<never, PageResult<Product>>("/api/products", { params });
}

export function getFilterOptions() {
  return publicClient.get<never, FilterOptions>("/api/filter-options");
}

export function getProcessingFilterOptions() {
  return publicClient.get<never, FilterOptions>("/api/processing-filter-options");
}

export function search(keyword: string, page = 1, pageSize = 10) {
  return publicClient.get<never, SearchPayload>("/api/search", { params: { keyword, page, pageSize } });
}
