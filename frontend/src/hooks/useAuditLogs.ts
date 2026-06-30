import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/lib/queryKeys";
import { auditService, type AuditQuery } from "@/services/audit.service";

export const useAuditLogs = (
  projectId: string | undefined,
  query: AuditQuery = {},
) =>
  useQuery({
    queryKey: queryKeys.projects.audit(projectId ?? "", query),
    queryFn: () => auditService.list(projectId as string, query),
    enabled: Boolean(projectId),
  });
