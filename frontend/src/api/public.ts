import { publicClient } from "./client";
import type { ContentPageRecord, FilterOptions, FriendLink, HomePayload, LayoutConfig, PageResult, Product, ProductSupplier, SearchPayload, SiteMeta, Vendor } from "../types/api";

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

export function getFriendLinks() {
  return publicClient.get<never, FriendLink[]>("/api/friend-links");
}

export function getVendor(id: string) {
  return publicClient.get<never, Vendor>(`/api/vendors/${id}`);
}

export function getProduct(id: string) {
  return publicClient.get<never, Product>(`/api/products/${id}`);
}

export function getProductSuppliers(id: string) {
  return publicClient.get<never, ProductSupplier[]>(`/api/products/${id}/suppliers`);
}

export interface VendorListParams {
  keyword?: string;
  province?: string;
  tagId?: number | string;
  categoryId?: number | string;
  sort?: "recommended" | "latest";
  page?: number;
  pageSize?: number;
}

export interface ProductListParams {
  keyword?: string;
  categoryId?: number | string;
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
