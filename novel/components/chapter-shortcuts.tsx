"use client"

import { useEffect } from "react"

type ChapterShortcutsProps = {
  catalogHref: string
  previousHref?: string
  nextHref?: string
}

function isEditingTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) {
    return false
  }
  const tagName = target.tagName.toLowerCase()
  return target.isContentEditable || tagName === "input" || tagName === "textarea" || tagName === "select"
}

export function ChapterShortcuts({ catalogHref, previousHref, nextHref }: ChapterShortcutsProps) {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.metaKey || event.ctrlKey || event.altKey || event.shiftKey || isEditingTarget(event.target)) {
        return
      }

      if (event.key === "ArrowLeft" && previousHref) {
        window.location.href = previousHref
      }
      if (event.key === "ArrowRight" && nextHref) {
        window.location.href = nextHref
      }
      if (event.key.toLowerCase() === "m") {
        window.location.href = catalogHref
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [catalogHref, nextHref, previousHref])

  return null
}
