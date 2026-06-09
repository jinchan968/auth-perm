# 小说前端 Cloudflare 部署清单

本文用于补充 `./docs/DEPLOYMENT.md` 中已有的 Cloudflare Pages 部署说明，聚焦 `./novel` 小说阅读前端。当前项目已有两个前端入口：

- `./ui`：管理后台，面向登录用户，包含小说导入、草稿审核、章节发布等能力。
- `./novel`：公开阅读站，面向读者，不要求登录，只展示已发布内容。

部署时建议把 `./ui` 和 `./novel` 配成两个 Cloudflare Pages 项目或至少两个独立域名，避免管理端的认证、Cloudflare Access、缓存策略影响公开阅读站。

## 一、部署前要准备什么

### 1. Cloudflare 侧

- 已有 Cloudflare 账号，并授权当前 Git 仓库。
- 为小说阅读站准备独立 Pages 项目名，例如 `{#novelProjectName}`。
- 为小说阅读站准备公开域名，例如 `novel.{#domain}` 或 Cloudflare 默认域名 `https://{#novelProjectName}.pages.dev`。
- 确认小说阅读站不要启用 Cloudflare Access 登录保护；如果 `./ui` 已启用 Access，`./novel` 需要独立策略或独立项目。
- 确认生产分支，例如 `main`。

### 2. 后端侧

- 后端 API 已部署并可通过 HTTPS 访问，例如 `https://api.{#domain}/api/v1`。
- 数据库迁移已完成，至少包含小说相关表、公开阅读接口和管理端权限资源。
- 后端公开阅读接口不要求登录：
  - `GET /api/v1/novels`
  - `GET /api/v1/novels/:id`
  - `GET /api/v1/novels/:id/chapters/:slug`
- 公开阅读接口只返回已发布内容：
  - 小说应处于可公开展示状态，例如连载或完结。
  - 章节必须是已发布状态。
  - 草稿、未发布章节、导入临时文件内容不能出现在公开接口响应中。
- 后端 `CORS_ORIGINS` 加入小说阅读站域名。即使当前 `./novel` 主要通过服务端渲染读取 API，也建议同步加上，避免后续客户端交互扩展时踩跨域问题。

示例：

```bash
CORS_ORIGINS=https://{#uiProjectName}.pages.dev,https://{#novelProjectName}.pages.dev,https://novel.{#domain}
```

### 3. 小说内容侧

- 小说已通过 `./ui` 管理后台导入。
- 导入后的小说处于草稿态时，不会在 `./novel` 首页展示。
- 在 `./ui` 小说详情页检查卷、单元、章的识别顺序。
- 使用章节树批量选择要发布的章节，确认已发布章节不可再次选中。
- 将需要公开的章节发布。
- 如小说已经完结，在 `./ui` 小说详情页将状态从连载切换为完结。

## 二、当前 `./novel` 的部署形态

`./novel` 是 Next.js App Router 应用，当前页面使用动态渲染读取后端公开接口：

- `./novel/app/page.tsx`
- `./novel/app/novels/[id]/page.tsx`
- `./novel/app/novels/[id]/chapters/[slug]/page.tsx`

因此不要按纯静态站点直接部署普通 `.next` 目录。参考 `./ui` 的现有做法，Cloudflare Pages 上应使用 Next.js 适配构建。

当前 `./ui` 已具备：

- `./ui/package.json` 中包含 `@cloudflare/next-on-pages`。
- `./ui/wrangler.toml` 中配置了 `nodejs_compat`。
- `./docs/DEPLOYMENT.md` 中已有 Cloudflare Pages 的构建命令与输出目录。

`./novel` 在部署前需要补齐同类配置：

- 在 `./novel/package.json` 增加 `@cloudflare/next-on-pages` 开发依赖，或在 Cloudflare 构建命令中使用 `npx @cloudflare/next-on-pages@1`。
- 新增 `./novel/wrangler.toml`，至少包含：

```toml
compatibility_flags = ["nodejs_compat"]
```

同时，所有非静态页面必须声明 Edge Runtime，否则 `@cloudflare/next-on-pages` 会在构建时拒绝生成 Pages 产物：

```ts
export const runtime = "edge"
```

## 三、Cloudflare Pages 配置

在 Cloudflare Dashboard 中创建新的 Pages 项目：

| 配置项 | 建议值 |
|--------|--------|
| Framework preset | `Next.js` |
| Root directory | `novel` |
| Build command | `npm install -g pnpm && pnpm install --frozen-lockfile && pnpm build && npx @cloudflare/next-on-pages@1` |
| Build output directory | `.vercel/output/static` |
| Deploy command | 留空 |
| Production branch | `main` |

如果 Cloudflare 构建环境未启用合适的 Node 版本，在 Pages 项目的环境变量中增加：

