import type { User } from "@/types/user";

/**
 * Token storage strategy.
 *
 * Tokens are kept in localStorage so a refresh survives reloads. The axios
 * interceptor reads the access token from here; AuthContext owns mutation.
 * (For a higher-security posture the access token could live in memory only,
 * but the refresh flow requires durable storage of the refresh token.)
 */
const ACCESS_TOKEN_KEY = "sm.access_token";
const REFRESH_TOKEN_KEY = "sm.refresh_token";
const USER_KEY = "sm.user";

export const tokenStore = {
  getAccessToken(): string | null {
    return localStorage.getItem(ACCESS_TOKEN_KEY);
  },
  getRefreshToken(): string | null {
    return localStorage.getItem(REFRESH_TOKEN_KEY);
  },
  setTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  },
  getUser(): User | null {
    const raw = localStorage.getItem(USER_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as User;
    } catch {
      return null;
    }
  },
  setUser(user: User): void {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  },
  clear(): void {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  },
} as const;
