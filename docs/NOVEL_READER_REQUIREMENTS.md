# 小说阅读前端需求说明

本文档记录 `./novel` 小说阅读前端的产品目标、阶段拆分、信息架构与后端契约方向。当前前端已覆盖只读、编辑、规则约束三阶段原型；后端已新增 `./internal/domain/novel/` 领域模块承载持久化与规则冲突基础逻辑。

## 1. 目标

以 `https://echo-novel.innei.in/` 为视觉参考，建立一个适合长篇小说阅读与创作管理的前端子应用：

- 第一阶段：只读模式，提供小说列表首页、小说详情与分卷目录、章节阅读、主题切换、字号调整、上下章导航。
- 第二阶段：可编辑模式，允许维护小说、分卷、单元、章节、草稿、发布状态与版本。
- 第三阶段：规则约束编辑模式，围绕人物线、百科、地理、价值观/三观、时间线、设定冲突检查建立可编辑资料库。

## 2. 当前第一阶段范围

### 2.1 页面

- `./novel/app/page.tsx`：小说列表首页，展示已公开作品书架。
- `./novel/app/novels/[id]/page.tsx`：指定小说详情页，展示作品介绍与分卷目录。
- `./novel/app/novels/[id]/chapters/[slug]/page.tsx`：指定小说章节阅读页。
- `./novel/app/chapters/[slug]/page.tsx`：旧章节链接兼容跳转页。
- `./novel/app/not-found.tsx`：缺失章节兜底页。
- `./novel/app/error.tsx`：后端读取失败兜底页。

### 2.2 体验

- 首页保留参考站的极简顶栏和文学式大标题，用作品列表承载多本小说入口。
- 小说详情页展示作品介绍、阅读入口、终端样本和分卷目录。
- 阅读页保留居中正文、细顶栏、元信息、引用、分隔符、终端代码块、首字下沉。
- 阅读页底部固定提供“上一章 / 目录 / 下一章”快捷导航；键盘左/右方向键切换章节，`m` 返回目录。
- 只读阶段不出现编辑入口，避免读者误以为内容可修改。
- 主题切换与字号调整仅保存在浏览器本地。

### 2.3 数据边界

当前读者侧 `./novel/lib/api/novel.ts` 已接入后端公开读取接口，并在该层把后端 snake_case VO 转换为前端页面使用的视图模型。页面不直接拼接 `fetch`，后续字段或接口调整仍应优先收敛在这一层。

联调配置：

- `./novel/.env.example` 提供 `NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1` 示例。
- 如需指定租户，可配置 `NEXT_PUBLIC_TENANT_ID`；未配置时公开列表沿用后端默认租户逻辑。

已实现的公开读取接口：

- `GET /api/v1/novels`：公开小说列表。
- `GET /api/v1/novels/:id`：公开小说详情，仅返回已发布章节。
- `GET /api/v1/novels/:id/chapters/:slug`：公开章节正文，仅返回已发布章节。

新增小说不会再被本地范例内容遮挡：只要后端小说状态为 `serial` 或 `completed`，并且章节已发布，`./novel` 首页和 `/novels/[id]` 路由会读取真实数据并生成目录链接。

## 3. 第二阶段：可编辑模式

### 3.1 角色与权限

- 读者：只能阅读已发布内容。
- 作者：可新建与编辑自己拥有的小说、分卷、章节、资料条目。
- 编辑：可审核、退回、发布、锁定章节。
- 管理员：可管理全局分类、可见性、敏感规则与权限资源。

### 3.2 核心功能

- 小说管理：标题、副标题、简介、封面、连载状态、标签、默认语言。
- 分卷管理：卷名、排序、描述、发布状态。
- 单元管理：单元归属于分卷，可维护单元名、排序、描述，用于承载“卷 / 单元 / 章”的长篇结构。
- 章节管理：正文编辑、摘要、字数统计、阅读时长、排序、草稿/已发布/归档状态。
- 版本管理：保存历史版本，支持查看差异、回滚、复制为新草稿。
- 发布流程：草稿保存、预览、提交审核、发布、撤回。
- 编辑体验：Markdown 或富文本二选一，优先支持 Markdown + 预览；保存时由后端做结构化校验。

### 3.3 当前前端落点

- `./novel/app/studio/page.tsx`：章节可编辑工作台入口。
- `./novel/components/studio-workspace.tsx`：章节列表、正文编辑、状态流转、版本记录、规则冲突提示。
- 当前为本地状态模拟，尚未持久化；后续接入 Go 后端后由 `./novel/lib/api/novel.ts` 替换数据源。

### 3.4 已实现接口

管理接口统一挂在 `./api/v1/novel-admin`，需要登录并经过 API 权限校验：

