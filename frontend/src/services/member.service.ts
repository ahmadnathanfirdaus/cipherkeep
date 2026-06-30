import { api } from "@/lib/api";
import type {
  AddMemberRequest,
  CollectionResponse,
  Member,
  UpdateMemberRequest,
} from "@/types";

export const memberService = {
  async list(projectId: string): Promise<Member[]> {
    const res = await api.get<CollectionResponse<Member>>(
      `/projects/${projectId}/members`,
    );
    return res.data.data;
  },

  async add(projectId: string, payload: AddMemberRequest): Promise<void> {
    await api.post(`/projects/${projectId}/members`, payload);
  },

  async updateRole(
    projectId: string,
    userId: string,
    payload: UpdateMemberRequest,
  ): Promise<void> {
    await api.patch(`/projects/${projectId}/members/${userId}`, payload);
  },

  async remove(projectId: string, userId: string): Promise<void> {
    await api.delete(`/projects/${projectId}/members/${userId}`);
  },
};
