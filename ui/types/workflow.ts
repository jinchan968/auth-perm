export type NodeType = 'trigger' | 'llm' | 'condition' | 'transform' | 'merge' | 'output'

export interface TriggerNodeData {
  label?: string
  input_mode?: 'text' | 'structured'
  input_schema?: Record<string, string>
  status?: string
}

export interface LLMNodeData {
  label?: string
  model_id?: string
  system_prompt?: string
  temperature?: number
  reasoning_mode?: 'low' | 'medium' | 'high'
  status?: string
}

export interface ConditionNodeData {
  label?: string
  branches?: BranchConfig[]
  default_handle?: string
  status?: string
}

export interface TransformNodeData {
  label?: string
  operation?: string
  pattern?: string
  replacement?: string
  max_length?: number
  status?: string
}

export interface MergeNodeData {
  label?: string
  strategy?: 'concat' | 'first' | 'join'
  delimiter?: string
  status?: string
}

export interface OutputNodeData {
  label?: string
  format?: 'raw' | 'json' | 'markdown'
  join_mode?: 'and' | 'or'
  status?: string
}

export type WorkflowNodeData = TriggerNodeData | LLMNodeData | ConditionNodeData | TransformNodeData | MergeNodeData | OutputNodeData

export interface FlowNode {
  id: string
  type: NodeType
  position: { x: number; y: number }
  data: WorkflowNodeData
}

export interface FlowEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
}

export interface FlowGraph {
  nodes: FlowNode[]
  edges: FlowEdge[]
  viewport: { x: number; y: number; zoom: number }
}

export interface Workflow {
  id: string
  tenant_id?: string
  name: string
  description?: string
  flow_json: FlowGraph
  template_id?: string
  status: 'draft' | 'published' | 'archived'
  created_at: string
  updated_at: string
}

export interface WorkflowListResponse {
  data: Workflow[]
  total: number
  page: number
  size: number
}

export interface WorkflowRun {
  id: string
  workflow_id: string
  execution_mode: 'sync' | 'async'
  input_text?: string
  input_json?: Record<string, unknown>
  result_json?: Record<string, unknown>
  status: 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
  started_at?: string
  finished_at?: string
  duration_ms: number
  error?: string
}

export interface WorkflowRunNode {
  id: string
  run_id: string
  node_id: string
  node_type: NodeType
  node_label?: string
  status: 'pending' | 'running' | 'success' | 'failed'
  input_json?: Record<string, unknown>
  output_json?: Record<string, unknown>
  error?: string
  started_at?: string
  finished_at?: string
  duration_ms: number
}

export interface RuleItem {
  field?: string
  operator?: string
  value?: string
  negate?: boolean
  sub_group?: RuleGroup
}

export interface RuleGroup {
  logic: 'AND' | 'OR'
  rules: RuleItem[]
}

export interface BranchConfig {
  handle: string
  rule: RuleGroup
}

export interface WSMessage {
  type: 'node_start' | 'node_end' | 'node_error' | 'run_end' | 'ping'
  run_id: string
  node_id?: string
  node_type?: NodeType
  duration_ms?: number
  output?: string
  error?: string
  status?: string
  result?: string
  ts: string
}
