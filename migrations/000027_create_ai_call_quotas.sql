-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS ai_call_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    model_id VARCHAR(64) NOT NULL,
    call_date DATE NOT NULL,
    call_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uk_quota_tenant_account_model_date UNIQUE (tenant_id, account_id, model_id, call_date)
);

CREATE INDEX IF NOT EXISTS idx_ai_call_quotas_lookup ON ai_call_quotas(tenant_id, account_id, call_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS ai_call_quotas;

-- +goose StatementEnd