import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/queryKeys";
import { environmentService } from "@/services/environment.service";
import type { CreateEnvironmentRequest } from "@/types";

export const useEnvironments = (projectId: string | undefined) =>
  useQuery({
    queryKey: queryKeys.environments.byProject(projectId ?? ""),
    queryFn: () => environmentService.listByProject(projectId as string),
    enabled: Boolean(projectId),
  });

export const useCreateEnvironment = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateEnvironmentRequest) =>
      environmentService.create(projectId, payload),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.environments.byProject(projectId),
      });
    },
  });
};

export const useDeleteEnvironment = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (environmentId: string) =>
      environmentService.remove(projectId, environmentId),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.environments.byProject(projectId),
      });
    },
  });
};
