-- +goose Up
-- +goose StatementBegin

-- 菜单权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000030',
    'default',
    'menu:workflow',
    '工作流菜单',
    '访问工作流编排菜单',
    'menu',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 读取权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000031',
    'default',
    'workflow.read',
    '查看工作流',
    '查看工作流列表和详情',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 写入权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000032',
    'default',
    'workflow.write',
    '创建/编辑工作流',
    '创建、编辑、执行、克隆工作流',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 删除权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000033',
    'default',
    'workflow.delete',
    '删除工作流',
    '删除工作流',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 资源绑定
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000030',
    'a0000001-0000-0000-0000-000000000030',
    'workflow',
    'menu',
    '工作流菜单',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000031',
    'a0000001-0000-0000-0000-000000000030',
    '/api/v1/workflow/*',
    'api_path',
    '工作流 API 通配',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000032',
    'a0000001-0000-0000-0000-000000000031',
    'GET /api/v1/workflow',
    'api_path',
    '查看工作流列表',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000033',
    'a0000001-0000-0000-0000-000000000031',
    'workflow.tab.designer',
    'button',
    '编排设计 Tab',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000034',
    'a0000001-0000-0000-0000-000000000031',
    'workflow.tab.runs',
    'button',
    '运行历史 Tab',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000035',
    'a0000001-0000-0000-0000-000000000032',
    'POST /api/v1/workflow',
    'api_path',
    '创建工作流',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000036',
    'a0000001-0000-0000-0000-000000000032',
    'PUT /api/v1/workflow/*',
    'api_path',
    '更新工作流',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000037',
    'a0000001-0000-0000-0000-000000000033',
    'DELETE /api/v1/workflow/*',
    'api_path',
    '删除工作流',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000030',
    'a0000001-0000-0000-0000-000000000031',
    'a0000001-0000-0000-0000-000000000032',
    'a0000001-0000-0000-0000-000000000033'
);
DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000030',
    'a0000001-0000-0000-0000-000000000031',
    'a0000001-0000-0000-0000-000000000032',
    'a0000001-0000-0000-0000-000000000033'
);

-- +goose StatementEnd
