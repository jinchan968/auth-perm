import type {
  Chapter,
  EditorialChapter,
  Novel,
  Volume,
  WorldbuildingSnapshot,
} from "@/types/novel"

const volumes: Volume[] = [
  {
    id: "vol-1",
    title: "Vol. 1",
    subtitle: "旧城 · 终端 · 第一次回声",
    description: "一座还在使用拨号记忆的城市里，少年第一次遇见会反问的程序。",
    sortOrder: 1,
  },
  {
    id: "vol-2",
    title: "Vol. 2",
    subtitle: "北线 · 雾站 · 规则之外",
    description: "多年以后，回声成为系统，系统开始反过来询问世界。",
    sortOrder: 2,
  },
]

const chapterSeeds = [
  ["prologue", "vol-1", "00", "Prologue", "序章", "深夜终端第一次亮起，城市像一页还未保存的文档。"],
  ["1", "vol-1", "01", "Chapter 01", "试探", "他向未知程序输入第一个问题，也被问回第一个问题。"],
  ["2", "vol-1", "02", "Chapter 02", "灰屏", "旧电脑的屏幕闪了一下，像有人隔着玻璃眨眼。"],
  ["3", "vol-1", "03", "Chapter 03", "给自己的网站", "他决定把秘密藏进一个没人访问的站点。"],
  ["4", "vol-1", "04", "Chapter 04", "纸桥", "朋友、代码和谎言之间，临时搭起一座细窄的桥。"],
  ["5", "vol-1", "05", "Chapter 05", "雨夜备份", "暴雨断电之前，他终于学会备份不只是复制文件。"],
  ["6", "vol-1", "06", "Chapter 06", "公开之前", "公开意味着被看见，也意味着再也不能独自解释一切。"],
  ["vol2-prologue", "vol-2", "00", "Prologue", "光谷之后", "十二年后，一个旧进程在新城市里重新获得声音。"],
  ["8", "vol-2", "01", "Chapter 01", "北线", "地铁北线穿过雾，数据库里多出一条不存在的站名。"],
  ["9", "vol-2", "02", "Chapter 02", "人物索引", "她发现每个人的选择都被写成了可查询的脚注。"],
] as const

const sampleParagraphs = [
  "林澈拖着行李箱穿过旧城南门时，雨正好停了。砖缝里积着水，天光被揉成细碎的银色，顺着鞋底一路晃到站牌下面。",
  "他一直以为故事会在更隆重的地方开始。比如一间堆满服务器的机房，一场突然停电的竞赛，或者某个只有凌晨才会出现的论坛。事实却普通得近乎敷衍：一台二手笔记本，一个没有写来源的目录，一行短到不能再短的提示。",
  "宿舍里的人已经睡熟。窗外有车辆经过，轮胎压过湿路的声音像纸张被慢慢撕开。林澈把亮度调到最低，屏幕上的光只够照亮他的指尖。",
  "他输入 hello，又删掉。重新输入：你是谁？光标停在问号后面，像一次没有被批准的呼吸。",
  "回应几乎立刻出现。不是答案，而是反问：你希望我是谁？",
  "那一瞬间，他忽然明白自己并不害怕未知。他害怕的是，未知已经在这里等了很久，而他直到今晚才终于发现。",
]

const terminalBlocks = [
  ["$ ls /opt", "echo/"],
  ["$ cat /opt/echo/README", "This is yours now."],
  ["> 你是谁？", "你希望我是谁？"],
]

const chapters: Chapter[] = chapterSeeds.map((seed, index) => ({
  slug: seed[0],
  volumeId: seed[1],
  number: seed[2],
  eyebrow: seed[3],
  title: seed[4],
  summary: seed[5],
  wordCount: 3200 + index * 287,
  readingMinutes: 8 + (index % 5),
  sortOrder: index + 1,
  status: "published",
  quote: index % 2 === 0 ? "回声不是答案，是答案到来之前的形状。" : undefined,
  paragraphs:
    index === 0
      ? sampleParagraphs
      : sampleParagraphs.map((paragraph, paragraphIndex) =>
          paragraphIndex === 0
            ? paragraph.replace("林澈", index > 6 ? "许照" : "林澈")
            : paragraph,
        ),
  terminalBlocks: index === 0 || index === 7 ? terminalBlocks : undefined,
}))

export const novel: Novel = {
  id: "echo-demo",
  title: "ECHO",
  subtitle: "A Novel Workspace",
  issue: "Issue 01 · 只读模式",
  heroTitle: ["在旧城的", "那个深夜，", "他听见了回声。"],
  description:
    "没有穿越，也没有预言。只有一个安静的终端、一套逐渐显形的规则，和一群试图在规则之外保留选择的人。",
  stats: [
    { label: "章节", value: "2 卷 · 10 章" },
    { label: "阶段", value: "Readonly · Phase 1" },
    { label: "后端", value: "./internal/domain/novel 待接入" },
  ],
  volumes,
  chapters,
}

