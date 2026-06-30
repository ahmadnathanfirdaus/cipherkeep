/**
 * Centralized, typed access to Vite environment variables.
 */
const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";

const rawBaseUrl = import.meta.env.VITE_API_BASE_URL;

export const env = {
  apiBaseUrl:
    typeof rawBaseUrl === "string" && rawBaseUrl.length > 0
      ? rawBaseUrl.replace(/\/$/, "")
      : DEFAULT_API_BASE_URL,
  isDev: import.meta.env.DEV,
} as const;
