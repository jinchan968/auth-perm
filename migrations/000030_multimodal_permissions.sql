-- +goose Up
-- +goose StatementBegin

-- 菜单权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000034',
    'default',
    'menu:multimodal',
    '多模态菜单',
    '访问多模态功能菜单',
    'menu',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 读取权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000035',
    'default',
    'multimodal.read',
    '使用多模态功能',
    '识图、生成提示词等多模态功能',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 资源绑定
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000038',
    'a0000001-0000-0000-0000-000000000034',
    'multimodal',
    'menu',
    '多模态菜单',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000039',
    'a0000001-0000-0000-0000-000000000034',
    '/api/v1/multimodal/*',
    'api_path',
    '多模态 API 通配',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000040',
    'a0000001-0000-0000-0000-000000000035',
    'POST /api/v1/multimodal/recognize',
    'api_path',
    '识图接口',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000041',
    'a0000001-0000-0000-0000-000000000035',
    'POST /api/v1/multimodal/generate',
    'api_path',
    '生成提示词接口',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000042',
    'a0000001-0000-0000-0000-000000000035',
    'multimodal.tab.recognize',
    'button',
    '识图 Tab',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000043',
    'a0000001-0000-0000-0000-000000000035',
    'multimodal.tab.generate',
    'button',
    '生成提示词 Tab',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000034',
    'a0000001-0000-0000-0000-000000000035'
);
DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000034',
    'a0000001-0000-0000-0000-000000000035'
);

-- +goose StatementEnd
