-- +migrate Up
CREATE TABLE IF NOT EXISTS org_settings (
    id BIGSERIAL PRIMARY KEY,
    org_name TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    primary_color TEXT NOT NULL DEFAULT '',
    secondary_color TEXT NOT NULL DEFAULT '',
    support_email TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS app_users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    role_id BIGINT REFERENCES roles(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_app_users_role_id ON app_users(role_id);

INSERT INTO org_settings (org_name)
SELECT 'CIS Benchmark Intelligence'
WHERE NOT EXISTS (SELECT 1 FROM org_settings);

INSERT INTO roles (name, description, is_system) VALUES
('Admin', 'Full administrative access', TRUE),
('Reviewer', 'Review and comment on diff reports', TRUE),
('Viewer', 'Read-only access', TRUE)
ON CONFLICT (name) DO NOTHING;
