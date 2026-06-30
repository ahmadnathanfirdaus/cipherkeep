import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/queryKeys";
import { memberService } from "@/services/member.service";
import type { AddMemberRequest, UpdateMemberRequest } from "@/types";

export const useMembers = (projectId: string | undefined) =>
  useQuery({
    queryKey: queryKeys.projects.members(projectId ?? ""),
    queryFn: () => memberService.list(projectId as string),
    enabled: Boolean(projectId),
  });

export const useAddMember = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: AddMemberRequest) =>
      memberService.add(projectId, payload),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.projects.members(projectId),
      });
    },
  });
};

export const useUpdateMemberRole = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { userId: string; payload: UpdateMemberRequest }) =>
      memberService.updateRole(projectId, vars.userId, vars.payload),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.projects.members(projectId),
      });
    },
  });
};

export const useRemoveMember = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) => memberService.remove(projectId, userId),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.projects.members(projectId),
      });
    },
  });
};
