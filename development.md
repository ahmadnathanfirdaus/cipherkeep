# Development Progress

This document is the single source of truth for build progress. Each phase has a
checklist. Agents must update the relevant checkboxes and the status table when work
is completed, and append notes under "Decision Log" when a non-obvious choice is made.

## Project Summary

A team-oriented **secret manager** service.

- **Backend**: Go + Gin, service-repository pattern, `database/sql` + pgx (manual), logrus.
- **Auth**: JWT access + refresh tokens. Per-user accounts. RBAC.
- **Crypto**: Server-side encryption via envelope encryption (Argon2id-derived KEK → DEK → AES-256-GCM).
- **Storage**: PostgreSQL.
- **Frontend**: React + Vite, TypeScript (strict, no `any`), Tailwind, lucide-react icons.
- **Infra**: Docker + Docker Compose.
- **Language**: All code, comments, docs, and UI text in English.

## Status Overview

| Phase | Title                                   | Status      |
| ----- | --------------------------------------- | ----------- |
| 0     | Planning & documentation                | Done        |
| 1     | Backend foundation & infra              | Done        |
| 2     | Crypto core (envelope encryption)       | Done        |
| 3     | Auth & users (JWT + RBAC)               | Done        |
| 4     | Secrets domain (projects, secrets, audit)| Done        |
| 5     | Frontend foundation & design system     | Done        |
| 6     | Frontend features                       | Done        |
| 7     | Integration & full-stack Docker Compose | Done        |
| 8     | Service tokens (API keys)               | Done        |

Status values: `Not Started`, `In Progress`, `Blocked`, `Done`.

---

## Phase 0 — Planning & documentation

- [x] Lock architectural decisions
- [x] `development.md` progress tracker
- [x] `docs/architecture.md`
- [x] `docs/database-schema.md`
- [x] `docs/api-spec.md`
- [x] `docs/security.md`
- [x] `docs/backend-guidelines.md`
- [x] `docs/frontend-guidelines.md`

## Phase 1 — Backend foundation & infra

- [x] `backend/` Go module scaffold (`go.mod`, folder layout per backend-guidelines)
- [x] Config loader (env-based) in `internal/config`
- [x] logrus logger setup in `internal/logger` (structured JSON, request id)
- [x] Postgres connection pool (`database/sql` + pgx stdlib driver)
- [x] Migration runner + initial migrations (golang-migrate)
- [x] Gin server bootstrap, health endpoint, graceful shutdown
- [x] Base middleware: request id, recovery, logging, CORS
- [x] `backend/Dockerfile` (multi-stage, non-root, distroless/alpine)
- [x] Standard error & response helpers

## Phase 2 — Crypto core (envelope encryption)

- [x] Argon2id KDF helper (password → KEK)
- [x] AES-256-GCM encrypt/decrypt helpers (with nonce handling)
- [x] Envelope service: unwrap DEK with KEK at startup; encrypt/decrypt secrets with DEK
- [x] DEK bootstrap flow (generate + wrap on first run; load + unwrap on subsequent runs)
- [x] Unit tests for crypto (round-trip, tamper detection)
- [x] Zeroization of sensitive buffers where feasible

## Phase 3 — Auth & users (JWT + RBAC)

- [x] `users` repository + service
- [x] Password hashing (Argon2id) for user login
- [x] Register / login handlers
- [x] JWT issue (access + refresh), refresh rotation, logout
- [x] Auth middleware (validate access token, inject user into context)
- [x] RBAC roles (owner / admin / member) + membership checks
- [x] Rate limiting on auth endpoints

## Phase 4 — Secrets domain

- [x] `projects` (+ environments) repository + service + handlers
- [x] Project membership / RBAC enforcement
- [x] `secrets` repository + service + handlers (CRUD, encrypted at rest)
- [x] Secret versioning (history) on update
- [x] Audit log on every read/write/delete
- [x] Never log plaintext secret values

## Phase 5 — Frontend foundation & design system

- [x] Vite + React + TS scaffold (strict mode, no `any`)
- [x] Tailwind config + design tokens (colors, typography, spacing)
- [x] Folder layers: `components`, `pages`, `services`, `hooks`, `utils`, `types`, `lib`, `layouts`
- [x] API client (axios) + interceptors (auth header, refresh-on-401)
- [x] Auth context / token storage strategy
- [x] Routing (react-router) + protected routes
- [x] Base UI primitives (Button, Input, Card, etc.) using lucide-react icons
- [x] App shell / layout (sidebar nav, top bar)

