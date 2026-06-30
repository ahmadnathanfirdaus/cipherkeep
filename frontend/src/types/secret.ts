export interface SecretMeta {
  id: string;
  key: string;
  version: number;
  updated_at: string;
  updated_by: string;
}

export interface Secret extends SecretMeta {
  /** Decrypted plaintext value; only returned by the single-secret GET. */
  value: string;
}

export interface SecretVersion {
  version: number;
  created_at: string;
  created_by: string;
}

export interface CreateSecretRequest {
  key: string;
  value: string;
}

export interface UpdateSecretRequest {
  value: string;
}

export type SecretFormat = "env" | "json" | "yaml";

export interface ImportSecretsRequest {
  format: SecretFormat;
  content: string;
  overwrite: boolean;
}

export interface ImportResult {
  created: number;
  updated: number;
  skipped: number;
  total: number;
}

export interface SecretExport {
  format: SecretFormat;
  filename: string;
  content: string;
}
