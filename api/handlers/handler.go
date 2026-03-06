package handlers

import (
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handler struct {
	DB        *gorm.DB
	Redis     *redis.Client
	UploadDir string
	ExportDir string
}

func NewHandler(db *gorm.DB, redisClient *redis.Client, uploadDir, exportDir string) *Handler {
	if db != nil {
		_ = db.Exec("ALTER TABLE diff_reports ADD COLUMN IF NOT EXISTS control_level TEXT NOT NULL DEFAULT 'ALL'").Error
		_ = db.Exec("ALTER TABLE uploaded_files ADD COLUMN IF NOT EXISTS file_hash TEXT DEFAULT ''").Error
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_uploaded_files_file_hash ON uploaded_files(file_hash)").Error
		_ = db.Exec("ALTER TABLE diff_items ADD COLUMN IF NOT EXISTS reviewed BOOLEAN NOT NULL DEFAULT FALSE").Error
		_ = db.Exec("ALTER TABLE diff_items ADD COLUMN IF NOT EXISTS review_comment TEXT DEFAULT ''").Error
		_ = db.Exec("ALTER TABLE diff_items ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ").Error
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_diff_items_report_reviewed ON diff_items(report_id, reviewed)").Error
		_ = db.Exec(`
CREATE TABLE IF NOT EXISTS org_settings (
	id BIGSERIAL PRIMARY KEY,
	org_name TEXT NOT NULL DEFAULT '',
	logo_url TEXT NOT NULL DEFAULT '',
	primary_color TEXT NOT NULL DEFAULT '',
	secondary_color TEXT NOT NULL DEFAULT '',
	support_email TEXT NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`).Error
		_ = db.Exec(`
CREATE TABLE IF NOT EXISTS roles (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL DEFAULT '',
	is_system BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`).Error
		_ = db.Exec(`
CREATE TABLE IF NOT EXISTS app_users (
	id BIGSERIAL PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL DEFAULT '',
	role_id BIGINT REFERENCES roles(id) ON DELETE SET NULL,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`).Error
		_ = db.Exec(`
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
)`).Error
		_ = db.Exec(`
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
)`).Error
		_ = db.Exec(`
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
)`).Error
		_ = db.Exec(`
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
)`).Error
		_ = db.Exec(`
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
)`).Error
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_app_users_role_id ON app_users(role_id)").Error
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_policy_settings_source_key ON policy_settings(policy_source_id, setting_key)").Error
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_benchmark_policy_rules_key ON benchmark_policy_rules(setting_key)").Error
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_assessment_runs_status_created ON assessment_runs(status, created_at DESC)").Error
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_assessment_results_run_status ON assessment_results(assessment_run_id, status)").Error
		_ = db.Exec(`
INSERT INTO org_settings (org_name)
SELECT ''
WHERE NOT EXISTS (SELECT 1 FROM org_settings)
`).Error
		_ = db.Exec(`
INSERT INTO roles (name, description, is_system) VALUES
('Admin', 'Full administrative access', TRUE),
('Reviewer', 'Review and comment on diff reports', TRUE),
('Viewer', 'Read-only access', TRUE)
ON CONFLICT (name) DO NOTHING
`).Error

		bootstrapAdminEmail := strings.TrimSpace(strings.ToLower(os.Getenv("AUTH_BOOTSTRAP_ADMIN_EMAIL")))
		if bootstrapAdminEmail != "" {
			_ = db.Exec(`
INSERT INTO app_users (email, display_name, role_id, is_active)
SELECT ?, ?, roles.id, TRUE
FROM roles
WHERE roles.name = 'Admin'
ON CONFLICT (email)
DO UPDATE SET role_id = EXCLUDED.role_id, is_active = TRUE
`, bootstrapAdminEmail, bootstrapAdminEmail).Error
		}
	}
	return &Handler{DB: db, Redis: redisClient, UploadDir: uploadDir, ExportDir: exportDir}
}
