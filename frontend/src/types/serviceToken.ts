export interface ServiceTokenMeta {
  id: string;
  name: string;
  project_id: string;
  environment_id: string;
  display_hint: string;
  created_by: string;
  expires_at: string | null;
  last_used_at: string | null;
  created_at: string;
}

export interface CreateServiceTokenRequest {
  name: string;
  environment_id: string;
  /** Omit/null = 90 days; 0 = never; >0 = that many days. */
  expires_in_days?: number | null;
}

export interface CreateServiceTokenResult {
  token: ServiceTokenMeta;
  /** Plaintext token, returned once at creation only. */
  plaintext: string;
}
