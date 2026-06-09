import { CodexWorkspace } from "@/components/codex-workspace"
import { SiteHeader } from "@/components/site-header"
import { getWorldbuildingSnapshot } from "@/lib/api/novel"

export default async function CodexPage() {
  const snapshot = await getWorldbuildingSnapshot()

  return (
    <>
      <SiteHeader volume="Codex" showCatalog showWorkspaceLinks />
      <CodexWorkspace initialSnapshot={snapshot} />
    </>
  )
}
