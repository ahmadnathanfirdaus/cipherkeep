import { useState } from "react";
import { useParams } from "react-router-dom";
import { KeySquare, Plus, Trash2 } from "lucide-react";
import { ProjectPageShell } from "@/components/projects/ProjectPageShell";
import { CreateTokenModal } from "@/components/tokens/CreateTokenModal";
import { TokenRevealModal } from "@/components/tokens/TokenRevealModal";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { EmptyState } from "@/components/ui/EmptyState";
import { ErrorState } from "@/components/ui/ErrorState";
import { TableSkeleton } from "@/components/ui/Skeleton";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import { Table, TBody, TD, TH, THead, TR } from "@/components/ui/Table";
import { useEnvironments } from "@/hooks/useEnvironments";
import { useCreateToken, useRevokeToken, useTokens } from "@/hooks/useTokens";
import { useToast } from "@/hooks/useToast";
import { getApiErrorMessage } from "@/utils/apiError";
import type {
  CreateServiceTokenRequest,
  CreateServiceTokenResult,
  Project,
  ServiceTokenMeta,
} from "@/types";

const formatDate = (value: string | null): string =>
  value ? new Date(value).toLocaleDateString() : "Never";

interface TokensPanelProps {
  project: Project;
}

const TokensPanel = ({ project }: TokensPanelProps) => {
  const { notify } = useToast();
  const canManage = project.role === "owner" || project.role === "admin";

  const { data: environments } = useEnvironments(project.id);
  const { data: tokens, isLoading, isError, error, refetch } = useTokens(
    project.id,
  );
  const createToken = useCreateToken(project.id);
  const revokeToken = useRevokeToken(project.id);

  const [createOpen, setCreateOpen] = useState(false);
  const [revealed, setRevealed] = useState<CreateServiceTokenResult | null>(
    null,
  );
  const [pendingRevoke, setPendingRevoke] = useState<ServiceTokenMeta | null>(
    null,
  );

  const envName = (environmentId: string): string =>
    environments?.find((env) => env.id === environmentId)?.name ?? "—";

  const handleCreate = (payload: CreateServiceTokenRequest): void => {
    createToken.mutate(payload, {
      onSuccess: (result) => {
        setCreateOpen(false);
        setRevealed(result);
      },
      onError: (err) =>
        notify({
          variant: "error",
          title: "Could not create token",
          description: getApiErrorMessage(err),
        }),
    });
  };

  const handleRevoke = (): void => {
    if (!pendingRevoke) return;
    revokeToken.mutate(pendingRevoke.id, {
      onSuccess: () => {
        notify({ variant: "success", title: "Token revoked" });
        setPendingRevoke(null);
      },
      onError: (err) =>
        notify({
          variant: "error",
          title: "Could not revoke token",
          description: getApiErrorMessage(err),
        }),
    });
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-neutral-500">
          Read-only API keys for apps and CI. Each token is scoped to a single
          environment.
        </p>
        {canManage && (
          <Button
            leftIcon={<Plus className="h-4 w-4" />}
            onClick={() => setCreateOpen(true)}
          >
            Create token
          </Button>
        )}
      </div>

      <Card>
        {isLoading ? (
          <TableSkeleton rows={3} columns={5} />
        ) : isError ? (
          <ErrorState error={error} onRetry={() => void refetch()} />
        ) : !tokens || tokens.length === 0 ? (
          <EmptyState
            icon={KeySquare}
            title="No tokens yet"
            description="Create a token to let an application read this project's secrets."
          />
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Name</TH>
                <TH>Environment</TH>
                <TH>Token</TH>
                <TH>Expires</TH>
                <TH>Last used</TH>
                {canManage && <TH className="text-right">Actions</TH>}
              </TR>
            </THead>
            <TBody>
              {tokens.map((token) => (
                <TR key={token.id}>
                  <TD className="font-medium text-neutral-900">{token.name}</TD>
                  <TD>{envName(token.environment_id)}</TD>
                  <TD>
                    <code className="font-mono text-xs text-neutral-500">
                      {token.display_hint}
                    </code>
                  </TD>
                  <TD>{formatDate(token.expires_at)}</TD>
                  <TD>{formatDate(token.last_used_at)}</TD>
                  {canManage && (
                    <TD className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label={`Revoke ${token.name}`}
                        onClick={() => setPendingRevoke(token)}
                      >
                        <Trash2 className="h-4 w-4 text-danger-600" />
                      </Button>
                    </TD>
                  )}
                </TR>
              ))}
            </TBody>
          </Table>
        )}
      </Card>

      <CreateTokenModal
        open={createOpen}
        environments={environments ?? []}
        isSubmitting={createToken.isPending}
        onSubmit={handleCreate}
        onClose={() => setCreateOpen(false)}
      />
      <TokenRevealModal
        open={revealed !== null}
        plaintext={revealed?.plaintext ?? ""}
        environmentId={revealed?.token.environment_id ?? ""}
        onClose={() => setRevealed(null)}
      />
      <ConfirmDialog
        open={pendingRevoke !== null}
        title={`Revoke "${pendingRevoke?.name ?? ""}"?`}
        description="Any application using this token will immediately lose access."
        confirmLabel="Revoke token"
        destructive
        isLoading={revokeToken.isPending}
        onConfirm={handleRevoke}
        onCancel={() => setPendingRevoke(null)}
      />
    </div>
  );
};

export const ProjectTokensPage = () => {
  const { projectId } = useParams<{ projectId: string }>();
  if (!projectId) return null;

  return (
    <ProjectPageShell projectId={projectId}>
      {(project) => <TokensPanel project={project} />}
    </ProjectPageShell>
  );
};
