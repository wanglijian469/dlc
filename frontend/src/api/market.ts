import { publicClient } from "./client";
import type { AuctionBid, AuctionInput, MarketPost, MarketPostContact, MarketPostInput, PageResult, ProcurementAuction, UserNotification } from "../types/api";

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

export function listAuctions(params: { keyword?: string } = {}) {
  return publicClient.get<never, ProcurementAuction[]>("/api/v1/auctions", { params });
}

export function getAuction(id: string | number) {
  return publicClient.get<never, ProcurementAuction>("/api/v1/auctions/" + id);
}

export function listOwnAuctions() {
  return publicClient.get<never, ProcurementAuction[]>("/api/v1/me/auctions");
}

export function getOwnAuction(id: string | number) {
  return publicClient.get<never, { auction: ProcurementAuction; bids: AuctionBid[] }>("/api/v1/me/auctions/" + id);
}

export function createAuction(payload: AuctionInput) {
  return publicClient.post<never, ProcurementAuction>("/api/v1/me/auctions", payload);
}

export function updateAuction(id: number, payload: AuctionInput) {
  return publicClient.put<never, ProcurementAuction>("/api/v1/me/auctions/" + id, payload);
}

export function publishAuction(id: number) {
  return publicClient.post<never, { status: string }>("/api/v1/me/auctions/" + id + "/publish");
}

export function cancelAuction(id: number, reason = "") {
  return publicClient.post<never, { status: string }>("/api/v1/me/auctions/" + id + "/cancel", { reason });
}

export function awardAuction(id: number, bidId: number) {
  return publicClient.post<never, { status: string; awardedBidId: number }>("/api/v1/me/auctions/" + id + "/award", { bidId });
}

export function placeAuctionBid(id: number, payload: { unitPriceCents: number; taxIncluded: boolean; freightNote?: string; deliveryDays: number; supplyNote?: string; promiseText?: string; expectedVersion: number }) {
  return publicClient.post<never, { bid: AuctionBid; auctionVersion: number; endAt: string; extended: boolean }>("/api/v1/auctions/" + id + "/bids", payload);
}

export function listNotifications(unread = false) {
  return publicClient.get<never, { items: UserNotification[]; unreadCount: number }>("/api/v1/me/notifications", { params: { unread } });
}

export function readNotification(id: number) {
  return publicClient.put<never, { read: boolean }>("/api/v1/me/notifications/" + id + "/read");
}

export function readAllNotifications() {
  return publicClient.put<never, { read: boolean }>("/api/v1/me/notifications/read-all");
}
