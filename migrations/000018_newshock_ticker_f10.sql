-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS newshock_ticker_f10 (
    id VARCHAR(36) PRIMARY KEY,
    ticker_id VARCHAR(36) NOT NULL,
    pe_ttm DOUBLE PRECISION NOT NULL DEFAULT 0,
    pe_static DOUBLE PRECISION NOT NULL DEFAULT 0,
    pb DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_mcap DOUBLE PRECISION NOT NULL DEFAULT 0,
    float_mcap DOUBLE PRECISION NOT NULL DEFAULT 0,
    turnover_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    volume_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,
    limit_up DOUBLE PRECISION NOT NULL DEFAULT 0,
    limit_down DOUBLE PRECISION NOT NULL DEFAULT 0,
    industry VARCHAR(64) DEFAULT '',
    total_shares DOUBLE PRECISION NOT NULL DEFAULT 0,
    float_shares DOUBLE PRECISION NOT NULL DEFAULT 0,
    eps DOUBLE PRECISION NOT NULL DEFAULT 0,
    bvps DOUBLE PRECISION NOT NULL DEFAULT 0,
    roe DOUBLE PRECISION NOT NULL DEFAULT 0,
    source VARCHAR(32) DEFAULT '',
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ticker_f10_unique ON newshock_ticker_f10(ticker_id);
CREATE INDEX IF NOT EXISTS idx_ticker_f10_industry ON newshock_ticker_f10(industry);
CREATE INDEX IF NOT EXISTS idx_ticker_f10_tenant ON newshock_ticker_f10(tenant_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS newshock_ticker_f10;
-- +goose StatementEnd
