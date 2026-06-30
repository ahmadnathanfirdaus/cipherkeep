import { useState } from "react";
import { useParams } from "react-router-dom";
import { Plus, Trash2, Users } from "lucide-react";
import { ProjectPageShell } from "@/components/projects/ProjectPageShell";
import { AddMemberModal } from "@/components/members/AddMemberModal";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { EmptyState } from "@/components/ui/EmptyState";
import { ErrorState } from "@/components/ui/ErrorState";
import { RoleBadge } from "@/components/ui/RoleBadge";
import { Select } from "@/components/ui/Select";
import { TableSkeleton } from "@/components/ui/Skeleton";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import { Table, TBody, TD, TH, THead, TR } from "@/components/ui/Table";
import {
  useAddMember,
  useMembers,
  useRemoveMember,
  useUpdateMemberRole,
} from "@/hooks/useMembers";
import { useToast } from "@/hooks/useToast";
import { getApiErrorMessage } from "@/utils/apiError";
import type { AddMemberRequest, Member, Project } from "@/types";

const ROLE_OPTIONS = [
  { value: "member", label: "Member" },
  { value: "admin", label: "Admin" },
];

interface MembersPanelProps {
  project: Project;
}

const MembersPanel = ({ project }: MembersPanelProps) => {
  const { notify } = useToast();
  const canManage = project.role === "owner" || project.role === "admin";

  const { data, isLoading, isError, error, refetch } = useMembers(project.id);
  const addMember = useAddMember(project.id);
  const updateRole = useUpdateMemberRole(project.id);
  const removeMember = useRemoveMember(project.id);

  const [addOpen, setAddOpen] = useState(false);
  const [pendingRemove, setPendingRemove] = useState<Member | null>(null);

  const handleAdd = (payload: AddMemberRequest): void => {
    addMember.mutate(payload, {
      onSuccess: () => {
        notify({ variant: "success", title: "Member added" });
        setAddOpen(false);
      },
      onError: (err) =>
        notify({
          variant: "error",
          title: "Could not add member",
          description: getApiErrorMessage(err),
        }),
    });
  };

  const handleRoleChange = (
    member: Member,
    role: AddMemberRequest["role"],
  ): void => {
    updateRole.mutate(
      { userId: member.user_id, payload: { role } },
      {
        onSuccess: () =>
          notify({ variant: "success", title: "Role updated" }),
        onError: (err) =>
          notify({
            variant: "error",
            title: "Could not update role",
            description: getApiErrorMessage(err),
          }),
      },
    );
  };

  const handleRemove = (): void => {
    if (!pendingRemove) return;
    removeMember.mutate(pendingRemove.user_id, {
      onSuccess: () => {
        notify({ variant: "success", title: "Member removed" });
        setPendingRemove(null);
      },
      onError: (err) =>
        notify({
          variant: "error",
          title: "Could not remove member",
          description: getApiErrorMessage(err),
        }),
    });
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-neutral-500">
          People with access to this project.
        </p>
        {canManage && (
          <Button
            leftIcon={<Plus className="h-4 w-4" />}
            onClick={() => setAddOpen(true)}
          >
            Add member
          </Button>
        )}
      </div>

      <Card>
        {isLoading ? (
          <TableSkeleton rows={4} columns={3} />
        ) : isError ? (
          <ErrorState error={error} onRetry={() => void refetch()} />
        ) : !data || data.length === 0 ? (
          <EmptyState
            icon={Users}
            title="No members"
            description="Add teammates to collaborate on this project."
          />
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Member</TH>
                <TH>Role</TH>
                {canManage && <TH className="text-right">Actions</TH>}
              </TR>
            </THead>
            <TBody>
              {data.map((member) => {
                const isOwner = member.role === "owner";
                const canEditThis = canManage && !isOwner;
                return (
                  <TR key={member.user_id}>
                    <TD>
                      <div className="flex flex-col">
                        <span className="font-medium text-neutral-900">
                          {member.name}
                        </span>
                        <span className="text-xs text-neutral-500">
                          {member.email}
                        </span>
                      </div>
                    </TD>
                    <TD>
                      {canEditThis ? (
                        <Select
                          className="h-9 w-32"
                          options={ROLE_OPTIONS}
                          value={member.role}
                          aria-label={`Role for ${member.name}`}
                          onChange={(e) =>
                            handleRoleChange(
                              member,
                              e.target.value as AddMemberRequest["role"],
                            )
                          }
                        />
                      ) : (
                        <RoleBadge role={member.role} />
                      )}
                    </TD>
                    {canManage && (
                      <TD className="text-right">
                        {canEditThis ? (
                          <Button
                            variant="ghost"
                            size="sm"
                            aria-label={`Remove ${member.name}`}
                            onClick={() => setPendingRemove(member)}
                          >
                            <Trash2 className="h-4 w-4 text-danger-600" />
                          </Button>
                        ) : (
                          <span className="text-xs text-neutral-400">
                            Owner
                          </span>
                        )}
                      </TD>
                    )}
                  </TR>
                );
              })}
            </TBody>
          </Table>
        )}
      </Card>

      <AddMemberModal
        open={addOpen}
        isSubmitting={addMember.isPending}
        onSubmit={handleAdd}
        onClose={() => setAddOpen(false)}
      />
      <ConfirmDialog
        open={pendingRemove !== null}
        title={`Remove ${pendingRemove?.name ?? ""}?`}
        description="They will immediately lose access to this project."
        confirmLabel="Remove member"
        destructive
        isLoading={removeMember.isPending}
        onConfirm={handleRemove}
        onCancel={() => setPendingRemove(null)}
      />
    </div>
  );
};

export const ProjectMembersPage = () => {
  const { projectId } = useParams<{ projectId: string }>();
  if (!projectId) return null;

  return (
    <ProjectPageShell projectId={projectId}>
      {(project) => <MembersPanel project={project} />}
    </ProjectPageShell>
  );
};
