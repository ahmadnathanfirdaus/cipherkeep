import { api } from "@/lib/api";
import type {
  AuditLog,
  CollectionResponse,
  PaginationMeta,
  PaginationParams,
} from "@/types";

export interface AuditQuery extends PaginationParams {
  action?: string;
}

export interface AuditResult {
  logs: AuditLog[];
  meta?: PaginationMeta;
}

export const auditService = {
  async list(projectId: string, query: AuditQuery = {}): Promise<AuditResult> {
    const res = await api.get<CollectionResponse<AuditLog>>(
      `/projects/${projectId}/audit-logs`,
      { params: query },
    );
    return { logs: res.data.data, meta: res.data.meta };
  },
};
