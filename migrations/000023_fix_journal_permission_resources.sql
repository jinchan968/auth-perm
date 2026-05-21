-- +goose Up
-- +goose StatementBegin

-- =====================================================
-- 修复：000021/000022 的 permission_resources UUID 与
-- 000013_newshock_permissions 冲突，导致 ON CONFLICT DO NOTHING
-- 静默跳过插入。这里用不冲突的 UUID 重新插入。
-- =====================================================

-- 1. 模板读取 API（原 b0000001-...-000000000020 → 000000000017）
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000017',
    'a0000001-0000-0000-0000-000000000017',
    'GET /api/v1/journal/templates',
    'api_path',
    '模板读取API',
    'default',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- 2. 模板创建 API（原 b0000001-...-000000000021 → 000000000018）
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000018',
    'a0000001-0000-0000-0000-000000000018',
    'POST /api/v1/journal/templates',
    'api_path',
    '模板创建API',
    'default',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- 3. 模板更新 API（原 b0000001-...-000000000022 → 000000000019）
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000019',
    'a0000001-0000-0000-0000-000000000018',
    'PUT /api/v1/journal/templates',
    'api_path',
    '模板更新API',
    'default',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- 4. 模板删除 API（原 b0000001-...-000000000023 → 000000000026）
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000026',
    'a0000001-0000-0000-0000-000000000019',
    'DELETE /api/v1/journal/templates',
    'api_path',
    '模板删除API',
    'default',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- 5. 札记条目列表 Tab 按钮（原 b0000001-...-000000000024 → 000000000027）
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

-- 6. 模板管理 Tab 按钮（原 b0000001-...-000000000025 → 000000000028）
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

DELETE FROM permission_resources WHERE id IN (
    'b0000001-0000-0000-0000-000000000017',
    'b0000001-0000-0000-0000-000000000018',
    'b0000001-0000-0000-0000-000000000019',
    'b0000001-0000-0000-0000-000000000026',
    'b0000001-0000-0000-0000-000000000027',
    'b0000001-0000-0000-0000-000000000028'
);

-- +goose StatementEnd
