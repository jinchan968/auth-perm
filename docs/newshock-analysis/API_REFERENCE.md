# Newshock API 端点参考

---

## 已验证的 API 端点

### 1. 全局搜索

```
GET /api/search?q={query}
```

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `q` | string | 是 | 搜索关键词 |

**响应** (200 OK):
```json
{
  "themes": [],
  "tickers": [],
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

**curl 示例**:
```bash
curl 'https://news.moneych.top/api/search?q=AI'
```

**调用方式**: 前端搜索弹窗 (Cmd+K) 使用 debounce 200ms 后调用。

---

### 2. 其他 API 端点 (推测)

根据前端 JS 代码和路由结构，推测存在以下端点，但直接 GET 返回空响应 (可能需要认证或特定 headers):

| 端点 | 方法 | 状态 | 说明 |
|------|------|------|------|
| `/api/themes` | GET | 返回空 | 主题列表 |
| `/api/events` | GET | 返回空 | 事件列表 |
| `/api/tickers` | GET | 返回空 | 股票列表 |
| `/api/theme/:id` | GET | 未确认 | 主题详情 |
| `/api/ticker/:symbol` | GET | 未确认 | 股票详情 |
| `/api/docs` | GET | 存在 | API 文档页面 |
| `/admin/login` | GET | 存在 | 管理后台登录 |

**注意**: 大部分数据通过 SSR (RSC payload) 直接嵌入 HTML，前端不需要单独调用这些 API。API 端点可能主要用于管理后台或外部集成。

---

## 前端数据获取模式

### 首页: SSR 数据注入

```javascript
// 服务端: Next.js RSC 将数据嵌入 HTML
self.__next_f.push([1, "7:[\"$\",\"$L12\",null,{\"data\":{...}}]"])

// 客户端: 组件通过 props 接收
function RadarPage({ data }) {
  const { themes, tickers, events, regime, stats, freshness } = data;
  // 直接使用，无需 fetch
}
```

### 搜索: 客户端 API 调用

```javascript
// 搜索弹窗组件
useEffect(() => {
  if (!query.trim()) return;
  const timer = setTimeout(async () => {
    const res = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
    const data = await res.json();
    setResults(data);
  }, 200); // debounce 200ms
  return () => clearTimeout(timer);
}, [query]);
```

---

## SSR 数据结构 (首页)

首页 HTML 中嵌入的完整数据结构:

```json
{
  "themes": [
    {
      "id": 13,
      "name": "中东地缘与能源",
      "description": "...",
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
    // ... 共 20 条
  ],
  "tickers": [
    {
      "id": 9,
      "symbol": "AAPL",
      "name": "苹果",
      "market": "us",
      "hot_score": 473,
      "mention_count": 64
    }
    // ... 共 20 条
  ],
  "events": [
    {
      "id": 32355,
      "title": "瑞杰金融上调Arm目标价",
      "summary": "...",
      "channel": "金融快报 Financial Express",
      "importance": 2,
      "theme_id": 43,
      "theme_name": "科技巨头资本运作与业务调整",
      "created_at": "2026-05-07 15:06:14",
      "event_time": "2026-05-07 22:57:01",
      "tickers": [{"symbol": "ARM"}]
    }
    // ... 共 30 条
  ],
  "regime": {
    "regime_type": "risk_on",
    "confidence": 0.7,
    "summary": "纳指创历史新高...",
    "created_at": "2026-05-07 15:06:24"
  },
  "stats": {
    "themeCount": 517,
    "tickerCount": 6702,
    "eventCount": 32372
  },
  "freshness": {
    "latest_news_time": "2026-05-07 23:03:39",
    "latest_event_time": "2026-05-07 22:57:01",
    "latest_score_run": "2026-05-07 15:07:10",
    "latest_deep_analysis": "2026-05-06 19:41:34",
    "latest_poly_ingest": "2026-05-07T15:06:28.794394Z"
  }
}
```

---

## 缓存与性能

| 头部 | 值 | 说明 |
|------|-----|------|
| `Cache-Control` | `s-maxage=30, stale-while-revalidate=31535970` | CDN 缓存 30s，长期 stale |
| `x-nextjs-cache` | `HIT` | Next.js ISR 缓存命中 |
| `x-nextjs-prerender` | `1` | 静态预渲染 |
| `x-nextjs-stale-time` | `300` | ISR 重新验证间隔 5 分钟 |
| `ETag` | `"sicnoqm9a8pd4"` | 条件请求支持 |
