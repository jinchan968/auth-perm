"use client"

import { AlertOctagon, BookMarked, CheckCircle2, Globe2, Network, Plus, Scale } from "lucide-react"
import type { ReactNode } from "react"
import { useMemo, useState } from "react"
import type {
  CharacterProfile,
  CodexKind,
  EncyclopediaEntry,
  GeographyEntry,
  WorldbuildingSnapshot,
  WorldviewRule,
} from "@/types/novel"

type CodexWorkspaceProps = {
  initialSnapshot: WorldbuildingSnapshot
}

const tabs: Array<{ id: CodexKind; label: string; icon: typeof Network }> = [
  { id: "character", label: "人物线", icon: Network },
  { id: "encyclopedia", label: "百科", icon: BookMarked },
  { id: "geography", label: "小说地理", icon: Globe2 },
  { id: "worldview", label: "小说三观", icon: Scale },
]

export function CodexWorkspace({ initialSnapshot }: CodexWorkspaceProps) {
  const [snapshot, setSnapshot] = useState(initialSnapshot)
  const [activeKind, setActiveKind] = useState<CodexKind>("character")
  const [activeIds, setActiveIds] = useState({
    character: initialSnapshot.characters[0]?.id,
    encyclopedia: initialSnapshot.encyclopedia[0]?.id,
    geography: initialSnapshot.geography[0]?.id,
    worldview: initialSnapshot.worldview[0]?.id,
  })

  const activeItem = useMemo(() => {
    if (activeKind === "character") {
      return snapshot.characters.find((item) => item.id === activeIds.character)
    }
    if (activeKind === "encyclopedia") {
      return snapshot.encyclopedia.find((item) => item.id === activeIds.encyclopedia)
    }
    if (activeKind === "geography") {
      return snapshot.geography.find((item) => item.id === activeIds.geography)
    }
    return snapshot.worldview.find((item) => item.id === activeIds.worldview)
  }, [activeIds, activeKind, snapshot])

  const conflicts = collectConflicts(snapshot)

  function setActiveId(id: string) {
    setActiveIds((current) => ({ ...current, [activeKind]: id }))
  }

  function addItem() {
    if (activeKind === "character") {
      const next: CharacterProfile = {
        id: `char-local-${snapshot.characters.length + 1}`,
        name: "新人物",
        aliases: [],
        faction: "未分配",
        role: "待定",
        firstChapter: "未登场",
        status: "未知",
        desire: "待补充",
        fear: "待补充",
        arc: "待补充",
        relations: [],
        constraints: [],
      }
      setSnapshot((current) => ({ ...current, characters: [next, ...current.characters] }))
      setActiveIds((current) => ({ ...current, character: next.id }))
    }

    if (activeKind === "encyclopedia") {
      const next: EncyclopediaEntry = {
        id: `entry-local-${snapshot.encyclopedia.length + 1}`,
        term: "新条目",
        kind: "术语",
        definition: "待补充",
        aliases: [],
        evidence: "待补充",
        related: [],
        confidence: "低",
      }
      setSnapshot((current) => ({ ...current, encyclopedia: [next, ...current.encyclopedia] }))
      setActiveIds((current) => ({ ...current, encyclopedia: next.id }))
    }

    if (activeKind === "geography") {
      const next: GeographyEntry = {
        id: `geo-local-${snapshot.geography.length + 1}`,
        name: "新地点",
        level: "地点",
        parent: "未归属",
        owner: "未知",
        route: "待补充",
        description: "待补充",
        scenes: [],
        constraints: [],
      }
      setSnapshot((current) => ({ ...current, geography: [next, ...current.geography] }))
      setActiveIds((current) => ({ ...current, geography: next.id }))
    }

    if (activeKind === "worldview") {
      const next: WorldviewRule = {
        id: `rule-local-${snapshot.worldview.length + 1}`,
        title: "新规则",
        strength: "软规则",
        domain: "待分类",
        statement: "待补充",
        exception: "待补充",
        cost: "待补充",
        constraints: [],
      }
      setSnapshot((current) => ({ ...current, worldview: [next, ...current.worldview] }))
      setActiveIds((current) => ({ ...current, worldview: next.id }))
    }
  }

  function resolveConstraints() {
    if (activeKind === "character") {
      setSnapshot((current) => ({
        ...current,
        characters: current.characters.map((item) =>
          item.id === activeIds.character ? { ...item, constraints: [] } : item,
        ),
      }))
    }
    if (activeKind === "geography") {
      setSnapshot((current) => ({
        ...current,
        geography: current.geography.map((item) =>
          item.id === activeIds.geography ? { ...item, constraints: [] } : item,
        ),
      }))
    }
    if (activeKind === "worldview") {
      setSnapshot((current) => ({
        ...current,
        worldview: current.worldview.map((item) =>
          item.id === activeIds.worldview ? { ...item, constraints: [] } : item,
        ),
      }))
    }
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-5 md:px-8 md:py-10">
      <header className="hairline pb-8 pt-8">
        <p className="micro-type text-mark">Worldbuilding Codex</p>
        <div className="mt-4 grid gap-6 lg:grid-cols-[1fr_28rem]">
          <div>
            <h1 className="font-display text-4xl leading-tight sm:text-5xl md:text-6xl">
              规则约束工作台
            </h1>
            <p className="mt-5 max-w-3xl text-base leading-8 text-dim sm:text-lg sm:leading-9">
              人物线、百科、地理和小说三观不是附录，而是写作时会反向约束章节的资料库。
            </p>
          </div>
          <div className="border border-line p-5">
            <div className="micro-type text-dim">Constraint Pulse</div>
            <div className="mt-5 grid grid-cols-3 gap-2 text-center sm:gap-4">
              <Metric label="阻断" value={conflicts.blocking} />
              <Metric label="警告" value={conflicts.warning} />
              <Metric label="提示" value={conflicts.hint} />
            </div>
          </div>
        </div>
      </header>

      <div className="mt-8 flex flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap gap-2">
          {tabs.map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => setActiveKind(tab.id)}
                className={`inline-flex min-h-11 items-center gap-2 border px-4 text-sm transition-colors ${
                  activeKind === tab.id
                    ? "border-mark text-mark"
                    : "border-line text-dim hover:border-mark hover:text-mark"
                }`}
              >
                <Icon size={16} />
                {tab.label}
              </button>
            )
          })}
        </div>
        <button
          type="button"
          onClick={addItem}
          className="inline-flex min-h-11 items-center gap-2 border border-ink px-4 text-sm transition-colors hover:border-mark hover:text-mark"
        >
          <Plus size={16} />
          新增条目
        </button>
      </div>

      <div className="mt-8 grid gap-8 lg:grid-cols-[20rem_1fr_22rem]">
        <ItemList
          kind={activeKind}
          snapshot={snapshot}
          activeId={activeIds[activeKind]}
          onSelect={setActiveId}
        />

        <section className="border border-line p-5 md:p-7">
          {activeKind === "character" && activeItem ? (
            <CharacterEditor
              item={activeItem as CharacterProfile}
              onChange={(next) =>
                setSnapshot((current) => ({
                  ...current,
                  characters: current.characters.map((item) => (item.id === next.id ? next : item)),
                }))
              }
            />
          ) : null}

          {activeKind === "encyclopedia" && activeItem ? (
            <EncyclopediaEditor
              item={activeItem as EncyclopediaEntry}
              onChange={(next) =>
                setSnapshot((current) => ({
                  ...current,
                  encyclopedia: current.encyclopedia.map((item) => (item.id === next.id ? next : item)),
                }))
              }
            />
          ) : null}

          {activeKind === "geography" && activeItem ? (
            <GeographyEditor
              item={activeItem as GeographyEntry}
              onChange={(next) =>
                setSnapshot((current) => ({
                  ...current,
                  geography: current.geography.map((item) => (item.id === next.id ? next : item)),
                }))
              }
            />
          ) : null}

          {activeKind === "worldview" && activeItem ? (
            <WorldviewEditor
              item={activeItem as WorldviewRule}
              onChange={(next) =>
                setSnapshot((current) => ({
                  ...current,
                  worldview: current.worldview.map((item) => (item.id === next.id ? next : item)),
                }))
              }
            />
          ) : null}
        </section>

        <ConstraintPanel item={activeItem} onResolve={resolveConstraints} />
      </div>
    </div>
  )
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div>
      <p className="font-display text-4xl">{value}</p>
      <p className="micro-type mt-2 text-dim">{label}</p>
    </div>
  )
}

