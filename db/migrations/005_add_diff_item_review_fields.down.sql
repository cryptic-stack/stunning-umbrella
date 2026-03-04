-- +migrate Down
DROP INDEX IF EXISTS idx_diff_items_report_reviewed;

ALTER TABLE diff_items
DROP COLUMN IF EXISTS reviewed_at;

ALTER TABLE diff_items
DROP COLUMN IF EXISTS review_comment;

ALTER TABLE diff_items
DROP COLUMN IF EXISTS reviewed;
