'use client'

import type { DragEvent } from 'react'
import { Brain, GitFork, Wrench, GitMerge, Workflow, Play } from 'lucide-react'

export const DRAG_MIME = 'application/x-workflow-node-type'

export interface SidebarItem {
  type: string
  label: string
  icon: React.ReactNode
  color: string
}

export const sidebarItems: SidebarItem[] = [
  { type: 'trigger', label: 'Trigger', icon: <Play className="h-4 w-4" />, color: 'text-green-600' },
  { type: 'llm', label: 'LLM', icon: <Brain className="h-4 w-4" />, color: 'text-blue-600' },
  { type: 'condition', label: 'Condition', icon: <GitFork className="h-4 w-4" />, color: 'text-amber-600' },
  { type: 'transform', label: 'Transform', icon: <Wrench className="h-4 w-4" />, color: 'text-purple-600' },
  { type: 'merge', label: 'Merge', icon: <GitMerge className="h-4 w-4" />, color: 'text-rose-600' },
  { type: 'output', label: 'Output', icon: <Workflow className="h-4 w-4" />, color: 'text-slate-600' },
]

interface WorkflowSidebarProps {
  onAddNode?: (type: string) => void
}

export function WorkflowSidebar({ onAddNode }: WorkflowSidebarProps) {
  const handleDragStart = (e: DragEvent<HTMLButtonElement>, type: string) => {
    e.dataTransfer.setData(DRAG_MIME, type)
    e.dataTransfer.effectAllowed = 'move'
  }

  return (
    <div className="w-48 border-r bg-slate-50 p-3 space-y-2">
      <p className="text-xs font-semibold text-slate-500 mb-3">节点类型</p>
      {sidebarItems.map((item) => (
        <button
          key={item.type}
          draggable
          onDragStart={(e) => handleDragStart(e, item.type)}
          onClick={() => onAddNode?.(item.type)}
          className="flex items-center gap-2 px-3 py-2 rounded-md border bg-white cursor-grab active:cursor-grabbing hover:shadow-sm text-sm w-full text-left"
        >
          <span className={item.color}>{item.icon}</span>
          <span>{item.label}</span>
        </button>
      ))}
    </div>
  )
}
