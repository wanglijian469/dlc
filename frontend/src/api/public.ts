import { publicClient } from "./client";
import type { FilterOptions, HomePayload, PageResult, Product, SearchPayload, Vendor } from "../types/api";

export function getHome() {
  return publicClient.get<never, HomePayload>("/api/home");
}

export function getVendor(id: string) {
  return publicClient.get<never, Vendor>(`/api/vendors/${id}`);
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

export function listProducts(params: ProductListParams = {}) {
  return publicClient.get<never, PageResult<Product>>("/api/products", { params });
}

export function getFilterOptions() {
  return publicClient.get<never, FilterOptions>("/api/filter-options");
}

export function search(keyword: string, page = 1, pageSize = 10) {
  return publicClient.get<never, SearchPayload>("/api/search", {
    params: { keyword, page, pageSize },
  });
}
