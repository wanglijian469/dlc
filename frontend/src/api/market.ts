import { publicClient } from "./client";
import type { MarketPost, MarketPostContact, MarketPostInput, PageResult } from "../types/api";

export interface MarketPostListParams {
  type?: "supply" | "demand";
  keyword?: string;
  categoryId?: number;
  province?: string;
  page?: number;
  pageSize?: number;
}

export interface BuyerProfile {
  id?: number;
  userId?: number;
  displayName: string;
  contactName: string;
  phone?: string;
  province: string;
  city: string;
}

export function listMarketPosts(params: MarketPostListParams = {}) {
  return publicClient.get<never, PageResult<MarketPost>>("/api/v1/market-posts", { params });
}

export function getMarketPost(id: string | number) {
  return publicClient.get<never, MarketPost>(`/api/v1/market-posts/${id}`);
}

export function getMarketPostContact(id: string | number) {
  return publicClient.get<never, MarketPostContact>(`/api/v1/market-posts/${id}/contact`);
}

export function listOwnMarketPosts(page = 1) {
  return publicClient.get<never, PageResult<MarketPost>>("/api/v1/me/market-posts", { params: { page, pageSize: 20 } });
}

export function getOwnMarketPost(id: string | number) {
  return publicClient.get<never, { post: MarketPost; contactName: string; contactPhone: string; assetIds: number[] }>(`/api/v1/me/market-posts/${id}`);
}

export function createMarketPost(payload: MarketPostInput) {
  return publicClient.post<never, MarketPost>("/api/v1/me/market-posts", payload);
}

export function updateMarketPost(id: string | number, payload: MarketPostInput) {
  return publicClient.put<never, MarketPost>(`/api/v1/me/market-posts/${id}`, payload);
}

export function withdrawMarketPost(id: string | number) {
  return publicClient.delete<never, { withdrawn: boolean }>(`/api/v1/me/market-posts/${id}`);
}

export function uploadMarketImage(file: File) {
  const form = new FormData();
  form.append("file", file);
  return publicClient.post<never, { assetId: number; url: string; previewUrl: string }>("/api/v1/media", form, { headers: { "Content-Type": "multipart/form-data" } });
}

export function getAccountProfile() {
  return publicClient.get<never, { username: string; role: "buyer" | "vendor"; vendorId?: number; profile?: BuyerProfile }>("/api/v1/me/profile");
}

export function updateBuyerProfile(payload: BuyerProfile) {
  return publicClient.put<never, { username: string; role: "buyer"; profile: BuyerProfile }>("/api/v1/me/profile", payload);
}
