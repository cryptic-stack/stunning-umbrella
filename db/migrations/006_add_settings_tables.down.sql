-- +migrate Down
DROP INDEX IF EXISTS idx_app_users_role_id;
DROP TABLE IF EXISTS app_users;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS org_settings;
