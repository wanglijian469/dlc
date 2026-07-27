import { adminClient, publicClient } from "./client";
import type { Banner, Category, ContentPageRecord, FriendLink, Menu, PageResult, Product, ProductSubmission, ProductSupplier, SiteConfig, Tag, Vendor, VendorOption, VendorProductRecord } from "../types/api";

export type AccountRole = "admin" | "editor" | "reviewer" | "vendor";

export interface LoginResponse {
  username: string;
  role: AccountRole;
  vendorId?: number;
  csrfToken?: string;
}

export interface CMSUser {
  id: number;
  username: string;
  role: AccountRole;
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

export interface VendorSEOSuggestion {
  seoTitle: string;
  seoDescription: string;
  sourceFields: string[];
}

export function suggestVendorSEO(payload: Partial<Vendor>) {
  return adminClient.post<never, VendorSEOSuggestion>("/api/admin/seo-suggestions/vendor", payload);
}

export function login(username: string, password: string) {
  return publicClient.post<never, LoginResponse>("/api/auth/login", { username, password });
}

export function staffLogin(username: string, password: string) {
  return publicClient.post<never, LoginResponse>("/api/admin/login", { username, password });
}

export function register(payload: { username: string; password: string; role: "vendor"; companyName: string }) {
  return publicClient.post<never, LoginResponse>("/api/auth/register", payload);
}

export function getCurrentSession() {
  return adminClient.get<never, LoginResponse>("/api/auth/session");
}

export function logoutSession() {
  return adminClient.post<never, { loggedOut: boolean }>("/api/auth/logout");
}

export function changePassword(currentPassword: string, newPassword: string) {
  return adminClient.put<never, { changed: boolean; reauthenticate: boolean }>("/api/auth/password", { currentPassword, newPassword });
}

export function getDashboardStats() {
  return adminClient.get<never, { vendors: number; products: number; pendingReviews: number; pendingProductReviews: number; missingImages: number }>("/api/admin/dashboard");
}

export interface AnalyticsSummary {
  days: number;
  pv: number;
  uv: number;
  conversions: number;
  conversionRate: number;
  trend: Array<{ date: string; pv: number; uv: number }>;
  provinces: Array<{ label: string; count: number }>;
  contents: Array<{ path: string; contentType: string; contentId: number; count: number }>;
  conversionEvents: Array<{ label: string; count: number }>;
}

export function getAnalyticsSummary(days: 7 | 30 | 90) {
  return adminClient.get<never, AnalyticsSummary>("/api/admin/analytics", { params: { days } });
}

export interface VendorAnalyticsSummary {
  days: number;
  vendor: {
    id: number;
    name: string;
    pv: number;
    uv: number;
    contacts: number;
    contactEvents: Array<{ label: string; count: number }>;
    trend: Array<{ date: string; pv: number; uv: number }>;
  };
  products: {
    pv: number;
    uv: number;
    trend: Array<{ date: string; pv: number; uv: number }>;
    items: Array<{ productId: number; name: string; path: string; pv: number; uv: number }>;
  };
}

export function getVendorAnalytics(days: 7 | 30 | 90) {
  return adminClient.get<never, VendorAnalyticsSummary>("/api/admin/vendor-analytics", { params: { days } });
}

export interface SEOStatus {
  issues: { missingTitle: number; missingDescription: number; missingImage: number; duplicateTitles: number; duplicateDescriptions: number; drafts: number; inReview: number; redirects: number; recent404: number; zeroResultSearches: number };
  sitemap: { index: string; sections: string[] };
  baiduSubmission: { enabled: boolean; configured: boolean };
}

export interface ContentRevision {
  id: number;
  resourceType: "vendor" | "product" | "category" | "page" | "article";
  resourceId: number;
  version: number;
  baseVersion: number;
  status: "draft" | "in_review" | "scheduled" | "published" | "rejected" | "archived";
  snapshot: string;
  authorUsername: string;
  reviewer?: string;
  reviewNote?: string;
  scheduledAt?: string;
  publishedAt?: string;
  updatedAt: string;
}

export function getSEOStatus() { return adminClient.get<never, SEOStatus>("/api/admin/seo-status"); }
export function submitBaiduURLs() { return adminClient.post<never, { submitted: number; response: string }>("/api/admin/seo-submit/baidu"); }
export function listRevisions(params: { page: number; pageSize: number; status?: string; resourceType?: string; resourceId?: number }) { return adminClient.get<never, PageResult<ContentRevision>>("/api/admin/revisions", { params }); }
export function saveRevision(payload: { resourceType: ContentRevision["resourceType"]; resourceId: number; baseVersion: number; snapshot: unknown }) { return adminClient.post<never, ContentRevision>("/api/admin/revisions", payload); }
export function submitRevision(id: number) { return adminClient.post<never, ContentRevision>(`/api/admin/revisions/${id}/submit`); }
export function approveRevision(id: number, note = "", scheduledAt?: string) { return adminClient.post<never, ContentRevision>(`/api/admin/revisions/${id}/approve`, { note, scheduledAt }); }
export function rejectRevision(id: number, note: string) { return adminClient.post<never, ContentRevision>(`/api/admin/revisions/${id}/reject`, { note }); }
export function archiveRevision(id: number) { return adminClient.post<never, { archived: boolean }>(`/api/admin/revisions/${id}/archive`); }
export function previewRevision(id: number) { return adminClient.get<never, { revision: ContentRevision; snapshot: unknown; baseSnapshot: unknown }>(`/api/admin/revisions/${id}/preview`); }

export interface OperationLog { id: number; username: string; action: string; resource: string; recordId: number; reason?: string; createdAt: string }
export function listOperationLogs(params: { page: number; pageSize: number; search?: string }) { return adminClient.get<never, PageResult<OperationLog>>("/api/admin/operation-logs", { params }); }

export type ResourceName = "menus" | "vendors" | "tags" | "categories" | "products" | "banners" | "pages" | "friend-links";
export type ResourceRecord = (Menu | Vendor | Tag | Category | Product | Banner | ContentPageRecord | FriendLink) & { id: number };

const editorialResources = new Set<ResourceName>(["vendors", "products", "categories", "pages", "tags"]);

function resourceListPath(resource: ResourceName) {
  const role = typeof window === "undefined" ? "admin" : window.localStorage.getItem("cms_role");
  return role && role !== "admin" && editorialResources.has(resource)
    ? `/api/admin/editorial/${resource}`
    : `/api/admin/${resource}`;
}

export function listResource<T extends ResourceRecord>(resource: ResourceName) {
  return adminClient.get<never, T[]>(resourceListPath(resource));
}

export function listResourcePage<T extends ResourceRecord>(resource: "vendors" | "products", params: { page: number; pageSize: number; search?: string; status?: string; publicationStatus?: string; province?: string; vendorId?: number; categoryId?: number; associationStatus?: "linked" | "unlinked" }) {
  return adminClient.get<never, PageResult<T>>(resourceListPath(resource), { params });
}

export function listVendorOptions(params: { page?: number; pageSize?: number; search?: string }) {
  const role = typeof window === "undefined" ? "admin" : window.localStorage.getItem("cms_role");
  const path = role && role !== "admin" ? "/api/admin/editorial/vendor-options" : "/api/admin/vendor-options";
  return adminClient.get<never, PageResult<VendorOption>>(path, { params: { page: 1, pageSize: 20, ...params } });
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

export function uploadFile(file: File, metadata: { altText?: string; caption?: string } = {}) {
  const form = new FormData();
  form.append("file", file);
  if (metadata.altText) form.append("altText", metadata.altText);
  if (metadata.caption) form.append("caption", metadata.caption);
  return adminClient.post<never, { assetId: number; status: "staged" | "published"; url: string; previewUrl: string; width: number; height: number; size: number; mime: string; sha256: string }>("/api/admin/uploads", form, { headers: { "Content-Type": "multipart/form-data" } });
}

export function downloadRemoteImage(url: string) {
  return adminClient.post<never, { assetId: number; status: "staged" | "published"; url: string; previewUrl: string; width: number; height: number; size: number; mime: string; sha256: string }>("/api/admin/remote-images", { url });
}

export interface BulkImportIssue { sheet: string; row: number; message: string }
export interface BulkImportResult {
  resource: "vendors" | "products";
  totalRows: number;
  created: number;
  updated: number;
  relationsCreated: number;
  relationsUpdated: number;
  imported: boolean;
  issues: BulkImportIssue[];
  warnings?: BulkImportIssue[];
}

export function importWorkbook(resource: "vendors" | "products", file: File) {
  const form = new FormData();
  form.append("file", file);
  return adminClient.post<never, BulkImportResult>(`/api/admin/imports/${resource}`, form, { headers: { "Content-Type": "multipart/form-data" } });
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

export function deleteProductSupplier(productId: number, supplierId: number) {
	return adminClient.delete<never, { deleted: boolean }>(`/api/admin/products/${productId}/suppliers/${supplierId}`);
}

export function batchSaveProductSuppliers(productIds: number[], vendorId: number) {
  return adminClient.post<never, { created: number; existing: number; total: number }>("/api/admin/product-suppliers/batch", { productIds, vendorId });
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

export function createCMSUser(payload: { username: string; password: string; role: AccountRole; vendorId?: number; companyName?: string; isEnabled: boolean }) {
  return adminClient.post<never, CMSUser>("/api/admin/users", payload);
}

export function updateCMSUser(id: number, payload: { username: string; password?: string; role: AccountRole; vendorId?: number; isEnabled: boolean }) {
  return adminClient.put<never, CMSUser>(`/api/admin/users/${id}`, payload);
}

export function deleteCMSUser(id: number) {
  return adminClient.delete<never, { deleted: boolean }>(`/api/admin/users/${id}`);
}
