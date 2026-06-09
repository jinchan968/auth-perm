import { NovelLibrary } from "@/components/novel-library"
import { SiteHeader } from "@/components/site-header"
import { listNovels, NovelApiError } from "@/lib/api/novel"

export const dynamic = "force-dynamic"

export default async function HomePage() {
  try {
    const novels = await listNovels()

    if (novels.length === 0) {
      return <EmptyLibrary />
    }

    return (
      <>
        <SiteHeader title="小说" volume="书架" />
        <NovelLibrary novels={novels} />
      </>
    )
  } catch (error) {
    return <LibraryError error={error} />
  }
}

function EmptyLibrary() {
  return (
    <>
      <SiteHeader title="小说" volume="Library" />
      <main className="mx-auto flex min-h-[calc(100svh-4rem)] max-w-3xl flex-col justify-center px-4 py-12 sm:px-5">
        <p className="micro-type text-mark">Library</p>
        <h1 className="mt-5 font-display text-4xl leading-tight sm:text-5xl">
          还没有公开小说。
        </h1>
        <p className="mt-6 leading-8 text-dim">
          当有连载中或已完结的作品发布后，这里会自动显示书架。
        </p>
      </main>
    </>
  )
}

function LibraryError({ error }: { error: unknown }) {
  const detail =
    error instanceof NovelApiError ? error.message : "请确认后端服务和 NEXT_PUBLIC_API_URL 配置。"

  return (
    <>
      <SiteHeader title="小说" volume="Library" />
      <main className="mx-auto flex min-h-[calc(100svh-4rem)] max-w-3xl flex-col justify-center px-4 py-12 sm:px-5">
        <p className="micro-type text-mark">API Error</p>
        <h1 className="mt-5 font-display text-4xl leading-tight sm:text-5xl">
          暂时无法读取小说。
        </h1>
        <p className="mt-6 leading-8 text-dim">{detail}</p>
      </main>
    </>
  )
}
