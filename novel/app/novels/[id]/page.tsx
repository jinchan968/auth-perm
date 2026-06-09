import type { Metadata } from "next"
import { notFound } from "next/navigation"
import { NovelLanding } from "@/components/novel-landing"
import { SiteHeader } from "@/components/site-header"
import { getNovel, listNovels } from "@/lib/api/novel"

type NovelPageProps = {
  params: {
    id: string
  }
}

export const dynamic = "force-dynamic"
export const runtime = "edge"

export async function generateMetadata({ params }: NovelPageProps): Promise<Metadata> {
  const novel = await getNovel(params.id).catch(() => undefined)

  return {
    title: novel ? `${novel.title} · 小说` : "小说不存在",
    description: novel?.description,
  }
}

export default async function NovelPage({ params }: NovelPageProps) {
  const [novel, novels] = await Promise.all([getNovel(params.id), listNovels()])

  if (!novel) {
    notFound()
  }

  return (
    <>
      <SiteHeader title={novel.title} volume="目录" homeHref="/" showCatalog catalogHref="#chapters" />
      <NovelLanding novel={novel} novels={novels} />
    </>
  )
}
