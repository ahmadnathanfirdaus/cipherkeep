import { api } from "@/lib/api";
import type {
  CollectionResponse,
  CreateProjectRequest,
  DataResponse,
  Project,
  UpdateProjectRequest,
} from "@/types";

export const projectService = {
  async list(): Promise<Project[]> {
    const res = await api.get<CollectionResponse<Project>>("/projects");
    return res.data.data;
  },

  async get(projectId: string): Promise<Project> {
    const res = await api.get<DataResponse<{ project: Project }>>(
      `/projects/${projectId}`,
    );
    return res.data.data.project;
  },

  async create(payload: CreateProjectRequest): Promise<Project> {
    const res = await api.post<DataResponse<{ project: Project }>>(
      "/projects",
      payload,
    );
    return res.data.data.project;
  },

  async update(
    projectId: string,
    payload: UpdateProjectRequest,
  ): Promise<Project> {
    const res = await api.patch<DataResponse<{ project: Project }>>(
      `/projects/${projectId}`,
      payload,
    );
    return res.data.data.project;
  },

  async remove(projectId: string): Promise<void> {
    await api.delete(`/projects/${projectId}`);
  },
};
