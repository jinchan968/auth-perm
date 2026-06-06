'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ShellLayout } from '@/components/layout/shell-layout'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { PermGuard } from '@/components/ui/perm-guard'
import { usePermissions } from '@/hooks/use-permissions'
import WorkflowDesigner from '@/components/workflow/workflow-designer'
import WorkflowRuns from '@/components/workflow/workflow-runs'

type TabType = 'designer' | 'runs'

export default function WorkflowPage() {
  const [activeTab, setActiveTab] = useState<TabType>('designer')
  const [selectedWorkflowId, setSelectedWorkflowId] = useState<string | null>(null)
  const { hasButton, isSuperAdmin, loading } = usePermissions()
  const canUseDesigner = isSuperAdmin || hasButton('workflow.tab.designer')
  const canUseRuns = isSuperAdmin || hasButton('workflow.tab.runs')
  const visibleTab: TabType | null = activeTab === 'designer'
    ? (canUseDesigner ? 'designer' : canUseRuns ? 'runs' : null)
    : (canUseRuns ? 'runs' : canUseDesigner ? 'designer' : null)

  return (
    <ShellLayout pathname="/workflow">
      <Breadcrumb
        items={[
          { label: '首页', href: '/home' },
          { label: '工作流', href: '/workflow' },
        ]}
      />

      <div className="flex flex-wrap gap-2 mb-6 mt-4">
        <PermGuard button="workflow.tab.designer">
          <Button
            variant={activeTab === 'designer' ? 'default' : 'outline'}
            onClick={() => setActiveTab('designer')}
          >
            编排设计
          </Button>
        </PermGuard>
        <PermGuard button="workflow.tab.runs">
          <Button
            variant={activeTab === 'runs' ? 'default' : 'outline'}
            onClick={() => setActiveTab('runs')}
          >
            运行历史
          </Button>
        </PermGuard>
      </div>

      {loading ? null : visibleTab === 'designer' ? (
        <PermGuard button="workflow.tab.designer">
          <WorkflowDesigner onWorkflowChange={setSelectedWorkflowId} />
        </PermGuard>
      ) : visibleTab === 'runs' ? (
        <PermGuard button="workflow.tab.runs">
          <WorkflowRuns workflowId={selectedWorkflowId ?? undefined} />
        </PermGuard>
      ) : (
        <div className="py-12 text-center text-sm text-slate-400">暂无工作流页面权限</div>
      )}
    </ShellLayout>
  )
}
