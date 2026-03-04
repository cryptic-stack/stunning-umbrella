-- +migrate Down
ALTER TABLE safeguards DROP COLUMN IF EXISTS level;
