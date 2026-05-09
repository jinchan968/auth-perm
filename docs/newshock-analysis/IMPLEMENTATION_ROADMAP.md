# Newshock 复刻实现路线图

---

## Phase 0: 项目初始化 (1-2 天)

### 前端
```bash
npx create-next-app@latest newshock --typescript --app --tailwind
cd newshock
npm install antd @ant-design/icons
```

### 后端
- 选择: Go (Gin/Fiber) 或 Python (FastAPI)
- 数据库: PostgreSQL + ClickHouse (事件分析)
- 缓存: Redis

### 目录结构
```
newshock/
├── app/                    # Next.js App Router
│   ├── layout.tsx
│   ├── page.tsx
│   ├── themes/
│   ├── tickers/
│   ├── events/
│   ├── markets/
│   ├── edge/
│   └── api/
├── components/
│   ├── layout/
│   │   ├── Sidebar.tsx
│   │   ├── Topbar.tsx
│   │   ├── MobileTabbar.tsx
│   │   └── SearchModal.tsx
│   ├── radar/
│   │   ├── StatCards.tsx
│   │   ├── EventStream.tsx
│   │   ├── ThematicLeaders.tsx
│   │   ├── MarketRegime.tsx
│   │   └── Watchlist.tsx
│   └── shared/
│       ├── ImportanceDots.tsx
│       ├── StrengthBar.tsx
│       └── TimeAgo.tsx
├── lib/
│   ├── api.ts              # API 客户端
│   ├── i18n.ts             # 国际化
│   └── theme.ts            # 主题配置
├── styles/
│   └── globals.css         # CSS 变量
└── public/
    └── newshock-logo.png
```

---

## Phase 1: 布局框架 (2-3 天)

### 1.1 CSS 变量系统
```css
:root {
  --nshock-primary: #4f46e5;
  --nshock-bg: #ffffff;
  --nshock-bg-card: #ffffff;
  --nshock-text: #1a1a2e;
  --nshock-text-muted: #6b7280;
  --nshock-border: #e5e7eb;
}

[data-theme="dark"] {
  --nshock-primary: #6366f1;
  --nshock-bg: #0a0a14;
  --nshock-bg-card: #14141f;
  --nshock-text: #e4e4ed;
  --nshock-text-muted: #8888a0;
  --nshock-border: rgba(255,255,255,0.06);
}
```

### 1.2 Ant Design 主题配置
```typescript
const darkToken = {
  colorPrimary: "#6366f1",
  colorBgContainer: "#14141f",
  colorBgElevated: "#1a1a2e",
  colorBorderSecondary: "rgba(255,255,255,0.06)",
  colorText: "#e4e4ed",
  colorTextSecondary: "#8888a0",
  borderRadius: 8,
};
```

### 1.3 布局组件
- Sidebar: Logo + 6 个导航项 + 活跃指示器 + About
- Topbar: 搜索框 + 语言切换 + 主题切换 + 更多菜单
- MobileTabbar: 底部 6 个图标导航
- 响应式: 移动端隐藏 Sidebar，显示 Topbar Mobile + Tabbar

### 1.4 国际化
```typescript
const dict: Record<string, { zh: string; en: string }> = {
  radar: { zh: "雷达", en: "Radar" },
  themes: { zh: "主题", en: "Themes" },
  // ... 67 个键
};

export function tt(key: string, lang: string): string {
  return dict[key]?.[lang as keyof typeof dict[string]] ?? key;
}
```

---

## Phase 2: 数据层 (3-5 天)

### 2.1 数据库 Schema

```sql
-- 主题表
CREATE TABLE themes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    strength FLOAT DEFAULT 0,
    strength_norm FLOAT DEFAULT 0,
    classification_confidence FLOAT DEFAULT 0.85,
    ticker_count INT DEFAULT 0,
    event_count INT DEFAULT 0,
    trend VARCHAR(20) DEFAULT 'stable',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 股票表
CREATE TABLE tickers (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255),
    market VARCHAR(10) NOT NULL, -- us/cn/hk/kr
    hot_score FLOAT DEFAULT 0,
    mention_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 事件表
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    summary TEXT,
    channel VARCHAR(100),
    importance INT DEFAULT 3, -- 1-5
    theme_id INT REFERENCES themes(id),
    event_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 事件-股票关联表
CREATE TABLE event_tickers (
    event_id INT REFERENCES events(id),
    ticker_id INT REFERENCES tickers(id),
    PRIMARY KEY (event_id, ticker_id)
);

-- 主题-股票关联表
CREATE TABLE theme_tickers (
    theme_id INT REFERENCES themes(id),
    ticker_id INT REFERENCES tickers(id),
    PRIMARY KEY (theme_id, ticker_id)
);
```

