export type ChapterStatus = "published" | "draft" | "review" | "locked"
export type EditorialStatus = "draft" | "review" | "published" | "locked"
export type ConflictLevel = "blocking" | "warning" | "hint"
export type CodexKind = "character" | "encyclopedia" | "geography" | "worldview"
export type NovelStatus = "draft" | "serial" | "completed" | "archived"

export type NovelSummary = {
  id: string
  title: string
  subtitle: string
  description: string
  status: NovelStatus
  tags: string[]
  updatedAt: string
}

export type Volume = {
  id: string
  title: string
  subtitle: string
  description: string
  sortOrder: number
}

export type Chapter = {
  slug: string
  volumeId: string
  number: string
  eyebrow: string
  title: string
  summary: string
  wordCount: number
  readingMinutes: number
  sortOrder: number
  status: ChapterStatus
  quote?: string
  paragraphs: string[]
  terminalBlocks?: string[][]
}

export type Novel = {
  id: string
  title: string
  subtitle: string
  issue: string
  heroTitle: string[]
  description: string
  stats: Array<{
    label: string
    value: string
  }>
  volumes: Volume[]
  chapters: Chapter[]
}

export type AdjacentChapter = {
  previous?: Pick<Chapter, "slug" | "title" | "number">
  next?: Pick<Chapter, "slug" | "title" | "number">
}

export type ChapterVersion = {
  id: string
  label: string
  savedAt: string
  author: string
  note: string
}

export type RuleConflict = {
  id: string
  level: ConflictLevel
  title: string
  detail: string
  target: string
}

export type EditorialChapter = Pick<
  Chapter,
  "slug" | "number" | "title" | "summary" | "wordCount" | "readingMinutes"
> & {
  status: EditorialStatus
  owner: string
  updatedAt: string
  body: string
  versions: ChapterVersion[]
  conflicts: RuleConflict[]
}

export type CharacterProfile = {
  id: string
  name: string
  aliases: string[]
  faction: string
  role: string
  firstChapter: string
  status: string
  desire: string
  fear: string
  arc: string
  relations: string[]
  constraints: RuleConflict[]
}

export type EncyclopediaEntry = {
  id: string
  term: string
  kind: string
  definition: string
  aliases: string[]
  evidence: string
  related: string[]
  confidence: string
}

export type GeographyEntry = {
  id: string
  name: string
  level: string
  parent: string
  owner: string
  route: string
  description: string
  scenes: string[]
  constraints: RuleConflict[]
}

export type WorldviewRule = {
  id: string
  title: string
  strength: string
  domain: string
  statement: string
  exception: string
  cost: string
  constraints: RuleConflict[]
}

export type WorldbuildingSnapshot = {
  characters: CharacterProfile[]
  encyclopedia: EncyclopediaEntry[]
  geography: GeographyEntry[]
  worldview: WorldviewRule[]
}
