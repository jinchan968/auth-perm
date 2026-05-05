'use client'

import { BookOpen, PenLine } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface JournalEmptyStateProps {
  onCreate: () => void
}

export function JournalEmptyState({ onCreate }: JournalEmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
      <div className="relative mb-6">
        <div className="w-24 h-24 rounded-2xl bg-gradient-to-br from-primary/10 to-accent/10 flex items-center justify-center">
          <BookOpen className="h-10 w-10 text-primary/60" />
        </div>
        <div className="absolute -bottom-1 -right-1 w-8 h-8 rounded-lg bg-gradient-to-r from-primary to-accent flex items-center justify-center shadow-lg shadow-primary/20">
          <PenLine className="h-4 w-4 text-white" />
        </div>
      </div>
      <h3 className="text-lg font-semibold text-slate-700 mb-1">
        这一天还没有札记
      </h3>
      <p className="text-sm text-slate-400 mb-5">
        记录此刻的想法、见闻与感悟
      </p>
      <Button
        onClick={onCreate}
        className="bg-gradient-to-r from-primary to-accent text-white shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30 hover:-translate-y-0.5 transition-all"
      >
        <PenLine className="h-4 w-4 mr-1.5" />
        写一篇札记
      </Button>
    </div>
  )
}