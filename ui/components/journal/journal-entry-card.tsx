'use client'

import {
  Clock, MapPin, MessageSquarePlus, Pencil, Trash2, X,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { Entry, Tag } from '@/types/journal'
import { PERIODS, WEATHERS } from './constants'

interface JournalEntryCardProps {
  entry: Entry
  onCorrection: (entryId: string) => void
  onEditTags: (entry: Entry) => void
  onDelete: (entryId: string) => void
}

export function JournalEntryCard({
  entry,
  onCorrection,
  onEditTags,
  onDelete,
}: JournalEntryCardProps) {
  const periodInfo = PERIODS.find(p => p.value === entry.period)
  const weatherInfo = WEATHERS.find(w => w.value === entry.weather)
  const borderColor = periodInfo?.color ?? '#6366f1'

  return (
    <div className="group relative animate-slide-up">
      <div
        className="
          relative rounded-xl overflow-hidden transition-all duration-300
          shadow-sm group-hover:shadow-lg group-hover:shadow-slate-200/60 group-hover:-translate-y-0.5
          bg-white/90 backdrop-blur-sm border border-slate-200/60
        "
      >
        {/* Left accent bar */}
        <div
          className="absolute left-0 top-0 bottom-0 w-1 rounded-l-xl"
          style={{ backgroundColor: borderColor }}
        />

        <div className="pl-5 pr-4 py-5">
          {/* Top meta row */}
          <div className="flex items-start justify-between mb-3">
            <div className="flex items-center gap-2 flex-wrap">
              {/* Period badge */}
              <span
                className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold border ${periodInfo?.bgClass ?? 'bg-slate-100 text-slate-600'}`}
              >
                <Clock className="h-3 w-3" />
                {entry.period}
              </span>

              {/* Weather badge */}
              {entry.weather && weatherInfo && (
                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-slate-50 text-slate-600 border border-slate-100">
                  {weatherInfo.icon}
                  {entry.weather}
                </span>
              )}

              {/* Location */}
              {entry.location && (
                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-slate-50 text-slate-600 border border-slate-100">
                  <MapPin className="h-3 w-3" />
                  {entry.location}
                </span>
              )}
            </div>

            {/* Action buttons - always visible on mobile, hover on desktop */}
            <div className="flex items-center gap-0.5 transition-opacity duration-200 opacity-100 sm:opacity-0 sm:group-hover:opacity-100">
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-slate-400 hover:text-primary hover:bg-primary/5"
                title="追加修正"
                onClick={() => onCorrection(entry.id)}
              >
                <MessageSquarePlus className="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-slate-400 hover:text-primary hover:bg-primary/5"
                title="编辑标签"
                onClick={() => onEditTags(entry)}
              >
                <Pencil className="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-slate-400 hover:text-red-500 hover:bg-red-50"
                title="删除"
                onClick={() => onDelete(entry.id)}
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          {/* Title */}
          {entry.title && (
            <h3 className="text-lg font-semibold text-slate-800 mb-2 leading-snug">
              {entry.title}
            </h3>
          )}

          {/* Content */}
          <div className="text-slate-600 text-[15px] leading-relaxed whitespace-pre-wrap">
            {entry.content}
          </div>

          {/* Char count */}
          {entry.content.length > 100 && (
            <div className="text-xs text-slate-300 mt-2 text-right">
              {Array.from(entry.content).length}字
            </div>
          )}

          {/* Corrections timeline */}
          {entry.corrections && entry.corrections.length > 0 && (
            <div className="mt-4 pt-3 border-t border-slate-100">
              <p className="text-xs font-semibold text-slate-400 mb-2 uppercase tracking-wider">
                修正记录
              </p>
              <div className="space-y-2">
                {entry.corrections.map((c, idx) => (
                  <div
                    key={c.id}
                    className="relative pl-4 before:absolute before:left-0 before:top-2 before:w-1.5 before:h-1.5 before:rounded-full before:bg-primary/50 before:ring-2 before:ring-primary/20"
                    style={{ animationDelay: `${idx * 50}ms` }}
                  >
                    <div className="bg-gradient-to-r from-primary/5 to-transparent rounded-lg px-3 py-2.5">
                      <div className="text-sm text-slate-600 whitespace-pre-wrap leading-relaxed">
                        {c.content}
                      </div>
                      <div className="text-[11px] text-slate-400 mt-1.5">
                        {new Date(c.created_at).toLocaleString('zh-CN', {
                          month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit',
                        })}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Tags */}
          {entry.tags && entry.tags.length > 0 && (
            <div className="flex items-center gap-1.5 mt-3.5 flex-wrap">
              {entry.tags.map(tag => (
                <TagPill key={tag.id} tag={tag} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function TagPill({ tag }: { tag: Tag }) {
  return (
    <span
      className="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium border transition-colors duration-150 hover:brightness-95"
      style={{
        backgroundColor: tag.color + '15',
        color: tag.color,
        borderColor: tag.color + '30',
      }}
    >
      {tag.name}
    </span>
  )
}

export function FormTagPill({
  tag,
  selected,
  onToggle,
}: {
  tag: Tag
  selected: boolean
  onToggle: () => void
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className={`
        inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium border
        transition-all duration-200 cursor-pointer
        ${selected
          ? 'text-white shadow-sm hover:shadow-md'
          : 'border-transparent hover:border-slate-300'
        }
      `}
      style={selected
        ? { backgroundColor: tag.color, borderColor: tag.color, boxShadow: `0 2px 8px ${tag.color}40` }
        : { backgroundColor: tag.color + '10', color: tag.color }
      }
    >
      {tag.name}
      {selected && <X className="h-3 w-3" />}
    </button>
  )
}