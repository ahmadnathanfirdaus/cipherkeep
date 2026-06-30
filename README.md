# Cipherkeep

A self-hosted, team-oriented secret manager. Users authenticate with per-user accounts,
organize secrets into projects and environments, and the server stores every secret
value encrypted at rest using envelope encryption. Applications and CI read secrets at
runtime with scoped, read-only API keys.

- **Backend** — Go + Gin, service-repository pattern, `database/sql` + pgx (manual SQL),
  logrus, JWT (access + refresh), RBAC. PostgreSQL storage.
- **Frontend** — React + Vite, TypeScript (strict, no `any`), Tailwind, lucide-react
  icons, clear layer separation, TanStack Query.
- **Crypto** — server-side envelope encryption: `MASTER_PASSWORD` → Argon2id → KEK →
  unwraps a random DEK → AES-256-GCM per secret. See [docs/security.md](docs/security.md).

## Features

- **Projects & environments** — group secrets per project, namespaced by environment
  (e.g. `production`, `staging`).
- **Secrets** — create/read/update/delete, values masked by default with reveal & copy,
  full **version history** on every change, and key **search/filter**.
- **Import / export** — bulk import and export in `.env`, **JSON**, and **YAML**. JSON and
  YAML are **hierarchical**: nested objects flatten to dotted keys on import and nest back
  on export. The export panel lets you **switch format, view raw, copy, and download**.
- **Service tokens (API keys)** — read-only, per-environment, revocable credentials for
  apps and CI. Plaintext shown once.
- **RBAC** — per-project roles (`owner` > `admin` > `member`).
- **Audit log** — every secret read/write/delete and admin action is recorded.
- **Accounts** — register, sign in, change password (revokes other sessions).
- **Hardening** — strong-secret enforcement, trusted-proxy IP handling, request body /
  import caps, login-timing equalization. See [docs/security.md](docs/security.md).

## Repository layout

```
cipherkeep/
├── backend/      Go API service (see docs/backend-guidelines.md)
├── frontend/     React SPA (see docs/frontend-guidelines.md)
├── docs/         Architecture, API spec, DB schema, security, guidelines
├── scripts/      smoke-test.sh (end-to-end API check)
├── docker-compose.yml          Full stack: postgres + api + frontend
├── docker-compose.override.yml Dev conveniences (exposed DB port, debug logs)
└── .env.example  Root environment for docker-compose
```

## Quick start (Docker)

Requires Docker with the Compose plugin.

```bash
cp .env.example .env

# Generate strong secrets (required — the server refuses weak/placeholder values):
#   MASTER_PASSWORD  >= 16 chars   ->  openssl rand -hex 24
#   JWT_SECRET       >= 32 chars   ->  openssl rand -hex 32
# Paste them into .env, then:

docker compose up --build
```

Then open:

- **Web UI** — http://localhost:5173
- **API** — http://localhost:8080/api/v1 (health: http://localhost:8080/healthz)

The database schema is migrated automatically on API startup. There is **no default
admin account** — create the first account from the Register screen (see below).

> **Important:** `MASTER_PASSWORD` is the root of the encryption hierarchy. If you lose
> it, every stored secret becomes permanently unrecoverable. Back it up securely and do
> not change it casually.

---

## How to use it

The walkthrough below shows the web UI flow, with the equivalent API call where useful.
For API calls, first capture an access token:

```bash
API=http://localhost:8080/api/v1
TOKEN=$(curl -fsS -X POST "$API/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"YourPassword1"}' | jq -r .data.access_token)
AUTH="Authorization: Bearer $TOKEN"
```

### 1. Create your account

There is no seeded admin. Open the UI and **Register** (email, name, password ≥ 8 chars);
you are signed in immediately. Via API:

```bash
curl -fsS -X POST "$API/auth/register" -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","name":"You","password":"YourPassword1"}'
```

### 2. Create a project

In the UI click **New project**. The creator becomes the project **owner**.

