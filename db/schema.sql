CREATE TABLE IF NOT EXISTS frameworks (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS versions (
    id BIGSERIAL PRIMARY KEY,
    framework_id BIGINT NOT NULL REFERENCES frameworks(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    release_date DATE,
    source_file TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (framework_id, version)
);

CREATE TABLE IF NOT EXISTS controls (
    id BIGSERIAL PRIMARY KEY,
    framework_id BIGINT NOT NULL REFERENCES frameworks(id) ON DELETE CASCADE,
    version_id BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    control_id TEXT NOT NULL,
    title TEXT,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (version_id, control_id)
);

CREATE TABLE IF NOT EXISTS safeguards (
    id BIGSERIAL PRIMARY KEY,
    control_id BIGINT NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
    version_id BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    safeguard_id TEXT NOT NULL,
    title TEXT,
    description TEXT,
    level TEXT DEFAULT '',
    ig1 BOOLEAN NOT NULL DEFAULT FALSE,
    ig2 BOOLEAN NOT NULL DEFAULT FALSE,
    ig3 BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (version_id, safeguard_id)
);

CREATE TABLE IF NOT EXISTS diff_reports (
    id BIGSERIAL PRIMARY KEY,
    framework_id BIGINT NOT NULL REFERENCES frameworks(id) ON DELETE CASCADE,
    version_a BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    version_b BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    control_level TEXT NOT NULL DEFAULT 'ALL',
    status TEXT NOT NULL DEFAULT 'queued',
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS diff_items (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES diff_reports(id) ON DELETE CASCADE,
    change_type TEXT NOT NULL,
    safeguard_old TEXT,
    safeguard_new TEXT,
    old_text TEXT,
    new_text TEXT,
    similarity NUMERIC(5,2) NOT NULL DEFAULT 0,
    reviewed BOOLEAN NOT NULL DEFAULT FALSE,
    review_comment TEXT DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS uploaded_files (
    id BIGSERIAL PRIMARY KEY,
    framework TEXT,
    version TEXT,
    filename TEXT NOT NULL,
    stored_path TEXT NOT NULL,
    file_type TEXT NOT NULL,
    file_hash TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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

CREATE INDEX IF NOT EXISTS idx_versions_framework ON versions(framework_id);
CREATE INDEX IF NOT EXISTS idx_controls_version ON controls(version_id);
CREATE INDEX IF NOT EXISTS idx_safeguards_version ON safeguards(version_id);
CREATE INDEX IF NOT EXISTS idx_diff_items_report ON diff_items(report_id);
CREATE INDEX IF NOT EXISTS idx_diff_items_report_reviewed ON diff_items(report_id, reviewed);
CREATE INDEX IF NOT EXISTS idx_uploaded_files_file_hash ON uploaded_files(file_hash);
CREATE INDEX IF NOT EXISTS idx_app_users_role_id ON app_users(role_id);

INSERT INTO org_settings (org_name)
SELECT 'CIS Benchmark Intelligence'
WHERE NOT EXISTS (SELECT 1 FROM org_settings);

INSERT INTO roles (name, description, is_system) VALUES
('Admin', 'Full administrative access', TRUE),
('Reviewer', 'Review and comment on diff reports', TRUE),
('Viewer', 'Read-only access', TRUE)
ON CONFLICT (name) DO NOTHING;
