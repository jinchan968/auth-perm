-- +goose Up
-- +goose StatementBegin

-- 工作流定义表
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    flow_json JSONB NOT NULL DEFAULT '{}',
    template_id UUID,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflows_tenant_account ON workflows(tenant_id, account_id);
CREATE INDEX IF NOT EXISTS idx_workflows_template ON workflows(template_id);

-- 工作流执行记录表
CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    execution_mode VARCHAR(8) NOT NULL DEFAULT 'sync',
    input_text TEXT,
    input_json JSONB,
    result_json JSONB,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_wf_runs_workflow ON workflow_runs(workflow_id);
CREATE INDEX IF NOT EXISTS idx_wf_runs_tenant ON workflow_runs(tenant_id, account_id, started_at DESC);

-- 工作流逐节点执行记录表
CREATE TABLE IF NOT EXISTS workflow_run_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    node_id VARCHAR(128) NOT NULL,
    node_type VARCHAR(32) NOT NULL,
    node_label VARCHAR(128),
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    input_json JSONB,
    output_json JSONB,
    error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    retry_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_wf_run_nodes_run ON workflow_run_nodes(run_id);
CREATE INDEX IF NOT EXISTS idx_wf_run_nodes_run_node ON workflow_run_nodes(run_id, node_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS workflow_run_nodes;
DROP TABLE IF EXISTS workflow_runs;
DROP TABLE IF EXISTS workflows;

-- +goose StatementEnd
