'use client'

import { useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { ShellLayout } from '@/components/layout/shell-layout'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { PermGuard } from '@/components/ui/perm-guard'
import { usePermissions } from '@/hooks/use-permissions'
import RecognizePanel from '@/components/multimodal/recognize-panel'
import GeneratePanel from '@/components/multimodal/generate-panel'
import ImageGeneratePanel from '@/components/multimodal/image-generate-panel'

type TabType = 'recognize' | 'prompt' | 'image'

const tabPermissions: Record<TabType, string> = {
  recognize: 'multimodal.tab.recognize',
  prompt: 'multimodal.tab.generate',
  image: 'multimodal.tab.image_generate',
}

export default function MultimodalPage() {
  const [activeTab, setActiveTab] = useState<TabType>('recognize')
  const { hasButton, isSuperAdmin, loading } = usePermissions()

  const visibleTabs = useMemo<TabType[]>(() => {
    const tabs: TabType[] = ['recognize', 'prompt', 'image']
    if (isSuperAdmin) return tabs
    return tabs.filter((tab) => hasButton(tabPermissions[tab]))
  }, [hasButton, isSuperAdmin])

  useEffect(() => {
    if (loading || visibleTabs.length === 0 || visibleTabs.includes(activeTab)) return
    setActiveTab(visibleTabs[0])
  }, [activeTab, loading, visibleTabs])

  const currentTab = visibleTabs.includes(activeTab) ? activeTab : visibleTabs[0]

  return (
    <ShellLayout pathname="/multimodal">
      <Breadcrumb
        items={[
          { label: '首页', href: '/home' },
          { label: '多模态', href: '/multimodal' },
        ]}
      />

      <div className="flex flex-wrap gap-2 mb-6 mt-4">
        <PermGuard button="multimodal.tab.recognize">
          <Button
            variant={activeTab === 'recognize' ? 'default' : 'outline'}
            onClick={() => setActiveTab('recognize')}
          >
            识图
          </Button>
        </PermGuard>
        <PermGuard button="multimodal.tab.generate">
          <Button
            variant={activeTab === 'prompt' ? 'default' : 'outline'}
            onClick={() => setActiveTab('prompt')}
          >
            生成提示词
          </Button>
        </PermGuard>
        <PermGuard button="multimodal.tab.image_generate">
          <Button
            variant={activeTab === 'image' ? 'default' : 'outline'}
            onClick={() => setActiveTab('image')}
          >
            生成图片
          </Button>
        </PermGuard>
      </div>

      {visibleTabs.length === 0 && !loading ? null : currentTab === 'recognize' ? (
        <RecognizePanel />
      ) : currentTab === 'prompt' ? (
        <GeneratePanel />
      ) : (
        <ImageGeneratePanel />
      )}
    </ShellLayout>
  )
}
