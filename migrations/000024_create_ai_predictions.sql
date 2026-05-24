-- +goose Up
-- +goose StatementBegin

-- =====================================================
-- AI 预测历史记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS ai_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    question TEXT NOT NULL,
    system_prompt TEXT,
    results JSONB NOT NULL DEFAULT '[]',
    model_snapshot JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_predictions_tenant_account ON ai_predictions(tenant_id, account_id);
CREATE INDEX IF NOT EXISTS idx_ai_predictions_created_at ON ai_predictions(created_at DESC);

-- =====================================================
-- AI 预测 按钮权限资源
-- =====================================================
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000029',
    'a0000001-0000-0000-0000-000000000009',
    'journal.tab.ai_predictions',
    'button',
    '多AI预测按钮',
    'default',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS ai_predictions;

DELETE FROM permission_resources WHERE resource_id = 'journal.tab.ai_predictions';

-- +goose StatementEnd
