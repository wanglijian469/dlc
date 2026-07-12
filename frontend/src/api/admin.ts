import { adminClient, publicClient } from "./client";
import type { Banner, Category, ContentPageRecord, FriendLink, Menu, PageResult, Product, ProductSubmission, ProductSupplier, SiteConfig, Tag, Vendor, VendorProductRecord } from "../types/api";

export interface LoginResponse {
  token: string;
  username: string;
  role: "admin" | "vendor";
  vendorId?: number;
}

export interface CMSUser {
  id: number;
  username: string;
  role: "admin" | "vendor";
  vendorId?: number;
  vendor?: Vendor;
  isEnabled: boolean;
}

export interface VendorSubmission {
  id: number;
  vendorId: number;
  vendor: Vendor;
  draft: Vendor;
  status: "pending" | "approved" | "rejected" | "superseded";
  reviewNote?: string;
  submittedBy?: string;
  reviewedBy?: string;
  createdAt: string;
  updatedAt: string;
	baseVersion: number;
}

export interface VendorProfileResponse {
  vendor: Vendor;
  draft: Vendor;
  submission: VendorSubmission | null;
}

export function login(username: string, password: string) {
  return publicClient.post<never, LoginResponse>("/api/admin/login", { username, password });
}

export function getDashboardStats() {
  return adminClient.get<never, { vendors: number; products: number; pendingReviews: number; pendingProductReviews: number; missingImages: number }>("/api/admin/dashboard");
}

export interface OperationLog { id: number; username: string; action: string; resource: string; recordId: number; createdAt: string }
export function listOperationLogs(params: { page: number; pageSize: number; search?: string }) { return adminClient.get<never, PageResult<OperationLog>>("/api/admin/operation-logs", { params }); }

export type ResourceName = "menus" | "vendors" | "tags" | "categories" | "products" | "banners" | "pages" | "friend-links";
export type ResourceRecord = (Menu | Vendor | Tag | Category | Product | Banner | ContentPageRecord | FriendLink) & { id: number };

export function listResource<T extends ResourceRecord>(resource: ResourceName) {
  return adminClient.get<never, T[]>(`/api/admin/${resource}`);
}

export function listResourcePage<T extends ResourceRecord>(resource: "vendors" | "products", params: { page: number; pageSize: number; search?: string; status?: string; publicationStatus?: string; province?: string; vendorId?: number; categoryId?: number }) {
  return adminClient.get<never, PageResult<T>>(`/api/admin/${resource}`, { params });
}

export function createResource<T extends ResourceRecord>(resource: ResourceName, payload: Partial<T>) {
  return adminClient.post<never, T>(`/api/admin/${resource}`, payload);
}

export function updateResource<T extends ResourceRecord>(resource: ResourceName, id: number, payload: Partial<T>) {
  return adminClient.put<never, T>(`/api/admin/${resource}/${id}`, payload);
}

export function deleteResource(resource: ResourceName, id: number) {
  return adminClient.delete<never, { deleted: boolean }>(`/api/admin/${resource}/${id}`);
}

export function listConfigs() {
  return adminClient.get<never, SiteConfig[]>("/api/admin/configs");
}

export function updateConfig(key: string, payload: Partial<SiteConfig>) {
  return adminClient.put<never, SiteConfig>(`/api/admin/configs/${key}`, payload);
}

export function uploadFile(file: File) {
  const form = new FormData();
  form.append("file", file);
  return adminClient.post<never, { assetId: number; status: "staged" | "published"; url: string; previewUrl: string; width: number; height: number; size: number; mime: string; sha256: string }>("/api/admin/uploads", form, { headers: { "Content-Type": "multipart/form-data" } });
}

export function getVendorProfile() {
  return adminClient.get<never, VendorProfileResponse>("/api/admin/vendor-profile");
}

export function submitVendorProfile(payload: Partial<Vendor>) {
  return adminClient.put<never, VendorSubmission>("/api/admin/vendor-profile", payload);
}

export function listOwnProducts() {
  return adminClient.get<never, VendorProductRecord[]>("/api/admin/vendor-products");
}

export function createOwnProduct(payload: Partial<Product>) {
  return adminClient.post<never, VendorProductRecord>("/api/admin/vendor-products", payload);
}

export function updateOwnProduct(id: number, payload: Partial<ProductSupplier>) {
  return adminClient.put<never, VendorProductRecord>(`/api/admin/vendor-products/${id}`, payload);
}

export function deleteOwnProduct(id: number) {
  return adminClient.delete<never, { deleted: boolean }>(`/api/admin/vendor-products/${id}`);
}

export function searchVendorProductCatalog(keyword = "") {
  return adminClient.get<never, Product[]>("/api/admin/vendor-product-catalog", { params: { keyword } });
}

export function linkOwnProduct(productId: number, payload: Partial<ProductSupplier>) {
  return adminClient.post<never, VendorProductRecord>("/api/admin/vendor-products/link", { ...payload, productId });
}

export function listProductSubmissions(params: { page: number; pageSize: number; status?: string }) {
  return adminClient.get<never, PageResult<ProductSubmission>>("/api/admin/product-submissions", { params });
}

export function reviewProductSubmission(id: number, status: "approved" | "rejected", reviewNote: string, mergeProductId?: number) {
  return adminClient.put<never, ProductSubmission>(`/api/admin/product-submissions/${id}/review`, { status, reviewNote, mergeProductId });
}

export function listProductSuppliers(productId: number) {
  return adminClient.get<never, ProductSupplier[]>(`/api/admin/products/${productId}/suppliers`);
}

export function saveProductSupplier(productId: number, payload: Partial<ProductSupplier>) {
  return adminClient.post<never, ProductSupplier>(`/api/admin/products/${productId}/suppliers`, payload);
}

export function disableProductSupplier(productId: number, supplierId: number) {
  return adminClient.delete<never, { disabled: boolean }>(`/api/admin/products/${productId}/suppliers/${supplierId}`);
}

export function mergeProducts(targetProductId: number, sourceProductId: number) {
  return adminClient.put<never, { merged: boolean }>(`/api/admin/products/${targetProductId}/merge`, { sourceProductId });
}

export function listVendorSubmissions(status = "") {
  return adminClient.get<never, VendorSubmission[]>("/api/admin/vendor-submissions", { params: status ? { status } : undefined });
}

export function listVendorSubmissionPage(params: { page: number; pageSize: number; status?: string }) {
  return adminClient.get<never, PageResult<VendorSubmission>>("/api/admin/vendor-submissions", { params });
}

export function reviewVendorSubmission(id: number, status: "approved" | "rejected", reviewNote: string) {
  return adminClient.put<never, VendorSubmission>(`/api/admin/vendor-submissions/${id}/review`, { status, reviewNote });
}

export function listCMSUsers() {
  return adminClient.get<never, CMSUser[]>("/api/admin/users");
}

export function listCMSUserPage(params: { page: number; pageSize: number; search?: string; role?: string }) {
  return adminClient.get<never, PageResult<CMSUser>>("/api/admin/users", { params });
}

export function createCMSUser(payload: { username: string; password: string; role: "admin" | "vendor"; vendorId?: number; isEnabled: boolean }) {
  return adminClient.post<never, CMSUser>("/api/admin/users", payload);
}

export function updateCMSUser(id: number, payload: { username: string; password?: string; role: "admin" | "vendor"; vendorId?: number; isEnabled: boolean }) {
  return adminClient.put<never, CMSUser>(`/api/admin/users/${id}`, payload);
}

export function deleteCMSUser(id: number) {
  return adminClient.delete<never, { deleted: boolean }>(`/api/admin/users/${id}`);
}
