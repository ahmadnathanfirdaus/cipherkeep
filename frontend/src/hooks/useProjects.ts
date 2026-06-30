import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/queryKeys";
import { projectService } from "@/services/project.service";
import type {
  CreateProjectRequest,
  Project,
  UpdateProjectRequest,
} from "@/types";

export const useProjects = () =>
  useQuery({
    queryKey: queryKeys.projects.list(),
    queryFn: () => projectService.list(),
  });

export const useProject = (projectId: string | undefined) =>
  useQuery({
    queryKey: queryKeys.projects.detail(projectId ?? ""),
    queryFn: () => projectService.get(projectId as string),
    enabled: Boolean(projectId),
  });

export const useCreateProject = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateProjectRequest) =>
      projectService.create(payload),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.projects.list() });
    },
  });
};

export const useUpdateProject = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: UpdateProjectRequest) =>
      projectService.update(projectId, payload),
    onSuccess: (updated: Project) => {
      qc.setQueryData(queryKeys.projects.detail(projectId), updated);
      void qc.invalidateQueries({ queryKey: queryKeys.projects.list() });
    },
  });
};

export const useDeleteProject = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (projectId: string) => projectService.remove(projectId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.projects.list() });
    },
  });
};
