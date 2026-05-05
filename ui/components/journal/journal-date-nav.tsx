'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import {
  format, startOfMonth, endOfMonth, startOfWeek, endOfWeek,
  addDays, isSameDay, isToday, isAfter, isBefore, addMonths, subMonths,
  addYears, subYears, setYear,
} from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { ChevronLeft, ChevronRight, CalendarDays, ChevronUp, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface JournalDateNavProps {
  currentDate: Date
  onDateChange: (date: Date) => void
}

const WEEKDAY_LABELS = ['一', '二', '三', '四', '五', '六', '日']

export function JournalDateNav({
  currentDate,
  onDateChange,
}: JournalDateNavProps) {
  const [calendarOpen, setCalendarOpen] = useState(false)
  const [viewMonth, setViewMonth] = useState(() => startOfMonth(currentDate))
  const [yearInputOpen, setYearInputOpen] = useState(false)
  const [yearInputValue, setYearInputValue] = useState('')
  const containerRef = useRef<HTMLDivElement>(null)
  const yearInputRef = useRef<HTMLInputElement>(null)

  const weekday = format(currentDate, 'EEEE', { locale: zhCN })
  const isTodayDate = isToday(currentDate)
  const today = new Date()
  const currentYear = today.getFullYear()

  useEffect(() => {
    setViewMonth(startOfMonth(currentDate))
  }, [currentDate])

  const handleClickOutside = useCallback((e: MouseEvent) => {
    if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
      setCalendarOpen(false)
      setYearInputOpen(false)
    }
  }, [])

  useEffect(() => {
    if (calendarOpen) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [calendarOpen, handleClickOutside])

  useEffect(() => {
    if (yearInputOpen && yearInputRef.current) {
      yearInputRef.current.focus()
      yearInputRef.current.select()
    }
  }, [yearInputOpen])

  const handleDateSelect = (date: Date) => {
    if (isAfter(date, today)) return
    onDateChange(date)
    setCalendarOpen(false)
    setYearInputOpen(false)
  }

  const handleYearSubmit = () => {
    const year = parseInt(yearInputValue, 10)
    if (isNaN(year) || year < 1900 || year > currentYear) return
    const newDate = setYear(startOfMonth(viewMonth), year)
    if (!isAfter(newDate, startOfMonth(today))) {
      setViewMonth(newDate)
    }
    setYearInputOpen(false)
  }

  const handleYearInputChange = (value: string) => {
    setYearInputValue(value)
  }

  const calendarDays = buildCalendarDays(viewMonth)

  return (
    <div className="flex items-center justify-between mb-8">
      <div className="relative flex items-center gap-3" ref={containerRef}>
        {/* Calendar icon block */}
        <div className="relative w-14 h-14 rounded-xl bg-gradient-to-br from-primary to-accent shadow-lg shadow-primary/20 flex flex-col items-center justify-center text-white overflow-hidden cursor-pointer hover:shadow-xl hover:-translate-y-0.5 transition-all"
          onClick={() => { setCalendarOpen(!calendarOpen); setYearInputOpen(false) }}
        >
          <span className="text-[10px] font-bold uppercase tracking-wider leading-none">
            {format(currentDate, 'M')}月
          </span>
          <span className="text-xl font-bold leading-none mt-0.5">
            {format(currentDate, 'd')}
          </span>
        </div>

        {/* Clickable date text */}
        <div
          className="flex flex-col cursor-pointer group"
          onClick={() => { setCalendarOpen(!calendarOpen); setYearInputOpen(false) }}
        >
          <div className="flex items-center gap-2">
            <span className="text-xl font-bold text-slate-800 group-hover:text-primary transition-colors">
              {format(currentDate, 'yyyy年M月d日')}
            </span>
            {isTodayDate && (
              <span className="px-2 py-0.5 rounded-full bg-primary/10 text-primary text-xs font-semibold">
                今天
              </span>
            )}
          </div>
          <span className="text-sm text-slate-400 group-hover:text-primary/60 transition-colors">
            {weekday}
          </span>
        </div>

        {/* Navigation arrows */}
        <div className="flex items-center gap-1 ml-2">
          <button
            type="button"
            className="h-8 w-8 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 hover:border-slate-300 flex items-center justify-center transition-colors"
            onClick={() => onDateChange(addDays(currentDate, -1))}
          >
            <ChevronLeft className="h-4 w-4 text-slate-500" />
          </button>
          <button
            type="button"
            className="h-8 w-8 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 hover:border-slate-300 flex items-center justify-center transition-colors"
            onClick={() => onDateChange(addDays(currentDate, 1))}
          >
            <ChevronRight className="h-4 w-4 text-slate-500" />
          </button>
        </div>

        {!isTodayDate && (
          <Button
            variant="ghost"
            size="sm"
            className="text-primary hover:text-primary hover:bg-primary/5 h-8"
            onClick={() => onDateChange(new Date())}
          >
            <CalendarDays className="h-3.5 w-3.5 mr-1" />
            回到今天
          </Button>
        )}

        {/* Calendar dropdown */}
        {calendarOpen && (
          <div className="absolute top-full left-0 mt-2 z-50 bg-white rounded-xl shadow-xl border border-slate-200/80 p-4 animate-slide-down min-w-[300px]">
            {/* Year & Month navigation */}
            <div className="flex items-center justify-between mb-3">
              {/* Year navigation with click-to-edit */}
              <div className="flex items-center gap-0.5">
                <button
                  type="button"
                  className="h-7 w-7 rounded-md border border-slate-200 hover:bg-slate-50 flex items-center justify-center transition-colors"
                  onClick={() => setViewMonth(subYears(viewMonth, 1))}
                >
                  <ChevronUp className="h-3 w-3 text-slate-500" />
                </button>
                {yearInputOpen ? (
                  <input
                    ref={yearInputRef}
                    type="text"
                    value={yearInputValue}
                    onChange={e => handleYearInputChange(e.target.value)}
                    onKeyDown={e => {
                      if (e.key === 'Enter') handleYearSubmit()
                      if (e.key === 'Escape') setYearInputOpen(false)
                    }}
                    onBlur={handleYearSubmit}
                    className="w-14 h-7 text-sm font-semibold text-slate-700 text-center border border-primary rounded-md focus:outline-none focus:ring-2 focus:ring-primary/20"
                    maxLength={4}
                  />
                ) : (
                  <button
                    type="button"
                    className="h-7 px-1.5 rounded-md hover:bg-slate-50 text-sm font-semibold text-slate-700 transition-colors"
                    onClick={() => { setYearInputOpen(true); setYearInputValue(format(viewMonth, 'yyyy')) }}
                    title="点击输入年份"
                  >
                    {format(viewMonth, 'yyyy')}年
                  </button>
                )}
                <button
                  type="button"
                  className="h-7 w-7 rounded-md border border-slate-200 hover:bg-slate-50 flex items-center justify-center transition-colors"
                  onClick={() => {
                    if (!isAfter(addYears(viewMonth, 1), startOfMonth(today))) {
                      setViewMonth(addYears(viewMonth, 1))
                    }
                  }}
                  disabled={viewMonth.getFullYear() >= currentYear}
                >
                  <ChevronDown className="h-3 w-3 text-slate-500" />
                </button>
              </div>

              {/* Month navigation */}
              <div className="flex items-center gap-0.5">
                <button
                  type="button"
                  className="h-7 w-7 rounded-md border border-slate-200 hover:bg-slate-50 flex items-center justify-center transition-colors"
                  onClick={() => setViewMonth(subMonths(viewMonth, 1))}
                >
                  <ChevronLeft className="h-3.5 w-3.5 text-slate-500" />
                </button>
                <span className="text-sm font-semibold text-slate-700 min-w-[56px] text-center">
                  {format(viewMonth, 'M月')}
                </span>
                <button
                  type="button"
                  className="h-7 w-7 rounded-md border border-slate-200 hover:bg-slate-50 flex items-center justify-center transition-colors"
                  onClick={() => setViewMonth(addMonths(viewMonth, 1))}
                  disabled={isAfter(addMonths(viewMonth, 1), addMonths(startOfMonth(today), 1))}
                >
                  <ChevronRight className="h-3.5 w-3.5 text-slate-500" />
                </button>
              </div>
            </div>

            {/* Weekday headers */}
            <div className="grid grid-cols-7 gap-0 mb-1">
              {WEEKDAY_LABELS.map(label => (
                <div key={label} className="h-8 flex items-center justify-center text-xs font-medium text-slate-400">
                  {label}
                </div>
              ))}
            </div>

            {/* Day grid */}
            <div className="grid grid-cols-7 gap-0">
              {calendarDays.map((day, idx) => {
                const isSelected = isSameDay(day, currentDate)
                const isCurrentDay = isToday(day)
                const isFuture = isAfter(day, today)
                const isCurrentMonth = day.getMonth() === viewMonth.getMonth()

                return (
                  <button
                    key={idx}
                    type="button"
                    disabled={isFuture || !isCurrentMonth}
                    className={`
                      h-8 w-full rounded-lg text-sm font-medium transition-all duration-150 relative
                      ${isFuture
                        ? 'text-slate-200 cursor-not-allowed'
                        : isSelected
                          ? 'bg-gradient-to-r from-primary to-accent text-white shadow-sm'
                          : isCurrentDay
                            ? 'text-primary font-bold hover:bg-primary/10'
                            : isCurrentMonth
                              ? 'text-slate-700 hover:bg-slate-100'
                              : 'text-slate-300'
                      }
                    `}
                    onClick={() => handleDateSelect(day)}
                  >
                    {isCurrentDay && !isSelected && (
                      <span className="absolute bottom-0.5 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-primary" />
                    )}
                    {format(day, 'd')}
                  </button>
                )
              })}
            </div>

            {/* Today shortcut */}
            {!isToday(currentDate) && (
              <div className="mt-2 pt-2 border-t border-slate-100">
                <button
                  type="button"
                  className="w-full text-center text-xs text-primary hover:text-primary/80 font-medium py-1.5 rounded-lg hover:bg-primary/5 transition-colors"
                  onClick={() => handleDateSelect(today)}
                >
                  回到今天
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function buildCalendarDays(monthStart: Date): Date[] {
  const start = startOfWeek(monthStart, { weekStartsOn: 1 })
  const end = endOfWeek(endOfMonth(monthStart), { weekStartsOn: 1 })
  const days: Date[] = []
  let day = start
  while (isBefore(day, end) || isSameDay(day, end)) {
    days.push(day)
    day = addDays(day, 1)
  }
  return days
}