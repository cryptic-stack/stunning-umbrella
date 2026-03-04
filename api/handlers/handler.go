package handlers

import (
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
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_app_users_role_id ON app_users(role_id)").Error
		_ = db.Exec(`
INSERT INTO org_settings (org_name)
SELECT 'CIS Benchmark Intelligence'
WHERE NOT EXISTS (SELECT 1 FROM org_settings)
`).Error
		_ = db.Exec(`
INSERT INTO roles (name, description, is_system) VALUES
('Admin', 'Full administrative access', TRUE),
('Reviewer', 'Review and comment on diff reports', TRUE),
('Viewer', 'Read-only access', TRUE)
ON CONFLICT (name) DO NOTHING
`).Error
	}
	return &Handler{DB: db, Redis: redisClient, UploadDir: uploadDir, ExportDir: exportDir}
}
