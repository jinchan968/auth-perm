# 文档总览

本文档用于说明 `[GoBase]/docs/` 的组织方式、阅读入口和维护约定。

## 推荐阅读顺序
1. [`AGENTS.md`](../AGENTS.md) — 团队共享的最小规则、命令和索引。
2. [`ARCHITECTURE.md`](./ARCHITECTURE.md) — 先理解分层边界、依赖装配和权限链路。
3. [`DEVELOPMENT.md`](./DEVELOPMENT.md) — 再看构建、测试、联调与提交流程。
4. [`AGENT_WORKFLOWS.md`](./AGENT_WORKFLOWS.md) — 多代理/多阶段任务如何拆分、交接和验收。

## 文档地图

### 核心规范
- [`ARCHITECTURE.md`](./ARCHITECTURE.md) — 后端 DDD 分层、前端 App Router 分层、权限数据流、常见改动落点。
- [`DEVELOPMENT.md`](./DEVELOPMENT.md) — 已验证命令、开发流程、测试建议、联调检查项。
- [`AGENT_WORKFLOWS.md`](./AGENT_WORKFLOWS.md) — 借鉴 Claude folder / harness agent 的协作方式，约束 planner / implementer / reviewer 的输入输出。

### 外部参考
- [`REFERENCES.md`](./REFERENCES.md) — 收录与文档规范、Claude folder、harness agent 相关的外部参考链接。

### 业务专题
- [`PERMISSION_SYSTEM_IMPLEMENTATION.md`](./PERMISSION_SYSTEM_IMPLEMENTATION.md) — 权限系统实现说明。
- [`USER_MANAGEMENT_IMPLEMENTATION.md`](./USER_MANAGEMENT_IMPLEMENTATION.md) — 用户管理实现说明。
- [`USER_MANAGEMENT_QUICKSTART.md`](./USER_MANAGEMENT_QUICKSTART.md) — 用户管理快速开始。

### 历史排查 / 临时专题
- [`IMPLEMENTATION_SUMMARY.md`](./IMPLEMENTATION_SUMMARY.md)
- [`FIX_NEW_USER_BUTTON_ERROR.md`](./FIX_NEW_USER_BUTTON_ERROR.md)

## 文档分层约定
- `AGENTS.md`：只放高频、稳定、必须遵守的共享规则，控制在短文本范围内。
- `AGENTS.local.md`：只放本地偏好、隐私相关约束、个人环境差异。
- `docs/*.md`：放详细规范、架构说明、工作流、专题设计和较长说明。
- 一次性排障记录可留在 `docs/` 根目录；若内容转化为长期规则，需回写到核心规范文档。

## 维护规则
1. 新增长文档时，先判断是“长期规范”还是“阶段性记录”。
2. 长期规范优先写入 `ARCHITECTURE`、`DEVELOPMENT`、`AGENT_WORKFLOWS`；业务实现说明则写入对应专题文档。
3. 新增文档后，需要同步更新 `AGENTS.md` 的快速入口，必要时补充本页索引。
4. 文档中的项目路径统一写成 `[GoBase]/相对路径`，不要写本机绝对路径。
5. 当认证、权限、租户、前后端 API 契约发生变化时，必须至少检查：
   - `AGENTS.md` 是否需要补充入口或规则；
   - `ARCHITECTURE.md` 的链路描述是否过时；
   - 对应专题文档是否需要更新接口、资源标识或流程说明。

## 当前仓库特别说明
- 根目录 `README.md` 已收敛为“项目入口 + 常用命令 + 文档索引”的轻量说明；更细的架构、流程与协作规则仍以本目录文档为准。
- 当前仓库已确认的规范入口是 `AGENTS.md + docs/*.md`，而不是隐藏目录形式的 `.claude/`；但协作思想仍可沿用“共享规则短、长说明下沉、本地偏好分离”的做法。