## Phase 6 — Frontend features

- [x] Login / register pages
- [x] Projects list + create/edit
- [x] Project detail with environment switcher
- [x] Secrets table: list, reveal, copy, create, edit, delete
- [x] Secret version history view
- [x] Audit log view
- [x] Members / RBAC management UI
- [x] Loading, empty, and error states for all views

## Phase 7 — Integration & full-stack Docker Compose

- [x] `docker-compose.yml` (api + postgres + frontend + healthchecks)
- [x] `docker-compose.override.yml` for dev
- [x] `.env.example` with all required vars (no real secrets)
- [x] Postgres healthcheck + `depends_on: service_healthy`
- [x] Migration step wired into startup
- [x] End-to-end smoke test: register → login → create project → create/read secret
- [x] Root `README.md` with run instructions

## Phase 8 — Service tokens (API keys)

See `docs/service-tokens.md`. Per-environment, read-only, optional expiry (default 90d).

- [x] Migration `0008_service_tokens` + domain entity & repository interface
- [x] `service_tokens` repository (create, get-by-hash, list-by-project, revoke, touch)
- [x] Token generation (`ck_live_` + 32 random bytes, SHA-256 hash, display hint)
- [x] `Principal` abstraction; authenticator resolves JWT *or* `ck_` token
- [x] Auth middleware sets principal; user-only routes reject token principals (403)
- [x] Secret read/list/export authorize a token principal by environment scope
- [x] Token management endpoints (list/create/revoke), admin-only, plaintext shown once
- [x] Audit `token.create` / `token.revoke`; token usage audited with token actor
- [x] Frontend: project "Tokens" tab — list, create (one-time reveal + copy), revoke
- [x] Docs: api-spec updated; verified end-to-end (create → use → scope/write denied → revoke)

---

## Decision Log

- 2026-06-30 (hierarchical JSON/YAML): import/export now nest dotted keys. Stored model
  stays flat string→string; `secretfmt` flattens nested JSON/YAML to dotted keys on
  import and nests them back on export (delimiter `.`, objects-only — numeric segments
  are string keys, no array reconstruction). `.env` stays flat (literal dotted keys).
  Key conflicts (a key that is both a value and a group) → `400 VALIDATION_ERROR` on
  JSON/YAML export (not 500), `.env` still works. Formats are a small registry in
  `secretfmt` (extension/content-type/hierarchical) so adding a format is one entry.
  FE: the Export modal became a view/convert panel — format switcher (env/json/yaml),
  raw preview, copy, download; one-time token reveal already had a usage snippet.
  Verified live: nested YAML import → flat dotted storage → nested YAML/JSON export +
  flat `.env`; conflict → 400. Unit tests: flatten, nested round-trip, conflict.
- 2026-06-30 (FE gaps + account): added change-password (backend `POST /auth/change-password`
  — verifies current, revokes all refresh tokens; FE Settings page at `/settings` reachable
  from the topbar), a global `ErrorBoundary`, a one-time token-reveal usage snippet
  (curl + `SECRETS_API`/`SECRETS_ENV_ID`/`SECRETS_TOKEN`), secret key search/filter, and a
  "Session expired" toast on forced logout. Change-password verified end-to-end (wrong
  current 401, short 400, success 204, old password 401, new 200, old refresh revoked 401).
  Not done: forgot-password/reset (requires email/SMTP infrastructure).
- 2026-06-30 (security fixes): closed the four validated pentest findings and
  re-tested each live. (1) `SetTrustedProxies` + `TRUSTED_PROXIES` (default trust none)
  — rotating `X-Forwarded-For` went from 0/30 to 13/30 blocked. (2) Startup validation
  of `JWT_SECRET` (≥32) / `MASTER_PASSWORD` (≥16) + placeholder rejection (fail-fast;
  unit-tested in `internal/config`). (3) `BodyLimit` middleware (`MAX_BODY_BYTES`, 1 MiB)
  + 1000-key import cap — oversized body and 1500-key import both rejected with 400.
  (4) Dummy Argon2 verify on unknown/inactive login — timing went from ~50ms-vs-~2ms to
  ~63ms-vs-~60ms (indistinguishable). New env vars: `TRUSTED_PROXIES`, `MAX_BODY_BYTES`.