function ItemList({
  kind,
  snapshot,
  activeId,
  onSelect,
}: {
  kind: CodexKind
  snapshot: WorldbuildingSnapshot
  activeId?: string
  onSelect: (id: string) => void
}) {
  const items = getItems(kind, snapshot)

  return (
    <aside className="border-y border-line lg:sticky lg:top-24 lg:self-start">
      {items.map((item) => (
        <button
          key={item.id}
          type="button"
          onClick={() => onSelect(item.id)}
          className={`chapter-row w-full px-2 py-4 text-left ${activeId === item.id ? "text-mark" : ""}`}
        >
          <span className="block">{item.title}</span>
          <span className="mt-1 block text-xs text-dim">{item.meta}</span>
        </button>
      ))}
    </aside>
  )
}

function CharacterEditor({
  item,
  onChange,
}: {
  item: CharacterProfile
  onChange: (item: CharacterProfile) => void
}) {
  return (
    <EditorFrame eyebrow="Character Arc" title={item.name}>
      <Input label="姓名" value={item.name} onChange={(value) => onChange({ ...item, name: value })} />
      <Input label="阵营" value={item.faction} onChange={(value) => onChange({ ...item, faction: value })} />
      <Input label="身份" value={item.role} onChange={(value) => onChange({ ...item, role: value })} />
      <Input label="当前状态" value={item.status} onChange={(value) => onChange({ ...item, status: value })} />
      <Textarea label="欲望" value={item.desire} onChange={(value) => onChange({ ...item, desire: value })} />
      <Textarea label="恐惧" value={item.fear} onChange={(value) => onChange({ ...item, fear: value })} />
      <Textarea label="人物弧线" value={item.arc} onChange={(value) => onChange({ ...item, arc: value })} />
    </EditorFrame>
  )
}

