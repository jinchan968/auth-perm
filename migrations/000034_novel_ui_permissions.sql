-- +goose Up
-- +goose StatementBegin

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES
    ('b0000001-0000-0000-0000-000000000055', 'a0000001-0000-0000-0000-000000000040', 'novel', 'menu', '小说管理菜单', 'default', NOW(), NOW()),
    ('b0000001-0000-0000-0000-000000000056', 'a0000001-0000-0000-0000-000000000041', 'novel.import', 'button', '小说导入按钮', 'default', NOW(), NOW()),
    ('b0000001-0000-0000-0000-000000000057', 'a0000001-0000-0000-0000-000000000040', 'POST /api/v1/novel-admin/import-md-bundle/inspect', 'api_path', '小说 Markdown 包识别 API', 'default', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM permission_resources WHERE id IN (
    'b0000001-0000-0000-0000-000000000055',
    'b0000001-0000-0000-0000-000000000056',
    'b0000001-0000-0000-0000-000000000057'
);

-- +goose StatementEnd
