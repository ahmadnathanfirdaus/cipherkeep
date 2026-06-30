import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/queryKeys";
import { tokenService } from "@/services/token.service";
import type { CreateServiceTokenRequest } from "@/types";

export const useTokens = (projectId: string) =>
  useQuery({
    queryKey: queryKeys.projects.tokens(projectId),
    queryFn: () => tokenService.list(projectId),
  });

export const useCreateToken = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateServiceTokenRequest) =>
      tokenService.create(projectId, payload),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.projects.tokens(projectId),
      });
    },
  });
};

export const useRevokeToken = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (tokenId: string) => tokenService.revoke(projectId, tokenId),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.projects.tokens(projectId),
      });
    },
  });
};
