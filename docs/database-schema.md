# Database Schema

PostgreSQL. Use `golang-migrate` with sequential, reversible migrations under
`backend/migrations/` (`NNNN_name.up.sql` / `NNNN_name.down.sql`).

Conventions:

- `id` is `UUID` (default `gen_random_uuid()`; enable `pgcrypto` extension).
- Timestamps are `TIMESTAMPTZ`, default `now()`.
- Soft-delete is **not** used for secrets; deletes are real but recorded in `audit_logs`.
- All foreign keys `ON DELETE` behavior is explicit (see below).

## Extensions

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

## Tables

### `users`

| column         | type        | notes                                   |
| -------------- | ----------- | --------------------------------------- |
| id             | uuid PK     |                                         |
| email          | text UNIQUE | citext-style lowercase enforced in app  |
| name           | text        |                                         |
| password_hash  | text        | Argon2id encoded hash                   |
| is_active      | boolean     | default true                            |
| created_at     | timestamptz | default now()                           |
| updated_at     | timestamptz | default now()                           |

### `refresh_tokens`

Stores hashed refresh tokens to support rotation and revocation.

| column      | type        | notes                                          |
| ----------- | ----------- | ---------------------------------------------- |
| id          | uuid PK     |                                                |
| user_id     | uuid FK     | → users(id) ON DELETE CASCADE                  |
| token_hash  | text        | SHA-256 of the refresh token (never store raw) |
| expires_at  | timestamptz |                                                |
| revoked_at  | timestamptz | nullable                                       |
| created_at  | timestamptz | default now()                                  |

Index: `(user_id)`, `(token_hash)`.

### `projects`

| column      | type        | notes                          |
| ----------- | ----------- | ------------------------------ |
| id          | uuid PK     |                                |
| name        | text        |                                |
| slug        | text UNIQUE | url-safe identifier            |
| description | text        | nullable                       |
| created_by  | uuid FK     | → users(id) ON DELETE RESTRICT |
| created_at  | timestamptz | default now()                  |
| updated_at  | timestamptz | default now()                  |

### `project_members`

RBAC membership. Role is per-project.

| column     | type        | notes                                       |
| ---------- | ----------- | ------------------------------------------- |
| id         | uuid PK     |                                             |
| project_id | uuid FK     | → projects(id) ON DELETE CASCADE            |
| user_id    | uuid FK     | → users(id) ON DELETE CASCADE               |
| role       | text        | enum: `owner` \| `admin` \| `member`        |
| created_at | timestamptz | default now()                               |

Unique: `(project_id, user_id)`.

### `environments`

Namespacing within a project (e.g. `production`, `staging`, `development`).

| column     | type        | notes                            |
| ---------- | ----------- | -------------------------------- |
| id         | uuid PK     |                                  |
| project_id | uuid FK     | → projects(id) ON DELETE CASCADE |
| name       | text        | e.g. "production"                |
| slug       | text        | url-safe                         |
| created_at | timestamptz | default now()                    |

Unique: `(project_id, slug)`.

### `secrets`

One row per secret key per environment. Value stored encrypted.

| column         | type        | notes                                             |
| -------------- | ----------- | ------------------------------------------------- |
| id             | uuid PK     |                                                   |
| environment_id | uuid FK     | → environments(id) ON DELETE CASCADE              |
| key            | text        | e.g. "DATABASE_URL"                               |
| ciphertext     | bytea       | AES-256-GCM ciphertext (incl. auth tag)           |
| nonce          | bytea       | 12-byte GCM nonce, unique per encryption          |
| version        | int         | current version number, starts at 1               |
| created_by     | uuid FK     | → users(id) ON DELETE RESTRICT                    |
| updated_by     | uuid FK     | → users(id) ON DELETE RESTRICT                    |
| created_at     | timestamptz | default now()                                     |
| updated_at     | timestamptz | default now()                                     |

Unique: `(environment_id, key)`.
Never store plaintext. `ciphertext`/`nonce` are produced by the envelope service.

### `secret_versions`

History of every value change (append-only).

| column      | type        | notes                                   |
| ----------- | ----------- | --------------------------------------- |
| id          | uuid PK     |                                         |
| secret_id   | uuid FK     | → secrets(id) ON DELETE CASCADE         |
| version     | int         |                                         |
| ciphertext  | bytea       |                                         |
| nonce       | bytea       |                                         |
| created_by  | uuid FK     | → users(id) ON DELETE RESTRICT          |
| created_at  | timestamptz | default now()                           |

Unique: `(secret_id, version)`.

### `encryption_keys`

Stores the wrapped Data Encryption Key (DEK). Single active row in v1.

| column        | type        | notes                                            |
| ------------- | ----------- | ------------------------------------------------ |
| id            | uuid PK     |                                                  |
| wrapped_dek   | bytea       | DEK encrypted with the KEK (AES-256-GCM)         |
| nonce         | bytea       | nonce used to wrap the DEK                        |
| kdf_salt      | bytea       | Argon2id salt used to derive the KEK             |
| is_active     | boolean     | default true                                     |
| created_at    | timestamptz | default now()                                    |

See `security.md` for the bootstrap/unwrap flow.

### `audit_logs`

Append-only record of security-relevant actions.

| column      | type        | notes                                                      |
| ----------- | ----------- | ---------------------------------------------------------- |
| id          | uuid PK     |                                                            |
| user_id     | uuid FK     | → users(id) ON DELETE SET NULL (nullable)                  |
| action      | text        | e.g. `secret.read`, `secret.create`, `auth.login`         |
| resource    | text        | e.g. `secret`, `project`                                   |
| resource_id | uuid        | nullable                                                   |
| metadata    | jsonb       | non-sensitive context (project slug, env, key name only)  |
| ip_address  | inet        | nullable                                                   |
| created_at  | timestamptz | default now()                                              |

**Never** put plaintext secret values in `metadata`.

## Migration order

1. `0001_extensions_and_users`
2. `0002_refresh_tokens`
3. `0003_encryption_keys`
4. `0004_projects_and_members`
5. `0005_environments`
6. `0006_secrets_and_versions`
7. `0007_audit_logs`
