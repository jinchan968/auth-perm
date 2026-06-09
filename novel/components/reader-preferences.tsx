"use client"

import { Minus, Plus } from "lucide-react"
import type { CSSProperties, ReactNode } from "react"
import { useMemo, useState } from "react"

type ReaderPreferencesProps = {
  children: ReactNode
}

export function ReaderPreferences({ children }: ReaderPreferencesProps) {
  const [readerScale, setReaderScale] = useState(1)
  const readerSize = useMemo(
    () =>
      `clamp(${(1.04 * readerScale).toFixed(2)}rem, ${(2.8 * readerScale).toFixed(2)}vw, ${(1.14 * readerScale).toFixed(2)}rem)`,
    [readerScale],
  )

  function decreaseSize() {
    setReaderScale((current) => Math.max(0.92, Number((current - 0.06).toFixed(2))))
  }

  function increaseSize() {
    setReaderScale((current) => Math.min(1.24, Number((current + 0.06).toFixed(2))))
  }

  return (
    <div style={{ "--reader-size": readerSize } as CSSProperties}>
      <div className="fixed bottom-[calc(0.9rem+env(safe-area-inset-bottom))] left-1/2 z-20 flex -translate-x-1/2 items-center gap-2 border border-line bg-paper/95 p-1 shadow-reader backdrop-blur md:bottom-5 md:left-auto md:right-5 md:translate-x-0 md:flex-col">
        <button
          type="button"
          className="icon-button"
          onClick={increaseSize}
          aria-label="放大字号"
          title="放大字号"
        >
          <Plus size={15} />
        </button>
        <button
          type="button"
          className="icon-button"
          onClick={decreaseSize}
          aria-label="缩小字号"
          title="缩小字号"
        >
          <Minus size={15} />
        </button>
      </div>
      {children}
    </div>
  )
}
