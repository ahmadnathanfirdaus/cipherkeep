import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { FolderOpen, MoreVertical, Pencil, Plus, Trash2 } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { RoleBadge } from "@/components/ui/RoleBadge";
import { EmptyState } from "@/components/ui/EmptyState";
import { ErrorState } from "@/components/ui/ErrorState";
import { TableSkeleton } from "@/components/ui/Skeleton";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import { Table, TBody, TD, TH, THead, TR } from "@/components/ui/Table";
import {
  ProjectFormModal,
  type ProjectFormValues,
} from "@/components/projects/ProjectFormModal";
import {
  useCreateProject,
  useDeleteProject,
  useProjects,
  useUpdateProject,
} from "@/hooks/useProjects";
import { useToast } from "@/hooks/useToast";
import { getApiErrorMessage } from "@/utils/apiError";
import { formatRelativeTime } from "@/utils/format";
import type { Project } from "@/types";

type ModalState =
  | { kind: "closed" }
  | { kind: "create" }
  | { kind: "edit"; project: Project };

export const ProjectsPage = () => {
  const navigate = useNavigate();
  const { notify } = useToast();
  const { data: projects, isLoading, isError, error, refetch } = useProjects();

  const createMutation = useCreateProject();
  const deleteMutation = useDeleteProject();
  const [modal, setModal] = useState<ModalState>({ kind: "closed" });
  const [pendingDelete, setPendingDelete] = useState<Project | null>(null);

  // Edit mutation is bound per-project; created lazily when editing.
  const editingId = modal.kind === "edit" ? modal.project.id : "";
  const updateMutation = useUpdateProject(editingId);

  const closeModal = (): void => setModal({ kind: "closed" });

  const handleCreate = (values: ProjectFormValues): void => {
    createMutation.mutate(
      { name: values.name, description: values.description || undefined },
      {
        onSuccess: (project) => {
          notify({ variant: "success", title: "Project created" });
          closeModal();
          navigate(`/projects/${project.id}`);
        },
        onError: (err) =>
          notify({
            variant: "error",
            title: "Could not create project",
            description: getApiErrorMessage(err),
          }),
      },
    );
  };

  const handleEdit = (values: ProjectFormValues): void => {
    updateMutation.mutate(
      { name: values.name, description: values.description || undefined },
      {
        onSuccess: () => {
          notify({ variant: "success", title: "Project updated" });
          closeModal();
        },
        onError: (err) =>
          notify({
            variant: "error",
            title: "Could not update project",
            description: getApiErrorMessage(err),
          }),
      },
    );
  };

  const handleDelete = (): void => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        notify({ variant: "success", title: "Project deleted" });
        setPendingDelete(null);
      },
      onError: (err) =>
        notify({
          variant: "error",
          title: "Could not delete project",
          description: getApiErrorMessage(err),
        }),
    });
  };

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Projects"
        description="Group environments and secrets by project."
        actions={
          <Button
            leftIcon={<Plus className="h-4 w-4" />}
            onClick={() => setModal({ kind: "create" })}
          >
            New project
          </Button>
        }
      />

      <Card>
        {isLoading ? (
          <TableSkeleton rows={4} columns={4} />
        ) : isError ? (
          <ErrorState error={error} onRetry={() => void refetch()} />
        ) : !projects || projects.length === 0 ? (
          <EmptyState
            icon={FolderOpen}
            title="No projects yet"
            description="Create your first project to start managing secrets."
            action={
              <Button
                leftIcon={<Plus className="h-4 w-4" />}
                onClick={() => setModal({ kind: "create" })}
              >
                New project
              </Button>
            }
          />
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Name</TH>
                <TH>Role</TH>
                <TH>Created</TH>
                <TH className="w-px text-right">Actions</TH>
              </TR>
            </THead>
            <TBody>
              {projects.map((project) => (
                <TR
                  key={project.id}
                  className="cursor-pointer"
                  onClick={() => navigate(`/projects/${project.id}`)}
                >
                  <TD>
                    <div className="flex flex-col">
                      <span className="font-medium text-neutral-900">
                        {project.name}
                      </span>
                      {project.description && (
                        <span className="text-xs text-neutral-500">
                          {project.description}
                        </span>
                      )}
                    </div>
                  </TD>
                  <TD>
                    <RoleBadge role={project.role} />
                  </TD>
                  <TD className="text-neutral-500">
                    {formatRelativeTime(project.created_at)}
                  </TD>
                  <TD className="text-right">
                    <div
                      className="flex items-center justify-end gap-1"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {project.role !== "member" && (
                        <Button
                          variant="ghost"
                          size="sm"
                          aria-label="Edit project"
                          onClick={() =>
                            setModal({ kind: "edit", project })
                          }
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                      )}
                      {project.role === "owner" && (
                        <Button
                          variant="ghost"
                          size="sm"
                          aria-label="Delete project"
                          onClick={() => setPendingDelete(project)}
                        >
                          <Trash2 className="h-4 w-4 text-danger-600" />
                        </Button>
                      )}
                      {project.role === "member" && (
                        <span className="px-2 text-neutral-300">
                          <MoreVertical className="h-4 w-4" />
                        </span>
                      )}
                    </div>
                  </TD>
                </TR>
              ))}
            </TBody>
          </Table>
        )}
      </Card>

      <ProjectFormModal
        open={modal.kind === "create"}
        mode="create"
        isSubmitting={createMutation.isPending}
        onSubmit={handleCreate}
        onClose={closeModal}
      />
      <ProjectFormModal
        open={modal.kind === "edit"}
        mode="edit"
        initial={
          modal.kind === "edit"
            ? {
                name: modal.project.name,
                description: modal.project.description,
              }
            : undefined
        }
        isSubmitting={updateMutation.isPending}
        onSubmit={handleEdit}
        onClose={closeModal}
      />
      <ConfirmDialog
        open={pendingDelete !== null}
        title={`Delete "${pendingDelete?.name ?? ""}"?`}
        description="All environments and secrets in this project will be removed."
        confirmLabel="Delete project"
        destructive
        isLoading={deleteMutation.isPending}
        onConfirm={handleDelete}
        onCancel={() => setPendingDelete(null)}
      />
    </div>
  );
};
