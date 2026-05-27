'use client'

import dynamic from 'next/dynamic'

const WorkflowCanvas = dynamic(
  () => import('./workflow-canvas'),
  { ssr: false }
)

export default function WorkflowDesigner() {
  return <WorkflowCanvas />
}
