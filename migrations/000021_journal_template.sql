-- +goose Up
-- +goose StatementBegin

-- =====================================================
-- 札记模板模块：表结构 + 权限种子数据
-- =====================================================

-- 1. 札记模板表
CREATE TABLE IF NOT EXISTS journal_templates (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    content TEXT,
    tags JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 索引优化
CREATE INDEX IF NOT EXISTS idx_journal_templates_tenant ON journal_templates(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_journal_templates_tenant_name ON journal_templates(tenant_id, name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_journal_templates_account ON journal_templates(account_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_journal_templates_created ON journal_templates(tenant_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_journal_templates_tags ON journal_templates USING GIN(tags) WHERE deleted_at IS NULL;

COMMENT ON TABLE  journal_templates               IS '札记模板表';
COMMENT ON COLUMN journal_templates.id            IS '模板UUID';
COMMENT ON COLUMN journal_templates.tenant_id      IS '租户ID';
COMMENT ON COLUMN journal_templates.account_id     IS '创建者账户ID';
COMMENT ON COLUMN journal_templates.name           IS '模板名称';
COMMENT ON COLUMN journal_templates.content        IS '模板内容';
COMMENT ON COLUMN journal_templates.tags           IS '标签ID数组';

-- =====================================================
-- 权限种子数据
-- =====================================================

-- 13. 札记模板读取权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000017',
    'default',
    'journal.template.read',
    '札记模板读取',
    '查看札记模板的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 14. 札记模板写入权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000018',
    'default',
    'journal.template.write',
    '札记模板写入',
    '创建和编辑札记模板的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 15. 札记模板删除权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000019',
    'default',
    'journal.template.delete',
    '札记模板删除',
    '删除札记模板的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- =====================================================
-- 权限资源关联
-- =====================================================

-- 模板读取 - API 资源（GET）
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

-- 模板写入 - API 资源（POST）
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

-- 模板写入 - API 资源（PUT）
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

-- 模板删除 - API 资源（DELETE）
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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 删除权限资源关联
DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000017',
    'a0000001-0000-0000-0000-000000000018',
    'a0000001-0000-0000-0000-000000000019'
);

-- 删除权限
DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000017',
    'a0000001-0000-0000-0000-000000000018',
    'a0000001-0000-0000-0000-000000000019'
);

-- 删除表
DROP TABLE IF EXISTS journal_templates;

-- +goose StatementEnd