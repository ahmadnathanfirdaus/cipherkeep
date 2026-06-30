CREATE TABLE encryption_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wrapped_dek BYTEA NOT NULL,
    nonce       BYTEA NOT NULL,
    kdf_salt    BYTEA NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Only one active encryption key row is permitted in v1.
CREATE UNIQUE INDEX idx_encryption_keys_single_active
    ON encryption_keys (is_active)
    WHERE is_active = TRUE;
