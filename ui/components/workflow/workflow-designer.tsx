'use client'

import dynamic from 'next/dynamic'

const WorkflowCanvas = dynamic(
  () => import('./workflow-canvas'),
  { ssr: false }
)

interface WorkflowDesignerProps {
  onWorkflowChange?: (id: string | null) => void
}

export default function WorkflowDesigner({ onWorkflowChange }: WorkflowDesignerProps) {
  return <WorkflowCanvas onWorkflowChange={onWorkflowChange} />
}