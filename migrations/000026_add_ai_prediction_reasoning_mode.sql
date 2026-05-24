-- +goose Up
ALTER TABLE ai_predictions ADD COLUMN IF NOT EXISTS reasoning_mode VARCHAR(16) NOT NULL DEFAULT 'low';

-- +goose Down
ALTER TABLE ai_predictions DROP COLUMN IF EXISTS reasoning_mode;