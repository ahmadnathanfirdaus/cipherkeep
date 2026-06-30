import { useContext } from "react";
import { AuthContext, type AuthContextValue } from "@/context/auth-context";

export const useAuth = (): AuthContextValue => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
};