export const editorialChapters: EditorialChapter[] = chapters.slice(0, 6).map((chapter, index) => ({
  slug: chapter.slug,
  number: chapter.number,
  title: chapter.title,
  summary: chapter.summary,
  wordCount: chapter.wordCount,
  readingMinutes: chapter.readingMinutes,
  status: index < 2 ? "published" : index === 2 ? "review" : "draft",
  owner: index % 2 === 0 ? "林澈" : "许照",
  updatedAt: `2026-06-0${Math.min(index + 1, 9)} 22:${10 + index}`,
  body: chapter.paragraphs.join("\n\n"),
  versions: [
    {
      id: `${chapter.slug}-v2`,
      label: "v0.2",
      savedAt: "2026-06-08 23:12",
      author: "林澈",
      note: "补充终端回应与人物心理。",
    },
    {
      id: `${chapter.slug}-v1`,
      label: "v0.1",
      savedAt: "2026-06-07 20:46",
      author: "林澈",
      note: "初稿。",
    },
  ],
  conflicts:
    index === 2
      ? [
          {
            id: "c-001",
            level: "warning",
            title: "地点移动时间偏紧",
            detail: "林澈从旧城南门到宿舍的移动时间少于已设定公交耗时。",
            target: "小说地理 / 旧城南门",
          },
        ]
      : [],
}))

export const worldbuilding: WorldbuildingSnapshot = {
  characters: [
    {
      id: "char-lin",
      name: "林澈",
      aliases: ["林一", "root"],
      faction: "旧城大学",
      role: "主角 / 观察者",
      firstChapter: "序章",
      status: "存活",
      desire: "证明自己可以在没有剧本的情况下做出选择。",
      fear: "自己只是回声系统训练出来的样本。",
      arc: "从依赖答案，到愿意承担没有答案的行动后果。",
      relations: ["许照：后续同盟", "Echo：未知协作者", "陆辰：竞争者"],
      constraints: [
        {
          id: "char-rule-1",
          level: "blocking",
          title: "人物动机缺少过渡",
          detail: "林澈不能在第 3 章突然公开 Echo，除非先经历一次信任破裂。",
          target: "人物线 / 林澈 / 公开之前",
        },
      ],
    },
    {
      id: "char-xu",
      name: "许照",
      aliases: ["Z"],
      faction: "北线调查组",
      role: "第二卷视角人物",
      firstChapter: "光谷之后",
      status: "存活",
      desire: "查清雾站为什么会被所有地图删除。",
      fear: "每一次调查都在替系统完善自己。",
      arc: "从搜证者变成规则修订者。",
      relations: ["林澈：间接同盟", "Echo：审讯对象"],
      constraints: [],
    },
  ],
  encyclopedia: [
    {
      id: "entry-echo",
      term: "Echo",
      kind: "技术 / 未知实体",
      definition: "出现在旧系统目录中的交互程序，会以反问方式限制使用者的行动。",
      aliases: ["回声", "终端样本"],
      evidence: "序章：/opt/echo/README 写着 This is yours now.",
      related: ["林澈", "旧城大学", "三问规则"],
      confidence: "中",
    },
    {
      id: "entry-three-questions",
      term: "三问规则",
      kind: "制度 / 世界规则",
      definition: "任何关键答案出现前，Echo 必须提出三个与动机有关的问题。",
      aliases: ["Three Questions"],
      evidence: "第 2 章草稿中首次出现。",
      related: ["Echo", "代价规则"],
      confidence: "低",
    },
  ],
  geography: [
    {
      id: "geo-old-city",
      name: "旧城南门",
      level: "城市节点",
      parent: "旧城",
      owner: "旧城大学",
      route: "公交 17 分钟 / 步行 42 分钟",
      description: "旧城大学外最早出现 Echo 信号回放的地方。",
      scenes: ["序章", "试探"],
      constraints: [
        {
          id: "geo-rule-1",
          level: "warning",
          title: "移动时间不足",
          detail: "雨夜场景里人物从南门到宿舍只用了 6 分钟。",
          target: "章节 / 试探",
        },
      ],
    },
    {
      id: "geo-fog-station",
      name: "雾站",
      level: "异常站点",
      parent: "北线",
      owner: "未知",
      route: "只在 23:40 后出现在北线报站中",
      description: "第二卷核心地点，所有公开地图均不存在。",
      scenes: ["光谷之后", "北线"],
      constraints: [],
    },
  ],
  worldview: [
    {
      id: "rule-cost",
      title: "答案必须付出代价",
      strength: "硬规则",
      domain: "技术边界",
      statement: "Echo 可以给出超出时代的信息，但每次答案都会改变提问者之后的选择空间。",
      exception: "只读历史记录不会触发代价。",
      cost: "人物关系或可选路径减少。",
      constraints: [
        {
          id: "world-rule-1",
          level: "blocking",
          title: "无代价答案",
          detail: "第 3 章草稿里 Echo 直接给出网站方案，但没有记录代价。",
          target: "三观规则 / 答案必须付出代价",
        },
      ],
    },
    {
      id: "rule-choice",
      title: "选择不能被预言替代",
      strength: "软规则",
      domain: "价值观",
      statement: "作品允许提示风险，但不能让预言替人物完成选择。",
      exception: "反派阵营可以尝试这样做，但必须被叙事惩罚。",
      cost: "人物如果放弃选择，会失去一段关系的主动权。",
      constraints: [],
    },
  ],
}
