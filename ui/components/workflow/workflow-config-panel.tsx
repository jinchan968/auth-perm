import { memo } from 'react'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { X } from 'lucide-react'
import type { Node } from '@xyflow/react'
import type { BranchConfig, RuleGroup } from '@/types/workflow'

interface Props {
  selectedNode: Node | null
  onNodeUpdate: (node: Node) => void
  mobile?: boolean
  onClose?: () => void
}

function ConfigContent({ selectedNode, onNodeUpdate }: { selectedNode: Node; onNodeUpdate: (node: Node) => void }) {
  const { type, data } = selectedNode

  const updateData = (updates: Record<string, unknown>) => {
    onNodeUpdate({
      ...selectedNode,
      data: { ...data, ...updates },
    })
  }

  return (
    <>
      <h3 className="text-sm font-semibold mb-4">
        {type === 'trigger' && 'Trigger 配置'}
        {type === 'llm' && 'LLM 配置'}
        {type === 'condition' && 'Condition 配置'}
        {type === 'transform' && 'Transform 配置'}
        {type === 'merge' && 'Merge 配置'}
        {type === 'output' && 'Output 配置'}
      </h3>

      {type === 'trigger' && (
        <div className="space-y-3">
          <div>
            <Label>输入模式</Label>
            <Select
              value={(data.input_mode as string) || 'text'}
              onValueChange={(v) => updateData({ input_mode: v })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="text">自由文本</SelectItem>
                <SelectItem value="structured">结构化字段</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      )}

      {type === 'llm' && (
        <div className="space-y-3">
          <div>
            <Label>模型</Label>
            <Select
              value={(data.model_id as string) || ''}
              onValueChange={(v) => updateData({ model_id: v })}
            >
              <SelectTrigger>
                <SelectValue placeholder="选择模型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="deepseek-v4-pro">DeepSeek V4 Pro</SelectItem>
                <SelectItem value="glm-5.1">GLM-5.1</SelectItem>
                <SelectItem value="kimi-k2.6">Kimi K2.6</SelectItem>
                <SelectItem value="qwen3.6-plus">Qwen3.6 Plus</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label>System Prompt</Label>
            <Textarea
              value={(data.system_prompt as string) || ''}
              onChange={(e) => updateData({ system_prompt: e.target.value })}
              placeholder="你是一个助手..."
            />
          </div>
          <div>
            <Label>Temperature</Label>
            <Input
              type="number"
              min={0}
              max={2}
              step={0.1}
              value={(data.temperature as number) ?? 0.7}
              onChange={(e) => updateData({ temperature: parseFloat(e.target.value) || 0 })}
            />
          </div>
          <div>
            <Label>思考程度</Label>
            <Select
              value={(data.reasoning_mode as string) || 'low'}
              onValueChange={(v) => updateData({ reasoning_mode: v })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="low">轻度</SelectItem>
                <SelectItem value="medium">中度</SelectItem>
                <SelectItem value="high">深度</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      )}

      {type === 'condition' && (
        <ConditionConfig data={data} updateData={updateData} />
      )}

      {type === 'transform' && (
        <TransformConfig data={data} updateData={updateData} />
      )}

      {type === 'merge' && (
        <MergeConfig data={data} updateData={updateData} />
      )}

      {type === 'output' && (
        <div className="space-y-3">
          <div>
            <Label>输出格式</Label>
            <Select
              value={(data.format as string) || 'raw'}
              onValueChange={(v) => updateData({ format: v })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="raw">原始文本</SelectItem>
                <SelectItem value="json">JSON</SelectItem>
                <SelectItem value="markdown">Markdown</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center justify-between">
            <Label>汇聚模式</Label>
            <div className="flex gap-2">
              <button
                className={`px-2 py-1 text-xs rounded ${
                  data.join_mode === 'and' ? 'bg-primary text-white' : 'bg-slate-100'
                }`}
                onClick={() => updateData({ join_mode: 'and' })}
              >
                会签
              </button>
              <button
                className={`px-2 py-1 text-xs rounded ${
                  data.join_mode === 'or' ? 'bg-primary text-white' : 'bg-slate-100'
                }`}
                onClick={() => updateData({ join_mode: 'or' })}
              >
                或签
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

function ConditionConfig({ data, updateData }: { data: Record<string, unknown>; updateData: (u: Record<string, unknown>) => void }) {
  const branches = (data.branches as BranchConfig[]) || []
  const defaultHandle = (data.default_handle as string) || 'default'

  const addBranch = () => {
    const handle = `branch_${branches.length + 1}`
    updateData({
      branches: [...branches, { handle, rule: { logic: 'AND', rules: [] } }],
    })
  }

  const updateBranch = (index: number, updates: Partial<BranchConfig>) => {
    const next = [...branches]
    next[index] = { ...next[index], ...updates }
    updateData({ branches: next })
  }

  const removeBranch = (index: number) => {
    updateData({ branches: branches.filter((_, i) => i !== index) })
  }

  return (
    <div className="space-y-3">
      {branches.map((branch, i) => (
        <div key={branch.handle} className="border rounded p-2 space-y-2">
          <div className="flex items-center justify-between">
            <Input
              className="h-7 text-xs w-32"
              value={branch.handle}
              onChange={(e) => updateBranch(i, { handle: e.target.value })}
              placeholder="分支标识"
            />
            <button onClick={() => removeBranch(i)} className="text-slate-400 hover:text-red-500 ml-2">
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
          <RuleGroupEditor
            group={branch.rule}
            onChange={(rule) => updateBranch(i, { rule })}
          />
        </div>
      ))}
      <button onClick={addBranch} className="text-xs text-primary hover:underline">
        + 添加分支
      </button>
      <div>
        <Label className="text-xs">默认分支标识</Label>
        <Input
          className="h-7 text-xs"
          value={defaultHandle}
          onChange={(e) => updateData({ default_handle: e.target.value })}
        />
      </div>
    </div>
  )
}

function RuleGroupEditor({ group, onChange, depth }: { group: RuleGroup; onChange: (g: RuleGroup) => void; depth?: number }) {
  const toggleLogic = () => {
    onChange({ ...group, logic: group.logic === 'AND' ? 'OR' : 'AND' })
  }

  const addRule = () => {
    onChange({ ...group, rules: [...group.rules, { field: 'content', operator: 'contains', value: '' }] })
  }

  const addSubGroup = () => {
    onChange({ ...group, rules: [...group.rules, { sub_group: { logic: 'AND', rules: [] } }] })
  }

  const updateRule = (index: number, updates: Record<string, string | boolean>) => {
    const next = [...group.rules]
    next[index] = { ...next[index], ...updates }
    onChange({ ...group, rules: next })
  }

  const removeRule = (index: number) => {
    onChange({ ...group, rules: group.rules.filter((_, i) => i !== index) })
  }

  const updateSubGroup = (index: number, subGroup: RuleGroup) => {
    const next = [...group.rules]
    next[index] = { ...next[index], sub_group: subGroup }
    onChange({ ...group, rules: next })
  }

  const currentDepth = depth ?? 0

  return (
    <div className="space-y-1.5">
      <div className="flex items-center gap-1.5">
        <button
          onClick={toggleLogic}
          className="text-xs font-medium px-1.5 py-0.5 rounded bg-slate-100 hover:bg-slate-200"
        >
          {group.logic}
        </button>
        {currentDepth < 2 && (
          <button onClick={addSubGroup} className="text-xs text-blue-500 hover:underline">
            + 嵌套组
          </button>
        )}
      </div>
      {group.rules.map((rule, i) => {
        if (rule.sub_group) {
          return (
            <div key={i} className="ml-2 border-l-2 border-slate-200 pl-2">
              <div className="flex items-center justify-end">
                <button onClick={() => removeRule(i)} className="text-slate-400 hover:text-red-500">
                  <X className="h-3 w-3" />
                </button>
              </div>
              <RuleGroupEditor
                group={rule.sub_group}
                onChange={(g) => updateSubGroup(i, g)}
                depth={currentDepth + 1}
              />
            </div>
          )
        }
        return (
          <div key={i} className="flex items-center gap-1">
            <Select
              value={rule.field || 'content'}
              onValueChange={(v) => updateRule(i, { field: v })}
            >
              <SelectTrigger className="h-6 text-xs w-20">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="content">内容</SelectItem>
                <SelectItem value="length">长度</SelectItem>
              </SelectContent>
            </Select>
            <Select
              value={rule.operator || 'contains'}
              onValueChange={(v) => updateRule(i, { operator: v })}
            >
              <SelectTrigger className="h-6 text-xs w-24">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="contains">包含</SelectItem>
                <SelectItem value="not_contains">不包含</SelectItem>
                <SelectItem value="equals">等于</SelectItem>
                <SelectItem value="matches">正则匹配</SelectItem>
                <SelectItem value="starts_with">开头是</SelectItem>
                <SelectItem value="ends_with">结尾是</SelectItem>
                <SelectItem value="gt">大于</SelectItem>
                <SelectItem value="gte">大于等于</SelectItem>
                <SelectItem value="lt">小于</SelectItem>
                <SelectItem value="lte">小于等于</SelectItem>
                <SelectItem value="is_empty">为空</SelectItem>
                <SelectItem value="not_empty">不为空</SelectItem>
              </SelectContent>
            </Select>
            {!['is_empty', 'not_empty'].includes(rule.operator || '') && (
              <Input
                className="h-6 text-xs flex-1 min-w-0"
                value={rule.value || ''}
                onChange={(e) => updateRule(i, { value: e.target.value })}
                placeholder="值"
              />
            )}
            <button onClick={() => removeRule(i)} className="text-slate-400 hover:text-red-500 shrink-0">
              <X className="h-3 w-3" />
            </button>
          </div>
        )
      })}
      <button onClick={addRule} className="text-xs text-primary hover:underline">
        + 规则
      </button>
    </div>
  )
}

function TransformConfig({ data, updateData }: { data: Record<string, unknown>; updateData: (u: Record<string, unknown>) => void }) {
  const operation = (data.operation as string) || ''
  const params = (data.params as Record<string, unknown> | undefined) || {}

  const updateParams = (updates: Record<string, unknown>) => {
    updateData({ params: { ...params, ...updates } })
  }

  const operationLabels: Record<string, string> = {
    regex_extract: '正则提取',
    regex_replace: '正则替换',
    trim: '去空白',
    markdown_to_text: '去Markdown',
    extract_json: '提取JSON',
    truncate: '截断',
    to_uppercase: '转大写',
    to_lowercase: '转小写',
  }

  const needsPattern = ['regex_extract', 'regex_replace'].includes(operation)
  const needsReplacement = operation === 'regex_replace'
  const needsMaxLength = operation === 'truncate'

  return (
    <div className="space-y-3">
      <div>
        <Label>操作类型</Label>
        <Select value={operation} onValueChange={(v) => updateData({ operation: v, params: {} })}>
          <SelectTrigger>
            <SelectValue placeholder="选择操作" />
          </SelectTrigger>
          <SelectContent>
            {Object.entries(operationLabels).map(([val, label]) => (
              <SelectItem key={val} value={val}>{label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      {needsPattern && (
        <div>
          <Label>正则表达式</Label>
            <Input
              value={(params.pattern as string) || ''}
              onChange={(e) => updateParams({ pattern: e.target.value })}
              placeholder="输入正则表达式，如 \\d+"
            />
        </div>
      )}
      {needsReplacement && (
        <div>
          <Label>替换文本</Label>
            <Input
              value={(params.replacement as string) || ''}
              onChange={(e) => updateParams({ replacement: e.target.value })}
              placeholder="替换后的文本"
            />
        </div>
      )}
      {needsMaxLength && (
        <div>
          <Label>最大长度</Label>
            <Input
              type="number"
              min={1}
              value={(params.max_length as number) ?? 100}
              onChange={(e) => updateParams({ max_length: parseInt(e.target.value) || 100 })}
            />
        </div>
      )}
    </div>
  )
}

function MergeConfig({ data, updateData }: { data: Record<string, unknown>; updateData: (u: Record<string, unknown>) => void }) {
  const strategy = (data.strategy as string) || 'concat'

  return (
    <div className="space-y-3">
      <div>
        <Label>合并策略</Label>
        <Select value={strategy} onValueChange={(v) => updateData({ strategy: v, delimiter: undefined })}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="concat">拼接</SelectItem>
            <SelectItem value="first">取首</SelectItem>
            <SelectItem value="join">自定义分隔</SelectItem>
          </SelectContent>
        </Select>
      </div>
      {strategy === 'join' && (
        <div>
          <Label>分隔符</Label>
          <Input
            value={(data.delimiter as string) || ''}
            onChange={(e) => updateData({ delimiter: e.target.value })}
            placeholder="自定义分隔符，如 \n 或 ,"
          />
        </div>
      )}
    </div>
  )
}

export default memo(function WorkflowConfigPanel({ selectedNode, onNodeUpdate, mobile, onClose }: Props) {
  if (mobile) {
    if (!selectedNode) return null
    return (
      <div className="fixed bottom-0 left-0 right-0 z-40 bg-white border-t shadow-2xl rounded-t-2xl max-h-[60vh] animate-slide-up">
        <div className="flex items-center justify-between px-4 pt-3 pb-2">
          <div className="w-10 h-1 rounded-full bg-slate-300 mx-auto absolute left-1/2 -translate-x-1/2 top-2" />
          <div className="w-6" />
          <button
            onClick={onClose}
            className="p-1 rounded-full hover:bg-slate-100 transition-colors"
          >
            <X className="h-4 w-4 text-slate-500" />
          </button>
        </div>
        <div className="px-4 pb-4 overflow-y-auto max-h-[calc(60vh-44px)]">
          <ConfigContent selectedNode={selectedNode} onNodeUpdate={onNodeUpdate} />
        </div>
      </div>
    )
  }

  if (!selectedNode) {
    return (
      <div className="w-80 border-l bg-slate-50 p-4 flex items-center justify-center">
        <p className="text-sm text-slate-400">点击节点编辑属性</p>
      </div>
    )
  }

  return (
    <div className="w-80 border-l bg-white p-4 overflow-y-auto">
      <ConfigContent selectedNode={selectedNode} onNodeUpdate={onNodeUpdate} />
    </div>
  )
})
