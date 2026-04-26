-- +goose Up
-- +goose StatementBegin

-- role_permissions 表需要 (role_id, permission_id) 唯一约束，
-- 以保证角色分配权限时的幂等写入与并发一致性。
CREATE UNIQUE INDEX IF NOT EXISTS idx_role_permissions_role_permission
    ON role_permissions (role_id, permission_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_role_permissions_role_permission;

-- +goose StatementEnd

