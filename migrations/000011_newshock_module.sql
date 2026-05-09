-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- 投资主题表
CREATE TABLE IF NOT EXISTS newshock_themes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    category VARCHAR(50) NOT NULL DEFAULT 'exploratory',
    strength DOUBLE PRECISION DEFAULT 0,
    strength_norm DOUBLE PRECISION DEFAULT 0,
    classification_confidence DOUBLE PRECISION DEFAULT 0.85,
    ticker_count INTEGER DEFAULT 0,
    event_count INTEGER DEFAULT 0,
    trend VARCHAR(20) NOT NULL DEFAULT 'stable',
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_newshock_themes_category ON newshock_themes(category);
CREATE INDEX idx_newshock_themes_strength ON newshock_themes(strength DESC);
CREATE INDEX idx_newshock_themes_trend ON newshock_themes(trend);
CREATE INDEX idx_newshock_themes_tenant ON newshock_themes(tenant_id);
CREATE INDEX idx_newshock_themes_deleted ON newshock_themes(deleted_at);

-- 股票标的表
CREATE TABLE IF NOT EXISTS newshock_tickers (
    id VARCHAR(36) PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    name VARCHAR(255) DEFAULT '',
    market VARCHAR(10) NOT NULL DEFAULT 'us',
    hot_score DOUBLE PRECISION DEFAULT 0,
    mention_count INTEGER DEFAULT 0,
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_newshock_tickers_symbol ON newshock_tickers(symbol, tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_newshock_tickers_market ON newshock_tickers(market);
CREATE INDEX idx_newshock_tickers_hot ON newshock_tickers(hot_score DESC);
CREATE INDEX idx_newshock_tickers_tenant ON newshock_tickers(tenant_id);

-- 事件表
CREATE TABLE IF NOT EXISTS newshock_events (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    summary TEXT DEFAULT '',
    channel VARCHAR(50) DEFAULT '',
    importance INTEGER NOT NULL DEFAULT 3,
    theme_id VARCHAR(36) DEFAULT '',
    theme_name VARCHAR(255) DEFAULT '',
    event_time TIMESTAMP,
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_newshock_events_theme ON newshock_events(theme_id);
CREATE INDEX idx_newshock_events_importance ON newshock_events(importance DESC);
CREATE INDEX idx_newshock_events_time ON newshock_events(event_time DESC);
CREATE INDEX idx_newshock_events_tenant ON newshock_events(tenant_id);
CREATE INDEX idx_newshock_events_deleted ON newshock_events(deleted_at);

-- 事件-股票关联表
CREATE TABLE IF NOT EXISTS newshock_event_tickers (
    event_id VARCHAR(36) NOT NULL,
    ticker_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (event_id, ticker_id)
);

CREATE INDEX idx_newshock_et_ticker ON newshock_event_tickers(ticker_id);

-- 主题-股票关联表
CREATE TABLE IF NOT EXISTS newshock_theme_tickers (
    theme_id VARCHAR(36) NOT NULL,
    ticker_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (theme_id, ticker_id)
);

CREATE INDEX idx_newshock_tt_ticker ON newshock_theme_tickers(ticker_id);

-- 市场环境表
CREATE TABLE IF NOT EXISTS newshock_regime (
    id VARCHAR(36) PRIMARY KEY,
    regime_type VARCHAR(20) NOT NULL DEFAULT 'neutral',
    confidence DOUBLE PRECISION DEFAULT 0.5,
    summary TEXT DEFAULT '',
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_newshock_regime_tenant ON newshock_regime(tenant_id);

-- 原始新闻表 (数据管线用)
CREATE TABLE IF NOT EXISTS newshock_news_raw (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    content TEXT DEFAULT '',
    source VARCHAR(100) DEFAULT '',
    channel VARCHAR(50) DEFAULT '',
    url VARCHAR(1000) DEFAULT '',
    published_at TIMESTAMP,
    content_hash VARCHAR(64) DEFAULT '',
    processed BOOLEAN DEFAULT FALSE,
    tenant_id VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_newshock_news_processed ON newshock_news_raw(processed);
CREATE INDEX idx_newshock_news_hash ON newshock_news_raw(content_hash);
CREATE INDEX idx_newshock_news_tenant ON newshock_news_raw(tenant_id);
CREATE INDEX idx_newshock_news_published ON newshock_news_raw(published_at DESC);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS newshock_news_raw;
DROP TABLE IF EXISTS newshock_regime;
DROP TABLE IF EXISTS newshock_theme_tickers;
DROP TABLE IF EXISTS newshock_event_tickers;
DROP TABLE IF EXISTS newshock_events;
DROP TABLE IF EXISTS newshock_tickers;
DROP TABLE IF EXISTS newshock_themes;
