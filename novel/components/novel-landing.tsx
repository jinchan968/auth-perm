import { ArrowRight, ScrollText } from "lucide-react"
import Link from "next/link"
import type { Novel, NovelStatus, NovelSummary } from "@/types/novel"

type NovelLandingProps = {
  novel: Novel
  novels: NovelSummary[]
}

export function NovelLanding({ novel, novels }: NovelLandingProps) {
  const firstChapter = novel.chapters[0]
  const secondStart = novel.chapters.find((chapter) => chapter.volumeId !== firstChapter?.volumeId)
  const chapterHref = (slug: string) => `/novels/${novel.id}/chapters/${slug}`

  return (
    <main>
      <section className="mx-auto grid min-h-[calc(100svh-4rem)] max-w-7xl items-center gap-10 px-4 py-10 sm:px-5 md:grid-cols-[1fr_0.72fr] md:px-8 md:py-20">
        <div>
          <div className="micro-type text-dim">{novel.subtitle}</div>
          <div className="mt-3 flex items-center gap-3 text-sm text-mark">
            <span className="h-1.5 w-1.5 rounded-full bg-mark" />
            {novel.issue}
          </div>

          <p className="mt-8 text-sm text-dim md:text-base">
            {novel.volumes.length} 卷 · {novel.chapters.length} 章
          </p>
          <h1 className="mt-5 max-w-3xl font-display text-4xl leading-[1.12] sm:text-5xl md:text-7xl">
            {novel.heroTitle.map((line) => (
              <span key={line} className="block">
                {line}
              </span>
            ))}
          </h1>
          <p className="mt-7 max-w-2xl text-base leading-8 text-dim sm:text-lg sm:leading-9">
            {novel.description}
          </p>

          <div className="mt-9 flex flex-wrap gap-3">
            {firstChapter ? (
              <Link
                href={chapterHref(firstChapter.slug)}
                className="inline-flex min-h-12 w-full items-center justify-between gap-3 border border-ink px-5 text-sm transition-colors hover:border-mark hover:text-mark sm:w-auto"
              >
                从第一章开始
                <ArrowRight size={16} />
              </Link>
            ) : null}
            {secondStart ? (
              <Link
                href={chapterHref(secondStart.slug)}
                className="inline-flex min-h-12 w-full items-center justify-between gap-3 border border-line px-5 text-sm text-dim transition-colors hover:border-mark hover:text-mark sm:w-auto"
              >
                继续下一卷
                <ArrowRight size={16} />
              </Link>
            ) : null}
            {novel.chapters.length > 0 ? (
              <Link
                href="#chapters"
                className="inline-flex min-h-12 items-center gap-2 px-1 text-sm text-dim transition-colors hover:text-mark sm:px-2"
              >
                查看目录
                <ScrollText size={16} />
              </Link>
            ) : null}
          </div>

          <dl className="mt-12 grid max-w-2xl grid-cols-1 gap-5 border-y border-line py-6 sm:grid-cols-3 sm:gap-6">
            {novel.stats.map((stat) => (
              <div key={stat.label}>
                <dt className="micro-type text-dim">{stat.label}</dt>
                <dd className="mt-2 text-sm leading-6">{stat.value}</dd>
              </div>
            ))}
          </dl>
        </div>

        <aside className="terminal-panel px-4 py-5 font-mono text-sm leading-7 sm:px-6 sm:py-7 md:justify-self-end">
          <div className="mb-5 flex flex-wrap items-center justify-between gap-2 border-b border-white/15 pb-3 text-xs text-white/60">
            <span>{novel.title.toLowerCase()} @ archive</span>
            <span>readonly</span>
          </div>
          <pre>{`$ query novel
title: ${novel.title}
chapters: ${novel.chapters.length}
issue: ${novel.issue}

$ open catalog
entry: ${firstChapter ? firstChapter.title : "等待章节发布"}`}</pre>
        </aside>
      </section>

      <NovelSwitcher currentId={novel.id} novels={novels} />
      <NovelCatalog novel={novel} chapterHref={chapterHref} />
    </main>
  )
}

