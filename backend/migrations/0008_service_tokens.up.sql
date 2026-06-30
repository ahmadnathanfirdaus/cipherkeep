CREATE TABLE service_tokens (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    token_hash     TEXT NOT NULL UNIQUE,
    display_hint   TEXT NOT NULL,
    created_by     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    expires_at     TIMESTAMPTZ,
    last_used_at   TIMESTAMPTZ,
    revoked_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_service_tokens_project ON service_tokens (project_id);
CREATE INDEX idx_service_tokens_hash ON service_tokens (token_hash);
