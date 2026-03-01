-- +goose Up
-- +goose StatementBegin
-- =====================================================
-- 系统权限与资源种子数据
-- =====================================================
--
-- 为 tenant_id='default' 插入系统菜单权限、按钮权限及其关联的 API 资源。
-- 使用 ON CONFLICT DO NOTHING 保证幂等性，可重复执行。
--
-- 注意：
-- - permissions.id 和 tenant_id 均为 VARCHAR(36)，使用固定 UUID 格式
-- - permissions 有 resource 列（NOT NULL），需提供值
-- - permission_resources 已有唯一约束 unique_permission_resource(permission_id, resource_id, resource_type, tenant_id)
--
-- =====================================================

-- =====================================================
-- 插入权限（Permissions）
-- =====================================================

-- 1. 租户管理菜单权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000001',
    'default',
    'menu:tenants',
    '租户管理',
    '访问租户管理菜单的权限',
    'menu',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 2. 权限管理菜单权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000002',
    'default',
    'menu:permissions',
    '权限管理',
    '访问权限管理菜单的权限',
    'menu',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 3. 待办事项菜单权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000003',
    'default',
    'menu:todos',
    '待办事项',
    '访问待办事项菜单的权限',
    'menu',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 4. 权限列表Tab权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000004',
    'default',
    'tab:perm.list',
    '权限列表Tab',
    '访问权限管理-权限列表Tab的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 5. 角色列表Tab权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000005',
    'default',
    'tab:perm.roles',
    '角色列表Tab',
    '访问权限管理-角色列表Tab的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 6. 用户列表Tab权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000006',
    'default',
    'tab:perm.users',
    '用户列表Tab',
    '访问权限管理-用户列表Tab的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 7. 显示全部租户按钮权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000007',
    'default',
    'button:tenants.show_all',
    '显示全部租户',
    '显示全部租户（含已删除）的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- =====================================================
-- 插入权限资源关联（Permission Resources）
-- =====================================================
-- 表已有唯一约束: unique_permission_resource(permission_id, resource_id, resource_type, tenant_id)
-- 使用 ON CONFLICT ON CONSTRAINT ... DO NOTHING 实现幂等

-- ──────────────────────────────────────────────────────
-- 1. 租户管理菜单 (menu:tenants) 的资源
-- ──────────────────────────────────────────────────────

-- 1.1 菜单资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000001',
    'a0000001-0000-0000-0000-000000000001',
    'tenants',
    'menu',
    '租户管理菜单',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 1.2 API 资源（租户相关所有接口）
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000002',
    'a0000001-0000-0000-0000-000000000001',
    '/api/v1/tenants/*',
    'api_path',
    '租户管理API',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- ──────────────────────────────────────────────────────
-- 2. 权限管理菜单 (menu:permissions) 的资源
-- ──────────────────────────────────────────────────────

-- 2.1 菜单资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000003',
    'a0000001-0000-0000-0000-000000000002',
    'permissions',
    'menu',
    '权限管理菜单',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 2.2 API 资源（权限相关所有接口）
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000004',
    'a0000001-0000-0000-0000-000000000002',
    '/api/v1/permissions/*',
    'api_path',
    '权限管理API',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- ──────────────────────────────────────────────────────
-- 3. 待办事项菜单 (menu:todos) 的资源
-- ──────────────────────────────────────────────────────

-- 3.1 菜单资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000005',
    'a0000001-0000-0000-0000-000000000003',
    'todos',
    'menu',
    '待办事项菜单',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 3.2 API 资源（待办相关所有接口）
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000006',
    'a0000001-0000-0000-0000-000000000003',
    '/api/v1/todos/*',
    'api_path',
    '待办事项API',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- ──────────────────────────────────────────────────────
-- 4. 权限列表Tab (tab:perm.list) 的资源
-- ──────────────────────────────────────────────────────

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000007',
    'a0000001-0000-0000-0000-000000000004',
    'perm.tab.list',
    'button',
    '权限列表Tab按钮',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- ──────────────────────────────────────────────────────
-- 5. 角色列表Tab (tab:perm.roles) 的资源
-- ──────────────────────────────────────────────────────

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000008',
    'a0000001-0000-0000-0000-000000000005',
    'perm.tab.roles',
    'button',
    '角色列表Tab按钮',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- ──────────────────────────────────────────────────────
-- 6. 用户列表Tab (tab:perm.users) 的资源
-- ──────────────────────────────────────────────────────

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000009',
    'a0000001-0000-0000-0000-000000000006',
    'perm.tab.users',
    'button',
    '用户列表Tab按钮',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- ──────────────────────────────────────────────────────
-- 7. 显示全部租户按钮 (button:tenants.show_all) 的资源
-- ──────────────────────────────────────────────────────

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000010',
    'a0000001-0000-0000-0000-000000000007',
    'tenants.show_all',
    'button',
    '显示全部租户按钮',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 删除权限资源关联
DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000001',
    'a0000001-0000-0000-0000-000000000002',
    'a0000001-0000-0000-0000-000000000003',
    'a0000001-0000-0000-0000-000000000004',
    'a0000001-0000-0000-0000-000000000005',
    'a0000001-0000-0000-0000-000000000006',
    'a0000001-0000-0000-0000-000000000007'
);

-- 删除权限
DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000001',
    'a0000001-0000-0000-0000-000000000002',
    'a0000001-0000-0000-0000-000000000003',
    'a0000001-0000-0000-0000-000000000004',
    'a0000001-0000-0000-0000-000000000005',
    'a0000001-0000-0000-0000-000000000006',
    'a0000001-0000-0000-0000-000000000007'
);

-- +goose StatementEnd
