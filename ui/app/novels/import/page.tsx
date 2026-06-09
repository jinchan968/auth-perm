'use client'

import { Breadcrumb } from '@/components/ui/breadcrumb'
import { ShellLayout } from '@/components/layout/shell-layout'
import { PermGuard } from '@/components/ui/perm-guard'
import { NovelImportWorkspace } from '@/components/novel/novel-import-workspace'

export default function NovelImportPage() {
  return (
    <ShellLayout pathname="/novels/import">
      <Breadcrumb
        items={[
          { label: '首页', href: '/home' },
          { label: '小说管理', href: '/novels' },
          { label: '导入', href: '/novels/import' },
        ]}
      />
      <div className="mt-4">
        <PermGuard button="novel.import">
          <NovelImportWorkspace />
        </PermGuard>
      </div>
    </ShellLayout>
  )
}
