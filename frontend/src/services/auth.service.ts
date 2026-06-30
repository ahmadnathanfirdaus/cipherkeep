import { api } from "@/lib/api";
import type {
  ChangePasswordRequest,
  DataResponse,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  User,
} from "@/types";

export const authService = {
  async register(payload: RegisterRequest): Promise<User> {
    const res = await api.post<DataResponse<{ user: User }>>(
      "/auth/register",
      payload,
    );
    return res.data.data.user;
  },

  async login(payload: LoginRequest): Promise<LoginResponse> {
    const res = await api.post<DataResponse<LoginResponse>>(
      "/auth/login",
      payload,
    );
    return res.data.data;
  },

  async logout(refreshToken: string): Promise<void> {
    await api.post("/auth/logout", { refresh_token: refreshToken });
  },

  async me(): Promise<User> {
    const res = await api.get<DataResponse<{ user: User }>>("/auth/me");
    return res.data.data.user;
  },

  async changePassword(payload: ChangePasswordRequest): Promise<void> {
    await api.post("/auth/change-password", payload);
  },
};
