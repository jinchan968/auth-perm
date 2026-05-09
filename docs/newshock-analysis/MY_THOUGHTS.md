# 分析思路与推理过程

---

## 1. 初始挑战

目标网站 https://news.moneych.top/ 是一个 **JavaScript 单页应用 (SPA)**，直接用 WebFetch 只能拿到一个标题：

```
Newshock — 事件驱动主题投研雷达
```

这是因为 Next.js 的页面内容由客户端 JavaScript 渲染，WebFetch 无法执行 JS。

---

## 2. 突破口: Next.js RSC (React Server Components)

关键发现：Next.js 15 使用 **RSC (React Server Components)**，会将服务端数据直接嵌入 HTML 的 `<script>` 标签中：

```html
<script>self.__next_f.push([1,"7:[\"$\",\"$L12\",null,{\"data\":{...}}]"])</script>
```

这意味着：
1. 页面 HTML 本身就包含了完整的 SSR 数据
2. 不需要执行 JavaScript 就能获取数据
3. 数据格式是 JSON，只是被转义了

---

## 3. 分析步骤

### Step 1: 获取原始 HTML

```bash
curl -s -L 'https://news.moneych.top/'
```

返回 40KB HTML，包含：
- `<head>` 中的 meta 标签 (SEO 信息)
- `<script>` 标签中的 JS bundle 引用
- RSC 数据 payload

### Step 2: 解析 meta 标签

从 meta 标签获取项目定位：
- `<title>`: "Newshock — 事件驱动主题投研雷达"
- `<meta name="description">`: "实时追踪市场主线、催化事件、标的暴露 — A股 / 美股 / 港股事件驱动主题投研工具。Polymarket 概率 × AI 深度分析 × 历史 archetype 模板。"
- `<meta name="keywords">`: "thematic investing, event-driven, polymarket, archetype..."

### Step 3: 提取 JS Bundle 列表

从 `<script src="...">` 标签识别出 Next.js 的 chunk 文件：
- `app/layout-*.js` — 布局组件
- `app/page-*.js` — 首页组件
- `app/themes/page-*.js` — 主题页组件
- 共享依赖 chunks

### Step 4: 下载并分析 JS 源码

下载关键 JS 文件：
- `layout-c8910d7b88ab3115.js` (17.7KB) — 包含 AppLayout、侧边栏、顶部栏、搜索弹窗、i18n 字典
- `page-0113e25affcad2df.js` (19.4KB) — 包含 Radar 首页组件

从 JS 源码中提取：
- **路由结构**: 6 个导航项 (/, /themes, /tickers, /markets, /edge, /events)
- **i18n 字典**: 67 个中英文翻译键
- **antd 主题 Token**: 深色/浅色主题的颜色配置
- **组件逻辑**: 数据展示、搜索、Watchlist 等

### Step 5: 解析 RSC 数据

使用 Python 脚本从 HTML 中提取嵌入的数据：

```python
# 1. 找到所有 RSC script 标签
scripts = re.findall(r'<script>self\.__next_f\.push\(\[1,"(.*?)"\]\)</script>', html)

# 2. 反转义
unesc = s.replace('\\"', '"').replace('\\n', '\n')

# 3. 找到 "themes":[ 开始的位置
# 4. 用括号匹配找到完整的 JSON 对象
# 5. json.loads 解析
```

提取结果：
- **themes**: 20 条主题 (按 strength 排序)
- **tickers**: 20 条股票 (按 hot_score 排序)
- **events**: 30 条事件 (按时间排序)
- **regime**: 市场环境判断
- **stats**: 总计 517 主题 / 6702 股票 / 32372 事件
- **freshness**: 各类数据的最后更新时间

### Step 6: 探测 API 端点

```bash
curl 'https://news.moneych.top/api/search?q=AI'
```

发现 `/api/search` 返回 JSON：
```json
{"themes":[], "tickers":[], "events":[{"id":12728, "title":"..."}]}
```

其他端点 (/api/themes, /api/events 等) 返回空响应，可能需要认证或特定 headers。

### Step 7: 下载子页面

- `/themes` — 主题列表页 (14KB)
- `/tickers` — 股票列表页 (14KB)
- `/events` — 事件列表页 (14KB)
- `/markets` — 盘口页 (333KB，大量数据)
- `/edge` — 边缘信号页 (202KB)
- `/themes/13` — 主题详情页
- `/tickers/AAPL` — 股票详情页

---

## 4. 关键推理

### 4.1 数据管线推断

从事件数据的 `channel` 字段识别出 5 个新闻源：
- 金融快报 Financial Express (主要来源)
- 参考消息 XHQ
- 驻华外电 PD China
- 财经日报 Finance News Daily
- CNBeta 科技

从 `freshness` 数据推断更新频率：
- 新闻采集: 实时/分钟级
- 事件提取: 实时/分钟级
- 评分计算: 小时级
- AI 深度分析: 天级

### 4.2 主题聚类推断

主题的 `classification_confidence` 统一为 0.85（部分为 0.7 或 0.6），说明：
- 分类模型可能是批量运行的
- 置信度阈值设在 0.6 以上
- 新主题默认置信度 0.85

### 4.3 强度计算推断

`strength` 值范围很大 (350-5500)，`strength_norm` 范围 0-100：
- strength 可能基于事件数 × 重要性 × 时效性
- strength_norm 是百分位排名
- Top 1 主题 strength_norm=100，说明是相对排名

### 4.4 前端架构推断

- 使用 Next.js 15 App Router (从 `_next/static/chunks/app/` 路径确认)
- 使用 RSC 进行 SSR (从 `self.__next_f.push` 确认)
- 使用 Ant Design 5 (从 JS 中的 `antd` 组件引用确认)
- 自实现 i18n (从 JS 中的翻译字典确认，未使用 next-intl)

---

## 5. 未解之谜

| 问题 | 线索 | 推测 |
|------|------|------|
| 后端用什么语言? | 无直接证据 | 可能是 Python (NLP 生态) 或 Go (性能) |
| 数据库是什么? | 无直接证据 | PostgreSQL (关系型) + ClickHouse (分析型) |
| AI 模型是什么? | description 字段是 AI 生成的 | 可能用 LLM API (GPT-4/Claude) |
| Polymarket 数据如何集成? | freshness 有 `latest_poly_ingest` | 定时采集 Polymarket API |
| 管理后台功能? | `/admin/login` 存在 | 主题管理、事件审核、数据源配置 |
| Markets 页面内容? | 333KB HTML | 可能包含实时行情数据或图表 |
| Edge 页面内容? | 202KB HTML | 可能包含边缘信号检测算法结果 |

---

## 6. 方法论总结

```
1. curl 获取原始 HTML
2. 解析 meta 标签 → 项目定位
3. 提取 JS bundle 列表 → 技术栈识别
4. 下载 JS 源码 → 组件结构、路由、i18n、主题配置
5. 解析 RSC 数据 → 数据模型、样本数据
6. 探测 API 端点 → 接口设计
7. 下载子页面 → 功能分析
8. 综合推理 → 架构、管线、算法
```

**核心洞察**: Next.js RSC 是逆向分析的金矿 — 服务端数据直接嵌入 HTML，无需执行 JS 即可获取完整数据结构和样本数据。
