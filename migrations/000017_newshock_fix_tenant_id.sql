-- +goose Up
-- +goose StatementBegin
-- 统一 newshock 模块的 tenant_id 为 'default'
-- 旧值 '00000000-0000-0000-0000-000000000001' → 新值 'default'

UPDATE newshock_themes       SET tenant_id = 'default' WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
UPDATE newshock_tickers      SET tenant_id = 'default' WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
UPDATE newshock_events       SET tenant_id = 'default' WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
UPDATE newshock_regime       SET tenant_id = 'default' WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
UPDATE newshock_news_raw     SET tenant_id = 'default' WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
UPDATE newshock_polymarket   SET tenant_id = 'default' WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
UPDATE newshock_ticker_daily SET tenant_id = 'default' WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
UPDATE newshock_ticker_concepts SET tenant_id = 'default' WHERE tenant_id = '00000000-0000-0000-0000-000000000001';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE newshock_themes       SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id = 'default';
UPDATE newshock_tickers      SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id = 'default';
UPDATE newshock_events       SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id = 'default';
UPDATE newshock_regime       SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id = 'default';
UPDATE newshock_news_raw     SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id = 'default';
UPDATE newshock_polymarket   SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id = 'default';
UPDATE newshock_ticker_daily SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id = 'default';
UPDATE newshock_ticker_concepts SET tenant_id = '00000000-0000-0000-0000-000000000001' WHERE tenant_id = 'default';
-- +goose StatementEnd
