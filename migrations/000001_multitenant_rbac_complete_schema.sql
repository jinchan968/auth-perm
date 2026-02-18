-- +goose Up
-- +goose StatementBegin
-- =====================================================
-- 多租户 RBAC 权限系统完整数据库结构
-- =====================================================
--
-- 此脚本创建完整的多租户 RBAC 权限系统数据库结构，包括：
-- 1. 用户认证系统（users, accounts, sessions）
-- 2. 多租户组织管理（organizations, account_org）
-- 3. 角色权限系统（roles, permissions, account_roles, role_permissions）
-- 4. 审计日志系统（audit_logs）
-- 5. 双因素认证（TOTP）
-- 6. 设备信任管理（device_trusts）
--
-- 核心特性：
-- - 基于账户的权限模型（account_id 而非 user_id）
-- - 完整的多租户支持（tenant_id）
-- - 冗余字段优化（中间表包含 tenant_id）
-- - 软删除支持
-- - 完整的索引策略
--
-- =====================================================

-- =====================================================
-- 1. 用户表 (users)
-- =====================================================
-- 用户表 - 全局用户信息，支持多种登录方式

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50),
    nickname VARCHAR(100),
    avatar VARCHAR(500),
    phone VARCHAR(20),
    email VARCHAR(255),
    password_hash VARCHAR(255),
    identifier_type VARCHAR(20) NOT NULL,
    identifier_value VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    profile JSONB DEFAULT '{}',
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 用户表索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_identifier_unique ON users(identifier_type, identifier_value);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 用户表注释
COMMENT ON TABLE users IS '用户表 - 全局用户信息，支持多种登录方式';
COMMENT ON COLUMN users.id IS '用户ID';
COMMENT ON COLUMN users.username IS '用户名';
COMMENT ON COLUMN users.nickname IS '昵称';
COMMENT ON COLUMN users.avatar IS '头像URL';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.email IS '邮箱地址';
COMMENT ON COLUMN users.password_hash IS '密码哈希';
COMMENT ON COLUMN users.identifier_type IS '标识类型：username/email/phone/oauth';
COMMENT ON COLUMN users.identifier_value IS '标识值';
COMMENT ON COLUMN users.status IS '用户状态：active/inactive/suspended';
COMMENT ON COLUMN users.profile IS '用户档案信息（JSON）';
COMMENT ON COLUMN users.preferences IS '用户偏好设置（JSON）';
COMMENT ON COLUMN users.deleted_at IS '删除时间（软删除）';

-- =====================================================
-- 2. 组织表 (organizations)
-- =====================================================
-- 组织表 - 多租户组织层级结构

CREATE TABLE IF NOT EXISTS organizations (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    parent_id VARCHAR(36),
    name VARCHAR(200) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    level INTEGER DEFAULT 1,
    path VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT unique_tenant_org_code UNIQUE (tenant_id, code)
);

-- 组织表索引
CREATE INDEX IF NOT EXISTS idx_organizations_parent_id ON organizations(parent_id);
CREATE INDEX IF NOT EXISTS idx_organizations_created_at ON organizations(created_at);
CREATE INDEX IF NOT EXISTS idx_organizations_updated_at ON organizations(updated_at);
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations(deleted_at);

-- 组织表注释
COMMENT ON TABLE organizations IS '组织表 - 多租户组织层级结构';
COMMENT ON COLUMN organizations.id IS '组织ID';
COMMENT ON COLUMN organizations.tenant_id IS '租户ID';
COMMENT ON COLUMN organizations.parent_id IS '父组织ID';
COMMENT ON COLUMN organizations.name IS '组织名称';
COMMENT ON COLUMN organizations.code IS '组织代码';
COMMENT ON COLUMN organizations.description IS '组织描述';
COMMENT ON COLUMN organizations.level IS '组织层级';
COMMENT ON COLUMN organizations.path IS '组织路径';
COMMENT ON COLUMN organizations.is_active IS '是否激活';

-- =====================================================
-- 3. 账户表 (accounts)
-- =====================================================
-- 账户表 - 用户在特定租户下的账户实例

CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    oauth_id VARCHAR(255),
    oauth_provider VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ,
    user_agent TEXT,
    ip_address INET,
    device_info JSONB DEFAULT '{}',
    email_verified BOOLEAN DEFAULT false,
    email_verified_at TIMESTAMPTZ,
    email_verification_token VARCHAR(255),
    email_verification_expires_at TIMESTAMPTZ,
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMPTZ,
    password_reset_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT unique_user_tenant UNIQUE (user_id, tenant_id)
);

-- 账户表索引
CREATE INDEX IF NOT EXISTS idx_accounts_oauth ON accounts(oauth_provider, oauth_id);
CREATE INDEX IF NOT EXISTS idx_accounts_deleted_at ON accounts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_accounts_email_verification_token ON accounts(email_verification_token);
CREATE INDEX IF NOT EXISTS idx_accounts_password_reset_token ON accounts(password_reset_token);

-- 账户表注释
COMMENT ON TABLE accounts IS '账户表 - 用户在特定租户下的账户实例';
COMMENT ON COLUMN accounts.id IS '账户ID';
COMMENT ON COLUMN accounts.user_id IS '关联的用户ID';
COMMENT ON COLUMN accounts.tenant_id IS '租户ID';
COMMENT ON COLUMN accounts.account_type IS '账户类型：email/oauth';
COMMENT ON COLUMN accounts.oauth_id IS 'OAuth用户ID';
COMMENT ON COLUMN accounts.oauth_provider IS 'OAuth提供商';
COMMENT ON COLUMN accounts.last_login_at IS '最后登录时间';
COMMENT ON COLUMN accounts.email_verified IS '邮箱是否验证';
COMMENT ON COLUMN accounts.email_verified_at IS '邮箱验证时间';

-- =====================================================
-- 4. 角色表 (roles)
-- =====================================================
-- 角色表 - 多租户角色定义

CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    org_id VARCHAR(36),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT unique_tenant_role_code UNIQUE (tenant_id, code)
);

-- 角色表索引
CREATE INDEX IF NOT EXISTS idx_roles_org_id ON roles(org_id);
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles(deleted_at);

-- 角色表注释
COMMENT ON TABLE roles IS '角色表 - 多租户角色定义';
COMMENT ON COLUMN roles.id IS '角色ID';
COMMENT ON COLUMN roles.tenant_id IS '租户ID';
COMMENT ON COLUMN roles.org_id IS '所属组织ID';
COMMENT ON COLUMN roles.name IS '角色名称';
COMMENT ON COLUMN roles.code IS '角色代码';
COMMENT ON COLUMN roles.is_system IS '是否系统角色';
COMMENT ON COLUMN roles.is_active IS '是否激活';
COMMENT ON COLUMN roles.priority IS '角色优先级';

-- =====================================================
-- 5. 权限表 (permissions)
-- =====================================================
-- 权限表 - 多租户权限定义

CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT unique_tenant_permission_code UNIQUE (tenant_id, code)
);

-- 权限表索引
CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action);
CREATE INDEX IF NOT EXISTS idx_permissions_deleted_at ON permissions(deleted_at);

-- 权限表注释
COMMENT ON TABLE permissions IS '权限表 - 多租户权限定义';
COMMENT ON COLUMN permissions.id IS '权限ID';
COMMENT ON COLUMN permissions.tenant_id IS '租户ID';
COMMENT ON COLUMN permissions.code IS '权限代码';
COMMENT ON COLUMN permissions.name IS '权限名称';
COMMENT ON COLUMN permissions.resource IS '资源类型';
COMMENT ON COLUMN permissions.action IS '操作类型';
COMMENT ON COLUMN permissions.is_system IS '是否系统权限';

-- =====================================================
-- 6. 会话表 (sessions)
-- =====================================================
-- 会话表 - 用户登录会话管理

CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL DEFAULT 'default',
    device_info JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 会话表索引
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_tenant_id ON sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_account_id ON sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_sessions_is_active ON sessions(is_active);

