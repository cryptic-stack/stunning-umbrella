CREATE TABLE IF NOT EXISTS policy_sources (
    id BIGSERIAL PRIMARY KEY,
    source_type TEXT NOT NULL,
    source_name TEXT NOT NULL DEFAULT '',
    hostname TEXT NOT NULL DEFAULT '',
    domain_name TEXT NOT NULL DEFAULT '',
    collected_at TIMESTAMPTZ,
    raw_path TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS policy_settings (
    id BIGSERIAL PRIMARY KEY,
    policy_source_id BIGINT NOT NULL REFERENCES policy_sources(id) ON DELETE CASCADE,
    setting_key TEXT NOT NULL,
    setting_name TEXT NOT NULL DEFAULT '',
    canonical_type TEXT NOT NULL DEFAULT '',
    scope TEXT NOT NULL DEFAULT '',
    value_text TEXT NOT NULL DEFAULT '',
    value_number NUMERIC,
    value_bool BOOLEAN,
    value_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS benchmark_policy_rules (
    id BIGSERIAL PRIMARY KEY,
    framework_id BIGINT REFERENCES frameworks(id) ON DELETE SET NULL,
    version_id BIGINT REFERENCES versions(id) ON DELETE SET NULL,
    rule_id TEXT NOT NULL,
    benchmark_ref TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    setting_key TEXT NOT NULL,
    check_type TEXT NOT NULL,
    expected_value JSONB NOT NULL DEFAULT '{}'::jsonb,
    severity TEXT NOT NULL DEFAULT '',
    source_label TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (framework_id, version_id, rule_id)
);

CREATE TABLE IF NOT EXISTS assessment_runs (
    id BIGSERIAL PRIMARY KEY,
    policy_source_id BIGINT NOT NULL REFERENCES policy_sources(id) ON DELETE CASCADE,
    framework_id BIGINT REFERENCES frameworks(id) ON DELETE SET NULL,
    version_id BIGINT REFERENCES versions(id) ON DELETE SET NULL,
    mapping_label TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'queued',
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS assessment_results (
    id BIGSERIAL PRIMARY KEY,
    assessment_run_id BIGINT NOT NULL REFERENCES assessment_runs(id) ON DELETE CASCADE,
    benchmark_policy_rule_id BIGINT REFERENCES benchmark_policy_rules(id) ON DELETE SET NULL,
    rule_id TEXT NOT NULL DEFAULT '',
    setting_key TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    actual_value JSONB NOT NULL DEFAULT '{}'::jsonb,
    expected_value JSONB NOT NULL DEFAULT '{}'::jsonb,
    details TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policy_settings_source_key ON policy_settings(policy_source_id, setting_key);
CREATE INDEX IF NOT EXISTS idx_benchmark_policy_rules_key ON benchmark_policy_rules(setting_key);
CREATE INDEX IF NOT EXISTS idx_assessment_runs_status_created ON assessment_runs(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_assessment_results_run_status ON assessment_results(assessment_run_id, status);
