-- +migrate Down
ALTER TABLE diff_reports
DROP COLUMN IF EXISTS control_level;
