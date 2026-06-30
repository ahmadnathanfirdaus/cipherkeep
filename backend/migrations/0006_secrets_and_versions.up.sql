CREATE TABLE secrets (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    key            TEXT NOT NULL,
    ciphertext     BYTEA NOT NULL,
    nonce          BYTEA NOT NULL,
    version        INT NOT NULL DEFAULT 1,
    created_by     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_by     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (environment_id, key)
);

CREATE INDEX idx_secrets_environment_id ON secrets (environment_id);

CREATE TABLE secret_versions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id  UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    version    INT NOT NULL,
    ciphertext BYTEA NOT NULL,
    nonce      BYTEA NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (secret_id, version)
);

CREATE INDEX idx_secret_versions_secret_id ON secret_versions (secret_id);
