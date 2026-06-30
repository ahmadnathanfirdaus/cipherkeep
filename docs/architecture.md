# Architecture

## Overview

A self-hosted, team-oriented secret manager. Users authenticate with per-user
accounts, organize secrets into projects and environments, and the server stores all
secret values encrypted at rest. The server is trusted to decrypt (server-side
encryption model) so it can serve plaintext to authorized users and enforce RBAC.

```
┌──────────────┐       HTTPS/JSON       ┌──────────────────────────┐      ┌────────────┐
│  Web UI      │ ─────────────────────▶ │  API Server (Go + Gin)   │ ───▶ │ PostgreSQL │
│  React + Vite│ ◀───────────────────── │  service-repository      │ ◀─── │            │
└──────────────┘                        │  JWT auth, RBAC, crypto  │      └────────────┘
                                        └──────────────────────────┘
                                                   │
                                          Master password (env/secret)
                                          → Argon2id → KEK → unwrap DEK
```

## Components

- **API Server (Go + Gin)** — REST API. Layered as handler → service → repository.
  Cross-cutting concerns (logging, auth, request id, recovery, CORS) via middleware.
- **PostgreSQL** — stores users, projects, environments, secrets (ciphertext),
  versions, audit logs, and the wrapped DEK.
- **Web UI (React)** — SPA consuming the REST API. Strict TypeScript, Tailwind,
  lucide-react icons, clear layer separation.
- **Crypto subsystem** — envelope encryption. See `security.md`.

## Backend layering (service-repository)

```
handler  (Gin)        HTTP concerns: bind/validate input, map to DTO, write response.
   │                  No business logic.
service               Business logic, authorization checks, orchestration, crypto calls.
   │                  Depends on repository INTERFACES, not concrete types.
repository            Data access. SQL via database/sql + pgx. One repo per aggregate.
   │
PostgreSQL
```

- **Dependency direction**: handler → service → repository. Inner layers never import
  outer layers.
- **Interfaces** for repositories (and services where useful) live in the domain layer
  so services depend on abstractions and are unit-testable with mocks.
- **Entities/DTOs**: domain entities are plain structs; request/response DTOs are
  separate from entities.

## Repository layout (monorepo)

```
cipherkeep/
├── development.md
├── docs/
├── docker-compose.yml
├── docker-compose.override.yml
├── .env.example
├── README.md
├── backend/
│   ├── go.mod
│   ├── Dockerfile
│   ├── cmd/server/main.go
│   ├── migrations/
│   └── internal/
│       ├── config/          # env config loading
│       ├── logger/          # logrus setup
│       ├── database/        # connection pool, migration runner
│       ├── server/          # gin router, middleware wiring, graceful shutdown
│       ├── middleware/      # request id, auth, logging, recovery, cors, ratelimit
│       ├── crypto/          # argon2id, aes-gcm, envelope service
│       ├── domain/          # entities + repository/service interfaces + errors
│       ├── repository/      # concrete repositories (users, projects, secrets, audit)
│       ├── service/         # business logic (auth, user, project, secret, audit)
│       ├── handler/         # gin handlers grouped by resource
│       └── httputil/        # response/error helpers, DTOs
└── frontend/
    ├── package.json
    ├── Dockerfile
    ├── vite.config.ts
    ├── tailwind.config.ts
    ├── tsconfig.json
    └── src/
        ├── main.tsx
        ├── App.tsx
        ├── components/      # reusable presentational + composite components
        ├── pages/           # route-level screens
        ├── layouts/         # app shell, auth layout
        ├── services/        # API clients (axios), one module per resource
        ├── hooks/           # custom hooks (data fetching, auth)
        ├── context/         # React contexts (auth)
        ├── types/           # shared TS types/interfaces (mirror API DTOs)
        ├── utils/           # pure helpers (formatters, validators)
        ├── lib/             # configured 3rd-party clients (axios instance, query client)
        └── routes/          # route definitions, protected route wrapper
```

## Request flow (example: read a secret)

1. FE sends `GET /api/v1/projects/{id}/secrets/{key}` with `Authorization: Bearer <access>`.
2. Auth middleware validates JWT, loads user into context.
3. Handler binds params, calls `SecretService.Get(ctx, user, projectID, key)`.
4. Service checks project membership/role (RBAC), loads ciphertext via repository,
   decrypts with the envelope service, writes an audit log entry.
5. Handler returns the plaintext value in the response DTO.

## Non-goals (initial version)

- Zero-knowledge / end-to-end encryption (server is trusted to decrypt).
- Dynamic secrets / leasing (Vault-style).
- HSM integration (master password via env/secret instead).
