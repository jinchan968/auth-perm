import { BookOpen, Network, PenLine } from "lucide-react"
import Link from "next/link"
import { ThemeToggle } from "@/components/theme-toggle"

type SiteHeaderProps = {
  title?: string
  volume?: string
  showCatalog?: boolean
  showWorkspaceLinks?: boolean
  homeHref?: string
  catalogHref?: string
}

export function SiteHeader({
  title = "小说",
  volume = "Library",
  showCatalog = false,
  showWorkspaceLinks = false,
  homeHref = "/",
  catalogHref = "/",
}: SiteHeaderProps) {
  return (
    <header className="top-rule sticky top-0 z-20">
      <nav className="mx-auto flex min-h-16 max-w-7xl items-center justify-between gap-3 px-4 py-2 md:px-8">
        <Link
          href={homeHref}
          className="flex min-w-0 items-baseline gap-2 font-display text-xl font-bold sm:text-2xl"
        >
          <span className="truncate">{title}</span>
          <span className="micro-type hidden shrink-0 text-dim sm:inline">- {volume}</span>
        </Link>

        <div className="flex shrink-0 items-center gap-2 sm:gap-3">
          {showWorkspaceLinks ? (
            <>
              <Link
                href="/studio"
                className="hidden items-center gap-2 text-sm text-dim transition-colors hover:text-mark sm:inline-flex"
              >
                <PenLine size={15} />
                编辑
              </Link>
              <Link
                href="/codex"
                className="hidden items-center gap-2 text-sm text-dim transition-colors hover:text-mark sm:inline-flex"
              >
                <Network size={15} />
                规则
              </Link>
            </>
          ) : null}
          {showCatalog ? (
            <Link
              href={catalogHref}
              className="inline-flex min-h-10 items-center gap-2 text-sm text-dim transition-colors hover:text-mark"
            >
              <BookOpen size={15} />
              <span className="hidden sm:inline">目录</span>
            </Link>
          ) : null}
          <ThemeToggle />
        </div>
      </nav>
    </header>
  )
}
