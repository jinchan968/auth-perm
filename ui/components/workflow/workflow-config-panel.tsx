import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { Node } from '@xyflow/react'

interface Props {
  selectedNode: Node | null
  onNodeUpdate: (node: Node) => void
}

export default function WorkflowConfigPanel({ selectedNode, onNodeUpdate }: Props) {
  if (!selectedNode) {
    return (
      <div className="w-80 border-l bg-slate-50 p-4 flex items-center justify-center">
        <p className="text-sm text-slate-400">点击节点编辑属性</p>
      </div>
    )
  }

  const { type, data } = selectedNode

  const updateData = (updates: Record<string, unknown>) => {
    onNodeUpdate({
      ...selectedNode,
      data: { ...data, ...updates },
    })
  }

  return (
    <div className="w-80 border-l bg-white p-4 overflow-y-auto">
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
        <div className="space-y-3">
          <p className="text-xs text-slate-400">分支配置与规则编辑（敬请期待）</p>
        </div>
      )}

      {type === 'transform' && (
        <div className="space-y-3">
          <div>
            <Label>操作类型</Label>
            <Select
              value={(data.operation as string) || ''}
              onValueChange={(v) => updateData({ operation: v })}
            >
              <SelectTrigger>
                <SelectValue placeholder="选择操作" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="regex_extract">正则提取</SelectItem>
                <SelectItem value="regex_replace">正则替换</SelectItem>
                <SelectItem value="trim">去首尾空白</SelectItem>
                <SelectItem value="markdown_to_text">去 Markdown</SelectItem>
                <SelectItem value="truncate">截断</SelectItem>
                <SelectItem value="to_uppercase">转大写</SelectItem>
                <SelectItem value="to_lowercase">转小写</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      )}

      {type === 'merge' && (
        <div className="space-y-3">
          <div>
            <Label>合并策略</Label>
            <Select
              value={(data.strategy as string) || 'concat'}
              onValueChange={(v) => updateData({ strategy: v })}
            >
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
        </div>
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
    </div>
  )
}
