-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS newshock_ticker_news (
    id VARCHAR(36) PRIMARY KEY,
    ticker_id VARCHAR(36) NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT DEFAULT '',
    source VARCHAR(128) DEFAULT '',
    publish_time TIMESTAMP NOT NULL DEFAULT NOW(),
    url VARCHAR(500) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ticker_news_unique ON newshock_ticker_news(ticker_id, url);
CREATE INDEX IF NOT EXISTS idx_ticker_news_time ON newshock_ticker_news(ticker_id, publish_time DESC);
CREATE INDEX IF NOT EXISTS idx_ticker_news_tenant ON newshock_ticker_news(tenant_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS newshock_ticker_news;
-- +goose StatementEnd
