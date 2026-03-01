-- +goose Up
-- +goose StatementBegin
-- =====================================================
-- 初始化默认租户
-- =====================================================
--
-- 插入 tenant_id = 'default' 的默认租户，供开发和演示使用。
-- 使用 ON CONFLICT DO NOTHING 保证脚本幂等，可重复执行。
--
-- =====================================================

INSERT INTO tenants (id, name, code, status, plan, settings, created_at, updated_at)
VALUES (
    'default',
    'admin',
    'default',
    'active',
    'enterprise',
    '{}',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM tenants WHERE id = 'default';

-- +goose StatementEnd
