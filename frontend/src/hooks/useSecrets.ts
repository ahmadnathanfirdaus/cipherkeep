import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/queryKeys";
import { secretService } from "@/services/secret.service";
import type {
  CreateSecretRequest,
  ImportSecretsRequest,
  SecretFormat,
  UpdateSecretRequest,
} from "@/types";

export const useSecrets = (environmentId: string | undefined) =>
  useQuery({
    queryKey: queryKeys.secrets.list(environmentId ?? ""),
    queryFn: () => secretService.list(environmentId as string),
    enabled: Boolean(environmentId),
  });

export const useSecretVersions = (
  environmentId: string | undefined,
  key: string | undefined,
) =>
  useQuery({
    queryKey: queryKeys.secrets.versions(environmentId ?? "", key ?? ""),
    queryFn: () =>
      secretService.versions(environmentId as string, key as string),
    enabled: Boolean(environmentId) && Boolean(key),
  });

export const useCreateSecret = (environmentId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateSecretRequest) =>
      secretService.create(environmentId, payload),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.secrets.list(environmentId),
      });
    },
  });
};

export const useUpdateSecret = (environmentId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { key: string; payload: UpdateSecretRequest }) =>
      secretService.update(environmentId, vars.key, vars.payload),
    onSuccess: (_data, vars) => {
      void qc.invalidateQueries({
        queryKey: queryKeys.secrets.list(environmentId),
      });
      void qc.invalidateQueries({
        queryKey: queryKeys.secrets.versions(environmentId, vars.key),
      });
    },
  });
};

export const useDeleteSecret = (environmentId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (key: string) => secretService.remove(environmentId, key),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.secrets.list(environmentId),
      });
    },
  });
};

export const useImportSecrets = (environmentId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: ImportSecretsRequest) =>
      secretService.import(environmentId, payload),
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: queryKeys.secrets.list(environmentId),
      });
    },
  });
};

export const useExportSecrets = (environmentId: string) =>
  useMutation({
    mutationFn: (format: SecretFormat) =>
      secretService.export(environmentId, format),
  });
