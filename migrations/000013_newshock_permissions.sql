-- +goose Up
-- Newshock 权限种子数据

-- 1. 新知菜单权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000020',
    'default',
    'menu:newshock',
    '新知',
    '访问新知菜单的权限',
    'menu',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 2. 新知读取权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000021',
    'default',
    'newshock.read',
    '新知读取',
    '查看新知数据的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 3. 新知写入权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000022',
    'default',
    'newshock.write',
    '新知写入',
    '创建和编辑新知数据的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- =====================================================
-- 权限资源关联
-- =====================================================

-- 4. 新知菜单资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000020',
    'a0000001-0000-0000-0000-000000000020',
    'newshock',
    'menu',
    '新知菜单',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 5. 新知菜单 — GET API 资源（通配）
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000021',
    'a0000001-0000-0000-0000-000000000020',
    'GET /api/v1/newshock',
    'api_path',
    '新知API（通配）',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 6. 新知读取 - GET API 资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000022',
    'a0000001-0000-0000-0000-000000000021',
    'GET /api/v1/newshock',
    'api_path',
    '新知读取API（GET）',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 7. 新知写入 — POST API 资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000023',
    'a0000001-0000-0000-0000-000000000022',
    'POST /api/v1/newshock',
    'api_path',
    '新知写入API（POST）',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 8. 新知写入 — PUT API 资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000024',
    'a0000001-0000-0000-0000-000000000022',
    'PUT /api/v1/newshock',
    'api_path',
    '新知写入API（PUT）',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 9. 新知写入 — DELETE API 资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000025',
    'a0000001-0000-0000-0000-000000000022',
    'DELETE /api/v1/newshock',
    'api_path',
    '新知写入API（DELETE）',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- +goose Down

-- 删除权限资源关联
DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000020',
    'a0000001-0000-0000-0000-000000000021',
    'a0000001-0000-0000-0000-000000000022'
);

-- 删除权限
DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000020',
    'a0000001-0000-0000-0000-000000000021',
    'a0000001-0000-0000-0000-000000000022'
);