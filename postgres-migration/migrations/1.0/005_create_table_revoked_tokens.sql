CREATE TABLE orbitum.revoked_tokens (
    token_id UUID PRIMARY KEY,
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