```bash
curl -fsS -X POST "$API/projects" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"name":"Payments API"}'
```

### 3. Add environments

Open the project and use the environment tabs to add `production`, `staging`, etc.
Secrets always live inside an environment.

```bash
curl -fsS -X POST "$API/projects/$PROJECT_ID/environments" -H "$AUTH" \
  -H 'Content-Type: application/json' -d '{"name":"production"}'
```

### 4. Add and read secrets

In the **Secrets** tab: **New secret** (key + value). Values are masked; click the eye to
reveal, the copy icon to copy. Editing a secret saves a **new version** (old versions are
kept under history). Use the search box to filter keys.

```bash
# create
curl -fsS -X POST "$API/environments/$ENV_ID/secrets" -H "$AUTH" \
  -H 'Content-Type: application/json' -d '{"key":"DATABASE_URL","value":"postgres://..."}'
# read (decrypted; audited)
curl -fsS "$API/environments/$ENV_ID/secrets/DATABASE_URL" -H "$AUTH"
```

### 5. Import / export / convert

**Import** (Import button): paste or upload a `.env`, JSON, or YAML file. Toggle
"overwrite existing keys" as needed.

Nested JSON/YAML is flattened to **dotted keys**. For example this YAML:

```yaml
database:
  master:
    host: localhost
    port: 5432
```

…is stored as the flat keys `database.master.host` and `database.master.port`.

**Export / convert** (Export button): pick a format and the panel renders the **raw**
content live — switch between `.env`, JSON, and YAML to convert, then **Copy** or
**Download**. JSON/YAML re-nest dotted keys into a tree; `.env` stays flat.

```bash
curl -fsS "$API/environments/$ENV_ID/export?format=yaml" -H "$AUTH"
```

### 6. Manage members (RBAC)

In the **Members** tab (admin+), add teammates by email and set their role. See the role
table below.

### 7. Service tokens — let an app read secrets

In the **Tokens** tab (admin+): **Create token**, choose the environment and expiry.
The token (`ck_live_…`) is **shown once** — copy it. Tokens are **read-only** and scoped
to a single environment, and can be revoked anytime.

Use it from another project to read secrets at runtime (see the next section). This is the
recommended way to consume secrets — your app's `.env` only holds the bootstrap values,
never the real secrets.

### 8. Account settings

Click your name in the top bar → **Settings** to change your password (which signs out
your other sessions).

### 9. Audit log

The **Audit log** tab (admin+) lists every read/write/delete and admin action, filterable
by action type.

---

## Roles & permissions

Roles are **per project**.

| Capability                              | member | admin | owner |
| --------------------------------------- | :----: | :---: | :---: |
| Read / list / export secrets            |   ✅   |  ✅   |  ✅   |
| Create / update / import secrets        |   ✅   |  ✅   |  ✅   |
| Delete secrets                          |   —    |  ✅   |  ✅   |
| Manage environments                     |   —    |  ✅   |  ✅   |
| Manage members                          |   —    |  ✅   |  ✅   |
| Manage service tokens                   |   —    |  ✅   |  ✅   |
| View audit log                          |   —    |  ✅   |  ✅   |
| Delete the project                      |   —    |   —   |  ✅   |

Service tokens are **read-only** regardless of who created them, and limited to their one
environment.

---

## Consuming secrets in another project

Best practice: keep only a tiny bootstrap in the consumer's `.env`, and fetch the real
secrets at runtime with a service token — they never touch disk.

`.env` in the consuming app (only these three, non-sensitive except the token):

```bash
SECRETS_API=http://localhost:8080/api/v1
SECRETS_ENV_ID=<environment id>
SECRETS_TOKEN=ck_live_xxxxxxxx
```

Fetch at startup (example shell entrypoint that injects into the process and runs your app):

