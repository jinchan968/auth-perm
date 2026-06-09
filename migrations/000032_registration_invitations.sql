-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS registration_invitations (
    id VARCHAR(36) PRIMARY KEY,
    code_hash VARCHAR(128) NOT NULL,
    code_preview VARCHAR(32) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL DEFAULT 'default',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    used_by_account_id VARCHAR(36),
    created_by_account_id VARCHAR(36) NOT NULL,
    invalidated_at TIMESTAMPTZ,
    invalidated_by_account_id VARCHAR(36),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_registration_invitations_code_hash ON registration_invitations(code_hash);
CREATE INDEX IF NOT EXISTS idx_registration_invitations_tenant_status ON registration_invitations(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_registration_invitations_expires_at ON registration_invitations(expires_at);
CREATE INDEX IF NOT EXISTS idx_registration_invitations_created_by ON registration_invitations(created_by_account_id);
CREATE INDEX IF NOT EXISTS idx_registration_invitations_used_by ON registration_invitations(used_by_account_id);

COMMENT ON TABLE registration_invitations IS '注册邀请码表';
COMMENT ON COLUMN registration_invitations.code_hash IS '邀请码哈希，禁止存储明文';
COMMENT ON COLUMN registration_invitations.code_preview IS '邀请码短预览，用于列表显示';
COMMENT ON COLUMN registration_invitations.status IS '状态：active, used, invalidated';

-- 邀请码管理权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000038',
    'default',
    'menu:invitations',
    '邀请管理菜单',
    '访问邀请码管理菜单',
    'menu',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000039',
    'default',
    'invitations.manage',
    '管理邀请码',
    '生成、查看、失效注册邀请码',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000048',
    'a0000001-0000-0000-0000-000000000038',
    'invitations',
    'menu',
    '邀请管理菜单',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000049',
    'a0000001-0000-0000-0000-000000000039',
    'GET /api/v1/auth/invitations',
    'api_path',
    '邀请码列表',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000050',
    'a0000001-0000-0000-0000-000000000039',
    'POST /api/v1/auth/invitations',
    'api_path',
    '生成邀请码',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000051',
    'a0000001-0000-0000-0000-000000000039',
    'POST /api/v1/auth/invitations/:id/invalidate',
    'api_path',
    '失效邀请码',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000052',
    'a0000001-0000-0000-0000-000000000039',
    'invitations.list',
    'button',
    '邀请码入口按钮',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000053',
    'a0000001-0000-0000-0000-000000000039',
    'invitations.create',
    'button',
    '生成邀请码按钮',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000054',
    'a0000001-0000-0000-0000-000000000039',
    'invitations.invalidate',
    'button',
    '失效邀请码按钮',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000038',
    'a0000001-0000-0000-0000-000000000039'
);
DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000038',
    'a0000001-0000-0000-0000-000000000039'
);
DROP TABLE IF EXISTS registration_invitations;

-- +goose StatementEnd
