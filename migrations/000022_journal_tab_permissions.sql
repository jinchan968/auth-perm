-- +goose Up
-- +goose StatementBegin

-- =====================================================
-- 札记 Tab 按钮权限资源
-- =====================================================

-- 1. 札记条目列表 Tab 按钮
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000027',
    'a0000001-0000-0000-0000-000000000009',
    'journal.tab.entries',
    'button',
    '札记条目列表Tab按钮',
    'default',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- 2. 模板管理 Tab 按钮
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000028',
    'a0000001-0000-0000-0000-000000000017',
    'journal.tab.templates',
    'button',
    '模板管理Tab按钮',
    'default',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000009',
    'a0000001-0000-0000-0000-000000000017'
) AND resource_type = 'button' AND resource_id IN ('journal.tab.entries', 'journal.tab.templates');

-- +goose StatementEnd
