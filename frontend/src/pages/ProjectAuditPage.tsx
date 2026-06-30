import { useState } from "react";
import { useParams } from "react-router-dom";
import { ChevronLeft, ChevronRight, ScrollText } from "lucide-react";
import { ProjectPageShell } from "@/components/projects/ProjectPageShell";
import { ActionBadge } from "@/components/audit/ActionBadge";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { EmptyState } from "@/components/ui/EmptyState";
import { ErrorState } from "@/components/ui/ErrorState";
import { Input } from "@/components/ui/Input";
import { TableSkeleton } from "@/components/ui/Skeleton";
import { Table, TBody, TD, TH, THead, TR } from "@/components/ui/Table";
import { useAuditLogs } from "@/hooks/useAuditLogs";
import { formatDateTime } from "@/utils/format";
import type { AuditLog, Project } from "@/types";

const PAGE_SIZE = 20;

interface AuditPanelProps {
  project: Project;
}

const formatMetadata = (metadata: AuditLog["metadata"]): string => {
  const entries = Object.entries(metadata);
  if (entries.length === 0) return "—";
  return entries.map(([key, value]) => `${key}=${value}`).join(", ");
};

const AuditPanel = ({ project }: AuditPanelProps) => {
  const [actionFilter, setActionFilter] = useState("");
  const [page, setPage] = useState(1);

  const query = {
    page,
    page_size: PAGE_SIZE,
    action: actionFilter.trim() || undefined,
  };
  const { data, isLoading, isError, error, refetch, isFetching } =
    useAuditLogs(project.id, query);

  const logs = data?.logs ?? [];
  const total = data?.meta?.total ?? logs.length;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-neutral-500">
          Every read, write, and delete is recorded here.
        </p>
        <div className="w-full max-w-xs">
          <Input
            placeholder="Filter by action (e.g. secret.read)"
            value={actionFilter}
            onChange={(e) => {
              setActionFilter(e.target.value);
              setPage(1);
            }}
            aria-label="Filter audit logs by action"
          />
        </div>
      </div>

      <Card>
        {isLoading ? (
          <TableSkeleton rows={8} columns={4} />
        ) : isError ? (
          <ErrorState error={error} onRetry={() => void refetch()} />
        ) : logs.length === 0 ? (
          <EmptyState
            icon={ScrollText}
            title="No audit entries"
            description={
              actionFilter
                ? "No entries match this action filter."
                : "Activity in this project will appear here."
            }
          />
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Action</TH>
                <TH>Resource</TH>
                <TH>Details</TH>
                <TH>When</TH>
              </TR>
            </THead>
            <TBody>
              {logs.map((log) => (
                <TR key={log.id}>
                  <TD>
                    <ActionBadge action={log.action} />
                  </TD>
                  <TD>
                    <div className="flex flex-col">
                      <span className="font-medium text-neutral-900">
                        {log.resource}
                      </span>
                      {log.resource_id && (
                        <span className="font-mono text-xs text-neutral-400">
                          {log.resource_id}
                        </span>
                      )}
                    </div>
                  </TD>
                  <TD className="max-w-xs truncate font-mono text-xs text-neutral-500">
                    {formatMetadata(log.metadata)}
                  </TD>
                  <TD className="whitespace-nowrap text-neutral-500">
                    {formatDateTime(log.created_at)}
                  </TD>
                </TR>
              ))}
            </TBody>
          </Table>
        )}
      </Card>

      {!isLoading && !isError && logs.length > 0 && (
        <div className="flex items-center justify-between text-sm text-neutral-500">
          <span>
            Page {page} of {totalPages}
            {total > 0 && ` · ${total} entries`}
          </span>
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="sm"
              leftIcon={<ChevronLeft className="h-4 w-4" />}
              disabled={page <= 1 || isFetching}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              Previous
            </Button>
            <Button
              variant="secondary"
              size="sm"
              disabled={page >= totalPages || isFetching}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

export const ProjectAuditPage = () => {
  const { projectId } = useParams<{ projectId: string }>();
  if (!projectId) return null;

  return (
    <ProjectPageShell projectId={projectId}>
      {(project) => <AuditPanel project={project} />}
    </ProjectPageShell>
  );
};
