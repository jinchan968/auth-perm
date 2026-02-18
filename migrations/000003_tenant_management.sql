-- +goose Up
-- +goose StatementBegin
-- =====================================================
-- 租户管理表
-- =====================================================
--
-- 此脚本创建租户管理相关表：
-- 1. tenants - 租户表
--
-- =====================================================

-- =====================================================
-- 1. 租户表 (tenants)
-- =====================================================
-- 租户表 - 多租户系统的基础，每个租户代表一个独立的组织或客户

CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    code VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    expire_at TIMESTAMPTZ,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 租户表唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_code ON tenants(code);

-- 租户表索引
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_tenants_plan ON tenants(plan);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);
CREATE INDEX IF NOT EXISTS idx_tenants_updated_at ON tenants(updated_at);

-- 租户表注释
COMMENT ON TABLE tenants IS '租户表 - 多租户系统的基础，每个租户代表一个独立的组织或客户';
COMMENT ON COLUMN tenants.id IS '租户ID (UUID)';
COMMENT ON COLUMN tenants.name IS '租户名称';
COMMENT ON COLUMN tenants.code IS '租户代码（唯一）';
COMMENT ON COLUMN tenants.status IS '租户状态：active/suspended/deleted';
COMMENT ON COLUMN tenants.plan IS '套餐：free/basic/pro/enterprise';
COMMENT ON COLUMN tenants.expire_at IS '过期时间';
COMMENT ON COLUMN tenants.settings IS '租户设置（JSON）：功能开关、配额限制等';
COMMENT ON COLUMN tenants.created_at IS '创建时间';
COMMENT ON COLUMN tenants.updated_at IS '更新时间';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS tenants;

-- +goose StatementEnd
