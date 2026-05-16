-- +goose Up
-- +goose StatementBegin
ALTER TABLE newshock_tickers ADD COLUMN IF NOT EXISTS security_type VARCHAR(20) NOT NULL DEFAULT 'stock';
CREATE INDEX IF NOT EXISTS idx_newshock_tickers_security_type ON newshock_tickers(security_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE newshock_tickers DROP COLUMN IF EXISTS security_type;
-- +goose StatementEnd
