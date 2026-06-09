"use client"

import { AlertTriangle, CheckCircle2, Eye, RotateCcw, Save, Send, Upload } from "lucide-react"
import { useMemo, useState } from "react"
import type { EditorialChapter, EditorialStatus } from "@/types/novel"

type StudioWorkspaceProps = {
  initialChapters: EditorialChapter[]
}

const statusLabels: Record<EditorialStatus, string> = {
  draft: "草稿",
  review: "审核中",
  published: "已发布",
  locked: "已锁定",
}

export function StudioWorkspace({ initialChapters }: StudioWorkspaceProps) {
  const [chapters, setChapters] = useState(initialChapters)
  const [activeSlug, setActiveSlug] = useState(initialChapters[0]?.slug)
  const [previewMode, setPreviewMode] = useState(false)

  const activeChapter = useMemo(
    () => chapters.find((chapter) => chapter.slug === activeSlug) ?? chapters[0],
    [activeSlug, chapters],
  )

  function updateActiveChapter(updater: (chapter: EditorialChapter) => EditorialChapter) {
    setChapters((current) =>
      current.map((chapter) => (chapter.slug === activeChapter.slug ? updater(chapter) : chapter)),
    )
  }

  function updateField(field: keyof Pick<EditorialChapter, "title" | "summary" | "body">, value: string) {
    updateActiveChapter((chapter) => ({
      ...chapter,
      [field]: value,
      wordCount: field === "body" ? countWords(value) : chapter.wordCount,
      readingMinutes: field === "body" ? Math.max(1, Math.ceil(countWords(value) / 420)) : chapter.readingMinutes,
      updatedAt: "本地未保存",
    }))
  }

  function setStatus(status: EditorialStatus) {
    updateActiveChapter((chapter) => ({ ...chapter, status, updatedAt: "本地未保存" }))
  }

  function saveDraft() {
    updateActiveChapter((chapter) => ({
      ...chapter,
      status: chapter.status === "published" ? "published" : "draft",
      updatedAt: "刚刚",
      versions: [
        {
          id: `${chapter.slug}-local-${chapter.versions.length + 1}`,
          label: `v0.${chapter.versions.length + 3}`,
          savedAt: "刚刚",
          author: chapter.owner,
          note: "本地保存草稿。",
        },
        ...chapter.versions,
      ],
    }))
  }

  function submitReview() {
    setStatus("review")
  }

  function publishChapter() {
    updateActiveChapter((chapter) => ({
      ...chapter,
      status: chapter.conflicts.some((conflict) => conflict.level === "blocking")
        ? "review"
        : "published",
      updatedAt: "刚刚",
    }))
  }

  if (!activeChapter) {
    return null
  }

  const blockingConflicts = activeChapter.conflicts.filter((conflict) => conflict.level === "blocking")
  const bodyParagraphs = activeChapter.body.split(/\n{2,}/).filter(Boolean)

  return (
    <div className="mx-auto grid max-w-7xl gap-8 px-5 py-10 md:px-8 lg:grid-cols-[18rem_1fr_22rem]">
      <aside className="lg:sticky lg:top-24 lg:self-start">
        <div className="micro-type text-dim">Chapter Desk</div>
        <h1 className="mt-3 font-display text-4xl">章节编辑</h1>
        <p className="mt-4 text-sm leading-7 text-dim">
          编辑模式以章节为中心，草稿、审核、发布和版本都在同一个工作台里完成。
        </p>

        <div className="mt-8 border-y border-line">
          {chapters.map((chapter) => (
            <button
              key={chapter.slug}
              type="button"
              className={`chapter-row grid w-full grid-cols-[3rem_1fr] items-center gap-3 py-4 text-left ${
                chapter.slug === activeChapter.slug ? "text-mark" : ""
              }`}
              onClick={() => setActiveSlug(chapter.slug)}
            >
              <span className="font-mono text-sm text-dim">{chapter.number}</span>
              <span>
                <span className="block">{chapter.title}</span>
                <span className="mt-1 block text-xs text-dim">{statusLabels[chapter.status]}</span>
              </span>
            </button>
          ))}
        </div>
      </aside>

      <section>
        <div className="hairline pb-6">
          <div className="flex flex-wrap items-center justify-between gap-4 pt-6">
            <div>
              <p className="micro-type text-mark">{activeChapter.number} · {statusLabels[activeChapter.status]}</p>
              <h2 className="mt-3 font-display text-4xl leading-tight sm:text-5xl">
                {activeChapter.title}
              </h2>
            </div>
            <div className="flex flex-wrap gap-2">
              <button type="button" className="icon-button" onClick={() => setPreviewMode((value) => !value)} title="预览">
                <Eye size={16} />
              </button>
              <button type="button" className="icon-button" onClick={saveDraft} title="保存草稿">
                <Save size={16} />
              </button>
              <button type="button" className="icon-button" onClick={submitReview} title="提交审核">
                <Send size={16} />
              </button>
              <button type="button" className="icon-button" onClick={publishChapter} title="发布">
                <Upload size={16} />
              </button>
            </div>
          </div>
        </div>

        {previewMode ? (
          <article className="reader-content mt-10">
            <p className="micro-type text-dim">Preview · {activeChapter.wordCount} 字</p>
            {bodyParagraphs.map((paragraph, index) => (
              <p key={`${activeChapter.slug}-preview-${index}`}>{paragraph}</p>
            ))}
          </article>
        ) : (
          <div className="mt-10 grid gap-6">
            <label className="grid gap-2">
              <span className="micro-type text-dim">Title</span>
              <input
                value={activeChapter.title}
                onChange={(event) => updateField("title", event.target.value)}
                className="border border-line bg-paper px-4 py-3 font-display text-2xl outline-none transition-colors focus:border-mark sm:text-3xl"
              />
            </label>

            <label className="grid gap-2">
              <span className="micro-type text-dim">Summary</span>
              <textarea
                value={activeChapter.summary}
                onChange={(event) => updateField("summary", event.target.value)}
                rows={3}
                className="resize-none border border-line bg-paper px-4 py-3 leading-7 outline-none transition-colors focus:border-mark"
              />
            </label>

            <label className="grid gap-2">
              <span className="micro-type text-dim">Manuscript</span>
              <textarea
                value={activeChapter.body}
                onChange={(event) => updateField("body", event.target.value)}
                rows={20}
                className="min-h-[28rem] resize-y border border-line bg-paper px-4 py-4 font-body text-base leading-8 outline-none transition-colors focus:border-mark sm:min-h-[36rem] sm:px-5 sm:text-lg sm:leading-9"
              />
            </label>
          </div>
        )}
      </section>

      <aside className="grid gap-6 lg:sticky lg:top-24 lg:self-start">
        <section className="border border-line p-5">
          <div className="micro-type text-dim">Workflow</div>
          <div className="mt-5 grid grid-cols-2 gap-2">
            {(Object.keys(statusLabels) as EditorialStatus[]).map((status) => (
              <button
                key={status}
                type="button"
                className={`min-h-11 border px-3 text-sm transition-colors ${
                  activeChapter.status === status
                    ? "border-mark text-mark"
                    : "border-line text-dim hover:border-mark hover:text-mark"
                }`}
                onClick={() => setStatus(status)}
              >
                {statusLabels[status]}
              </button>
            ))}
          </div>
          <dl className="mt-6 grid grid-cols-2 gap-4 text-sm">
            <div>
              <dt className="micro-type text-dim">Words</dt>
              <dd className="mt-1">{activeChapter.wordCount}</dd>
            </div>
            <div>
              <dt className="micro-type text-dim">Read</dt>
              <dd className="mt-1">{activeChapter.readingMinutes} 分钟</dd>
            </div>
            <div>
              <dt className="micro-type text-dim">Owner</dt>
              <dd className="mt-1">{activeChapter.owner}</dd>
            </div>
            <div>
              <dt className="micro-type text-dim">Updated</dt>
              <dd className="mt-1">{activeChapter.updatedAt}</dd>
            </div>
          </dl>
        </section>

        <section className="border border-line p-5">
          <div className="flex items-center justify-between gap-3">
            <div>
              <div className="micro-type text-dim">Constraint</div>
              <h3 className="mt-2 font-display text-2xl">规则校验</h3>
            </div>
            {blockingConflicts.length > 0 ? (
              <AlertTriangle className="text-mark" size={22} />
            ) : (
              <CheckCircle2 className="text-dim" size={22} />
            )}
          </div>
          <div className="mt-5 grid gap-3">
            {activeChapter.conflicts.length > 0 ? (
              activeChapter.conflicts.map((conflict) => (
                <div key={conflict.id} className="border-l-2 border-mark bg-wash px-4 py-3">
                  <p className="text-sm font-semibold">{conflict.title}</p>
                  <p className="mt-2 text-sm leading-6 text-dim">{conflict.detail}</p>
                  <p className="micro-type mt-3 text-mark">{conflict.target}</p>
                </div>
              ))
            ) : (
              <p className="text-sm leading-7 text-dim">当前章节没有阻断级冲突，可以进入发布流程。</p>
            )}
          </div>
        </section>

        <section className="border border-line p-5">
          <div className="flex items-center justify-between">
            <div className="micro-type text-dim">Versions</div>
            <RotateCcw size={16} className="text-dim" />
          </div>
          <div className="mt-5 grid gap-4">
            {activeChapter.versions.map((version) => (
              <div key={version.id} className="border-t border-line pt-4">
                <p className="font-mono text-sm">{version.label} · {version.savedAt}</p>
                <p className="mt-2 text-sm leading-6 text-dim">{version.note}</p>
                <p className="micro-type mt-2 text-dim">{version.author}</p>
              </div>
            ))}
          </div>
        </section>
      </aside>
    </div>
  )
}

function countWords(value: string) {
  return value.replace(/\s+/g, "").length
}
