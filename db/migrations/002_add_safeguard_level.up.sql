-- +migrate Up
ALTER TABLE safeguards ADD COLUMN IF NOT EXISTS level TEXT DEFAULT '';
