'use client'

import { Breadcrumb } from '@/components/ui/breadcrumb'
import { ShellLayout } from '@/components/layout/shell-layout'
import { PermGuard } from '@/components/ui/perm-guard'
import { NovelDetailWorkspace } from '@/components/novel/novel-detail-workspace'

export default function NovelDetailPage({ params }: { params: { id: string } }) {
  return (
    <ShellLayout pathname="/novels">
      <Breadcrumb
        items={[
          { label: '首页', href: '/home' },
          { label: '小说管理', href: '/novels' },
          { label: '明细', href: `/novels/${params.id}` },
        ]}
      />
      <div className="mt-4">
        <PermGuard menu="novel">
          <NovelDetailWorkspace novelId={params.id} />
        </PermGuard>
      </div>
    </ShellLayout>
  )
}
