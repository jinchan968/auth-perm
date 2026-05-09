# Newshock 数据模型分析

---

## 实体关系图

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│    Theme     │       │    Event     │       │   Ticker     │
│  (投资主题)   │◄──────│  (市场事件)   │──────►│  (股票标的)   │
│              │  1:N  │              │  N:M  │              │
│ id           │       │ id           │       │ id           │
│ name         │       │ title        │       │ symbol       │
│ description  │       │ summary      │       │ name         │
│ category     │       │ channel      │       │ market       │
│ strength     │       │ importance   │       │ hot_score    │
│ trend        │       │ theme_id     │       │ mention_count│
│ ticker_count │       │ tickers[]    │       │              │
│ event_count  │       │ event_time   │       │              │
└──────────────┘       └──────────────┘       └──────────────┘
                                                     │
┌──────────────┐                                     │
│    Regime    │       ┌──────────────┐              │
│ (市场环境)   │       │  Watchlist   │              │
│              │       │ (我的关注)    │              │
│ regime_type  │       │              │              │
│ confidence   │       │ theme_ids[]  │──────────────┘
│ summary      │       │ (localStorage)│
└──────────────┘       └──────────────┘
```

---

## 1. Theme (投资主题)

主题是 Newshock 的核心实体，代表一个投资叙事或市场主线。

### 字段定义

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `id` | int | 主题唯一标识 | `13` |
| `name` | string | 主题名称 | `"中东地缘与能源"` |
| `description` | string | 主题描述 (AI 生成) | `"科威特遭伊朗导弹攻击..."` |
| `category` | enum | 主题分类 | `"geopolitical"` |
| `strength` | float | 主题强度原始值 | `5524.3` |
| `strength_norm` | float | 强度归一化值 (0-100) | `100` |
| `classification_confidence` | float | 分类置信度 (0-1) | `0.85` |
| `ticker_count` | int | 关联股票数量 | `499` |
| `event_count` | int | 关联事件数量 | `6397` |
| `trend` | enum | 趋势方向 | `"rising"` / `"stable"` / `"declining"` |
| `created_at` | datetime | 创建时间 | `"2026-05-01 19:45:37"` |
| `updated_at` | datetime | 最后更新时间 | `"2026-05-07 15:06:24"` |

### Category 枚举值

| 值 | 含义 | 中文翻译 |
|----|------|----------|
| `geopolitical` | 地缘政治 | 地缘 |
| `ai_semi` | AI 与半导体 | AI |
| `macro_monetary` | 宏观货币政策 | 宏观 |
| `supply_chain` | 供应链 | 供应链 |
| `defense` | 国防 | 国防 |
| `energy` | 能源 | 能源 |
| `earnings_event` | 财报事件 | 财报 |
| `exploratory` | 探索性主题 | 综合 |
| `pharma` | 医药 | 医药 |
| `regulatory` | 监管合规 | 监管 |

### Trend 枚举值

| 值 | 中文 | 图标 |
|----|------|------|
| `rising` | ▲ 上升 | 上升趋势 |
| `stable` | — 稳定 | 持平 |
| `declining` | ▼ 下降 | 下降趋势 |

### 示例数据

```json
{
  "id": 13,
  "name": "中东地缘与能源",
  "description": "科威特遭伊朗导弹攻击、伊朗议长和外长暂被移出美以清除名单、阿联酋ADNOC CEO批评伊朗将霍尔木兹海峡武器化...",
  "category": "geopolitical",
  "strength": 5524.3,
  "strength_norm": 100,
  "classification_confidence": 0.85,
  "ticker_count": 499,
  "event_count": 6397,
  "trend": "rising",
  "created_at": "2026-05-01 19:45:37",
  "updated_at": "2026-05-07 15:06:24"
}
```

### 强度排名 Top 10

| 排名 | 主题 | 强度 | 归一化 | 趋势 |
|------|------|------|--------|------|
| 1 | 中东地缘与能源 | 5524.3 | 100 | ▲ |
| 2 | AI算力与半导体芯片 | 3604.6 | 97 | ▲ |
| 3 | 央行货币政策与通胀 | 2198.5 | 94 | ▲ |
| 4 | 新能源汽车 | 1584.6 | 91 | ▲ |
| 5 | 地缘贸易与制裁 | 1229.5 | 88 | ▲ |
| 6 | 全球国防支出超级周期 | 948.8 | 85 | ▲ |
| 7 | 黄金与贵金属投资 | 883.9 | 82 | ▲ |
| 8 | 新能源与储能 | 864.2 | 79 | ▲ |
| 9 | 财报季与业绩 | 792.2 | 76 | ▲ |
| 10 | 消费电子与科技巨头 | 759.2 | 74 | ▲ |

---

## 2. Ticker (股票标的)

### 字段定义

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `id` | int | 标的唯一标识 | `9` |
| `symbol` | string | 股票代码 | `"AAPL"` |
| `name` | string | 公司名称 | `"苹果"` |
| `market` | enum | 所属市场 | `"us"` / `"cn"` / `"hk"` / `"kr"` |
| `hot_score` | float | 热度评分 | `473` |
| `mention_count` | int | 新闻提及次数 | `64` |

### Market 枚举值

| 值 | 市场 | 代码格式示例 |
|----|------|-------------|
| `us` | 美股 | `AAPL`, `GOOGL`, `NVDA` |
| `cn` | A股 | `600418.SH`, `300308.SZ` |
| `hk` | 港股 | (未在样本中出现) |
| `kr` | 韩股 | `005930.KS`, `000660.KS` |

### 热度排名 Top 10

| 排名 | 代码 | 名称 | 市场 | 热度 | 提及数 |
|------|------|------|------|------|--------|
| 1 | AAPL | 苹果 | US | 473 | 64 |
| 2 | GOOGL | 谷歌 | US | 378.5 | 49 |
| 3 | NVDA | 英伟达 | US | 329 | 40 |
| 4 | META | Meta | US | 300.5 | 40 |
| 5 | MSFT | 微软 | US | 279.5 | 37 |
| 6 | TSLA | 特斯拉 | US | 231 | 33 |
| 7 | AMZN | 亚马逊 | US | 203 | 25 |
| 8 | AMD | 超威半导体 | US | 169.5 | 21 |
| 9 | XOM | 埃克森美孚 | US | 154 | 17 |
| 10 | INTC | 英特尔 | US | 133 | 17 |

---

## 3. Event (市场事件)

### 字段定义

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `id` | int | 事件唯一标识 | `32355` |
| `title` | string | 事件标题 | `"瑞杰金融上调Arm目标价"` |
| `summary` | string | 事件摘要 | `"瑞杰金融将Arm目标价从166美元上调至244美元..."` |
| `channel` | string | 新闻来源频道 | `"金融快报 Financial Express"` |
| `importance` | int | 重要性等级 (1-5) | `2` |
| `theme_id` | int | 关联主题 ID | `43` |
| `theme_name` | string | 关联主题名称 | `"科技巨头资本运作与业务调整"` |
| `created_at` | datetime | 入库时间 | `"2026-05-07 15:06:14"` |
| `event_time` | datetime | 事件发生时间 | `"2026-05-07 22:57:01"` |
| `tickers` | array | 关联股票列表 | `[{"symbol": "ARM"}]` |

### Importance 等级

| 等级 | 含义 | UI 表示 |
|------|------|---------|
| 1 | 低重要性 | ●○○○○ |
| 2 | 一般 | ●●○○○ |
| 3 | 中等 | ●●●○○ |
| 4 | 重要 | ●●●●○ |
| 5 | 极重要 | ●●●●● |

### Channel (新闻源) 枚举

| 频道 | 说明 |
|------|------|
| `金融快报 Financial Express` | 主要来源 |
| `参考消息 XHQ` | 国际新闻 |
| `驻华外电 PD China` | 外电中文报道 |
| `财经日报 Finance News Daily` | 财经新闻 |
| `CNBeta 科技` | 科技新闻 |

### 示例数据

```json
{
  "id": 32355,
  "title": "瑞杰金融上调Arm目标价",
  "summary": "瑞杰金融将Arm目标价从166美元上调至244美元，反映AI芯片需求乐观。",
  "channel": "金融快报 Financial Express",
  "importance": 2,
  "theme_id": 43,
  "theme_name": "科技巨头资本运作与业务调整",
  "created_at": "2026-05-07 15:06:14",
  "event_time": "2026-05-07 22:57:01",
  "tickers": [{"symbol": "ARM"}]
}
```

---

## 4. Regime (市场环境)

### 字段定义

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `regime_type` | enum | 市场环境类型 | `"risk_on"` |
| `confidence` | float | 判断置信度 (0-1) | `0.7` |
| `summary` | string | 环境描述 | `"纳指创历史新高..."` |
| `created_at` | datetime | 判断时间 | `"2026-05-07 15:06:24"` |

### Regime Type 枚举

| 值 | 中文 | UI 样式 |
|----|------|---------|
| `risk_on` | 风险偏好 | `.regime-badge.risk_on` |
| `risk_off` | 风险规避 | `.regime-badge.risk_off` |
| `neutral` | 中性 | `.regime-badge.neutral` |

### 当前 Regime

```json
{
  "regime_type": "risk_on",
  "confidence": 0.7,
  "summary": "纳指创历史新高，A股总市值突破120万亿元，黄金避险需求与科技股风险偏好并存，但中东地缘和通胀压力带来不确定性"
}
```

---

## 5. Stats (统计概览)

```json
{
  "themeCount": 517,
  "tickerCount": 6702,
  "eventCount": 32372
}
```

---

## 6. Freshness (数据新鲜度)

```json
{
  "latest_news_time": "2026-05-07 23:03:39",
  "latest_event_time": "2026-05-07 22:57:01",
  "latest_score_run": "2026-05-07 15:07:10",
  "latest_deep_analysis": "2026-05-06 19:41:34",
  "latest_poly_ingest": "2026-05-07T15:06:28.794394Z"
}
```

---

## 7. Search Result (搜索结果)

```json
{
  "themes": [...],
  "tickers": [...],
  "events": [
    {
      "id": 12728,
      "title": "埃森哲以12亿美元收购Ookla",
      "importance": 3,
      "created_at": "2026-05-05 13:47:38"
    }
  ]
}
```

搜索结果中 themes 和 tickers 为空数组（可能是搜索词不匹配），events 返回精简字段。

---

## 8. Watchlist (我的关注)

存储在客户端 `localStorage('newshock-watchlist')`，格式为 theme ID 数组：

```json
[13, 61, 25]
```

前端过滤 themes 列表获取关注主题的完整数据。
