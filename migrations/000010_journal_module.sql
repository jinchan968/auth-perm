-- +goose Up
-- +goose StatementBegin

-- =====================================================
-- 札记模块：表结构 + 权限种子数据
-- =====================================================

-- 1. 札记条目表
CREATE TABLE IF NOT EXISTS journal_entries (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    parent_id VARCHAR(36),          -- NULL=主条目，非NULL=修正条目
    title TEXT,
    content TEXT NOT NULL,
    weather VARCHAR(20),
    location VARCHAR(200),
    period VARCHAR(10) NOT NULL,
    entry_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_journal_entries_parent FOREIGN KEY (parent_id) REFERENCES journal_entries(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_journal_entries_tenant_account
    ON journal_entries(tenant_id, account_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_journal_entries_parent
    ON journal_entries(parent_id)
    WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_journal_entries_date
    ON journal_entries(entry_date);
CREATE INDEX IF NOT EXISTS idx_journal_entries_deleted_at
    ON journal_entries(deleted_at);

COMMENT ON TABLE  journal_entries                IS '札记条目表（含修正条目）';
COMMENT ON COLUMN journal_entries.id             IS '条目UUID';
COMMENT ON COLUMN journal_entries.tenant_id      IS '租户ID';
COMMENT ON COLUMN journal_entries.account_id     IS '所属账户ID';
COMMENT ON COLUMN journal_entries.parent_id      IS '父条目ID，NULL表示主条目，非NULL表示修正条目';
COMMENT ON COLUMN journal_entries.title          IS '标题（可选）';
COMMENT ON COLUMN journal_entries.content        IS '正文内容，最多800字';
COMMENT ON COLUMN journal_entries.weather        IS '天气：晴/多云/雨/雪/雾/风';
COMMENT ON COLUMN journal_entries.location       IS '位置描述（可选）';
COMMENT ON COLUMN journal_entries.period         IS '时段：晨/上午/下午/晚/夜';
COMMENT ON COLUMN journal_entries.entry_date     IS '条目日期';

-- 2. 标签表
CREATE TABLE IF NOT EXISTS journal_tags (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    name VARCHAR(30) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_journal_tags_tenant_account
    ON journal_tags(tenant_id, account_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_journal_tags_deleted_at
    ON journal_tags(deleted_at);

-- 标签名唯一（未删除范围内）
CREATE UNIQUE INDEX IF NOT EXISTS uq_journal_tags_name_active
    ON journal_tags(tenant_id, account_id, name)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE  journal_tags                IS '札记标签表';
COMMENT ON COLUMN journal_tags.name            IS '标签名称';
COMMENT ON COLUMN journal_tags.color           IS '标签颜色，HEX格式';

-- 3. 札记-标签关联表
CREATE TABLE IF NOT EXISTS diary_tags (
    diary_id VARCHAR(36) NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    tag_id VARCHAR(36) NOT NULL REFERENCES journal_tags(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_diary_tags PRIMARY KEY (diary_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_diary_tags_tag_id ON diary_tags(tag_id);

COMMENT ON TABLE  diary_tags              IS '札记-标签关联表';

-- =====================================================
-- 权限种子数据
-- =====================================================

-- 4. 札记菜单权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000008',
    'default',
    'menu:journal',
    '札记',
    '访问札记菜单的权限',
    'menu',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 5. 札记读取权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000009',
    'default',
    'journal.read',
    '札记读取',
    '查看札记条目的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 6. 札记写入权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000010',
    'default',
    'journal.write',
    '札记写入',
    '创建和编辑札记的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 7. 札记删除权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000011',
    'default',
    'journal.delete',
    '札记删除',
    '删除札记条目的权限',
    'button',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- =====================================================
-- 权限资源关联
-- =====================================================

-- 8. 札记菜单资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000011',
    'a0000001-0000-0000-0000-000000000008',
    'journal',
    'menu',
    '札记菜单',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 9. 札记 API 资源 (通配)
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000012',
    'a0000001-0000-0000-0000-000000000008',
    '/api/v1/journal/*',
    'api_path',
    '札记API',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 10. 札记读取 - API 资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000013',
    'a0000001-0000-0000-0000-000000000009',
    '/api/v1/journal',
    'api_path',
    '札记读取API（GET）',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 11. 札记写入 - API 资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000014',
    'a0000001-0000-0000-0000-000000000010',
    '/api/v1/journal',
    'api_path',
    '札记写入API（POST/PUT）',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- 12. 札记删除 - API 资源
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000015',
    'a0000001-0000-0000-0000-000000000011',
    '/api/v1/journal',
    'api_path',
    '札记删除API（DELETE）',
    'default',
    NOW(),
    NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 删除权限资源关联
DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000008',
    'a0000001-0000-0000-0000-000000000009',
    'a0000001-0000-0000-0000-000000000010',
    'a0000001-0000-0000-0000-000000000011'
);

-- 删除权限
DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000008',
    'a0000001-0000-0000-0000-000000000009',
    'a0000001-0000-0000-0000-000000000010',
    'a0000001-0000-0000-0000-000000000011'
);

-- 删除关联表
DROP TABLE IF EXISTS diary_tags CASCADE;
-- 删除标签表
DROP TABLE IF EXISTS journal_tags CASCADE;
-- 删除札记条目表
DROP TABLE IF EXISTS journal_entries CASCADE;

-- +goose StatementEnd