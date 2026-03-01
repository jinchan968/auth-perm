-- +goose Up
-- +goose StatementBegin
-- =====================================================
-- Todo 待办系统数据库结构
-- =====================================================

-- =====================================================
-- 1. 待办分类表 (todo_categories)
-- =====================================================
CREATE TABLE IF NOT EXISTS todo_categories (
    id          VARCHAR(36)  PRIMARY KEY,
    tenant_id   VARCHAR(36)  NOT NULL,
    account_id  VARCHAR(36)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    color       VARCHAR(7)   NOT NULL DEFAULT '#6366f1',  -- hex 颜色
    icon        VARCHAR(50),                               -- lucide 图标名（可选）
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_todo_categories_account_tenant
    ON todo_categories(account_id, tenant_id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE  todo_categories                IS '待办分类表（Project 概念）';
COMMENT ON COLUMN todo_categories.tenant_id      IS '租户ID（冗余，方便隔离）';
COMMENT ON COLUMN todo_categories.account_id     IS '所属账户ID';
COMMENT ON COLUMN todo_categories.color          IS '分类颜色，HEX格式';
COMMENT ON COLUMN todo_categories.icon           IS '分类图标，lucide 图标名';
COMMENT ON COLUMN todo_categories.sort_order     IS '排序权重，越小越靠前';

-- =====================================================
-- 2. 待办表 (todos)
-- =====================================================
CREATE TABLE IF NOT EXISTS todos (
    id           VARCHAR(36)   PRIMARY KEY,
    tenant_id    VARCHAR(36)   NOT NULL,
    account_id   VARCHAR(36)   NOT NULL,
    category_id  VARCHAR(36),                                -- 可为空（未分类）
    title        VARCHAR(500)  NOT NULL,
    description  TEXT,
    status       VARCHAR(20)   NOT NULL DEFAULT 'pending',   -- pending / in_progress / completed / cancelled
    priority     VARCHAR(20)   NOT NULL DEFAULT 'medium',    -- low / medium / high / urgent
    deadline     TIMESTAMPTZ,                                -- 截止时间（可选）
    completed_at TIMESTAMPTZ,                                -- 完成时间
    sort_order   INT           NOT NULL DEFAULT 0,           -- 排序权重
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

-- 核心查询索引
CREATE INDEX IF NOT EXISTS idx_todos_account_tenant
    ON todos(account_id, tenant_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_todos_status
    ON todos(status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_todos_priority
    ON todos(priority)
    WHERE deleted_at IS NULL;

-- 到期升级定时任务专用索引：只扫描活跃且有 deadline 的待办
CREATE INDEX IF NOT EXISTS idx_todos_deadline_active
    ON todos(deadline)
    WHERE deadline IS NOT NULL
      AND status   IN ('pending', 'in_progress')
      AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_todos_category
    ON todos(category_id)
    WHERE category_id IS NOT NULL AND deleted_at IS NULL;

COMMENT ON TABLE  todos              IS '待办事项表';
COMMENT ON COLUMN todos.tenant_id   IS '租户ID（冗余，方便数据隔离）';
COMMENT ON COLUMN todos.account_id  IS '所属账户ID（必填）';
COMMENT ON COLUMN todos.category_id IS '分类ID，NULL 表示未分类';
COMMENT ON COLUMN todos.status      IS '状态：pending/in_progress/completed/cancelled';
COMMENT ON COLUMN todos.priority    IS '优先级：low/medium/high/urgent';
COMMENT ON COLUMN todos.deadline    IS '截止时间，可选';
COMMENT ON COLUMN todos.completed_at IS '实际完成时间，完成时自动记录';
COMMENT ON COLUMN todos.sort_order  IS '用户自定义排序权重';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS todo_categories;
-- +goose StatementEnd
