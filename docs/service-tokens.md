# Service Tokens (API Keys)

Non-interactive credentials that let an application or CI job read secrets from a
single environment without using a human user's password. This is the recommended way
to consume secrets from another project (keeps the consumer's `.env` to a tiny
bootstrap and never stores plaintext secrets on disk).

## Design decisions (locked)

- **Scope: per-environment.** A token is bound to exactly one environment (least
  privilege; smallest blast radius).
- **Access: read-only.** Tokens may only read/list/export secrets. They can never
  write, delete, manage projects/environments/members, read audit logs, or create
  other tokens.
- **Expiry: optional, default 90 days.** Creator may choose a custom expiry or "never".
- A token is a **distinct principal**, not a user. It is not added to `project_members`.

## Token anatomy

- Format: `ck_live_<base64url(32 random bytes)>`. The `ck_` prefix makes it
  recognizable and detectable by secret scanners.
- Stored **hashed** (SHA-256) in `service_tokens.token_hash`; the plaintext is shown
  **once** at creation and never again.
- A non-secret display hint (`ck_live_AbC1…`, prefix + last 4) is stored for the UI
  list so tokens can be told apart and revoked.

## Database: `service_tokens`

| column          | type        | notes                                            |
| --------------- | ----------- | ------------------------------------------------ |
| id              | uuid PK     |                                                  |
| project_id      | uuid FK     | → projects(id) ON DELETE CASCADE                 |
| environment_id  | uuid FK     | → environments(id) ON DELETE CASCADE (required)  |
| name            | text        | human label, e.g. "ci-backend-prod"              |
| token_hash      | text UNIQUE | SHA-256 hex of the token (never store raw)       |
| display_hint    | text        | non-secret: prefix + last 4 chars                |
| created_by      | uuid FK     | → users(id) ON DELETE RESTRICT                   |
| expires_at      | timestamptz | nullable (null = never)                          |
| last_used_at    | timestamptz | nullable; best-effort update on use              |
| revoked_at      | timestamptz | nullable                                         |
| created_at      | timestamptz | default now()                                    |

Migration: `0008_service_tokens`. Index on `token_hash`, `(project_id)`.

A token is **valid** iff `revoked_at IS NULL` AND (`expires_at IS NULL` OR
`expires_at > now()`).

## Authentication

The `Authorization: Bearer <credential>` header carries either a user JWT or a service
token. The authenticator distinguishes by prefix:

- credential starts with `ck_` → resolve as a **service token** → `Principal{Kind: token}`.
- otherwise → parse as a **user JWT** → `Principal{Kind: user}`.

```go
type PrincipalKind int
const ( PrincipalUser PrincipalKind = iota; PrincipalToken )

type Principal struct {
    Kind  PrincipalKind
    User  *domain.User          // set when Kind == PrincipalUser
    Token *domain.ServiceToken  // set when Kind == PrincipalToken
}
```

- User-only routes (everything except the three read endpoints below) require
  `Kind == PrincipalUser`; a token principal there returns **403 FORBIDDEN**.
- The secret read/list/export handlers accept either principal. For a token principal
  the secret service authorizes by checking `token.EnvironmentID == :environmentId`
  (plus validity); for a user principal it uses the existing project-membership RBAC.

## Authorization matrix

| Operation                                   | Service token        | User (JWT)        |
| ------------------------------------------- | -------------------- | ----------------- |
| `GET /environments/:id/secrets`             | ✅ if `:id` == scope | ✅ member+        |
| `GET /environments/:id/secrets/:key`        | ✅ if `:id` == scope | ✅ member+        |
| `GET /environments/:id/export`              | ✅ if `:id` == scope | ✅ member+        |
| create / update / delete / import secret    | ❌ 403               | ✅ per role       |
| environments / projects / members / audit   | ❌ 403               | ✅ per role       |
| manage service tokens                       | ❌ 403               | ✅ admin+         |

## API

Token management is user-only and requires project **admin or owner**.

### GET `/projects/{projectId}/tokens`
Auth (user) + role ≥ admin. Lists token metadata (never the secret value).
→ `200 { "data": [ServiceTokenMeta] }`

### POST `/projects/{projectId}/tokens`
Auth (user) + role ≥ admin.
Body: `{ "name": string, "environment_id": string, "expires_in_days"?: number | null }`
(`expires_in_days` omitted → default 90; explicit `null` → never).
Returns the plaintext token **once**.
→ `201 { "data": { "token": ServiceTokenMeta, "plaintext": string } }`

### DELETE `/projects/{projectId}/tokens/{tokenId}`
Auth (user) + role ≥ admin. Revokes immediately (sets `revoked_at`).
→ `204`

### Consumption (by the token holder)
Uses the existing read endpoints with the token as the bearer credential:
```
GET /api/v1/environments/{environmentId}/export?format=env
Authorization: Bearer ck_live_...
```

## Audit

Token usage is audited with the existing `secret.read` / `secret.export` actions; the
actor is the token (audit `user_id` is null, `metadata` carries `token_id` and
`environment_id`). Token management is audited as `token.create` / `token.revoke`.

## Type definitions (shared contract)

```ts
interface ServiceTokenMeta {
  id: string;
  name: string;
  project_id: string;
  environment_id: string;
  display_hint: string;       // e.g. "ck_live_AbC1…"
  created_by: string;
  expires_at: string | null;
  last_used_at: string | null;
  created_at: string;
}

interface CreateServiceTokenRequest {
  name: string;
  environment_id: string;
  expires_in_days?: number | null;   // omit = 90; null = never
}
```

## Security properties

- Hashed at rest; plaintext shown once. Compromise of the DB does not reveal tokens.
- Scoped to one environment + read-only → minimal blast radius if leaked.
- Revocable instantly and independently per token; rotation is create-new → deploy →
  revoke-old (zero downtime).
- `ck_` prefix aids secret scanning. Always used over TLS in production.
- Tokens never gain destructive or administrative powers.
