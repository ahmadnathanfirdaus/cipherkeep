import { api } from "@/lib/api";
import type {
  CollectionResponse,
  CreateSecretRequest,
  DataResponse,
  ImportResult,
  ImportSecretsRequest,
  Secret,
  SecretExport,
  SecretFormat,
  SecretMeta,
  SecretVersion,
  UpdateSecretRequest,
} from "@/types";

export const secretService = {
  async list(environmentId: string): Promise<SecretMeta[]> {
    const res = await api.get<CollectionResponse<SecretMeta>>(
      `/environments/${environmentId}/secrets`,
    );
    return res.data.data;
  },

  /** Returns the decrypted value. Writes an audit log on the server. */
  async reveal(environmentId: string, key: string): Promise<Secret> {
    const res = await api.get<DataResponse<{ secret: Secret }>>(
      `/environments/${environmentId}/secrets/${encodeURIComponent(key)}`,
    );
    return res.data.data.secret;
  },

  async create(
    environmentId: string,
    payload: CreateSecretRequest,
  ): Promise<SecretMeta> {
    const res = await api.post<DataResponse<{ secret: SecretMeta }>>(
      `/environments/${environmentId}/secrets`,
      payload,
    );
    return res.data.data.secret;
  },

  async update(
    environmentId: string,
    key: string,
    payload: UpdateSecretRequest,
  ): Promise<SecretMeta> {
    const res = await api.put<DataResponse<{ secret: SecretMeta }>>(
      `/environments/${environmentId}/secrets/${encodeURIComponent(key)}`,
      payload,
    );
    return res.data.data.secret;
  },

  async remove(environmentId: string, key: string): Promise<void> {
    await api.delete(
      `/environments/${environmentId}/secrets/${encodeURIComponent(key)}`,
    );
  },

  async versions(
    environmentId: string,
    key: string,
  ): Promise<SecretVersion[]> {
    const res = await api.get<CollectionResponse<SecretVersion>>(
      `/environments/${environmentId}/secrets/${encodeURIComponent(key)}/versions`,
    );
    return res.data.data;
  },

  /** Bulk-imports secrets from .env, JSON, or YAML content. */
  async import(
    environmentId: string,
    payload: ImportSecretsRequest,
  ): Promise<ImportResult> {
    const res = await api.post<DataResponse<{ result: ImportResult }>>(
      `/environments/${environmentId}/import`,
      payload,
    );
    return res.data.data.result;
  },

  /** Exports all decrypted secrets in the given format. Audited on the server. */
  async export(
    environmentId: string,
    format: SecretFormat,
  ): Promise<SecretExport> {
    const res = await api.get<DataResponse<{ export: SecretExport }>>(
      `/environments/${environmentId}/export`,
      { params: { format } },
    );
    return res.data.data.export;
  },
};
