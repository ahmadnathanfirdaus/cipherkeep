# Security Design

## Threat model (scope)

- **Trusted**: the API server process and its memory. Server-side encryption means the
  server can decrypt secret values to serve authorized users.
- **Protected against**: database compromise / disk theft (secrets are ciphertext at
  rest, the DEK is wrapped), token theft mitigation, and accidental plaintext leakage
  via logs.
- **Out of scope (v1)**: a malicious server operator with live memory access; HSM-grade
  key protection; end-to-end (zero-knowledge) encryption.

## Envelope encryption

Two-layer key hierarchy so the master password can be rotated without re-encrypting
every secret.

```
MASTER_PASSWORD (env / docker secret, never persisted)
      │  Argon2id(salt = encryption_keys.kdf_salt)
      ▼
   KEK  (Key Encryption Key, 32 bytes, memory only)
      │  AES-256-GCM decrypt(encryption_keys.wrapped_dek, nonce)
      ▼
   DEK  (Data Encryption Key, 32 bytes, memory only)
      │  AES-256-GCM
      ▼
 secret value  ⇄  (ciphertext, nonce) stored in `secrets`
```

### Bootstrap (first run)

1. Read `MASTER_PASSWORD` from env / mounted secret.
2. If no active row in `encryption_keys`:
   - Generate random `kdf_salt` (16 bytes).
   - Derive KEK = Argon2id(MASTER_PASSWORD, kdf_salt).
   - Generate random DEK (32 bytes via `crypto/rand`).
   - Wrap: `wrapped_dek, nonce = AES-256-GCM.Seal(KEK, DEK)`.
   - Insert `encryption_keys { wrapped_dek, nonce, kdf_salt, is_active: true }`.
3. Hold the DEK in memory for the process lifetime.

### Startup (subsequent runs)

1. Load active `encryption_keys` row.
2. Derive KEK = Argon2id(MASTER_PASSWORD, kdf_salt).
3. DEK = AES-256-GCM.Open(KEK, wrapped_dek, nonce). If this fails, the master password
   is wrong → **fail fast, do not start**.

### Secret encryption

- Encrypt: generate a fresh random 12-byte nonce, `ciphertext = AES-256-GCM.Seal(DEK, plaintext, nonce)`.
- Store `(ciphertext, nonce)`. The GCM auth tag is appended to ciphertext.
- Decrypt: `plaintext = AES-256-GCM.Open(DEK, ciphertext, nonce)`. A failure means
  tampering or wrong key → return error, never partial data.
- **Nonce uniqueness**: always `crypto/rand`; never reuse a nonce with the same key.

## Parameters

- **Argon2id** (KEK derivation and user password hashing): time=1–3, memory=64 MiB,
  threads=number of cores (tune for ~target latency), key length 32 bytes. Store the
  encoded parameters with user password hashes so they can be upgraded later.
- **AES-256-GCM**: 32-byte keys, 12-byte nonces, 16-byte tag.
- **JWT access token**: short-lived (e.g. 15 min), signed HS256 (or RS256 if asymmetric
  needed later) with `JWT_SECRET`. Claims: `sub` (user id), `iat`, `exp`, `jti`.
- **Refresh token**: opaque random 32+ bytes, stored **hashed** (SHA-256) in
  `refresh_tokens`. Rotation on every refresh; old token revoked. Longer TTL (e.g. 7 days).

## Authentication & authorization

- Passwords hashed with Argon2id; never stored or logged in plaintext.
- Auth middleware validates the access token and loads the user; rejects expired/invalid.
- **RBAC**: roles `owner > admin > member`, enforced per project in the service layer
  (not just handlers). Service methods take the acting user and verify membership/role
  before any data access.
- Rate limit auth endpoints (login/register/refresh) to slow brute force.

## Secrets handling rules (mandatory)

- Plaintext secret values exist only transiently in memory and in the single-secret
  GET response. **Never** log them, never put them in audit `metadata`, error messages,
  or URL/query strings.
- List endpoints return metadata only (no values).
- Zero out sensitive byte slices after use where practical (`crypto` package buffers,
  decrypted plaintext) — best-effort in Go.
- Reading a secret value is itself audited (`secret.read`).

## Abuse resistance (hardening)

These mitigations were added after a security assessment:

- **Trusted proxies**: `X-Forwarded-For` is honored only from CIDRs listed in
  `TRUSTED_PROXIES` (empty by default → trust none). This prevents a client from
  spoofing its IP to evade the per-IP auth rate limiter. Set it to your reverse
  proxy's network in production.
- **Secret strength**: startup fails fast unless `JWT_SECRET` ≥ 32 chars and
  `MASTER_PASSWORD` ≥ 16 chars, and neither is a recognizable placeholder. A weak
  `JWT_SECRET` would allow token forgery; a weak `MASTER_PASSWORD` would make the
  wrapped DEK brute-forceable.
- **Request limits**: a body-size cap (`MAX_BODY_BYTES`, default 1 MiB) and a
  per-import key cap (1000) bound memory and database work per request.
- **Login timing**: a dummy Argon2 verification runs when the email is unknown or the
  account is inactive, so response time does not reveal whether an account exists.

## Transport & deployment

- TLS terminates at a reverse proxy (Caddy/Traefik/nginx) in front of the API in
  production. The API itself speaks HTTP inside the trusted network.
- Containers run as non-root.
- `MASTER_PASSWORD` and `JWT_SECRET` are provided via env or Docker secrets, never
  committed. `.env.example` documents them with placeholder values only.
- CORS restricted to the known frontend origin(s).

## Backup & recovery

- Losing `MASTER_PASSWORD` makes the DEK unrecoverable → all secrets unreadable.
  Document this prominently. Operators must back up the master password securely.
- Database backups contain only ciphertext + the wrapped DEK, which are useless without
  the master password.
