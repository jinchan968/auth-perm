"use client"

import { SiteHeader } from "@/components/site-header"

export default function ErrorPage() {
  return (
    <>
      <SiteHeader title="小说" volume="Error" />
      <main className="mx-auto flex min-h-[calc(100svh-4rem)] max-w-3xl flex-col justify-center px-4 py-12 sm:px-5">
        <p className="micro-type text-mark">API Error</p>
        <h1 className="mt-5 font-display text-4xl leading-tight sm:text-5xl">
          暂时无法读取小说。
        </h1>
        <p className="mt-6 leading-8 text-dim">
          请确认后端服务已启动，并检查 NEXT_PUBLIC_API_URL 是否指向后端的 /api/v1。
        </p>
      </main>
    </>
  )
}
