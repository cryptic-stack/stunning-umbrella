-- +migrate Up
ALTER TABLE diff_items
ADD COLUMN IF NOT EXISTS reviewed BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE diff_items
ADD COLUMN IF NOT EXISTS review_comment TEXT DEFAULT '';

ALTER TABLE diff_items
ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_diff_items_report_reviewed
ON diff_items(report_id, reviewed);
