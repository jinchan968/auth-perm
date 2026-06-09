import Link from "next/link"
import { SiteHeader } from "@/components/site-header"

export default function NotFoundPage() {
  return (
    <>
      <SiteHeader showCatalog />
      <main className="mx-auto flex min-h-[calc(100svh-4rem)] max-w-3xl flex-col justify-center px-4 py-12 sm:px-5">
        <p className="micro-type text-mark">404 · Missing Chapter</p>
        <h1 className="mt-5 font-display text-4xl leading-tight sm:text-5xl">
          这一章还没有写入档案。
        </h1>
        <p className="mt-6 leading-8 text-dim">当前只读模式仅开放已发布章节。</p>
        <Link
          href="/"
          className="mt-10 inline-flex min-h-12 w-fit items-center border border-ink px-5 text-sm transition-colors hover:border-mark hover:text-mark"
        >
          返回目录
        </Link>
      </main>
    </>
  )
}