- `GET /api/v1/novel-admin/mine`：我的小说列表。
- `POST /api/v1/novel-admin`：创建小说。
- `GET /api/v1/novel-admin/:id/manage`：管理端小说详情，包含草稿章节。
- `PUT /api/v1/novel-admin/:id`：更新小说，可将小说状态从 `serial` 改为 `completed`。
- `DELETE /api/v1/novel-admin/:id`：删除小说。
- `GET /api/v1/novel-admin/:id/volumes`：分卷列表。
- `POST /api/v1/novel-admin/:id/volumes`：创建分卷。
- `PUT /api/v1/novel-admin/volumes/:id`：更新分卷。
- `GET /api/v1/novel-admin/:id/units`：单元列表。
- `POST /api/v1/novel-admin/:id/units`：创建单元。
- `PUT /api/v1/novel-admin/units/:id`：更新单元。
- `GET /api/v1/novel-admin/:id/chapters`：章节列表。
- `POST /api/v1/novel-admin/:id/chapters`：创建章节。
- `POST /api/v1/novel-admin/:id/chapters/import-md`：导入 Markdown 章节，支持 JSON 文本和 multipart `.md` 文件。
- `POST /api/v1/novel-admin/import-md-bundle/inspect`：识别 Markdown zip / 文件树目录结构，不写数据库。
- `POST /api/v1/novel-admin/:id/import-md-bundle`：批量导入 Markdown 文件树，支持 JSON 文件数组和 multipart `.zip` 文件。
- `GET /api/v1/novel-admin/chapters/:id`：章节详情。
- `PUT /api/v1/novel-admin/chapters/:id`：更新章节并保存版本快照。
- `PATCH /api/v1/novel-admin/chapters/:id/status`：更新章节状态；发布前会检查未处理阻断级冲突。
- `PATCH /api/v1/novel-admin/:id/chapters/status`：批量更新章节状态，请求体传 `ids/status/note`；发布前会检查未处理阻断级冲突，状态更新不返回正文。
- `GET /api/v1/novel-admin/chapters/:id/versions`：章节版本列表。

### 3.5 `./ui` 管理后台发布路径

- `./ui/app/novels/page.tsx`：小说列表页，可进入明细或导入页。
- `./ui/app/novels/[id]/page.tsx`：小说明细页，按卷、单元、章节树状展示导入结果，支持折叠卷/单元。
- 小说明细页提供“设为完结 / 恢复连载”操作，调用 `PUT /api/v1/novel-admin/:id` 更新小说状态，不改变章节发布状态。
- Markdown zip 导入后的章节默认保持草稿态，避免误发布。
- 管理端和阅读端都必须显式按 `volume.sort_order -> unit.sort_order -> chapter.sort_order` 展示章节；范围选择使用当前展开可见的排序结果。
- 管理员在小说明细页勾选部分章节后，可执行“发布选中”或“退回草稿”；连续章节可开启范围选择，从未选章节拖动为批量勾选，从已选章节拖动为批量反选；已发布章节不可被选中。发布操作使用 `PATCH /api/v1/novel-admin/:id/chapters/status` 批量提交章节 IDs。
- 阅读端只展示已发布章节；如果小说本身不是连载中或已完结，即使章节已发布也不会出现在公开阅读页。

## 4. 第三阶段：规则约束编辑模式

第三阶段的核心不是“资料展示”，而是把资料变成写作时可校验的约束。

### 4.1 人物线

- 人物档案：姓名、别名、阵营、身份、首次登场章节、当前状态。
- 关系图谱：亲属、盟友、对手、师承、债务、情感关系。
- 人物弧线：目标、恐惧、误解、关键选择、阶段变化。
- 约束校验：死亡人物不得在后续章节无解释出现；人物称谓随关系变化提示不一致。

### 4.2 百科

- 条目类型：组织、技术、物品、事件、制度、术语、历史。
- 条目字段：定义、别名、出现章节、相关人物、证据片段、可信度。
- 约束校验：同名条目冲突、设定定义前后矛盾、术语未解释即大量出现。

### 4.3 小说地理

- 地点档案：国家/城市/街区/建筑/房间等层级。
- 地理关系：包含、相邻、距离、通行方式、所属势力。
- 场景记录：章节发生地点、时间、在场人物、天气与环境。
- 约束校验：人物跨地点移动时间不合理；地点归属和章节描述冲突。

### 4.4 小说三观与规则

- 世界规则：物理规则、社会规则、技术边界、超自然规则。
- 价值观主题：作品立场、人物信念、叙事禁区、必须付出代价的行为。
- 规则强度：硬规则、软规则、可破例规则；每条规则记录例外条件。
- 约束校验：关键能力无代价使用；世界规则被破坏但没有叙事解释；人物选择偏离既定信念。

### 4.5 写作校验工作流

1. 章节保存时抽取人物、地点、时间、事件、术语。
2. 与人物线、百科、地理、三观规则做交叉校验。
3. 返回冲突列表，按阻断、警告、提示分级。
4. 作者可修正文稿、更新资料库，或登记“有意破例”的解释。
5. 发布前必须处理所有阻断级冲突。

### 4.6 当前前端落点

- `./novel/app/codex/page.tsx`：规则约束工作台入口。
- `./novel/components/codex-workspace.tsx`：人物线、百科、地理、小说三观四类资料编辑。
- 约束结果以阻断、警告、提示聚合展示；条目级支持登记破例/处理完成的本地交互。

