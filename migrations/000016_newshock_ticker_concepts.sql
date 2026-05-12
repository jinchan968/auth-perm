-- +goose Up
-- +goose StatementBegin
-- A股个股题材概念/行业板块/地域板块

CREATE TABLE IF NOT EXISTS newshock_ticker_concepts (
    id          VARCHAR(36) PRIMARY KEY,
    ticker_id   VARCHAR(36) NOT NULL,
    name        VARCHAR(255) NOT NULL,           -- 板块名称，如"人工智能"、"锂电池"
    type        VARCHAR(20) NOT NULL DEFAULT 'concept', -- concept/industry/region
    source_code VARCHAR(50) DEFAULT '',           -- 东方财富板块代码，如 BK0800
    tenant_id   VARCHAR(36) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_newshock_tc_ticker_name ON newshock_ticker_concepts(ticker_id, name, tenant_id);
CREATE INDEX idx_newshock_tc_type ON newshock_ticker_concepts(type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS newshock_ticker_concepts;
-- +goose StatementEnd