function NovelSwitcher({ currentId, novels }: { currentId: string; novels: NovelSummary[] }) {
  if (novels.length <= 1) {
    return null
  }

  return (
    <section className="mx-auto max-w-7xl px-4 pb-14 sm:px-5 md:px-8 md:pb-16">
      <div className="hairline pt-7">
        <div className="micro-type text-dim">Library</div>
        <h2 className="mt-3 font-display text-3xl">作品</h2>
      </div>
      <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {novels.map((item) => {
          const active = item.id === currentId

          return (
            <Link
              key={item.id}
              href={`/novels/${item.id}`}
              className={`min-h-32 border p-4 transition-colors sm:min-h-36 sm:p-5 ${
                active
                  ? "border-mark text-mark"
                  : "border-line text-ink hover:border-mark hover:text-mark"
              }`}
            >
              <span className="micro-type text-dim">{formatNovelStatus(item.status)}</span>
              <span className="mt-3 block font-display text-2xl">{item.title}</span>
              <span className="mt-3 line-clamp-2 block text-sm leading-6 text-dim">
                {item.description || item.subtitle || "暂无简介"}
              </span>
            </Link>
          )
        })}
      </div>
    </section>
  )
}

function formatNovelStatus(status: NovelStatus) {
  const labels: Record<NovelStatus, string> = {
    draft: "草稿",
    serial: "连载中",
    completed: "已完结",
    archived: "已归档",
  }
  return labels[status] ?? status
}

function NovelCatalog({
  novel,
  chapterHref,
}: {
  novel: Novel
  chapterHref: (slug: string) => string
}) {
  return (
    <section id="chapters" className="mx-auto max-w-7xl px-4 pb-28 sm:px-5 md:px-8 md:pb-24">
      <div className="hairline pt-7">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <div className="micro-type text-dim">Contents</div>
            <h2 className="mt-3 font-display text-3xl sm:text-4xl">目录</h2>
          </div>
          <p className="text-sm text-dim">{novel.chapters.length} chapters · 由序章入</p>
        </div>
      </div>

      {novel.chapters.length === 0 ? (
        <div className="mt-10 border border-line p-8 text-sm leading-7 text-dim">
          当前小说还没有已发布章节。
        </div>
      ) : (
        <div className="mt-10 grid gap-12 lg:grid-cols-2">
          {novel.volumes.map((volume) => {
            const chapters = novel.chapters.filter((chapter) => chapter.volumeId === volume.id)

            if (chapters.length === 0) {
              return null
            }

            return (
              <section key={volume.id} aria-label={volume.title}>
                <div className="mb-5">
                  <h3 className="font-display text-2xl">
                    <span className="text-mark">{volume.title}</span>
                    <span className="mt-1 block text-dim sm:ml-3 sm:mt-0 sm:inline">
                      {volume.subtitle}
                    </span>
                  </h3>
                  <p className="mt-3 text-sm leading-7 text-dim">{volume.description}</p>
                </div>

                <div>
                  {chapters.map((chapter) => (
                    <Link
                      key={chapter.slug}
                      href={chapterHref(chapter.slug)}
                      className="chapter-row grid min-h-16 grid-cols-[2.75rem_minmax(0,1fr)_1.25rem] items-center gap-3 py-3 sm:grid-cols-[3.5rem_1fr_1.5rem]"
                    >
                      <span className="font-mono text-sm text-dim">{chapter.number}</span>
                      <span className="min-w-0">
                        <span className="block truncate text-base">{chapter.title}</span>
                        <span className="mt-1 line-clamp-2 block text-xs leading-5 text-dim">
                          {chapter.summary}
                        </span>
                      </span>
                      <ArrowRight size={16} />
                    </Link>
                  ))}
                </div>
              </section>
            )
          })}
        </div>
      )}
    </section>
  )
}
