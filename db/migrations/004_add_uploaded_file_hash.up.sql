-- +migrate Up
ALTER TABLE uploaded_files
ADD COLUMN IF NOT EXISTS file_hash TEXT DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_uploaded_files_file_hash
ON uploaded_files(file_hash);
