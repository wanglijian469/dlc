import axios from "axios";
import type { ApiResponse } from "../types/api";

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8080";

export const publicClient = axios.create({
  baseURL: API_BASE_URL,
});

export const adminClient = axios.create({
  baseURL: API_BASE_URL,
});

publicClient.interceptors.response.use((response) => unwrap(response.data));

adminClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("admin_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

adminClient.interceptors.response.use(
  (response) => unwrap(response.data),
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("admin_token");
      if (window.location.pathname !== "/admin/login") {
        window.location.href = "/admin/login";
      }
    }
    return Promise.reject(error);
  },
);

function unwrap<T>(payload: ApiResponse<T>): T {
  if (payload.code !== 0) {
    throw new Error(payload.message || "Request failed");
  }
  return payload.data;
}
