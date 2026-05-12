-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS newshock_ticker_daily (
    id VARCHAR(36) PRIMARY KEY,
    ticker_id VARCHAR(36) NOT NULL,
    trade_date DATE NOT NULL,
    open DOUBLE PRECISION NOT NULL DEFAULT 0,
    high DOUBLE PRECISION NOT NULL DEFAULT 0,
    low DOUBLE PRECISION NOT NULL DEFAULT 0,
    close DOUBLE PRECISION NOT NULL DEFAULT 0,
    volume BIGINT NOT NULL DEFAULT 0,
    amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    change_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    turnover DOUBLE PRECISION NOT NULL DEFAULT 0,
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ticker_daily_unique ON newshock_ticker_daily(ticker_id, trade_date);
CREATE INDEX IF NOT EXISTS idx_ticker_daily_date ON newshock_ticker_daily(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_ticker_daily_tenant ON newshock_ticker_daily(tenant_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS newshock_ticker_daily;
-- +goose StatementEnd
