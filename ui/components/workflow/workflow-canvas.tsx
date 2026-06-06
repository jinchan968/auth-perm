'use client'

import { useCallback, useEffect, useState, useRef, type DragEvent } from 'react'
import {
  ReactFlow,
  ReactFlowProvider,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  useReactFlow,
  addEdge,
  type Connection,
  type Node,
  type Edge,
  type NodeChange,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { Plus, ChevronDown, ChevronUp } from 'lucide-react'
import { WorkflowSidebar, sidebarItems, DRAG_MIME } from './workflow-sidebar'
import { getNodeColor, getNodeStrokeColor } from './nodes/node-utils'
import WorkflowConfigPanel from './workflow-config-panel'
import WorkflowToolbar from './workflow-toolbar'
import WorkflowSelector from './workflow-selector'
import TriggerNode from './nodes/trigger-node'
import LLMNode from './nodes/llm-node'
import ConditionNode from './nodes/condition-node'
import TransformNode from './nodes/transform-node'
import MergeNode from './nodes/merge-node'
import OutputNode from './nodes/output-node'
import { useTenant } from '@/lib/tenant-context'
import { createWorkflow, updateWorkflow, validateWorkflow, executeWorkflow } from '@/lib/api/workflow'
import { showError, showSuccess } from '@/lib/toast'
import type { FlowGraph, Workflow } from '@/types/workflow'

const nodeTypes = {
  trigger: TriggerNode,
  llm: LLMNode,
  condition: ConditionNode,
  transform: TransformNode,
  merge: MergeNode,
  output: OutputNode,
}

interface WorkflowCanvasProps {
  onWorkflowChange?: (id: string | null) => void
}

export default function WorkflowCanvas({ onWorkflowChange }: WorkflowCanvasProps) {
  return (
    <ReactFlowProvider>
      <WorkflowCanvasInner onWorkflowChange={onWorkflowChange} />
    </ReactFlowProvider>
  )
}

function WorkflowCanvasInner({ onWorkflowChange }: WorkflowCanvasProps) {
  const [nodes, setNodes, onNodesChangeRaw] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false)
  const [workflowId, setWorkflowId] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [validating, setValidating] = useState(false)
  const [running, setRunning] = useState(false)
  const [dirty, setDirty] = useState(false)
  const [currentWorkflow, setCurrentWorkflow] = useState<Workflow | null>(null)
  const [selectorRefreshKey, setSelectorRefreshKey] = useState(0)
  const idCounterRef = useRef(0)
  const suppressDirtyRef = useRef(false)
  const { screenToFlowPosition, fitView, getViewport } = useReactFlow()
  const { tenantId } = useTenant()
  const firstNodeId = nodes.length === 1 ? nodes[0].id : null

  const nextNodeId = useCallback((type: string): string => {
    idCounterRef.current++
    return `${type}-${idCounterRef.current}-${Date.now()}`
  }, [])

  const onConnect = useCallback(
    (params: Connection) => {
      setEdges((eds) => addEdge(params, eds))
      setDirty(true)
    },
    [setEdges]
  )

  const onNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node)
  }, [])

  const onPaneClick = useCallback(() => {
    setSelectedNode(null)
  }, [])

  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      const removedIds = new Set(
        changes.filter((c) => c.type === 'remove').map((c) => ('id' in c ? c.id : ''))
      )
      if (removedIds.size > 0 && selectedNode && removedIds.has(selectedNode.id)) {
        setSelectedNode(null)
      }
      onNodesChangeRaw(changes)
      if (!suppressDirtyRef.current && changes.some((change) => change.type !== 'select')) {
        setDirty(true)
      }
    },
    [onNodesChangeRaw, selectedNode]
  )

  const isValidConnection = useCallback(
    (connection: Edge | Connection) => {
      if (connection.source === connection.target) return false
      if (connection.sourceHandle) {
        const existing = (edges as Edge[]).find(
          (e) => e.source === connection.source && e.sourceHandle === connection.sourceHandle
        )
        if (existing) return false
      }
      const sourceNode = nodes.find((n) => n.id === connection.source)
      const targetNode = nodes.find((n) => n.id === connection.target)
      if (sourceNode?.type === 'output') return false
      if (targetNode?.type === 'trigger') return false
      return true
    },
    [edges, nodes]
  )

  const addNodeAt = useCallback(
    (type: string, position: { x: number; y: number }) => {
      const newNode: Node = {
        id: nextNodeId(type),
        type,
        position,
        data: { label: type },
      }
      setNodes((nds) => [...nds, newNode])
      setSelectedNode(null)
      setDirty(true)
    },
    [setNodes, nextNodeId]
  )

  const handleAddNode = useCallback(
    (type: string) => {
      const vp = getViewport()
      addNodeAt(type, { x: 100 - vp.x / vp.zoom, y: 100 - vp.y / vp.zoom })
      setMobileSidebarOpen(false)
    },
    [addNodeAt, getViewport]
  )

  const onDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }, [])

  const onDrop = useCallback(
    (e: DragEvent<HTMLDivElement>) => {
      e.preventDefault()
      const type = e.dataTransfer.getData(DRAG_MIME)
      if (!type || !(type in nodeTypes)) return
      addNodeAt(type, screenToFlowPosition({ x: e.clientX, y: e.clientY }))
    },
    [addNodeAt, screenToFlowPosition]
  )

  useEffect(() => {
    if (!firstNodeId) return
    const id = requestAnimationFrame(() => {
      fitView({ duration: 0, padding: 0.2, nodes: [{ id: firstNodeId }] })
    })
    return () => cancelAnimationFrame(id)
  }, [firstNodeId, fitView])

  const handleReset = useCallback(() => {
    setNodes([])
    setEdges([])
    setSelectedNode(null)
    setWorkflowId(null)
    setCurrentWorkflow(null)
    setDirty(false)
    idCounterRef.current = 0
    onWorkflowChange?.(null)
  }, [setNodes, setEdges, onWorkflowChange])

  const handleConfigUpdate = useCallback((updatedNode: Node) => {
    setNodes((nds) => nds.map((n) => (n.id === updatedNode.id ? updatedNode : n)))
    setSelectedNode(updatedNode)
    setDirty(true)
  }, [setNodes])

  const loadWorkflow = useCallback(
    (workflow: Workflow) => {
      if (dirty && !window.confirm('当前工作流有未保存修改，确认切换吗？')) {
        return
      }
      const graph = workflow.flow_json
      suppressDirtyRef.current = true
      setNodes((graph.nodes || []) as Node[])
      setEdges((graph.edges || []) as Edge[])
      setSelectedNode(null)
      setWorkflowId(workflow.id)
      setCurrentWorkflow(workflow)
      setDirty(false)
      onWorkflowChange?.(workflow.id)
      requestAnimationFrame(() => {
        fitView({ duration: 200, padding: 0.2 })
        suppressDirtyRef.current = false
      })
    },
    [dirty, fitView, onWorkflowChange, setEdges, setNodes]
  )

  const buildFlowGraph = useCallback((): FlowGraph => {
    const flowNodes = nodes.map((n) => ({
      id: n.id,
      type: (n.type || 'trigger') as 'trigger' | 'llm' | 'condition' | 'transform' | 'merge' | 'output',
      position: n.position,
      data: n.data as Record<string, unknown>,
    }))
    const flowEdges = edges.map((e) => ({
      id: e.id,
      source: e.source,
      target: e.target,
      sourceHandle: e.sourceHandle || undefined,
      targetHandle: e.targetHandle || undefined,
    }))
    return { nodes: flowNodes, edges: flowEdges, viewport: getViewport() }
  }, [nodes, edges, getViewport])

  const saveWorkflow = useCallback(async () => {
    if (!tenantId) {
      showError('请先选择租户')
      return null
    }
    setSaving(true)
    try {
      const flowGraph = buildFlowGraph()
      if (workflowId) {
        const result = await updateWorkflow(workflowId, {
          tenant_id: tenantId,
          name: currentWorkflow?.name || `工作流-${workflowId.slice(0, 8)}`,
          description: currentWorkflow?.description,
          flow_json: flowGraph,
        })
        setCurrentWorkflow(result)
        setDirty(false)
        showSuccess('保存成功')
        return result
      }

      const result = await createWorkflow({
        tenant_id: tenantId,
        name: `工作流-${Date.now()}`,
        flow_json: flowGraph,
      })
      const newId = result.id
      if (newId) {
        setWorkflowId(newId)
        setCurrentWorkflow(result)
        setDirty(false)
        onWorkflowChange?.(newId)
        setSelectorRefreshKey((key) => key + 1)
      }
      showSuccess('创建成功')
      return result
    } catch {
      showError('保存失败')
      return null
    } finally {
      setSaving(false)
    }
  }, [tenantId, workflowId, buildFlowGraph, currentWorkflow, onWorkflowChange])

  const handleSave = useCallback(async () => {
    await saveWorkflow()
  }, [saveWorkflow])

  const handleValidate = useCallback(async () => {
    if (!tenantId || !workflowId) {
      showError('请先保存工作流')
      return
    }
    setValidating(true)
    try {
      const result = await validateWorkflow(workflowId, tenantId)
      if (result.valid) {
        showSuccess('校验通过')
      } else {
        const errors = result.errors || []
        showError(`校验失败: ${errors.map((e: { message: string }) => e.message).join('; ') || '未知错误'}`)
      }
    } catch {
      showError('校验请求失败')
    } finally {
      setValidating(false)
    }
  }, [tenantId, workflowId])

  const handleRun = useCallback(async () => {
    if (!tenantId) {
      showError('请先选择租户')
      return
    }
    let targetWorkflowId = workflowId
    if (dirty || !targetWorkflowId) {
      const saved = await saveWorkflow()
      targetWorkflowId = saved?.id || null
    }
    if (!targetWorkflowId) {
      showError('请先保存工作流')
      return
    }
    setRunning(true)
    try {
      await executeWorkflow(targetWorkflowId, { tenant_id: tenantId }, 'async')
      showSuccess('已提交执行')
    } catch {
      showError('执行失败')
    } finally {
      setRunning(false)
    }
  }, [tenantId, workflowId, dirty, saveWorkflow])

  const onEdgesChangeSafe = useCallback(
    (changes: Parameters<typeof onEdgesChange>[0]) => {
      onEdgesChange(changes)
      if (!suppressDirtyRef.current && changes.some((change) => change.type !== 'select')) {
        setDirty(true)
      }
    },
    [onEdgesChange]
  )

  const renderCanvas = (
    <div className="flex-1 relative" onDragOver={onDragOver} onDrop={onDrop}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChangeSafe}
        onConnect={onConnect}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        isValidConnection={isValidConnection}
        nodeTypes={nodeTypes}
      >
        <Background />
        <Controls />
        <MiniMap
          nodeColor={getNodeColor}
          nodeStrokeColor={getNodeStrokeColor}
          nodeBorderRadius={4}
          nodeStrokeWidth={2}
          bgColor="#f8fafc"
          maskColor="rgba(15, 23, 42, 0.55)"
          maskStrokeColor="#0f172a"
          maskStrokeWidth={1.5}
          pannable
          zoomable
          position="bottom-right"
          ariaLabel="工作流画布缩略图"
        />
      </ReactFlow>
      <WorkflowToolbar
        nodes={nodes}
        edges={edges}
        setNodes={setNodes}
        setEdges={setEdges}
        onReset={handleReset}
        onSave={handleSave}
        onValidate={handleValidate}
        onRun={handleRun}
	        saving={saving}
	        validating={validating}
	        running={running}
	        dirty={dirty}
	      />
    </div>
  )

  return (
    <>
	      <WorkflowSelector
	        tenantId={tenantId}
	        selectedWorkflowId={workflowId}
	        refreshKey={selectorRefreshKey}
	        onSelect={loadWorkflow}
	        onNew={handleReset}
	      />

	      {/* Desktop: three-panel horizontal layout */}
      <div className="hidden lg:flex h-[calc(100vh-200px)]">
        <WorkflowSidebar onAddNode={handleAddNode} />
        <div className="flex-1 flex">
          {renderCanvas}
          <WorkflowConfigPanel selectedNode={selectedNode} onNodeUpdate={handleConfigUpdate} />
        </div>
      </div>

      {/* Mobile: vertical layout with collapsible node bar + bottom config drawer */}
      <div className="flex flex-col h-[calc(100vh-200px)] lg:hidden">
        {/* Mobile node bar */}
        <div className="border-b bg-slate-50">
          <button
            onClick={() => setMobileSidebarOpen(!mobileSidebarOpen)}
            className="w-full flex items-center justify-between px-4 py-2.5 text-sm font-medium text-slate-700 hover:bg-slate-100 transition-colors"
          >
            <span className="flex items-center gap-2">
              <Plus className="h-4 w-4" />
              添加节点
            </span>
            {mobileSidebarOpen ? (
              <ChevronUp className="h-4 w-4 text-slate-400" />
            ) : (
              <ChevronDown className="h-4 w-4 text-slate-400" />
            )}
          </button>
          {mobileSidebarOpen && (
            <div className="flex gap-2 px-4 pb-3 overflow-x-auto">
              {sidebarItems.map((item) => (
                <button
                  key={item.type}
                  onClick={() => handleAddNode(item.type)}
                  className="flex items-center gap-1.5 px-3 py-2 rounded-md border bg-white cursor-pointer hover:shadow-sm text-sm whitespace-nowrap shrink-0"
                >
                  <span className={item.color}>{item.icon}</span>
                  <span>{item.label}</span>
                </button>
              ))}
            </div>
          )}
        </div>

        {renderCanvas}

        {/* Mobile config drawer */}
        <WorkflowConfigPanel
          selectedNode={selectedNode}
          onNodeUpdate={handleConfigUpdate}
          mobile
          onClose={() => setSelectedNode(null)}
        />
      </div>
    </>
  )
}
