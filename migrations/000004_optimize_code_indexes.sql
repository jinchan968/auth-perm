-- +goose Up
-- +goose StatementBegin
-- 添加角色表 (tenant_id, code) 联合唯一索引
-- 权限表的联合索引已通过 GORM 的 uniqueIndex:idx_tenant_code 自动创建
-- 角色表需要修复为联合唯一索引

-- 创建角色表 (tenant_id, code) 联合唯一索引
-- 如果 idx_tenant_code 存在（只有 code 字段），先删除
DROP INDEX IF EXISTS idx_tenant_code;
-- 创建新的联合唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_tenant_code ON roles(tenant_id, code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 回滚索引更改
DROP INDEX IF EXISTS idx_roles_tenant_code;
-- 注意：无法完全恢复旧状态，因为旧索引是通过 GORM 自动创建的

-- +goose StatementEnd