```bash
#!/usr/bin/env bash
set -euo pipefail
eval "$(curl -fsS "$SECRETS_API/environments/$SECRETS_ENV_ID/export?format=env" \
  -H "Authorization: Bearer $SECRETS_TOKEN" \
  | jq -r '.data.export.content' | sed 's/^/export /')"
exec "$@"
```

Or read them in code (Node example):

```ts
const res = await fetch(`${process.env.SECRETS_API}/environments/${process.env.SECRETS_ENV_ID}/export?format=json`, {
  headers: { Authorization: `Bearer ${process.env.SECRETS_TOKEN}` },
});
const secrets = JSON.parse((await res.json()).data.export.content);
```

> The token itself ("secret zero") must be stored in your CI/secret store. It is
> read-only, scoped to one environment, and revocable — far safer than a user password.

---

## Verify the deployment

With the stack running:

```bash
./scripts/smoke-test.sh
```

Runs the full path: health → register → login → create project → create environment →
create secret → read secret (verifies the decryption round-trip) → refresh token.

## Local development (without Docker)

### Backend — requires Go 1.25+ and a running PostgreSQL

```bash
cd backend
cp .env.example .env        # set DATABASE_URL, MASTER_PASSWORD, JWT_SECRET
go run ./cmd/server         # migrations run on startup, serves :8080
```

### Frontend — requires Node 20+

```bash
cd frontend
cp .env.example .env        # VITE_API_BASE_URL defaults to http://localhost:8080/api/v1
npm install
npm run dev                 # serves :5173
```

## Configuration

Backend (see [docs/backend-guidelines.md](docs/backend-guidelines.md) and
`backend/.env.example`):

| Variable               | Required | Default                  | Purpose                                        |
| ---------------------- | -------- | ------------------------ | ---------------------------------------------- |
| `DATABASE_URL`         | yes      | —                        | PostgreSQL DSN                                 |
| `MASTER_PASSWORD`      | yes      | —                        | Derives the KEK (envelope encryption)          |
| `JWT_SECRET`           | yes      | —                        | Signs JWT access tokens                        |
| `HTTP_PORT`            | no       | `8080`                   | API listen port                                |
| `ACCESS_TOKEN_TTL`     | no       | `15m`                    | Access token lifetime                          |
| `REFRESH_TOKEN_TTL`    | no       | `168h`                   | Refresh token lifetime                         |
| `CORS_ALLOWED_ORIGINS` | no       | `http://localhost:5173`  | Comma-separated allowed origins                |
| `TRUSTED_PROXIES`      | no       | _(none)_                 | Proxy CIDRs whose `X-Forwarded-For` is trusted |
| `MAX_BODY_BYTES`       | no       | `1048576`                | Max request body size (1 MiB)                  |
| `LOG_LEVEL`            | no       | `info`                   | `trace`…`error`                                |

> `JWT_SECRET` (≥32 chars) and `MASTER_PASSWORD` (≥16 chars) must be strong, non-placeholder
> values or the server refuses to start. Generate with `openssl rand -hex 32`.

Frontend: `VITE_API_BASE_URL` (baked at build time).

## Documentation

- [Architecture](docs/architecture.md)
- [API specification](docs/api-spec.md) — the backend/frontend contract
- [Database schema](docs/database-schema.md)
- [Security design](docs/security.md)
- [Service tokens](docs/service-tokens.md)
- [Backend guidelines](docs/backend-guidelines.md)
- [Frontend guidelines](docs/frontend-guidelines.md)
- [Development progress](development.md)

## Security notes

- Secret values are never logged and are returned only by the single-secret read
  endpoint; list endpoints return metadata only. Every secret read/write/delete is
  audited.
- Run behind a TLS-terminating reverse proxy in production; the API speaks plain HTTP
  inside the trusted network. Set `TRUSTED_PROXIES` to the proxy's network so client IPs
  (and rate limiting) are accurate.
- Containers run as a non-root user.
- Losing `MASTER_PASSWORD` is unrecoverable — back it up securely.
