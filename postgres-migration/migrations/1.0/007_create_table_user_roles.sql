CREATE TABLE orbitum.user_roles (
    user_id UUID REFERENCES orbitum.users(id) ON DELETE CASCADE,
    role_id INT REFERENCES orbitum.roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role_id ON orbitum.user_roles(role_id);
