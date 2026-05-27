'use client'

import { useCallback, useMemo, useState } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  type Connection,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { DndContext, type DragEndEvent, useSensor, useSensors, PointerSensor } from '@dnd-kit/core'
import { WorkflowSidebar } from './workflow-sidebar'
import WorkflowConfigPanel from './workflow-config-panel'
import WorkflowToolbar from './workflow-toolbar'
import TriggerNode from './nodes/trigger-node'
import LLMNode from './nodes/llm-node'
import ConditionNode from './nodes/condition-node'
import TransformNode from './nodes/transform-node'
import MergeNode from './nodes/merge-node'
import OutputNode from './nodes/output-node'

const nodeTypes = {
  trigger: TriggerNode,
  llm: LLMNode,
  condition: ConditionNode,
  transform: TransformNode,
  merge: MergeNode,
  output: OutputNode,
}

export default function WorkflowCanvas() {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
  )

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  )

  const onNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node)
  }, [])

  const onPaneClick = useCallback(() => {
    setSelectedNode(null)
  }, [])

  const isValidConnection = useCallback(
    (connection: Edge | Connection) => {
      if (connection.source === connection.target) return false
      if (connection.sourceHandle) {
        const existing = (edges as Edge[]).find(
          (e) => e.source === connection.source && e.sourceHandle === connection.sourceHandle
        )
        if (existing) return false
      }
      return true
    },
    [edges]
  )

  const onDragEnd = useCallback((_event: DragEndEvent) => {
    // TODO: handle node drop from sidebar
  }, [])

  return (
    <DndContext sensors={sensors} onDragEnd={onDragEnd}>
      <div className="flex h-[calc(100vh-200px)]">
        <WorkflowSidebar />
        <div className="flex-1 flex">
          <div className="flex-1 relative">
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onConnect={onConnect}
              onNodeClick={onNodeClick}
              onPaneClick={onPaneClick}
              isValidConnection={isValidConnection}
              nodeTypes={nodeTypes}
              fitView
            >
              <Background />
              <Controls />
              <MiniMap />
            </ReactFlow>
            <WorkflowToolbar
              nodes={nodes}
              edges={edges}
              setNodes={setNodes}
              setEdges={setEdges}
            />
          </div>
          <WorkflowConfigPanel
            selectedNode={selectedNode}
            onNodeUpdate={(updatedNode: Node) => {
              setNodes((nds) =>
                nds.map((n) => (n.id === updatedNode.id ? updatedNode : n))
              )
            }}
          />
        </div>
      </div>
    </DndContext>
  )
}
