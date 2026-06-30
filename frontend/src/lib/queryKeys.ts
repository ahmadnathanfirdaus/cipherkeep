import type { PaginationParams } from "@/types";

/**
 * Centralized, typed React Query keys. Keep all keys here so invalidation is
 * consistent across hooks.
 */
export const queryKeys = {
  me: ["me"] as const,

  projects: {
    all: ["projects"] as const,
    list: () => [...queryKeys.projects.all, "list"] as const,
    detail: (projectId: string) =>
      [...queryKeys.projects.all, "detail", projectId] as const,
    members: (projectId: string) =>
      [...queryKeys.projects.all, projectId, "members"] as const,
    audit: (projectId: string, params?: { action?: string } & PaginationParams) =>
      [...queryKeys.projects.all, projectId, "audit", params ?? {}] as const,
    tokens: (projectId: string) =>
      [...queryKeys.projects.all, projectId, "tokens"] as const,
  },

  environments: {
    byProject: (projectId: string) =>
      ["environments", "byProject", projectId] as const,
  },

  secrets: {
    list: (environmentId: string) =>
      ["secrets", environmentId, "list"] as const,
    detail: (environmentId: string, key: string) =>
      ["secrets", environmentId, "detail", key] as const,
    versions: (environmentId: string, key: string) =>
      ["secrets", environmentId, "versions", key] as const,
  },
} as const;
