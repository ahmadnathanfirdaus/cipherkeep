# API Specification

Base path: `/api/v1`. All request and response bodies are JSON. This document is the
**contract** between backend and frontend — both sides must conform to it.

## Conventions

- **Auth**: `Authorization: Bearer <access_token>` for protected endpoints.
- **Content-Type**: `application/json`.
- **IDs**: UUID strings.
- **Timestamps**: RFC 3339 / ISO 8601 (e.g. `2026-06-28T10:00:00Z`).
- **Errors**: consistent envelope (see below). HTTP status reflects the error class.
- **Pagination** (list endpoints): query `?page=1&page_size=20`; response includes
  `meta: { page, page_size, total }`.

### Error envelope

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable message",
    "details": [{ "field": "email", "message": "must be a valid email" }]
  }
}
```

Common codes: `VALIDATION_ERROR` (400), `UNAUTHORIZED` (401), `FORBIDDEN` (403),
`NOT_FOUND` (404), `CONFLICT` (409), `RATE_LIMITED` (429), `INTERNAL` (500).

### Success envelope

Single resource: `{ "data": { ... } }`. Collection: `{ "data": [ ... ], "meta": { ... } }`.

## Auth

### POST `/auth/register`

Body: `{ "email": string, "name": string, "password": string }`
→ `201 { "data": { "user": User } }`

### POST `/auth/login`

Body: `{ "email": string, "password": string }`
→ `200 { "data": { "user": User, "access_token": string, "refresh_token": string, "expires_in": number } }`

### POST `/auth/refresh`

Body: `{ "refresh_token": string }`
→ `200 { "data": { "access_token": string, "refresh_token": string, "expires_in": number } }`
(Refresh tokens rotate: the old one is revoked.)

### POST `/auth/logout`

Auth required. Body: `{ "refresh_token": string }` → `204`.

### GET `/auth/me`

Auth required. → `200 { "data": { "user": User } }`

### POST `/auth/change-password`

Auth required (user only). Body: `{ "current_password": string, "new_password": string }`
(min 8). Verifies the current password, sets the new one, and revokes all refresh
tokens. → `204`. Wrong current password → `401`.

## Projects

### GET `/projects`

Auth. Lists projects the user is a member of. → `200 { "data": [Project], "meta": {...} }`

### POST `/projects`

Auth. Body: `{ "name": string, "description"?: string }`
Creates project; caller becomes `owner`. → `201 { "data": { "project": Project } }`

### GET `/projects/{projectId}`

Auth + membership. → `200 { "data": { "project": Project } }`

### PATCH `/projects/{projectId}`

Auth + role ≥ admin. Body: `{ "name"?: string, "description"?: string }` → `200`.

### DELETE `/projects/{projectId}`

Auth + role = owner. → `204`.

## Members (RBAC)

### GET `/projects/{projectId}/members`

Auth + membership. → `200 { "data": [Member] }`

### POST `/projects/{projectId}/members`

Auth + role ≥ admin. Body: `{ "email": string, "role": "admin" | "member" }` → `201`.

### PATCH `/projects/{projectId}/members/{userId}`

Auth + role ≥ admin. Body: `{ "role": "admin" | "member" }` → `200`.

### DELETE `/projects/{projectId}/members/{userId}`

Auth + role ≥ admin. → `204`.

## Environments

### GET `/projects/{projectId}/environments`

Auth + membership. → `200 { "data": [Environment] }`

### POST `/projects/{projectId}/environments`

Auth + role ≥ admin. Body: `{ "name": string }` → `201 { "data": { "environment": Environment } }`

### DELETE `/projects/{projectId}/environments/{environmentId}`

Auth + role ≥ admin. → `204`.

## Secrets

Secrets are scoped to an environment.

### GET `/environments/{environmentId}/secrets`

Auth + membership. Returns metadata only (no values) by default.
→ `200 { "data": [SecretMeta] }`

### GET `/environments/{environmentId}/secrets/{key}`

Auth + membership. Returns the **decrypted** value. Writes an audit log.
→ `200 { "data": { "secret": Secret } }`

### POST `/environments/{environmentId}/secrets`

Auth + role ≥ member (configurable). Body: `{ "key": string, "value": string }`
→ `201 { "data": { "secret": SecretMeta } }`

### PUT `/environments/{environmentId}/secrets/{key}`

Auth + role ≥ member. Body: `{ "value": string }`. Increments version, records history.
→ `200 { "data": { "secret": SecretMeta } }`

### DELETE `/environments/{environmentId}/secrets/{key}`

Auth + role ≥ admin. → `204`.

### GET `/environments/{environmentId}/secrets/{key}/versions`

Auth + membership. Lists version metadata (no values). → `200 { "data": [SecretVersion] }`

## Import / Export

Bulk operations over all secrets in an environment. `format` is one of
`env` | `json` | `yaml`.

Secrets are stored flat (string→string). **JSON and YAML are hierarchical**: on
import, nested objects are flattened to dotted keys (`database.master.host`); on
export, dotted keys are nested back into a tree. Numeric path segments are treated as
ordinary string keys (objects-only; no array reconstruction). **`.env` is always
flat** — dotted keys are written literally. If stored keys cannot form a tree (a key
that is both a value and a group, e.g. `app` and `app.name`), JSON/YAML export returns
`400 VALIDATION_ERROR`; `.env` export still works.

### POST `/environments/{environmentId}/import`

Auth + role ≥ member. Body: `{ "format": "env" | "json" | "yaml", "content": string, "overwrite"?: boolean }`.
Parses `content` and upserts each key in a single transaction: new keys are created,
existing keys are updated (new version) when `overwrite` is true (default) or skipped
when false. Audited as `secret.import`.
→ `200 { "data": { "result": ImportResult } }`

### GET `/environments/{environmentId}/export?format=env|json|yaml`

Auth + membership. Returns all secrets **decrypted** and serialized in the requested
format (default `env`). Audited as `secret.export`.
→ `200 { "data": { "export": SecretExport } }`

## Service tokens (API keys)

Read-only, per-environment credentials for non-interactive consumers. Management is
user-only (admin+); the token itself is a distinct principal. See
[service-tokens.md](service-tokens.md) for the full design.

### GET `/projects/{projectId}/tokens`
Auth (user) + role ≥ admin. → `200 { "data": [ServiceTokenMeta] }`

### POST `/projects/{projectId}/tokens`
Auth (user) + role ≥ admin.
Body: `{ "name": string, "environment_id": string, "expires_in_days"?: number | null }`
(omit/null → 90 days; `0` → never; `>0` → that many days). Returns the plaintext once.
→ `201 { "data": { "token": ServiceTokenMeta, "plaintext": string } }`

### DELETE `/projects/{projectId}/tokens/{tokenId}`
Auth (user) + role ≥ admin. Revokes immediately. → `204`

### Consumption
A token authenticates the three read endpoints as `Authorization: Bearer ck_live_...`,
but only for the environment it is scoped to:
`GET /environments/{environmentId}/secrets`, `.../secrets/{key}`, and `.../export`.
Any other endpoint returns `403 FORBIDDEN`.

## Audit

### GET `/projects/{projectId}/audit-logs`

Auth + role ≥ admin. Supports pagination + `?action=` filter.
→ `200 { "data": [AuditLog], "meta": {...} }`

## Health

### GET `/healthz` → `200 { "status": "ok" }` (no auth, no envelope).

---

## Type definitions (shared contract)

These mirror the frontend `src/types`. Field names are `snake_case` on the wire.

```ts
interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

type Role = "owner" | "admin" | "member";

interface Project {
  id: string;
  name: string;
  slug: string;
  description: string | null;
  role: Role;            // caller's role in this project
  created_at: string;
}

interface Member {
  user_id: string;
  email: string;
  name: string;
  role: Role;
}

interface Environment {
  id: string;
  project_id: string;
  name: string;
  slug: string;
  created_at: string;
}

interface SecretMeta {
  id: string;
  key: string;
  version: number;
  updated_at: string;
  updated_by: string;
}

interface Secret extends SecretMeta {
  value: string;         // decrypted; only returned by single-secret GET
}

interface SecretVersion {
  version: number;
  created_at: string;
  created_by: string;
}

type SecretFormat = "env" | "json" | "yaml";

interface ImportResult {
  created: number;
  updated: number;
  skipped: number;
  total: number;
}

interface SecretExport {
  format: SecretFormat;
  filename: string;
  content: string;       // serialized, decrypted secrets
}

interface AuditLog {
  id: string;
  user_id: string | null;
  action: string;
  resource: string;
  resource_id: string | null;
  metadata: Record<string, string>;
  created_at: string;
}
```
