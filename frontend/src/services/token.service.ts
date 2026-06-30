import { api } from "@/lib/api";
import type {
  CollectionResponse,
  CreateServiceTokenRequest,
  CreateServiceTokenResult,
  DataResponse,
  ServiceTokenMeta,
} from "@/types";

export const tokenService = {
  async list(projectId: string): Promise<ServiceTokenMeta[]> {
    const res = await api.get<CollectionResponse<ServiceTokenMeta>>(
      `/projects/${projectId}/tokens`,
    );
    return res.data.data;
  },

  async create(
    projectId: string,
    payload: CreateServiceTokenRequest,
  ): Promise<CreateServiceTokenResult> {
    const res = await api.post<DataResponse<CreateServiceTokenResult>>(
      `/projects/${projectId}/tokens`,
      payload,
    );
    return res.data.data;
  },

  async revoke(projectId: string, tokenId: string): Promise<void> {
    await api.delete(`/projects/${projectId}/tokens/${tokenId}`);
  },
};
