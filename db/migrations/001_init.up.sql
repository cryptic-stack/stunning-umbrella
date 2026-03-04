-- +migrate Up
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS uploaded_files (
    id BIGSERIAL PRIMARY KEY,
    framework TEXT,
    version TEXT,
    filename TEXT NOT NULL,
    stored_path TEXT NOT NULL,
    file_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_versions_framework ON versions(framework_id);
CREATE INDEX IF NOT EXISTS idx_controls_version ON controls(version_id);
CREATE INDEX IF NOT EXISTS idx_safeguards_version ON safeguards(version_id);
CREATE INDEX IF NOT EXISTS idx_diff_items_report ON diff_items(report_id);