function EncyclopediaEditor({
  item,
  onChange,
}: {
  item: EncyclopediaEntry
  onChange: (item: EncyclopediaEntry) => void
}) {
  return (
    <EditorFrame eyebrow="Encyclopedia" title={item.term}>
      <Input label="术语" value={item.term} onChange={(value) => onChange({ ...item, term: value })} />
      <Input label="类型" value={item.kind} onChange={(value) => onChange({ ...item, kind: value })} />
      <Input label="可信度" value={item.confidence} onChange={(value) => onChange({ ...item, confidence: value })} />
      <Textarea label="定义" value={item.definition} onChange={(value) => onChange({ ...item, definition: value })} />
      <Textarea label="证据片段" value={item.evidence} onChange={(value) => onChange({ ...item, evidence: value })} />
    </EditorFrame>
  )
}

function GeographyEditor({
  item,
  onChange,
}: {
  item: GeographyEntry
  onChange: (item: GeographyEntry) => void
}) {
  return (
    <EditorFrame eyebrow="Geography" title={item.name}>
      <Input label="地点名" value={item.name} onChange={(value) => onChange({ ...item, name: value })} />
      <Input label="层级" value={item.level} onChange={(value) => onChange({ ...item, level: value })} />
      <Input label="上级地点" value={item.parent} onChange={(value) => onChange({ ...item, parent: value })} />
      <Input label="所属势力" value={item.owner} onChange={(value) => onChange({ ...item, owner: value })} />
      <Textarea label="通行方式" value={item.route} onChange={(value) => onChange({ ...item, route: value })} />
      <Textarea label="描述" value={item.description} onChange={(value) => onChange({ ...item, description: value })} />
    </EditorFrame>
  )
}

function WorldviewEditor({
  item,
  onChange,
}: {
  item: WorldviewRule
  onChange: (item: WorldviewRule) => void
}) {
  return (
    <EditorFrame eyebrow="Worldview Rule" title={item.title}>
      <Input label="规则名" value={item.title} onChange={(value) => onChange({ ...item, title: value })} />
      <Input label="强度" value={item.strength} onChange={(value) => onChange({ ...item, strength: value })} />
      <Input label="领域" value={item.domain} onChange={(value) => onChange({ ...item, domain: value })} />
      <Textarea label="规则陈述" value={item.statement} onChange={(value) => onChange({ ...item, statement: value })} />
      <Textarea label="例外条件" value={item.exception} onChange={(value) => onChange({ ...item, exception: value })} />
      <Textarea label="代价" value={item.cost} onChange={(value) => onChange({ ...item, cost: value })} />
    </EditorFrame>
  )
}

