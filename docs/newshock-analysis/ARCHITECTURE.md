# Newshock 技术架构分析

---

## 1. 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        Nginx 1.20.1                         │
│                   (反向代理 + HTTPS + 静态资源)                │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                    Next.js 15 (Node.js)                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐ │
│  │ App Router│  │ RSC SSR  │  │ API Routes│  │ Static Gen  │ │
│  │ (Pages)  │  │ (数据注入) │  │ (/api/*) │  │ (OG Image)  │ │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────────────┘ │
└────────┼─────────────┼─────────────┼────────────────────────┘
         │             │             │
┌────────▼─────────────▼─────────────▼────────────────────────┐
│                    后端数据服务                                │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐ │
│  │ 新闻采集管线  │  │ NLP 处理引擎  │  │ 主题评分/聚类引擎   │ │
│  │ (多源 RSS/API)│  │ (NER+分类)   │  │ (强度/趋势计算)    │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬─────────────┘ │
└─────────┼─────────────────┼─────────────────┼───────────────┘
          │                 │                 │
┌─────────▼─────────────────▼─────────────────▼───────────────┐
│                      数据库层                                 │
│         PostgreSQL / ClickHouse (推测，未确认)                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐ │
│  │ themes   │  │ tickers  │  │ events   │  │ news_raw    │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                    外部数据源                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐ │
│  │ Polymarket│  │ 金融快报  │  │ 参考消息  │  │ 驻华外电    │ │
│  │ (概率数据) │  │ CNBeta   │  │ 财经日报  │  │ 更多...     │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 前端架构

### 2.1 Next.js App Router 结构

```
app/
├── layout.tsx              # AppLayout (侧边栏 + 顶部栏 + 搜索弹窗)
├── page.tsx                # Radar 首页 (SSR 数据注入)
├── themes/
│   ├── page.tsx            # 主题列表页
│   └── [id]/page.tsx       # 主题详情页
├── tickers/
│   ├── page.tsx            # 股票列表页
│   └── [symbol]/page.tsx   # 股票详情页
├── events/
│   └── page.tsx            # 事件列表页
├── markets/
│   └── page.tsx            # 盘口数据页
├── edge/
│   └── page.tsx            # 边缘信号页
├── changelog/
│   └── page.tsx            # 更新日志
├── api/
│   ├── search/route.ts     # 全局搜索 API
│   ├── themes/route.ts     # 主题 API
│   ├── events/route.ts     # 事件 API
│   ├── tickers/route.ts    # 股票 API
│   └── docs/               # API 文档
├── admin/
│   └── login/              # 管理后台登录
├── error.tsx               # 错误页面
└── not-found.tsx           # 404 页面
```

### 2.2 JS Bundle 分析

| Chunk | 文件 | 大小 | 内容 |
|-------|------|------|------|
| layout | `app/layout-c8910d7b88ab3115.js` | 17.7KB | AppLayout, 侧边栏, 顶部栏, 搜索弹窗, i18n, antd 主题配置 |
| page | `app/page-0113e25affcad2df.js` | 19.4KB | Radar 首页组件, 统计卡片, 事件流, 主题/股票领涨榜 |
| themes | `app/themes/page-0c7734f2ea2a2a5e.js` | 13.4KB | 主题列表页组件 |
| 3997 | `3997-d722117e863861f3.js` | - | 共享依赖 |
| 388 | `388-444152af4dd703e5.js` | - | 共享依赖 |
| 2277 | `2277-18ea9e7f47625a6d.js` | - | 共享依赖 |
| 27 | `27-665c860534195b66.js` | - | 共享依赖 |
| 6548 | `6548-8ccaf64cfc932a83.js` | - | 共享依赖 |

### 2.3 数据传递机制

首页使用 **React Server Components (RSC)** 将数据直接嵌入 HTML：

```javascript
// 数据通过 self.__next_f.push 注入
self.__next_f.push([1, "7:[\"$\",\"$L12\",null,{\"data\":{...}}]"])

// 客户端组件通过 props 接收
function Page({ data }) {
  const { themes, tickers, events, regime, stats, freshness } = data;
  // ...
}
```

**优势**: 首页无需客户端 API 调用，首屏渲染速度极快。

### 2.4 主题系统

使用 CSS 自定义属性 + antd ConfigProvider 实现双主题：

```javascript
// 深色主题 Token
const darkToken = {
  colorPrimary: "#6366f1",
  colorBgContainer: "#14141f",
  colorBgElevated: "#1a1a2e",
  colorBorderSecondary: "rgba(255,255,255,0.06)",
  colorText: "#e4e4ed",
  colorTextSecondary: "#8888a0",
  borderRadius: 8,
};

// 浅色主题 Token
const lightToken = {
  colorPrimary: "#4f46e5",
  borderRadius: 8,
};
```

CSS 变量命名: `--nshock-*` (28 个自定义变量)

持久化: `localStorage('theme-mode')` + cookie `theme-mode`

### 2.5 国际化

自实现 i18n，67 个翻译键，支持 zh/en：

```javascript
// 翻译字典 (部分)
const dict = {
  radar: { zh: "雷达", en: "Radar" },
  themes: { zh: "主题", en: "Themes" },
  tickers: { zh: "股票", en: "Tickers" },
  events: { zh: "事件", en: "Events" },
  edge: { zh: "边缘信号", en: "Edge" },
  markets: { zh: "盘口", en: "Markets" },
  // ... 共 67 个
};

function tt(key, lang) {
  return dict[key]?.[lang] ?? key;
}
```

持久化: `localStorage('newshock-lang')` + cookie `locale`

---

## 3. 数据管线 (后端推测)

### 3.1 新闻源

从事件数据的 `channel` 字段识别出以下新闻源：

| 频道 | 英文名 | 类型 |
|------|--------|------|
| 金融快报 | Financial Express | 主要来源，覆盖最广 |
| 参考消息 | XHQ | 国际新闻 |
| 驻华外电 | PD China | 外电中文报道 |
| 财经日报 | Finance News Daily | 财经新闻 |
| CNBeta 科技 | CNBeta | 科技新闻 |

### 3.2 处理管线

```
[新闻源采集]
    │
    ▼
[原始新闻存储] ──→ news_raw 表
    │
    ▼
[NER 实体提取] ──→ 识别股票代码 (AAPL, 600418.SH, 000660.KS 等)
    │
    ▼
[事件分类] ──→ 归入已有主题 (theme_id) 或创建新主题
    │
    ▼
[重要性评估] ──→ importance 1-5 评分
    │
    ▼
[主题强度计算] ──→ strength 基于事件数、股票数、时效性等
    │
    ▼
[趋势计算] ──→ rising / stable / declining
    │
    ▼
[Polymarket 集成] ──→ 概率数据补充
    │
    ▼
[AI 深度分析] ──→ archetype 模板匹配
    │
    ▼
[Regime 判断] ──→ risk_on / risk_off / neutral
```

### 3.3 数据更新频率

从 freshness 数据推断：

| 数据类型 | 最新更新时间 | 推测频率 |
|----------|-------------|----------|
| 新闻采集 | 2026-05-07 23:03:39 | 实时/分钟级 |
| 事件提取 | 2026-05-07 22:57:01 | 实时/分钟级 |
| 评分计算 | 2026-05-07 15:07:10 | 小时级 |
| AI 深度分析 | 2026-05-06 19:41:34 | 天级 |
| Polymarket 采集 | 2026-05-07 15:06:28 | 小时级 |

---

## 4. 部署架构

```
                    ┌─────────────┐
                    │   用户浏览器  │
                    └──────┬──────┘
                           │ HTTPS
                    ┌──────▼──────┐
                    │  Nginx 1.20 │
                    │  (反向代理)   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
       ┌──────▼──────┐ ┌──▼──┐ ┌──────▼──────┐
       │  Next.js    │ │Static│ │  API 服务   │
       │  (SSR)      │ │Files │ │  (/api/*)   │
       └──────┬──────┘ └─────┘ └──────┬──────┘
              │                       │
       ┌──────▼───────────────────────▼──────┐
       │           数据库 / 缓存              │
       └─────────────────────────────────────┘
```

**安全头**:
- `Strict-Transport-Security: max-age=31536000; includeSubDomains; preload`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), microphone=(), camera=()`

**缓存策略**:
- HTML: `s-maxage=30, stale-while-revalidate=31535970`
- Next.js ISR: `x-nextjs-prerender: 1`, `x-nextjs-stale-time: 300`
