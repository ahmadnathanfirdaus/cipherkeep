export interface Environment {
  id: string;
  project_id: string;
  name: string;
  slug: string;
  created_at: string;
}

export interface CreateEnvironmentRequest {
  name: string;
}
