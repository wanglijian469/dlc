import axios, { type InternalAxiosRequestConfig } from "axios";
import type { ApiResponse } from "../types/api";

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export const publicClient = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
});

export const adminClient = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
});

const csrfRequest = (config: InternalAxiosRequestConfig) => {
  const method = (config.method || "get").toLowerCase();
  if (!["get", "head", "options"].includes(method)) {
    const token = document.cookie.split("; ").find((item) => item.startsWith("dlc_csrf="))?.split("=").slice(1).join("=");
    if (token) config.headers.set("X-CSRF-Token", decodeURIComponent(token));
  }
  return config;
};

publicClient.interceptors.request.use(csrfRequest);

publicClient.interceptors.response.use((response) => unwrap(response.data));

adminClient.interceptors.request.use(csrfRequest);

adminClient.interceptors.response.use(
  (response) => unwrap(response.data),
  (error) => {
    if (error.response?.status === 401) {
      const vendorRoute = localStorage.getItem("cms_role") === "vendor" || /\/admin\/(vendor-|capture)/.test(window.location.pathname);
      localStorage.removeItem("cms_authenticated");
      localStorage.removeItem("cms_role");
      localStorage.removeItem("cms_username");
      if (window.location.pathname !== "/admin/login") {
        window.location.href = (vendorRoute ? "/account/login" : "/admin/login") + "?returnTo=" + encodeURIComponent(window.location.pathname + window.location.search);
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

export function getApiErrorMessage(error: unknown, fallback: string) {
  const response = (error as { response?: { data?: { message?: string } } })?.response;
  const message = (error as { message?: string })?.message;
  return response?.data?.message || message || fallback;
}