### 2.2 API 层

```
GET  /api/search?q=          # 全局搜索
GET  /api/themes              # 主题列表 (分页、排序)
GET  /api/themes/:id          # 主题详情
GET  /api/tickers             # 股票列表 (分页、筛选)
GET  /api/tickers/:symbol     # 股票详情
GET  /api/events              # 事件列表 (分页)
GET  /api/regime              # 市场环境
GET  /api/stats               # 统计数据
GET  /api/freshness           # 数据新鲜度
```

### 2.3 SSR 数据注入

```typescript
// app/page.tsx
async function RadarPage() {
  const [themes, tickers, events, regime, stats, freshness] = await Promise.all([
    getTopThemes(20),
    getTopTickers(20),
    getRecentEvents(30),
    getCurrentRegime(),
    getStats(),
    getFreshness(),
  ]);

  return (
    <RadarView
      data={{ themes, tickers, events, regime, stats, freshness }}
    />
  );
}
```

---

## Phase 3: 核心页面 (5-7 天)

### 3.1 Radar 首页
- 统计卡片 (3 个)
- 事件流 (可滚动，最多 15 条)
- 主题/股票领涨榜 (Tab 切换)
- 市场环境卡片
- 我的关注区域
- 侧边栏热门股票

### 3.2 主题列表页
- 搜索框
- 主题卡片列表 (强度条、分类 Tag)

### 3.3 主题详情页 (`/themes/:id`)
- 主题信息 (名称、描述、分类、强度、趋势)
- 关联股票列表
- 近期事件列表

### 3.4 股票列表页
- 搜索框
- 市场筛选 (All/US/CN/HK)
- 股票卡片列表

### 3.5 股票详情页 (`/tickers/:symbol`)
- 股票信息 (代码、名称、市场、热度)
- 关联主题列表
- 相关事件列表

### 3.6 事件列表页
- 完整事件流
- 时间、频道、重要性、标题、摘要

---

## Phase 4: 增值功能 (持续迭代)

### 4.1 全局搜索 (Cmd+K)
- Modal 弹窗
- Debounce 搜索
- 分类展示结果 (themes/tickers/events)

### 4.2 Watchlist
- localStorage 存储 theme IDs
- 首页展示关注主题

### 4.3 数据管线 (后端)
- 新闻采集 (RSS/API 爬虫)
- NER 实体提取 (股票代码识别)
- 事件分类 (归入主题)
- 主题强度评分算法
- 趋势计算

### 4.4 Polymarket 集成
- 概率数据采集
- 与主题/事件关联

### 4.5 AI 深度分析
- 主题描述生成
- Archetype 模板匹配
- Regime 判断

---

## 技术选型对比

| 维度 | 推荐方案 | 备选方案 |
|------|----------|----------|
| 前端框架 | Next.js 15 (App Router) | Nuxt 3, SvelteKit |
| UI 库 | Ant Design 5 | MUI, Chakra UI |
| 后端语言 | Go (Gin) | Python (FastAPI), Node.js |
| 数据库 | PostgreSQL | MySQL |
| 分析数据库 | ClickHouse | TimescaleDB |
| 缓存 | Redis | Memcached |
| 新闻采集 | Scrapy + RSS | Puppeteer, Playwright |
| NER | spaCy / LLM API | 自训练模型 |
| 部署 | Docker + Nginx | Vercel + Supabase |

---

## 参考资源

- 原始资料: `docs/newshock-analysis/raw/` 目录
- SSR 数据样例: `raw/homepage-ssr-data.json`
- JS 源码: `raw/page-js.js`, `raw/layout-js.js`
- i18n 字典: `raw/i18n-dictionary.json`
- CSS 变量: `raw/css-variables.txt`
- API 响应: `raw/api-search-response.json`
