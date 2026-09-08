import { adminClient, publicClient } from "./client";
import type { PageResult, Product, ProductSupplier, Vendor } from "../types/api";
export interface WorkDraft<T = Record<string, unknown>> {
 id: number; clientKey: string; kind: "product" | "profile"; targetType: string; targetId: number;
 version: number; payload: T; submissionId: number; committedAt?: string; updatedAt: string;
}
export interface WorkspaceSummary {
 vendor: Vendor; counts: Record<string, number>; draftCount: number; profileDraftCount?: number; unreadCount: number;
 profileStatus: string; profileReviewNote: string;
}
export const listWorkDrafts = <T,>() => adminClient.get<never, WorkDraft<T>[]>("/api/admin/vendor-work-drafts");
export const saveWorkDraft = <T,>(key: string, input: { kind: string; targetType: string; targetId: number; version: number; payload: T }) => adminClient.put<never, WorkDraft<T>>("/api/admin/vendor-work-drafts/" + key, input);
export const commitWorkDraft = (id: number, version: number) => adminClient.post<never, WorkDraft>("/api/admin/vendor-work-drafts/" + id + "/commit", { version });
export const deleteWorkDraft = (row: WorkDraft<unknown>) => adminClient.delete("/api/admin/vendor-work-drafts/" + row.id, { params: { version: row.version } });
export const getWorkspace = () => adminClient.get<never, WorkspaceSummary>("/api/admin/vendor-workspace");
export const updateShowroomOrder = (id: number, featured: boolean, order: number) => adminClient.put("/api/admin/vendor-products/" + id + "/showroom", { featured, order });
export const getShowroomProduct = (slug: string, id: string) => publicClient.get<never, ProductSupplier>("/api/showrooms/" + encodeURIComponent(slug) + "/products/" + encodeURIComponent(id));
export const getShowroomProducts = (slug: string | number, params: { keyword?: string; categoryId?: number; page?: number; pageSize?: number } = {}) => publicClient.get<never, PageResult<Product>>("/api/showrooms/" + encodeURIComponent(slug) + "/products", { params });
