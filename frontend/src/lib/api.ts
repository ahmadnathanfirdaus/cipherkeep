import axios, {
  type AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios";
import { env } from "@/lib/env";
import { tokenStore } from "@/lib/tokenStore";
import type { AuthTokens, DataResponse } from "@/types";

/**
 * Single configured axios instance for the whole app.
 *
 * Request interceptor: attaches `Authorization: Bearer <access>`.
 * Response interceptor: on 401, attempts ONE `/auth/refresh`, then retries the
 * original request. Concurrent 401s share a single in-flight refresh. If refresh
 * fails, the session is cleared and a logout event is emitted.
 */

/** Dispatched when refresh fails; AuthContext listens and clears state. */
export const AUTH_LOGOUT_EVENT = "sm:auth-logout";

const emitLogout = (): void => {
  window.dispatchEvent(new CustomEvent(AUTH_LOGOUT_EVENT));
};

export const api: AxiosInstance = axios.create({
  baseURL: env.apiBaseUrl,
  headers: { "Content-Type": "application/json" },
});

// A bare client without interceptors, used to perform the refresh call so it
// cannot itself trigger the 401-refresh loop.
const refreshClient: AxiosInstance = axios.create({
  baseURL: env.apiBaseUrl,
  headers: { "Content-Type": "application/json" },
});

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = tokenStore.getAccessToken();
  if (token) {
    config.headers.set("Authorization", `Bearer ${token}`);
  }
  return config;
});

interface RetriableConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
}

let refreshPromise: Promise<string> | null = null;

const performRefresh = async (): Promise<string> => {
  const refreshToken = tokenStore.getRefreshToken();
  if (!refreshToken) {
    throw new Error("No refresh token available");
  }
  const response = await refreshClient.post<DataResponse<AuthTokens>>(
    "/auth/refresh",
    { refresh_token: refreshToken },
  );
  const tokens = response.data.data;
  tokenStore.setTokens(tokens.access_token, tokens.refresh_token);
  return tokens.access_token;
};

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const config = error.config as RetriableConfig | undefined;
    const status = error.response?.status;

    const isRefreshCall = config?.url?.includes("/auth/refresh") ?? false;

    if (status !== 401 || !config || config._retry || isRefreshCall) {
      return Promise.reject(error);
    }

    config._retry = true;

    try {
      // Coalesce concurrent refreshes into a single request.
      refreshPromise ??= performRefresh().finally(() => {
        refreshPromise = null;
      });
      const newAccessToken = await refreshPromise;
      config.headers.set("Authorization", `Bearer ${newAccessToken}`);
      return api.request(config);
    } catch (refreshError) {
      tokenStore.clear();
      emitLogout();
      return Promise.reject(
        refreshError instanceof Error ? refreshError : error,
      );
    }
  },
);
