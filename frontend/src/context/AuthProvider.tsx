import { useCallback, useEffect, useMemo, useState } from "react";
import { AUTH_LOGOUT_EVENT } from "@/lib/api";
import { tokenStore } from "@/lib/tokenStore";
import { queryClient } from "@/lib/queryClient";
import { authService } from "@/services/auth.service";
import { useToast } from "@/hooks/useToast";
import { AuthContext, type AuthContextValue } from "@/context/auth-context";
import type { LoginRequest, RegisterRequest, User } from "@/types";

interface AuthProviderProps {
  children: React.ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const { notify } = useToast();
  const [user, setUser] = useState<User | null>(() => tokenStore.getUser());
  const [isInitializing, setIsInitializing] = useState<boolean>(true);

  const clearSession = useCallback((): void => {
    tokenStore.clear();
    setUser(null);
    queryClient.clear();
  }, []);

  // Validate any persisted session against /auth/me on first mount.
  useEffect(() => {
    let active = true;
    const bootstrap = async (): Promise<void> => {
      if (!tokenStore.getAccessToken()) {
        setIsInitializing(false);
        return;
      }
      try {
        const me = await authService.me();
        if (!active) return;
        tokenStore.setUser(me);
        setUser(me);
      } catch {
        if (active) clearSession();
      } finally {
        if (active) setIsInitializing(false);
      }
    };
    void bootstrap();
    return () => {
      active = false;
    };
  }, [clearSession]);

  // React to forced logout from the axios interceptor (failed refresh).
  useEffect(() => {
    const handleLogout = (): void => {
      const hadSession = tokenStore.getAccessToken() !== null || user !== null;
      clearSession();
      if (hadSession) {
        notify({
          variant: "error",
          title: "Session expired",
          description: "Please sign in again.",
        });
      }
    };
    window.addEventListener(AUTH_LOGOUT_EVENT, handleLogout);
    return () => window.removeEventListener(AUTH_LOGOUT_EVENT, handleLogout);
  }, [clearSession, notify, user]);

  const login = useCallback(async (payload: LoginRequest): Promise<void> => {
    const result = await authService.login(payload);
    tokenStore.setTokens(result.access_token, result.refresh_token);
    tokenStore.setUser(result.user);
    setUser(result.user);
  }, []);

  const register = useCallback(
    async (payload: RegisterRequest): Promise<User> => {
      return authService.register(payload);
    },
    [],
  );

  const logout = useCallback(async (): Promise<void> => {
    const refreshToken = tokenStore.getRefreshToken();
    if (refreshToken) {
      try {
        await authService.logout(refreshToken);
      } catch {
        // Best-effort; clear local state regardless.
      }
    }
    clearSession();
  }, [clearSession]);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: user !== null,
      isInitializing,
      login,
      register,
      logout,
    }),
    [user, isInitializing, login, register, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
