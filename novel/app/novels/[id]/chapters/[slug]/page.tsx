import type { Metadata } from "next"
import { ArrowLeft, ArrowRight, BookOpen } from "lucide-react"
import Link from "next/link"
import { notFound } from "next/navigation"
import { ReaderPreferences } from "@/components/reader-preferences"
import { ChapterShortcuts } from "@/components/chapter-shortcuts"
import { SiteHeader } from "@/components/site-header"
import { getChapterBySlug, getNovel } from "@/lib/api/novel"

type ChapterPageProps = {
  params: {
    id: string
    slug: string
  }
}

export const dynamic = "force-dynamic"
export const runtime = "edge"

export async function generateMetadata({ params }: ChapterPageProps): Promise<Metadata> {
  const [novel, chapter] = await Promise.all([
    getNovel(params.id).catch(() => undefined),
    getChapterBySlug(params.id, params.slug).catch(() => undefined),
  ])

  return {
    title: chapter && novel ? `${chapter.title} · ${novel.title}` : "章节不存在",
  }
}

export default async function ChapterPage({ params }: ChapterPageProps) {
  const [novel, chapter] = await Promise.all([
    getNovel(params.id),
    getChapterBySlug(params.id, params.slug),
  ])

  if (!novel || !chapter) {
    notFound()
  }

  const chapterIndex = novel.chapters.findIndex((item) => item.slug === chapter.slug)
  const previousChapter = chapterIndex > 0 ? novel.chapters[chapterIndex - 1] : undefined
  const nextChapter =
    chapterIndex >= 0 && chapterIndex < novel.chapters.length - 1
      ? novel.chapters[chapterIndex + 1]
      : undefined
  const volume = novel.volumes.find((item) => item.id === chapter.volumeId)
  const novelHref = `/novels/${novel.id}`
  const chapterHref = (slug: string) => `/novels/${novel.id}/chapters/${slug}`
  const catalogHref = `${novelHref}#chapters`
  const previousHref = previousChapter ? chapterHref(previousChapter.slug) : undefined
  const nextHref = nextChapter ? chapterHref(nextChapter.slug) : undefined

  return (
    <>
      <ChapterShortcuts catalogHref={catalogHref} previousHref={previousHref} nextHref={nextHref} />
      <SiteHeader
        title={novel.title}
        volume={volume?.title}
        showCatalog
        homeHref={novelHref}
        catalogHref={catalogHref}
      />
      <ReaderPreferences>
        <main className="mx-auto max-w-4xl px-4 pb-36 pt-10 sm:px-5 md:px-8 md:pb-24 md:pt-20">
          <article>
            <header>
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="micro-type text-dim">
                  {novel.title} · {volume?.title ?? "Vol."}
                </div>
                <div className="micro-type text-mark">
                  <span className="mr-3">●</span>
                  {chapter.eyebrow} · {chapter.title}
                </div>
              </div>
              <div className="hairline mt-4" />

              <p className="micro-type mt-7 text-mark md:mt-8">{chapter.eyebrow}</p>
              <h1 className="mt-4 font-display text-4xl leading-tight sm:text-5xl md:mt-5 md:text-6xl">
                {chapter.title}
              </h1>
              <p className="mt-6 font-mono text-sm text-dim">
                {chapter.wordCount.toLocaleString("zh-CN")} 字 · 约 {chapter.readingMinutes} 分钟
              </p>

              {chapter.quote ? (
                <blockquote className="mt-8 border-l-2 border-mark pl-4 text-base italic leading-8 text-dim sm:mt-10 sm:pl-5 sm:text-lg sm:leading-9">
                  “{chapter.quote}”
                </blockquote>
              ) : null}

              <div className="micro-type my-10 text-center text-dim md:my-12">- {novel.title} -</div>
            </header>

            <div className="reader-content">
              {chapter.paragraphs.length > 0 ? (
                chapter.paragraphs.map((paragraph, index) => (
                  <p key={`${chapter.slug}-${index}`}>{paragraph}</p>
                ))
              ) : (
                <p>这一章还没有公开正文。</p>
              )}

              {chapter.terminalBlocks?.map((block, index) => (
                <pre key={`${chapter.slug}-terminal-${index}`} className="reader-code">
                  {block.join("\n")}
                </pre>
              ))}
            </div>
          </article>

          <nav className="mt-16 grid gap-3 border-t border-line pt-6 sm:grid-cols-3 md:mt-20 md:gap-4 md:pt-8">
            {previousChapter ? (
              <Link
                href={chapterHref(previousChapter.slug)}
                className="group min-h-20 border border-line p-4 transition-colors hover:border-mark hover:text-mark sm:min-h-24 sm:p-5"
              >
                <span className="inline-flex items-center gap-2 text-sm text-dim group-hover:text-mark">
                  <ArrowLeft size={15} />
                  上一章
                </span>
                <span className="mt-3 block font-display text-xl sm:text-2xl">
                  {previousChapter.number} · {previousChapter.title}
                </span>
              </Link>
            ) : (
              <div
                aria-disabled="true"
                className="min-h-20 border border-line/70 p-4 text-dim/60 sm:min-h-24 sm:p-5"
              >
                <span className="inline-flex items-center gap-2 text-sm">
                  <ArrowLeft size={15} />
                  上一章
                </span>
                <span className="mt-3 block font-display text-xl sm:text-2xl">没有上一章</span>
              </div>
            )}

            <Link
              href={catalogHref}
              className="group flex min-h-20 flex-col justify-center border border-line p-4 text-center transition-colors hover:border-mark hover:text-mark sm:min-h-24 sm:p-5"
            >
              <span className="inline-flex items-center justify-center gap-2 text-sm text-dim group-hover:text-mark">
                <BookOpen size={15} />
                目录
              </span>
              <span className="mt-3 block font-display text-xl sm:text-2xl">回到目录</span>
            </Link>

            {nextChapter ? (
              <Link
                href={chapterHref(nextChapter.slug)}
                className="group min-h-20 border border-line p-4 text-left transition-colors hover:border-mark hover:text-mark sm:min-h-24 sm:p-5 sm:text-right"
              >
                <span className="inline-flex items-center gap-2 text-sm text-dim group-hover:text-mark sm:justify-end">
                  下一章
                  <ArrowRight size={15} />
                </span>
                <span className="mt-3 block font-display text-xl sm:text-2xl">
                  {nextChapter.number} · {nextChapter.title}
                </span>
              </Link>
            ) : (
              <div
                aria-disabled="true"
                className="min-h-20 border border-line/70 p-4 text-left text-dim/60 sm:min-h-24 sm:p-5 sm:text-right"
              >
                <span className="inline-flex items-center gap-2 text-sm sm:justify-end">
                  下一章
                  <ArrowRight size={15} />
                </span>
                <span className="mt-3 block font-display text-xl sm:text-2xl">没有下一章</span>
              </div>
            )}
          </nav>
        </main>
      </ReaderPreferences>
    </>
  )
}
