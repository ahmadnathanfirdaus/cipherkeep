import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { Download, Layers, Plus, Search, Upload } from "lucide-react";
import { ProjectPageShell } from "@/components/projects/ProjectPageShell";
import { EnvironmentSwitcher } from "@/components/secrets/EnvironmentSwitcher";
import { EnvironmentFormModal } from "@/components/secrets/EnvironmentFormModal";
import { SecretsTable } from "@/components/secrets/SecretsTable";
import { SecretFormModal, type SecretFormValues } from "@/components/secrets/SecretFormModal";
import { SecretVersionsModal } from "@/components/secrets/SecretVersionsModal";
import { ImportSecretsModal } from "@/components/secrets/ImportSecretsModal";
import { ExportSecretsModal } from "@/components/secrets/ExportSecretsModal";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Card } from "@/components/ui/Card";
import { EmptyState } from "@/components/ui/EmptyState";
import { ErrorState } from "@/components/ui/ErrorState";
import { Skeleton } from "@/components/ui/Skeleton";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import {
  useCreateEnvironment,
  useDeleteEnvironment,
  useEnvironments,
} from "@/hooks/useEnvironments";
import {
  useCreateSecret,
  useDeleteSecret,
  useImportSecrets,
  useSecrets,
  useUpdateSecret,
} from "@/hooks/useSecrets";
import { useToast } from "@/hooks/useToast";
import { getApiErrorMessage } from "@/utils/apiError";
import type { ImportSecretsRequest, Project, SecretMeta } from "@/types";

type SecretModalState =
  | { kind: "closed" }
  | { kind: "create" }
  | { kind: "edit"; secret: SecretMeta };

interface SecretsPanelProps {
  project: Project;
}

