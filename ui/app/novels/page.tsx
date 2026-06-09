'use client'

import { Breadcrumb } from '@/components/ui/breadcrumb'
import { ShellLayout } from '@/components/layout/shell-layout'
import { PermGuard } from '@/components/ui/perm-guard'
import { NovelListWorkspace } from '@/components/novel/novel-list-workspace'

export default function NovelListPage() {
  return (
    <ShellLayout pathname="/novels">
      <Breadcrumb
        items={[
          { label: '首页', href: '/home' },
          { label: '小说管理', href: '/novels' },
        ]}
      />
      <div className="mt-4">
        <PermGuard menu="novel">
          <NovelListWorkspace />
        </PermGuard>
      </div>
    </ShellLayout>
  )
}
