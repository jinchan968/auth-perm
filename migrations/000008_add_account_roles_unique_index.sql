-- +goose Up
-- +goose StatementBegin

-- account_roles 表缺少 (account_id, role_id) 唯一约束，
-- 导致 ON CONFLICT (account_id, role_id) DO NOTHING 无法执行。
-- 同一个账户在同一租户下不应该重复分配同一角色。
CREATE UNIQUE INDEX IF NOT EXISTS idx_account_roles_account_role
    ON account_roles (account_id, role_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_account_roles_account_role;

-- +goose StatementEnd
