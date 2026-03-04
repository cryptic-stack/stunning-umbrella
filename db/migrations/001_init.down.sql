-- +migrate Down
DROP TABLE IF EXISTS diff_items;
DROP TABLE IF EXISTS diff_reports;
DROP TABLE IF EXISTS safeguards;
DROP TABLE IF EXISTS controls;
DROP TABLE IF EXISTS versions;
DROP TABLE IF EXISTS frameworks;
DROP TABLE IF EXISTS uploaded_files;