后端已实现的规则资料库接口：

- `GET /api/v1/novel-admin/:id/codex`：资料条目列表，支持 `kind=character|encyclopedia|geography|worldview`。
- `POST /api/v1/novel-admin/:id/codex`：创建资料条目。
- `PUT /api/v1/novel-admin/:id/codex/:entryId`：更新资料条目。
- `GET /api/v1/novel-admin/:id/conflicts`：规则冲突列表。
- `POST /api/v1/novel-admin/:id/conflicts`：创建规则冲突。
- `PATCH /api/v1/novel-admin/conflicts/:id`：处理规则冲突，支持 resolved/waived。

## 5. Markdown 导入模式

当前采用“导入模式”：`.md` 文件只作为导入来源，导入成功后章节正文、元信息和版本流转以数据库为准，不直接在源 `.md` 文件上修改。

管理入口位于主后台 `./ui/app/novels/import/page.tsx`：

- 选择或创建目标小说。
- 上传 zip 后先调用预检接口识别目录结构，展示卷、单元、章节、slug、字数和跳过原因。
- 确认后调用批量导入接口写入数据库。
- 单章 Markdown 可在已有分卷下导入。
- 上传 zip / md 当前由后端直接从请求流读入内存处理，不保存到磁盘，因此导入完成后没有临时文件残留；若后续改为落盘缓存，必须在成功、失败和 panic 兜底路径清理临时文件。
- 目录识别当前采用确定性路径规则；可在此预检链路上扩展 OpenCode AI 辅助识别，但 AI 结果必须先展示给用户确认，不应绕过人工确认直接写库。

支持两种请求方式：

- JSON：`{"volume_id":"...","markdown":"# 序章\n\n正文..."}`
- multipart：字段 `volume_id` + 文件字段 `file`，文件内容为 Markdown。

Markdown 可选 front matter：

```markdown
---
title: 序章
slug: prologue
number: "00"
summary: 第一次回声。
status: draft
sort_order: 1
---

正文内容...
```

导入优先级：请求字段优先于 front matter；若都没有 `title`，则使用 Markdown 第一个一级标题。

### 5.1 单章导入

单章导入要求显式传入 `volume_id`，可选传入 `unit_id`：

- JSON：`{"volume_id":"...","unit_id":"...","markdown":"# 序章\n\n正文..."}`
- multipart：字段 `volume_id`、可选字段 `unit_id` + 文件字段 `file`，文件内容为 Markdown。

### 5.2 文件树批量导入

批量导入用于支持小说已经按“卷 / 单元 / 章”拆分为多个 `.md` 文件的场景。目录约定如下：

```text
第一卷-旧城/
  01-初遇/
    001-序章.md
    002-试探.md
  02-远行/
    003-夜车.md
第二卷-群星/
  001-抵达.md
```

解析规则：

- `卷/单元/章.md`：自动创建或复用分卷、单元和章节。
- `卷/章.md`：自动创建或复用分卷，章节不绑定单元。
- `章.md`：自动归入“默认卷”。
- 目录或文件名前缀为纯数字时会用作排序与章节号，例如 `001-序章.md` 的章节号为 `001`、标题回退为 `序章`。
- Markdown front matter 的 `title`、`slug`、`number`、`summary`、`status`、`sort_order` 优先于文件名推断。
- 同一本小说内相同 `slug` 再次导入会更新已有章节并保存版本快照；新 `slug` 会创建章节。

支持两种请求方式：

- JSON：`{"files":[{"path":"第一卷/01-初遇/001-序章.md","content":"# 序章\n\n正文..."}]}`
- multipart：文件字段 `file` 上传 `.zip`，zip 内仅读取 `.md` 文件。

## 6. 后端落点建议

后端已新增领域模块 `./internal/domain/novel/`，并按当前项目分层落位：

- `constant`：章节状态、条目类型、规则强度、冲突等级。
- `dm`：小说、分卷、单元、章节、版本、人物、地点、百科、规则、冲突记录。
- `dto`：创建/更新/查询请求。
- `vo`：前端展示结构、枚举校验器、dm 到 vo 转换器。
- `repo`：数据访问。
- `service`：发布流程、版本管理、规则校验。
- `handler`：HTTP 入参绑定与响应。
- `module.go`：领域注册。

同时已接入 `./internal/container/` 装配与 `./internal/controller/http/route.go` 路由注册；数据库迁移位于 `./migrations/000033_create_novels.sql`。

## 7. 验收标准

- 第一阶段：无需登录即可阅读样例小说，目录、章节、主题、字号、上下章导航可用。
- 第二阶段：作者可以在 `./novel/app/studio/page.tsx` 完成章节从草稿到发布的前端闭环，并保留版本记录；接入后端后需要补登录权限与持久化。
- 第三阶段：作者可以在 `./novel/app/codex/page.tsx` 编辑规则库并看到结构化冲突提示；接入后端后需要补权限控制与真实校验。
- 所有阶段：前端 API 调用集中在 `./novel/lib/api/`，后端业务规则不进入页面组件。
