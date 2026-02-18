-- +goose Up
-- +goose StatementBegin
-- =====================================================
-- 移除 permissions 表的 action 字段
-- =====================================================
--
-- 此迁移移除权限表中的 action 字段，
-- 权限检查现在只基于 permission_code，不再使用 resource.action 格式
--
-- =====================================================

-- 1. 删除 action 字段的索引
DROP INDEX IF EXISTS idx_permissions_action;

-- 2. 删除 action 字段
ALTER TABLE permissions DROP COLUMN IF EXISTS action;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- =====================================================
-- 回滚: 恢复 permissions 表的 action 字段
-- =====================================================

-- 1. 添加 action 字段
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS action VARCHAR(50) NOT NULL DEFAULT 'read';

-- 2. 重建索引
CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action);

-- +goose StatementEnd
