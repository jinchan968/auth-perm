import type { Metadata } from "next"
import { notFound, redirect } from "next/navigation"
import { getNovel } from "@/lib/api/novel"

type LegacyChapterPageProps = {
  params: {
    slug: string
  }
}

export const dynamic = "force-dynamic"

export async function generateMetadata(): Promise<Metadata> {
  return {
    title: "章节跳转 · 小说",
  }
}

export default async function LegacyChapterPage({ params }: LegacyChapterPageProps) {
  const novel = await getNovel()

  if (!novel) {
    notFound()
  }

  redirect(`/novels/${novel.id}/chapters/${params.slug}`)
}
