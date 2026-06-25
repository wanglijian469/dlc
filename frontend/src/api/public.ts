import { publicClient } from "./client";
import type { HomePayload, Product, Vendor } from "../types/api";

export function getHome() {
  return publicClient.get<never, HomePayload>("/api/home");
}

export function getVendor(id: string) {
  return publicClient.get<never, Vendor>(`/api/vendors/${id}`);
}

export function search(keyword: string) {
  return publicClient.get<never, { vendors: Vendor[]; products: Product[]; categories: unknown[] }>("/api/search", {
    params: { keyword },
  });
}
