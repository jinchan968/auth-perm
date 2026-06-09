import { ArrowRight, BookOpen, Clock3 } from "lucide-react"
import Link from "next/link"
import type { NovelStatus, NovelSummary } from "@/types/novel"

type NovelLibraryProps = {
  novels: NovelSummary[]
}

const statusLabels: Record<NovelStatus, string> = {
  draft: "草稿",
  serial: "连载中",
  completed: "已完结",
  archived: "已归档",
}

function formatDate(value: string) {
  if (!value) return "等待更新"
  return new Date(value).toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  })
}

export function NovelLibrary({ novels }: NovelLibraryProps) {
  const featured = novels[0]

  return (
    <main>
      <section className="mx-auto min-h-[calc(100svh-4rem)] max-w-7xl px-4 py-12 sm:px-5 md:px-8 md:py-20">
        <div className="grid gap-10 lg:grid-cols-[0.82fr_1.18fr] lg:items-end">
          <div>
            <p className="micro-type text-mark">Library</p>
            <h1 className="mt-5 max-w-3xl font-display text-5xl leading-[1.05] sm:text-6xl md:text-7xl">
              小说书架
            </h1>
            <p className="mt-7 max-w-2xl text-base leading-8 text-dim sm:text-lg sm:leading-9">
              这里收录已经公开的作品。选择一本小说进入详情页，再查看目录和开始阅读。
            </p>
          </div>

          {featured ? (
            <Link
              href={`/novels/${featured.id}`}
              className="terminal-panel block px-5 py-6 font-mono text-sm leading-7 transition-transform hover:-translate-y-1 sm:px-7 sm:py-8"
            >
              <div className="mb-5 flex flex-wrap items-center justify-between gap-2 border-b border-white/15 pb-3 text-xs text-white/60">
                <span>latest selection</span>
                <span>{statusLabels[featured.status]}</span>
              </div>
              <pre>{`$ open novel
title: ${featured.title}
updated: ${formatDate(featured.updatedAt)}

$ enter detail
catalog: /novels/${featured.id}`}</pre>
            </Link>
          ) : null}
        </div>

        <section className="mt-16 md:mt-24">
          <div className="hairline pt-7">
            <div className="flex flex-wrap items-end justify-between gap-4">
              <div>
                <div className="micro-type text-dim">Works</div>
                <h2 className="mt-3 font-display text-3xl sm:text-4xl">全部作品</h2>
              </div>
              <p className="text-sm text-dim">{novels.length} 部作品</p>
            </div>
          </div>

          <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {novels.map((novel, index) => (
              <Link
                key={novel.id}
                href={`/novels/${novel.id}`}
                className="group flex min-h-64 flex-col justify-between border border-line bg-paper/70 p-5 transition-colors hover:border-mark hover:bg-wash sm:p-6"
              >
                <div>
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <span className="micro-type text-dim">No. {String(index + 1).padStart(2, "0")}</span>
                    <span className="border border-line px-2 py-1 text-xs text-dim group-hover:border-mark group-hover:text-mark">
                      {statusLabels[novel.status]}
                    </span>
                  </div>
                  <h3 className="mt-6 font-display text-3xl leading-tight">{novel.title}</h3>
                  <p className="mt-4 line-clamp-3 text-sm leading-7 text-dim">
                    {novel.description || novel.subtitle || "这部小说还没有简介。"}
                  </p>
                </div>

                <div className="mt-8 flex flex-wrap items-center justify-between gap-4 border-t border-line pt-4 text-sm text-dim">
                  <span className="inline-flex items-center gap-2">
                    <Clock3 size={15} />
                    {formatDate(novel.updatedAt)}
                  </span>
                  <span className="inline-flex items-center gap-2 text-ink group-hover:text-mark">
                    详情
                    <ArrowRight size={16} />
                  </span>
                </div>
              </Link>
            ))}
          </div>
        </section>
      </section>
    </main>
  )
}
