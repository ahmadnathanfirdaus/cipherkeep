# Backend Guidelines

Go + Gin, service-repository pattern, `database/sql` + pgx (manual SQL), logrus.

## Principles

- **Layering**: `handler → service → repository`. Dependencies point inward only.
- **Depend on interfaces**: services depend on repository interfaces (declared in
  `internal/domain`); concrete implementations live in `internal/repository`. Wire
  concrete types together in `cmd/server/main.go` (manual dependency injection).
- **No business logic in handlers**: handlers bind/validate input, call a service, and
  map the result (or error) to an HTTP response.
- **No HTTP in services/repositories**: services and repos must not import `gin` or
  reference `http`. They take/return domain types and `context.Context`.
- **Errors**: use sentinel domain errors (e.g. `domain.ErrNotFound`,
  `domain.ErrForbidden`, `domain.ErrConflict`) returned from services; a central error
  mapper in `httputil` translates them to status codes + the error envelope.

## Package responsibilities

| Package              | Responsibility                                                        |
| -------------------- | --------------------------------------------------------------------- |
| `cmd/server`         | `main()`: load config, init logger/db/crypto, wire deps, start server. |
| `internal/config`    | Load + validate env config into a typed struct. Fail fast if missing. |
| `internal/logger`    | logrus setup (JSON formatter, level from config, fields helper).      |
| `internal/database`  | `*sql.DB` pool via pgx stdlib driver; migration runner.               |
| `internal/server`    | Build the Gin engine, register routes, middleware, graceful shutdown. |
| `internal/middleware`| request id, recovery, request logging, CORS, auth, rate limit.        |
| `internal/crypto`    | Argon2id, AES-256-GCM, envelope service (KEK/DEK).                     |
| `internal/domain`    | Entities, repository & service interfaces, sentinel errors.           |
| `internal/repository`| Concrete repositories implementing domain interfaces (SQL).           |
| `internal/service`   | Business logic, RBAC checks, orchestration, crypto usage.             |
| `internal/handler`   | Gin handlers grouped per resource; input DTOs + validation.           |
| `internal/httputil`  | Response/error envelope helpers, error→status mapping, DTO helpers.   |

## Conventions

- **Context**: every repository and service method takes `ctx context.Context` first.
- **Transactions**: when a service needs atomicity across repos, expose a unit-of-work
  / `WithTx` helper on the database layer; repositories accept a `Querier` interface
  (`*sql.DB` or `*sql.Tx`) so they work inside or outside a transaction.
- **SQL**: parameterized queries only (`$1, $2`) — never string-concatenate input.
  Keep SQL in the repository files as named constants for readability.
- **Validation**: use struct tags + `binding` in handlers for shape validation; deeper
  business rules validated in the service.
- **Logging**: structured logrus with fields (`request_id`, `user_id`, `action`).
  **Never** log secret values, passwords, tokens, or the master password/DEK. Log at
  `Info` for normal flow, `Warn` for handled client errors, `Error` for unexpected.
- **Config**: all configuration via env vars; provide sane defaults only for
  non-sensitive values. Required secrets (DB DSN, `MASTER_PASSWORD`, `JWT_SECRET`)
  must be present or the app refuses to start.
- **IDs/time**: generate UUIDs in DB (`gen_random_uuid()`); use `time.Time` (UTC).
- **Testing**: unit-test services with mocked repositories; unit-test the crypto
  package thoroughly (round-trip, tamper, wrong-key). Table-driven tests.

## Naming

- Interfaces: `UserRepository`, `SecretService`, etc.
- Concrete repos: `userRepository` (unexported struct) + `NewUserRepository(...) UserRepository`.
- Handlers: `UserHandler` with methods `Register`, `Login`, etc.
- One file per resource within a package (e.g. `service/secret.go`, `repository/secret.go`).

## HTTP response helpers (`httputil`)

- `Respond(c, status, data)` → wraps in `{ "data": ... }`.
- `RespondList(c, data, meta)` → `{ "data": [...], "meta": {...} }`.
- `RespondError(c, err)` → maps domain error to status + error envelope.

## Required env vars

See `.env.example`. At minimum:

```
APP_ENV=development
HTTP_PORT=8080
DATABASE_URL=postgres://user:pass@postgres:5432/cipherkeep?sslmode=disable
MASTER_PASSWORD=...           # required, used to derive KEK
JWT_SECRET=...                # required, signs access tokens
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h
CORS_ALLOWED_ORIGINS=http://localhost:5173
LOG_LEVEL=info
```

## Dockerfile expectations

- Multi-stage: `golang:1.x` build stage → minimal runtime (`gcr.io/distroless/static`
  or `alpine`). Static build (`CGO_ENABLED=0`).
- Run as non-root user.
- Migrations run on startup (or via a one-shot entrypoint step) before serving.
