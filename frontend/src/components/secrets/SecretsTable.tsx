import { History, KeyRound, Pencil, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { ErrorState } from "@/components/ui/ErrorState";
import { TableSkeleton } from "@/components/ui/Skeleton";
import { Table, TBody, TD, TH, THead, TR } from "@/components/ui/Table";
import { SecretValueCell } from "@/components/secrets/SecretValueCell";
import { formatRelativeTime } from "@/utils/format";
import type { Role, SecretMeta } from "@/types";

interface SecretsTableProps {
  environmentId: string;
  secrets: SecretMeta[] | undefined;
  role: Role;
  isLoading: boolean;
  isError: boolean;
  error: unknown;
  onRetry: () => void;
  onCreate: () => void;
  onEdit: (secret: SecretMeta) => void;
  onDelete: (secret: SecretMeta) => void;
  onViewHistory: (secret: SecretMeta) => void;
}

export const SecretsTable = ({
  environmentId,
  secrets,
  role,
  isLoading,
  isError,
  error,
  onRetry,
  onCreate,
  onEdit,
  onDelete,
  onViewHistory,
}: SecretsTableProps) => {
  const canWrite = role === "owner" || role === "admin" || role === "member";
  const canDelete = role === "owner" || role === "admin";

  if (isLoading) return <TableSkeleton rows={5} columns={4} />;
  if (isError) return <ErrorState error={error} onRetry={onRetry} />;

  if (!secrets || secrets.length === 0) {
    return (
      <EmptyState
        icon={KeyRound}
        title="No secrets in this environment"
        description="Add a secret to make it available to your team."
        action={
          canWrite ? (
            <Button leftIcon={<Plus className="h-4 w-4" />} onClick={onCreate}>
              New secret
            </Button>
          ) : undefined
        }
      />
    );
  }

  return (
    <Table>
      <THead>
        <TR>
          <TH className="w-1/4">Key</TH>
          <TH className="w-1/2">Value</TH>
          <TH>Updated</TH>
          <TH className="text-right">Actions</TH>
        </TR>
      </THead>
      <TBody>
        {secrets.map((secret) => (
          <TR key={secret.id}>
            <TD>
              <div className="flex items-center gap-2">
                <KeyRound
                  className="h-4 w-4 shrink-0 text-neutral-400"
                  aria-hidden="true"
                />
                <span className="font-mono text-sm font-medium text-neutral-900">
                  {secret.key}
                </span>
                <span className="text-xs text-neutral-400">
                  v{secret.version}
                </span>
              </div>
            </TD>
            <TD>
              <SecretValueCell
                environmentId={environmentId}
                secretKey={secret.key}
              />
            </TD>
            <TD className="whitespace-nowrap text-neutral-500">
              {formatRelativeTime(secret.updated_at)}
            </TD>
            <TD className="text-right">
              <div className="flex items-center justify-end gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  aria-label="View history"
                  onClick={() => onViewHistory(secret)}
                >
                  <History className="h-4 w-4" />
                </Button>
                {canWrite && (
                  <Button
                    variant="ghost"
                    size="sm"
                    aria-label="Edit secret"
                    onClick={() => onEdit(secret)}
                  >
                    <Pencil className="h-4 w-4" />
                  </Button>
                )}
                {canDelete && (
                  <Button
                    variant="ghost"
                    size="sm"
                    aria-label="Delete secret"
                    onClick={() => onDelete(secret)}
                  >
                    <Trash2 className="h-4 w-4 text-danger-600" />
                  </Button>
                )}
              </div>
            </TD>
          </TR>
        ))}
      </TBody>
    </Table>
  );
};
