'use client'

import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ShellLayout } from '@/components/layout/shell-layout'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import WorkflowDesigner from '@/components/workflow/workflow-designer'
import WorkflowRuns from '@/components/workflow/workflow-runs'

export default function WorkflowPage() {
  const [activeTab, setActiveTab] = useState('designer')

  return (
    <ShellLayout pathname="/workflow">
      <Breadcrumb
        items={[
          { label: '首页', href: '/home' },
          { label: '工作流', href: '/workflow' },
        ]}
      />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="designer">编排设计</TabsTrigger>
          <TabsTrigger value="runs">运行历史</TabsTrigger>
        </TabsList>
        <TabsContent value="designer">
          <WorkflowDesigner />
        </TabsContent>
        <TabsContent value="runs">
          <WorkflowRuns />
        </TabsContent>
      </Tabs>
    </ShellLayout>
  )
}