| 变量名 | 建议值 | 说明 |
|--------|--------|------|
| `NODE_VERSION` | `20` | 避免默认 Node 版本过旧 |
| `PNPM_VERSION` | `9.0.0` | 与 `./novel/package.json` 中 `packageManager` 对齐 |

Cloudflare Pages 的 Next.js 构建命令和输出目录以官方文档为准；截至本文编写时，Cloudflare Pages 文档给出的 Next.js 构建命令为 `npx @cloudflare/next-on-pages@1`，输出目录为 `.vercel/output/static`。

## 四、`./novel` 环境变量

在 Cloudflare Pages 项目的 Settings → Variables and Secrets 中配置 Production 和 Preview 环境变量。

| 变量名 | 必填 | 示例 | 说明 |
|--------|------|------|------|
| `NEXT_PUBLIC_API_URL` | 是 | `https://api.{#domain}/api/v1` | 后端 API 基地址，必须带 `/api/v1` |
| `NEXT_PUBLIC_TENANT_ID` | 否 | `default` | 多租户场景下指定租户；不设置则不附加 `tenant_id` 查询参数 |

注意：`NEXT_PUBLIC_*` 会进入前端构建产物，不要放密钥。公开阅读站不需要登录态，也不应该配置管理端 token。

## 五、部署前本地检查

在提交部署前，本地至少执行：

```bash
cd ./novel
pnpm install --frozen-lockfile
pnpm lint
pnpm type-check
pnpm build
```

如果要模拟 Cloudflare Pages 构建，再执行：

```bash
cd ./novel
npx @cloudflare/next-on-pages@1
```

后端公开接口检查：

```bash
curl "https://api.{#domain}/api/v1/novels"
curl "https://api.{#domain}/api/v1/novels/{#novelId}"
curl "https://api.{#domain}/api/v1/novels/{#novelId}/chapters/{#chapterSlug}"
```

检查重点：

- 未登录请求应返回公开小说数据。
- 首页列表不应返回草稿小说。
- 小说详情不应返回未发布章节。
- 章节详情不应返回未发布正文。
- 章节顺序应符合卷、单元、章的排序。

## 六、上线验收清单

### 读者访问

- 打开 `https://{#novelProjectName}.pages.dev` 或正式域名，不需要登录即可进入小说列表。
- 首页展示小说列表，不直接展开某本小说的完整目录。
- 点击小说进入详情页后能看到目录。
- 点击章节进入阅读页后能看到正文。
- 阅读页底部快捷操作包含上一章、目录、下一章。
- 第一章只禁用上一章；最后一章只禁用下一章；中间章节两侧导航都可用。
- 移动端宽度下首页、详情页、章节页均不溢出、不遮挡正文。

### 内容边界

- 新导入但未发布的小说不出现在公开阅读站。
- 草稿章节不出现在公开目录中。
- 已发布章节内容正常显示。
- 小说从连载改为完结后，阅读站状态展示同步更新。

### 运维边界

- `./ui` 管理后台仍要求登录。
- `./novel` 阅读站不要求登录，也不被 Cloudflare Access 拦截。
- 后端日志不打印小说正文。
- Cloudflare Pages 生产部署和预览部署的环境变量都已配置。
- 后端 `CORS_ORIGINS` 包含小说正式域名和预览域名。

## 七、回滚与故障排查

| 现象 | 检查项 |
|------|--------|
| Pages 构建失败 | `Root directory` 是否为 `novel`；是否能安装 pnpm；是否执行了 `npx @cloudflare/next-on-pages@1` |
| 部署后 500 | `NEXT_PUBLIC_API_URL` 是否正确；后端 API 是否 HTTPS 可访问；Cloudflare 构建日志是否缺少 Node 兼容配置 |
| 首页为空 | 后端是否已有处于连载或完结状态的小说；章节是否已发布；租户参数是否正确 |
| 小说详情有目录但章节打不开 | 章节 slug 是否与公开接口一致；章节是否已发布 |
| 公开站要求登录 | 是否误用了 `./ui` 域名、Cloudflare Access 策略或管理端路由 |
| 预览环境正常、生产异常 | Production 和 Preview 环境变量是否分别配置 |

Cloudflare Pages 支持在项目的 Deployments 页面回滚到之前的成功部署。回滚前先确认是否只是环境变量或后端 CORS 配置问题，避免回滚前端但问题仍在后端。

## 八、参考资料

- `./docs/DEPLOYMENT.md`：当前项目已有部署指南。
- `./ui/wrangler.toml`：现有 UI 的 Cloudflare Pages Node 兼容配置。
- `./ui/package.json`：现有 UI 的 `@cloudflare/next-on-pages` 依赖参考。
- Cloudflare Pages Build configuration: https://developers.cloudflare.com/pages/configuration/build-configuration/
- Cloudflare Pages Next.js guide: https://developers.cloudflare.com/pages/framework-guides/nextjs/
- Cloudflare Pages environment variables: https://developers.cloudflare.com/pages/functions/bindings/#environment-variables