- 2026-06-30 (Phase 8): Service tokens are per-environment, read-only, distinct
  principals (not users), with optional expiry (default 90 days; `expires_in_days=0`
  means never). Prefix `ck_live_`, stored SHA-256-hashed, plaintext shown once.
  Auth middleware resolves a `Principal` (user JWT or `ck_` token); read endpoints
  (`secrets list/get`, `export`) accept either, everything else requires a user via
  `RequireUser` (token → 403). Verified end-to-end: scoped read/export 200; write,
  cross-environment, and management all 403; revoke → 401.
- 2026-06-28: Persistence = `database/sql` + pgx manual (no ORM/sqlc) per user choice.
- 2026-06-28: Auth = JWT access + refresh.
- 2026-06-28: Server-side encryption with envelope encryption (KEK from master password).
- 2026-06-28: Frontend icon set = lucide-react (real icons, never emoji).
- 2026-06-28: Frontend palette = neutral slate gray scale + single teal accent
  (tokens in `tailwind.config.ts`). No gradients/glassmorphism.
- 2026-06-28: Frontend server state = TanStack Query over one axios instance
  (`lib/api.ts`); centralized typed query keys in `lib/queryKeys.ts`.
- 2026-06-28: Token storage = localStorage (`lib/tokenStore.ts`); axios 401
  interceptor performs a single coalesced `/auth/refresh` + retry, emits a
  logout event on failure which `AuthProvider` consumes to clear the session.
- 2026-06-28: ESLint uses flat config (`eslint.config.js`) with
  `no-explicit-any`, `no-floating-promises`, and `consistent-type-imports`.
- 2026-06-28 (Phase 7): `docker-compose.yml` publishes the API on host `8080` and the
  SPA on `5173`; the SPA's `VITE_API_BASE_URL` is baked at build time to the host API
  URL since the browser calls the API directly. Postgres has a `pg_isready` healthcheck
  and the API uses `depends_on: service_healthy`.
- 2026-06-28 (Phase 7 fix): hardened `database.Connect` with bounded ping retry/backoff
  (~30s). The previous single-shot ping made the API exit in milliseconds when the DB
  was not yet reachable; under Compose that tight crash loop left the container detached
  from the user network (host DNS resolver, unreachable), so it never recovered.
- 2026-06-28 (feature): added bulk import/export of secrets (.env / JSON / YAML).
  Backend: `internal/secretfmt` codec (+ unit tests), `SecretService.Import`
  (single-tx upsert with `overwrite` flag, audited `secret.import`) and `Export`
  (audited `secret.export`); routes `POST /environments/:id/import` and
  `GET /environments/:id/export`. Frontend: `ImportSecretsModal` (paste or file upload,
  format auto-detect, overwrite toggle), `ExportSecretsModal` (format + plaintext
  warning, client-side download), service/hooks/`download` util, toolbar buttons.
  Verified end-to-end against the live stack across all three formats. `go build`/`go vet`/
  `go test` and `npm run build`/`npm run lint` all clean.
- 2026-06-28 (post-build fix): wired up environment deletion in the UI. The backend
  endpoint, `environment.service.remove`, and `useDeleteEnvironment` hook already
  existed, but no control invoked them. Added an admin-only delete (Trash2) button to
  `EnvironmentSwitcher` + a `ConfirmDialog` in `ProjectDetailPage`; build and lint clean.
- 2026-06-28 (Phase 7): full stack verified via `scripts/smoke-test.sh` — health,
  register, login, create project, create environment, create secret, decrypt
  round-trip (value matches), and refresh-token rotation all pass against the running
  containers. Frontend serves built assets and CORS preflight from the SPA origin succeeds.
- 2026-06-28 (backend): Migrations embedded via `go:embed` in
  `backend/migrations` and run on startup with golang-migrate's iofs source —
  no external migrate CLI required.
- 2026-06-28 (backend): Repositories take a `domain.Querier` (satisfied by
  `*sql.DB` or `*sql.Tx`); `database.WithTx` provides the unit-of-work used for
  atomic flows (token rotation, project+owner, secret+version).
- 2026-06-28 (backend): Audit logs are scoped to a project via
  `metadata->>'project_id'` (jsonb), since `audit_logs` has no direct
  `project_id` column in the schema. The audit list endpoint filters on it.
- 2026-06-28 (backend): Project `slug` is `slugify(name)` plus a short uuid
  suffix to guarantee global uniqueness without a user-supplied slug field
  (the API create body only accepts `name`).
- 2026-06-28 (backend): Refresh tokens hashed with SHA-256 (hex) for
  constant-cost lookup by hash, per security.md.

## Open Questions

_(Agents append blockers / questions here for the user.)_