function EditorFrame({
  eyebrow,
  title,
  children,
}: {
  eyebrow: string
  title: string
  children: ReactNode
}) {
  return (
    <div>
      <p className="micro-type text-mark">{eyebrow}</p>
      <h2 className="mt-3 font-display text-4xl">{title}</h2>
      <div className="mt-8 grid gap-5">{children}</div>
    </div>
  )
}

function Input({
  label,
  value,
  onChange,
}: {
  label: string
  value: string
  onChange: (value: string) => void
}) {
  return (
    <label className="grid gap-2">
      <span className="micro-type text-dim">{label}</span>
      <input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="border border-line bg-paper px-4 py-3 outline-none transition-colors focus:border-mark"
      />
    </label>
  )
}

function Textarea({
  label,
  value,
  onChange,
}: {
  label: string
  value: string
  onChange: (value: string) => void
}) {
  return (
    <label className="grid gap-2">
      <span className="micro-type text-dim">{label}</span>
      <textarea
        value={value}
        rows={4}
        onChange={(event) => onChange(event.target.value)}
        className="resize-y border border-line bg-paper px-4 py-3 leading-7 outline-none transition-colors focus:border-mark"
      />
    </label>
  )
}

function ConstraintPanel({
  item,
  onResolve,
}: {
  item:
    | CharacterProfile
    | EncyclopediaEntry
    | GeographyEntry
    | WorldviewRule
    | undefined
  onResolve: () => void
}) {
  const constraints = getConstraints(item)

  return (
    <aside className="border border-line p-5 lg:sticky lg:top-24 lg:self-start">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="micro-type text-dim">Rule Check</p>
          <h3 className="mt-2 font-display text-2xl">约束结果</h3>
        </div>
        {constraints.length > 0 ? (
          <AlertOctagon size={22} className="text-mark" />
        ) : (
          <CheckCircle2 size={22} className="text-dim" />
        )}
      </div>

      <div className="mt-6 grid gap-4">
        {constraints.length > 0 ? (
          constraints.map((constraint) => (
            <div key={constraint.id} className="border-l-2 border-mark bg-wash px-4 py-3">
              <p className="text-sm font-semibold">{constraint.title}</p>
              <p className="mt-2 text-sm leading-6 text-dim">{constraint.detail}</p>
              <p className="micro-type mt-3 text-mark">{constraint.target}</p>
            </div>
          ))
        ) : (
          <p className="text-sm leading-7 text-dim">当前条目没有未处理冲突。</p>
        )}
      </div>

      <button
        type="button"
        onClick={onResolve}
        className="mt-6 min-h-11 w-full border border-ink px-4 text-sm transition-colors hover:border-mark hover:text-mark"
      >
        登记破例 / 处理完成
      </button>
    </aside>
  )
}

function getConstraints(
  item:
    | CharacterProfile
    | EncyclopediaEntry
    | GeographyEntry
    | WorldviewRule
    | undefined,
) {
  if (!item || !("constraints" in item)) {
    return []
  }

  return item.constraints
}

function getItems(kind: CodexKind, snapshot: WorldbuildingSnapshot) {
  if (kind === "character") {
    return snapshot.characters.map((item) => ({
      id: item.id,
      title: item.name,
      meta: `${item.faction} · ${item.status}`,
    }))
  }
  if (kind === "encyclopedia") {
    return snapshot.encyclopedia.map((item) => ({
      id: item.id,
      title: item.term,
      meta: `${item.kind} · 可信度 ${item.confidence}`,
    }))
  }
  if (kind === "geography") {
    return snapshot.geography.map((item) => ({
      id: item.id,
      title: item.name,
      meta: `${item.level} · ${item.owner}`,
    }))
  }
  return snapshot.worldview.map((item) => ({
    id: item.id,
    title: item.title,
    meta: `${item.strength} · ${item.domain}`,
  }))
}

function collectConflicts(snapshot: WorldbuildingSnapshot) {
  const all = [
    ...snapshot.characters.flatMap((item) => item.constraints),
    ...snapshot.geography.flatMap((item) => item.constraints),
    ...snapshot.worldview.flatMap((item) => item.constraints),
  ]

  return {
    blocking: all.filter((item) => item.level === "blocking").length,
    warning: all.filter((item) => item.level === "warning").length,
    hint: all.filter((item) => item.level === "hint").length,
  }
}
