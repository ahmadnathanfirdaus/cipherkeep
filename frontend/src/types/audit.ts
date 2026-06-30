export interface AuditLog {
  id: string;
  user_id: string | null;
  action: string;
  resource: string;
  resource_id: string | null;
  metadata: Record<string, string>;
  created_at: string;
}
