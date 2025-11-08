CREATE TABLE orbitum.audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES orbitum.users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON orbitum.audit_logs(user_id);
