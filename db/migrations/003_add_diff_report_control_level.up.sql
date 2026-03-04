-- +migrate Up
ALTER TABLE diff_reports
ADD COLUMN IF NOT EXISTS control_level TEXT NOT NULL DEFAULT 'ALL';
