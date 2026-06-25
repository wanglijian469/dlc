import { adminClient, publicClient } from "./client";
import type { Banner, Category, Menu, Product, SiteConfig, Tag, Vendor } from "../types/api";

export interface LoginResponse {
  token: string;
  username: string;
}

export function login(username: string, password: string) {
  return publicClient.post<never, LoginResponse>("/api/admin/login", { username, password });
}

export type ResourceName = "menus" | "vendors" | "tags" | "categories" | "products" | "banners";
export type ResourceRecord = (Menu | Vendor | Tag | Category | Product | Banner) & { id: number };

export function listResource<T extends ResourceRecord>(resource: ResourceName) {
  return adminClient.get<never, T[]>(`/api/admin/${resource}`);
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