const SecretsPanel = ({ project }: SecretsPanelProps) => {
  const { notify } = useToast();
  const {
    data: environments,
    isLoading: envLoading,
    isError: envError,
    error: envErr,
    refetch: refetchEnvs,
  } = useEnvironments(project.id);

  const [activeEnvId, setActiveEnvId] = useState<string | null>(null);
  const [envModalOpen, setEnvModalOpen] = useState(false);
  const [secretModal, setSecretModal] = useState<SecretModalState>({
    kind: "closed",
  });
  const [pendingDelete, setPendingDelete] = useState<SecretMeta | null>(null);
  const [pendingDeleteEnvId, setPendingDeleteEnvId] = useState<string | null>(
    null,
  );
  const [historyKey, setHistoryKey] = useState<string | null>(null);
  const [importOpen, setImportOpen] = useState(false);
  const [exportOpen, setExportOpen] = useState(false);
  const [search, setSearch] = useState("");

  // Default to the first environment once they load.
  useEffect(() => {
    if (!activeEnvId && environments && environments.length > 0) {
      setActiveEnvId(environments[0]?.id ?? null);
    }
  }, [environments, activeEnvId]);

  const createEnv = useCreateEnvironment(project.id);
  const deleteEnv = useDeleteEnvironment(project.id);
  const activeEnvironmentId = activeEnvId ?? "";
  const pendingDeleteEnv =
    environments?.find((env) => env.id === pendingDeleteEnvId) ?? null;

  const {
    data: secrets,
    isLoading: secretsLoading,
    isError: secretsIsError,
    error: secretsError,
    refetch: refetchSecrets,
  } = useSecrets(activeEnvId ?? undefined);

  const createSecret = useCreateSecret(activeEnvironmentId);
  const updateSecret = useUpdateSecret(activeEnvironmentId);
  const deleteSecret = useDeleteSecret(activeEnvironmentId);
  const importSecrets = useImportSecrets(activeEnvironmentId);

  const handleCreateEnv = (name: string): void => {
    createEnv.mutate(
      { name },
      {
        onSuccess: (env) => {
          notify({ variant: "success", title: "Environment created" });
          setEnvModalOpen(false);
          setActiveEnvId(env.id);
        },
        onError: (err) =>
          notify({
            variant: "error",
            title: "Could not create environment",
            description: getApiErrorMessage(err),
          }),
      },
    );
  };

  const handleDeleteEnv = (): void => {
    if (!pendingDeleteEnvId) return;
    deleteEnv.mutate(pendingDeleteEnvId, {
      onSuccess: () => {
        notify({ variant: "success", title: "Environment deleted" });
        if (pendingDeleteEnvId === activeEnvId) {
          // Let the effect re-select the first remaining environment.
          setActiveEnvId(null);
        }
        setPendingDeleteEnvId(null);
      },
      onError: (err) =>
        notify({
          variant: "error",
          title: "Could not delete environment",
          description: getApiErrorMessage(err),
        }),
    });
  };

  const handleSecretSubmit = (values: SecretFormValues): void => {
    if (secretModal.kind === "create") {
      createSecret.mutate(
        { key: values.key, value: values.value },
        {
          onSuccess: () => {
            notify({ variant: "success", title: "Secret created" });
            setSecretModal({ kind: "closed" });
          },
          onError: (err) =>
            notify({
              variant: "error",
              title: "Could not create secret",
              description: getApiErrorMessage(err),
            }),
        },
      );
    } else if (secretModal.kind === "edit") {
      updateSecret.mutate(
        { key: secretModal.secret.key, payload: { value: values.value } },
        {
          onSuccess: () => {
            notify({ variant: "success", title: "New version saved" });
            setSecretModal({ kind: "closed" });
          },
          onError: (err) =>
            notify({
              variant: "error",
              title: "Could not update secret",
              description: getApiErrorMessage(err),
            }),
        },
      );
    }
  };

  const handleDelete = (): void => {
    if (!pendingDelete) return;
    deleteSecret.mutate(pendingDelete.key, {
      onSuccess: () => {
        notify({ variant: "success", title: "Secret deleted" });
        setPendingDelete(null);
      },
      onError: (err) =>
        notify({
          variant: "error",
          title: "Could not delete secret",
          description: getApiErrorMessage(err),
        }),
    });
  };

  const handleImport = (payload: ImportSecretsRequest): void => {
    importSecrets.mutate(payload, {
      onSuccess: (result) => {
        notify({
          variant: "success",
          title: "Import complete",
          description: `${result.created} created, ${result.updated} updated, ${result.skipped} skipped.`,
        });
        setImportOpen(false);
      },
      onError: (err) =>
        notify({
          variant: "error",
          title: "Could not import secrets",
          description: getApiErrorMessage(err),
        }),
    });
  };

  // Per api-spec, creating/updating secrets requires role >= member, so every
  // project member may write. (Deletion is gated to admin/owner in the table.)
  const canWrite = true;

  if (envLoading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-10 w-64" />
        <Card>
          <div className="p-4">
            <Skeleton className="h-40 w-full" />
          </div>
        </Card>
      </div>
    );
  }

  if (envError) {
    return <ErrorState error={envErr} onRetry={() => void refetchEnvs()} />;
  }

  if (!environments || environments.length === 0) {
    return (
      <Card>
        <EmptyState
          icon={Layers}
          title="No environments yet"
          description="Create an environment (e.g. production) to start adding secrets."
          action={
            project.role !== "member" ? (
              <Button
                leftIcon={<Plus className="h-4 w-4" />}
                onClick={() => setEnvModalOpen(true)}
              >
                New environment
              </Button>
            ) : undefined
          }
        />
        <EnvironmentFormModal
          open={envModalOpen}
          isSubmitting={createEnv.isPending}
          onSubmit={handleCreateEnv}
          onClose={() => setEnvModalOpen(false)}
        />
      </Card>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <EnvironmentSwitcher
          environments={environments}
          activeId={activeEnvId}
          role={project.role}
          onSelect={setActiveEnvId}
          onCreate={() => setEnvModalOpen(true)}
          onDelete={(id) => setPendingDeleteEnvId(id)}
        />
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            leftIcon={<Download className="h-4 w-4" />}
            onClick={() => setExportOpen(true)}
          >
            Export
          </Button>
          {canWrite && (
            <Button
              variant="secondary"
              leftIcon={<Upload className="h-4 w-4" />}
              onClick={() => setImportOpen(true)}
            >
              Import
            </Button>
          )}
          {canWrite && (
            <Button
              leftIcon={<Plus className="h-4 w-4" />}
              onClick={() => setSecretModal({ kind: "create" })}
            >
              New secret
            </Button>
          )}
        </div>
      </div>

      {secrets && secrets.length > 0 && (
        <Input
          aria-label="Search secrets"
          placeholder="Search keys…"
          leftIcon={<Search className="h-4 w-4" />}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="max-w-xs"
        />
      )}

      <Card>
        <SecretsTable
          environmentId={activeEnvironmentId}
          secrets={
            search.trim()
              ? secrets?.filter((s) =>
                  s.key.toLowerCase().includes(search.trim().toLowerCase()),
                )
              : secrets
          }
          role={project.role}
          isLoading={secretsLoading}
          isError={secretsIsError}
          error={secretsError}
          onRetry={() => void refetchSecrets()}
          onCreate={() => setSecretModal({ kind: "create" })}
          onEdit={(secret) => setSecretModal({ kind: "edit", secret })}
          onDelete={(secret) => setPendingDelete(secret)}
          onViewHistory={(secret) => setHistoryKey(secret.key)}
        />
      </Card>

      <EnvironmentFormModal
        open={envModalOpen}
        isSubmitting={createEnv.isPending}
        onSubmit={handleCreateEnv}
        onClose={() => setEnvModalOpen(false)}
      />
      <SecretFormModal
        open={secretModal.kind !== "closed"}
        mode={secretModal.kind === "edit" ? "edit" : "create"}
        initialKey={
          secretModal.kind === "edit" ? secretModal.secret.key : undefined
        }
        isSubmitting={createSecret.isPending || updateSecret.isPending}
        onSubmit={handleSecretSubmit}
        onClose={() => setSecretModal({ kind: "closed" })}
      />
      <SecretVersionsModal
        open={historyKey !== null}
        environmentId={activeEnvironmentId}
        secretKey={historyKey}
        onClose={() => setHistoryKey(null)}
      />
      <ImportSecretsModal
        open={importOpen}
        isSubmitting={importSecrets.isPending}
        onSubmit={handleImport}
        onClose={() => setImportOpen(false)}
      />
      <ExportSecretsModal
        open={exportOpen}
        environmentId={activeEnvironmentId}
        secretCount={secrets?.length ?? 0}
        onClose={() => setExportOpen(false)}
      />
      <ConfirmDialog
        open={pendingDelete !== null}
        title={`Delete "${pendingDelete?.key ?? ""}"?`}
        description="This permanently removes the secret and its history."
        confirmLabel="Delete secret"
        destructive
        isLoading={deleteSecret.isPending}
        onConfirm={handleDelete}
        onCancel={() => setPendingDelete(null)}
      />
      <ConfirmDialog
        open={pendingDeleteEnvId !== null}
        title={`Delete environment "${pendingDeleteEnv?.name ?? ""}"?`}
        description="This permanently removes the environment and all of its secrets."
        confirmLabel="Delete environment"
        destructive
        isLoading={deleteEnv.isPending}
        onConfirm={handleDeleteEnv}
        onCancel={() => setPendingDeleteEnvId(null)}
      />
    </div>
  );
};

export const ProjectDetailPage = () => {
  const { projectId } = useParams<{ projectId: string }>();
  if (!projectId) return null;

  return (
    <ProjectPageShell projectId={projectId}>
      {(project) => <SecretsPanel project={project} />}
    </ProjectPageShell>
  );
};
