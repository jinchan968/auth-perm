-- +goose Up
-- Polymarket 概率市场表

CREATE TABLE IF NOT EXISTS newshock_polymarket (
    id VARCHAR(36) PRIMARY KEY,
    condition_id VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT DEFAULT '',
    outcome VARCHAR(50) DEFAULT '',
    probability DOUBLE PRECISION DEFAULT 0,
    volume DOUBLE PRECISION DEFAULT 0,
    theme_id VARCHAR(36) DEFAULT '',
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_newshock_pm_theme ON newshock_polymarket(theme_id);
CREATE INDEX idx_newshock_pm_tenant ON newshock_polymarket(tenant_id);
CREATE INDEX idx_newshock_pm_prob ON newshock_polymarket(probability DESC);

-- +goose Down

DROP TABLE IF EXISTS newshock_polymarket;