-- 会话表注释
COMMENT ON TABLE sessions IS '会话表 - 用户登录会话管理';
COMMENT ON COLUMN sessions.id IS '会话ID';
COMMENT ON COLUMN sessions.user_id IS '用户ID';
COMMENT ON COLUMN sessions.account_id IS '账户ID';
COMMENT ON COLUMN sessions.token_hash IS '令牌哈希';
COMMENT ON COLUMN sessions.tenant_id IS '租户ID';
COMMENT ON COLUMN sessions.device_info IS '设备信息（JSON）';
COMMENT ON COLUMN sessions.expires_at IS '过期时间';
COMMENT ON COLUMN sessions.is_active IS '是否激活';

-- =====================================================
-- 7. TOTP密钥表 (totp_secrets)
-- =====================================================
-- TOTP密钥表 - 双因素认证

CREATE TABLE IF NOT EXISTS totp_secrets (
    id VARCHAR(36) PRIMARY KEY,
    account_id VARCHAR(36) NOT NULL,
    secret VARCHAR(255) NOT NULL,
    algorithm VARCHAR(20) DEFAULT 'SHA1',
    digits INTEGER DEFAULT 6,
    period INTEGER DEFAULT 30,
    is_enabled BOOLEAN DEFAULT false,
    backup_codes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- TOTP密钥表索引
CREATE INDEX IF NOT EXISTS idx_totp_secrets_account_id ON totp_secrets(account_id);
CREATE INDEX IF NOT EXISTS idx_totp_secrets_enabled ON totp_secrets(is_enabled);

-- TOTP密钥表注释
COMMENT ON TABLE totp_secrets IS 'TOTP密钥表 - 双因素认证';
COMMENT ON COLUMN totp_secrets.id IS '记录ID';
COMMENT ON COLUMN totp_secrets.account_id IS '账户ID';
COMMENT ON COLUMN totp_secrets.secret IS 'TOTP密钥';
COMMENT ON COLUMN totp_secrets.algorithm IS '哈希算法';
COMMENT ON COLUMN totp_secrets.digits IS '验证码位数';
COMMENT ON COLUMN totp_secrets.period IS '时间周期（秒）';
COMMENT ON COLUMN totp_secrets.is_enabled IS '是否启用';
COMMENT ON COLUMN totp_secrets.backup_codes IS '备用码';

-- =====================================================
-- 8. TOTP备份码使用记录表 (totp_backup_code_usage)
-- =====================================================

CREATE TABLE IF NOT EXISTS totp_backup_code_usage (
    id VARCHAR(36) PRIMARY KEY,
    account_id VARCHAR(36) NOT NULL,
    code VARCHAR(36) NOT NULL,
    used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(45),
    user_agent TEXT,
    success BOOLEAN DEFAULT true
);

-- TOTP备份码使用记录表索引
CREATE INDEX IF NOT EXISTS idx_backup_code_usage_account ON totp_backup_code_usage(account_id);
CREATE INDEX IF NOT EXISTS idx_backup_code_usage_code ON totp_backup_code_usage(code);

-- TOTP备份码使用记录表注释
COMMENT ON TABLE totp_backup_code_usage IS 'TOTP备份码使用记录表 - 记录备份码使用情况';

-- =====================================================
-- 9. 审计日志表 (audit_logs)
-- =====================================================
-- 审计日志表 - 记录系统中所有重要操作

CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36), -- 可选字段，记录操作对应的账户
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255),
    old_values JSONB DEFAULT '{}',
    new_values JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 审计日志表索引
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_account_id ON audit_logs(account_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_deleted_at ON audit_logs(deleted_at);

-- 审计日志表注释
COMMENT ON TABLE audit_logs IS '审计日志表 - 记录系统中所有重要操作';
COMMENT ON COLUMN audit_logs.id IS '审计日志ID';
COMMENT ON COLUMN audit_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN audit_logs.user_id IS '操作用户ID';
COMMENT ON COLUMN audit_logs.account_id IS '操作账户ID（可选）';
COMMENT ON COLUMN audit_logs.action IS '操作类型';
COMMENT ON COLUMN audit_logs.resource_type IS '资源类型';
COMMENT ON COLUMN audit_logs.resource_id IS '资源ID';
COMMENT ON COLUMN audit_logs.old_values IS '操作前的旧值（JSON）';
COMMENT ON COLUMN audit_logs.new_values IS '操作后的新值（JSON）';
COMMENT ON COLUMN audit_logs.success IS '操作是否成功';
COMMENT ON COLUMN audit_logs.error_message IS '错误信息';

-- =====================================================
-- 10. 设备信任表 (device_trusts)
-- =====================================================
-- 设备信任表 - 管理用户设备信任状态

CREATE TABLE IF NOT EXISTS device_trusts (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    device_info JSONB DEFAULT '{}',
    trusted BOOLEAN DEFAULT true NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 设备信任表索引
CREATE INDEX IF NOT EXISTS idx_device_trusts_user_id ON device_trusts(user_id);
CREATE INDEX IF NOT EXISTS idx_device_trusts_tenant_id ON device_trusts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_device_trusts_device_id ON device_trusts(device_id);
CREATE INDEX IF NOT EXISTS idx_device_trusts_deleted_at ON device_trusts(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_device_trusts_user_device ON device_trusts(user_id, device_id) WHERE deleted_at IS NULL;

-- 设备信任表注释
COMMENT ON TABLE device_trusts IS '设备信任表 - 管理用户设备信任状态';
COMMENT ON COLUMN device_trusts.id IS '记录ID';
COMMENT ON COLUMN device_trusts.tenant_id IS '租户ID';
COMMENT ON COLUMN device_trusts.user_id IS '用户ID';
COMMENT ON COLUMN device_trusts.device_id IS '设备ID';
COMMENT ON COLUMN device_trusts.device_info IS '设备信息（JSON）';
COMMENT ON COLUMN device_trusts.trusted IS '是否信任';
COMMENT ON COLUMN device_trusts.reason IS '信任/不信任原因';

-- =====================================================
-- 11. 账户-组织关联表 (account_org)
-- =====================================================
-- 账户组织关联表 - 替代 user_org，使用 account_id 和 tenant_id

CREATE TABLE IF NOT EXISTS account_org (
    id VARCHAR(36) PRIMARY KEY, -- 主键ID
    account_id VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL, -- 冗余字段，用于快速过滤
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 账户-组织关联表索引
CREATE INDEX IF NOT EXISTS idx_account_org_account_id ON account_org(account_id);
CREATE INDEX IF NOT EXISTS idx_account_org_organization_id ON account_org(organization_id);
CREATE INDEX IF NOT EXISTS idx_account_org_tenant_id ON account_org(tenant_id);

-- 账户-组织关联表注释
COMMENT ON TABLE account_org IS '账户组织关联表 - 替代 user_org，使用 account_id 和 tenant_id';
COMMENT ON COLUMN account_org.account_id IS '账户ID，关联 accounts 表';
COMMENT ON COLUMN account_org.organization_id IS '组织ID，关联 organizations 表';
COMMENT ON COLUMN account_org.tenant_id IS '租户ID，冗余字段用于快速过滤';

-- =====================================================
-- 12. 账户-角色关联表 (account_roles)
-- =====================================================
-- 账户角色关联表 - 替代 user_roles，使用 account_id 和 tenant_id

CREATE TABLE IF NOT EXISTS account_roles (
    id VARCHAR(36) PRIMARY KEY, -- 主键ID
    account_id VARCHAR(36) NOT NULL, -- 使用 account_id 替代 user_id
    role_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL, -- 冗余字段，用于快速过滤
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 账户-角色关联表索引
CREATE INDEX IF NOT EXISTS idx_account_roles_account_id ON account_roles(account_id);
CREATE INDEX IF NOT EXISTS idx_account_roles_role_id ON account_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_account_roles_tenant_id ON account_roles(tenant_id);

-- 账户-角色关联表注释
COMMENT ON TABLE account_roles IS '账户角色关联表 - 替代 user_roles，使用 account_id 和 tenant_id';
COMMENT ON COLUMN account_roles.id IS '关联记录ID';
COMMENT ON COLUMN account_roles.account_id IS '账户ID，关联 accounts 表';
COMMENT ON COLUMN account_roles.role_id IS '角色ID，关联 roles 表';
COMMENT ON COLUMN account_roles.tenant_id IS '租户ID，冗余字段用于快速过滤';
COMMENT ON COLUMN account_roles.created_at IS '创建时间';

-- =====================================================
-- 13. 角色-权限关联表 (role_permissions)
-- =====================================================
-- 角色权限关联表 - 包含 tenant_id 冗余字段

CREATE TABLE IF NOT EXISTS role_permissions (
    id VARCHAR(36) PRIMARY KEY, -- 主键ID
    role_id VARCHAR(36) NOT NULL,
    permission_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL DEFAULT 'default', -- 冗余字段，用于快速过滤
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 角色-权限关联表索引
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_tenant_id ON role_permissions(tenant_id);

-- 角色-权限关联表注释
COMMENT ON TABLE role_permissions IS '角色权限关联表 - 包含 tenant_id 冗余字段';
COMMENT ON COLUMN role_permissions.id IS '关联记录ID';
COMMENT ON COLUMN role_permissions.role_id IS '角色ID';
COMMENT ON COLUMN role_permissions.permission_id IS '权限ID';
COMMENT ON COLUMN role_permissions.tenant_id IS '租户ID，冗余字段用于快速过滤';
COMMENT ON COLUMN role_permissions.created_at IS '创建时间';

-- =====================================================
-- 14. 权限-资源关联表 (permission_resources)
-- =====================================================
-- 权限资源关联表 - 权限持有的具体资源，支持多种资源类型

CREATE TABLE IF NOT EXISTS permission_resources (
    id VARCHAR(36) PRIMARY KEY,
    permission_id VARCHAR(36) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_name VARCHAR(255),
    tenant_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT unique_permission_resource UNIQUE (permission_id, resource_id, resource_type, tenant_id)
);

-- 权限-资源关联表索引
CREATE INDEX IF NOT EXISTS idx_permission_resources_permission_id ON permission_resources(permission_id);
CREATE INDEX IF NOT EXISTS idx_permission_resources_resource_type ON permission_resources(resource_type);
CREATE INDEX IF NOT EXISTS idx_permission_resources_tenant_id ON permission_resources(tenant_id);
CREATE INDEX IF NOT EXISTS idx_permission_resources_deleted_at ON permission_resources(deleted_at);

-- 权限-资源关联表注释
COMMENT ON TABLE permission_resources IS '权限资源关联表 - 权限持有的具体资源';
COMMENT ON COLUMN permission_resources.id IS '关联记录ID';
COMMENT ON COLUMN permission_resources.permission_id IS '权限ID';
COMMENT ON COLUMN permission_resources.resource_id IS '资源ID（API路径、菜单标识等）';
COMMENT ON COLUMN permission_resources.resource_type IS '资源类型：api_path, menu, button';
COMMENT ON COLUMN permission_resources.resource_name IS '资源名称（冗余字段）';
COMMENT ON COLUMN permission_resources.tenant_id IS '租户ID，冗余字段用于快速过滤';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- =====================================================
-- 回滚多租户 RBAC 权限系统数据库结构
-- =====================================================

-- 删除关联表（按依赖关系顺序）
DROP TABLE IF EXISTS permission_resources CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS account_roles CASCADE;
DROP TABLE IF EXISTS account_org CASCADE;
DROP TABLE IF EXISTS device_trusts CASCADE;

-- 删除审计日志表
DROP TABLE IF EXISTS audit_logs CASCADE;

-- 删除TOTP相关表
DROP TABLE IF EXISTS totp_backup_code_usage CASCADE;
DROP TABLE IF EXISTS totp_secrets CASCADE;

-- 删除主表
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;

-- 删除用户表（最后删除）
DROP TABLE IF EXISTS users CASCADE;

-- +goose StatementEnd
