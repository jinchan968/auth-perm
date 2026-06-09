-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS novels (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    subtitle VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    cover_url TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'serial',
    tags JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_novels_tenant_account ON novels(tenant_id, account_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novels_status ON novels(tenant_id, status) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS novel_volumes (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    novel_id VARCHAR(36) NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    subtitle VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_novel_volumes_novel ON novel_volumes(novel_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novel_volumes_tenant_account ON novel_volumes(tenant_id, account_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_novel_volumes_title ON novel_volumes(novel_id, title) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS novel_units (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    novel_id VARCHAR(36) NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    volume_id VARCHAR(36) NOT NULL REFERENCES novel_volumes(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    subtitle VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_novel_units_novel ON novel_units(novel_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novel_units_volume ON novel_units(volume_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novel_units_tenant_account ON novel_units(tenant_id, account_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_novel_units_title ON novel_units(volume_id, title) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS novel_chapters (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    novel_id VARCHAR(36) NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    volume_id VARCHAR(36) NOT NULL REFERENCES novel_volumes(id) ON DELETE CASCADE,
    unit_id VARCHAR(36) REFERENCES novel_units(id) ON DELETE SET NULL,
    slug VARCHAR(128) NOT NULL,
    number VARCHAR(32) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    word_count INTEGER NOT NULL DEFAULT 0,
    reading_minutes INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_novel_chapters_slug ON novel_chapters(novel_id, slug) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novel_chapters_novel_status ON novel_chapters(novel_id, status, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novel_chapters_volume ON novel_chapters(volume_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novel_chapters_unit ON novel_chapters(unit_id, sort_order) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS novel_chapter_versions (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    chapter_id VARCHAR(36) NOT NULL REFERENCES novel_chapters(id) ON DELETE CASCADE,
    label VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_novel_chapter_versions_chapter ON novel_chapter_versions(chapter_id, created_at DESC);

CREATE TABLE IF NOT EXISTS novel_codex_entries (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    novel_id VARCHAR(36) NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    kind VARCHAR(32) NOT NULL,
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    aliases JSONB NOT NULL DEFAULT '[]',
    properties JSONB NOT NULL DEFAULT '{}',
    relations JSONB NOT NULL DEFAULT '[]',
    evidence TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_novel_codex_entries_novel_kind ON novel_codex_entries(novel_id, kind, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novel_codex_entries_tenant_account ON novel_codex_entries(tenant_id, account_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS novel_rule_conflicts (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    novel_id VARCHAR(36) NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    target_id VARCHAR(36) NOT NULL,
    target_type VARCHAR(64) NOT NULL,
    level VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    title VARCHAR(255) NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    resolution TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_novel_rule_conflicts_novel ON novel_rule_conflicts(novel_id, status, level) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_novel_rule_conflicts_target ON novel_rule_conflicts(target_id, target_type) WHERE deleted_at IS NULL;

INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES
    ('a0000001-0000-0000-0000-000000000040', 'default', 'novel.read', '小说读取', '查看小说管理数据的权限', 'api', true, true, NOW(), NOW()),
    ('a0000001-0000-0000-0000-000000000041', 'default', 'novel.write', '小说写入', '创建和编辑小说、章节、资料库的权限', 'api', true, true, NOW(), NOW()),
    ('a0000001-0000-0000-0000-000000000042', 'default', 'novel.delete', '小说删除', '删除小说的权限', 'api', true, true, NOW(), NOW())
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES
    ('b0000001-0000-0000-0000-000000000040', 'a0000001-0000-0000-0000-000000000040', 'GET /api/v1/novel-admin/*', 'api_path', '小说管理读取 API', 'default', NOW(), NOW()),
    ('b0000001-0000-0000-0000-000000000041', 'a0000001-0000-0000-0000-000000000041', 'POST /api/v1/novel-admin/*', 'api_path', '小说管理创建 API', 'default', NOW(), NOW()),
    ('b0000001-0000-0000-0000-000000000042', 'a0000001-0000-0000-0000-000000000041', 'PUT /api/v1/novel-admin/*', 'api_path', '小说管理更新 API', 'default', NOW(), NOW()),
    ('b0000001-0000-0000-0000-000000000043', 'a0000001-0000-0000-0000-000000000041', 'PATCH /api/v1/novel-admin/*', 'api_path', '小说管理状态 API', 'default', NOW(), NOW()),
    ('b0000001-0000-0000-0000-000000000044', 'a0000001-0000-0000-0000-000000000042', 'DELETE /api/v1/novel-admin/*', 'api_path', '小说管理删除 API', 'default', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM permission_resources WHERE id IN (
    'b0000001-0000-0000-0000-000000000040',
    'b0000001-0000-0000-0000-000000000041',
    'b0000001-0000-0000-0000-000000000042',
    'b0000001-0000-0000-0000-000000000043',
    'b0000001-0000-0000-0000-000000000044'
);

DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000040',
    'a0000001-0000-0000-0000-000000000041',
    'a0000001-0000-0000-0000-000000000042'
);

DROP TABLE IF EXISTS novel_rule_conflicts;
DROP TABLE IF EXISTS novel_codex_entries;
DROP TABLE IF EXISTS novel_chapter_versions;
DROP TABLE IF EXISTS novel_chapters;
DROP TABLE IF EXISTS novel_units;
DROP TABLE IF EXISTS novel_volumes;
DROP TABLE IF EXISTS novels;

-- +goose StatementEnd
