'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ShellLayout } from '@/components/layout/shell-layout'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { PermGuard } from '@/components/ui/perm-guard'
import RecognizePanel from '@/components/multimodal/recognize-panel'
import GeneratePanel from '@/components/multimodal/generate-panel'
import ImageGeneratePanel from '@/components/multimodal/image-generate-panel'

type TabType = 'recognize' | 'prompt' | 'image'

export default function MultimodalPage() {
  const [activeTab, setActiveTab] = useState<TabType>('recognize')

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
        <PermGuard button="multimodal.tab.generate">
          <Button
            variant={activeTab === 'image' ? 'default' : 'outline'}
            onClick={() => setActiveTab('image')}
          >
            生成图片
          </Button>
        </PermGuard>
      </div>

      {activeTab === 'recognize' ? (
        <RecognizePanel />
      ) : activeTab === 'prompt' ? (
        <GeneratePanel />
      ) : (
        <ImageGeneratePanel />
      )}
    </ShellLayout>
  )
}
