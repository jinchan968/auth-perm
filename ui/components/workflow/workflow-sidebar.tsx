'use client'

import { useDraggable, type DragEndEvent } from '@dnd-kit/core'
import { Brain, GitFork, Wrench, GitMerge, Workflow, Play } from 'lucide-react'

interface SidebarItem {
  type: string
  label: string
  icon: React.ReactNode
  color: string
}

const items: SidebarItem[] = [
  { type: 'trigger', label: 'Trigger', icon: <Play className="h-4 w-4" />, color: 'text-green-600' },
  { type: 'llm', label: 'LLM', icon: <Brain className="h-4 w-4" />, color: 'text-blue-600' },
  { type: 'condition', label: 'Condition', icon: <GitFork className="h-4 w-4" />, color: 'text-amber-600' },
  { type: 'transform', label: 'Transform', icon: <Wrench className="h-4 w-4" />, color: 'text-purple-600' },
  { type: 'merge', label: 'Merge', icon: <GitMerge className="h-4 w-4" />, color: 'text-rose-600' },
  { type: 'output', label: 'Output', icon: <Workflow className="h-4 w-4" />, color: 'text-slate-600' },
]

function DraggableItem({ item }: { item: SidebarItem }) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: `sidebar-${item.type}`,
    data: { nodeType: item.type },
  })

  return (
    <div
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      className={`flex items-center gap-2 px-3 py-2 rounded-md border bg-white cursor-grab hover:shadow-sm text-sm ${
        isDragging ? 'opacity-50' : ''
      }`}
    >
      <span className={item.color}>{item.icon}</span>
      <span>{item.label}</span>
    </div>
  )
}

export function WorkflowSidebar() {
  return (
    <div className="w-48 border-r bg-slate-50 p-3 space-y-2">
      <p className="text-xs font-semibold text-slate-500 mb-3">节点类型</p>
      {items.map((item) => (
        <DraggableItem key={item.type} item={item} />
      ))}
    </div>
  )
}
