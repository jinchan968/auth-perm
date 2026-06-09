import { SiteHeader } from "@/components/site-header"
import { StudioWorkspace } from "@/components/studio-workspace"
import { getEditorialChapters } from "@/lib/api/novel"

export default async function StudioPage() {
  const chapters = await getEditorialChapters()

  return (
    <>
      <SiteHeader volume="Studio" showCatalog showWorkspaceLinks />
      <StudioWorkspace initialChapters={chapters} />
    </>
  )
}
