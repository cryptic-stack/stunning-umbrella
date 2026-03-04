-- +migrate Down
DROP INDEX IF EXISTS idx_uploaded_files_file_hash;

ALTER TABLE uploaded_files
DROP COLUMN IF EXISTS file_hash;
