import { api } from "@/lib/api";
import type {
  CollectionResponse,
  CreateEnvironmentRequest,
  DataResponse,
  Environment,
} from "@/types";

export const environmentService = {
  async listByProject(projectId: string): Promise<Environment[]> {
    const res = await api.get<CollectionResponse<Environment>>(
      `/projects/${projectId}/environments`,
    );
    return res.data.data;
  },

  async create(
    projectId: string,
    payload: CreateEnvironmentRequest,
  ): Promise<Environment> {
    const res = await api.post<DataResponse<{ environment: Environment }>>(
      `/projects/${projectId}/environments`,
      payload,
    );
    return res.data.data.environment;
  },

  async remove(projectId: string, environmentId: string): Promise<void> {
    await api.delete(`/projects/${projectId}/environments/${environmentId}`);
  },
};
