export type Role = "owner" | "admin" | "member";

export interface Project {
  id: string;
  name: string;
  slug: string;
  description: string | null;
  role: Role;
  created_at: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
}

export interface Member {
  user_id: string;
  email: string;
  name: string;
  role: Role;
}

export interface AddMemberRequest {
  email: string;
  role: Exclude<Role, "owner">;
}

export interface UpdateMemberRequest {
  role: Exclude<Role, "owner">;
}
