-- +goose Up
-- +goose StatementBegin

-- 将“生成提示词”和“生成图片”拆成独立可分配的按钮权限。
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000036',
    'default',
    'multimodal.generate_prompt',
    '生成提示词',
    '使用多模态生成提示词功能',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000037',
    'default',
    'multimodal.generate_image',
    '生成图片',
    '使用多模态生成图片功能',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 旧的 multimodal.read 只保留识图能力，生成类能力迁移到独立权限。
UPDATE permissions
SET name = '识图',
    description = '使用多模态识图功能',
    updated_at = NOW()
WHERE id = 'a0000001-0000-0000-0000-000000000035';

-- 菜单权限只控制入口可见性，API 访问交给具体功能权限。
DELETE FROM permission_resources
WHERE permission_id = 'a0000001-0000-0000-0000-000000000034'
  AND resource_type = 'api_path'
  AND resource_id = '/api/v1/multimodal/*';

DELETE FROM permission_resources
WHERE permission_id = 'a0000001-0000-0000-0000-000000000035'
  AND resource_type IN ('api_path', 'button')
  AND resource_id IN (
      'POST /api/v1/multimodal/generate',
      'multimodal.tab.generate'
  );

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000044',
    'a0000001-0000-0000-0000-000000000036',
    'POST /api/v1/multimodal/generate',
    'api_path',
    '生成提示词接口',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000045',
    'a0000001-0000-0000-0000-000000000036',
    'multimodal.tab.generate',
    'button',
    '生成提示词 Tab',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000046',
    'a0000001-0000-0000-0000-000000000037',
    'POST /api/v1/multimodal/image-generate',
    'api_path',
    '生成图片接口',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000047',
    'a0000001-0000-0000-0000-000000000037',
    'multimodal.tab.image_generate',
    'button',
    '生成图片 Tab',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM permission_resources
WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000036',
    'a0000001-0000-0000-0000-000000000037'
);

DELETE FROM permissions
WHERE id IN (
    'a0000001-0000-0000-0000-000000000036',
    'a0000001-0000-0000-0000-000000000037'
);

UPDATE permissions
SET name = '使用多模态功能',
    description = '识图、生成提示词等多模态功能',
    updated_at = NOW()
WHERE id = 'a0000001-0000-0000-0000-000000000035';

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
    'b0000001-0000-0000-0000-000000000041',
    'a0000001-0000-0000-0000-000000000035',
    'POST /api/v1/multimodal/generate',
    'api_path',
    '生成提示词接口',
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
