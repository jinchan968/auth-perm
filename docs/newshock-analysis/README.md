# Newshock — 事件驱动主题投研雷达 逆向分析

> 分析目标: https://news.moneych.top/
> 分析日期: 2026-05-08
> 分析方法: 前端逆向 (HTML/JS/CSS 提取 + API 探测 + SSR 数据解析)

---

## 项目定位

Newshock 是一个**事件驱动的主题投资研究雷达**，面向 A股/美股/港股市场。核心理念是：

```
新闻流 → 事件提取 → 主题聚类 → 强度评分 → 投资信号
```

**副标题描述**: "实时追踪市场主线 + 催化事件 + Polymarket 概率 + 标的暴露 + AI 深度分析"

**关键词**: thematic investing, event-driven, polymarket, archetype, 主题投资, 事件驱动, 美股主线, A股主题, 港股催化, AI 算力, 中东油价

---

## 核心功能一览

| 功能模块 | 说明 |
|----------|------|
| **Radar (雷达)** | 首页仪表盘，展示统计卡片、事件流、主题/股票领涨榜、市场环境 |
| **Themes (主题)** | 投资主题列表，按强度排序，可搜索，点击查看详情 |
| **Tickers (股票)** | 股票列表，支持市场筛选(US/CN/HK/KR)，热度排序 |
| **Events (事件)** | 从新闻流中提取的结构化事件流，按重要性(1-5)标记 |
| **Markets (盘口)** | 盘口数据展示 |
| **Edge (边缘信号)** | 低关注度但有潜力的边缘信号 |
| **全局搜索** | Cmd+K 全局搜索主题、股票、事件 |
| **我的关注** | 本地 Watchlist，关注特定主题 |
| **双语支持** | 中文/英文切换 |
| **深色/浅色主题** | CSS 变量实现的主题切换 |

---

## 技术栈

| 层级 | 技术 |
|------|------|
| **前端框架** | Next.js 15 (App Router, React Server Components) |
| **UI 组件库** | Ant Design 5 (antd) |
| **样式方案** | CSS 自定义属性 (CSS Variables) + antd ConfigProvider token 主题 |
| **路由** | Next.js 文件系统路由 |
| **状态管理** | React Context (主题/语言) + localStorage (watchlist) |
| **国际化** | 自实现翻译函数 `tt(key, lang)`，67 个翻译键 |
| **数据传递** | SSR 通过 RSC payload 嵌入页面 (self.__next_f.push) |
| **Web 服务器** | Nginx 1.20.1 反代 |
| **缓存策略** | s-maxage=30, stale-while-revalidate=31535970 |

---

## 数据规模 (截至 2026-05-07)

| 指标 | 数值 |
|------|------|
| 主题总数 | 517 |
| 股票总数 | 6,702 |
| 事件总数 | 32,372 |

---

## 文档目录

| 文件 | 内容 |
|------|------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 技术架构详解：前后端、数据管线、部署 |
| [DATA_MODEL.md](./DATA_MODEL.md) | 数据模型：Theme/Ticker/Event/Regime 字段详解 |
| [API_REFERENCE.md](./API_REFERENCE.md) | API 端点文档（含 curl 示例和响应样例） |
| [UI_COMPONENTS.md](./UI_COMPONENTS.md) | UI 组件分析：页面结构、交互特性、i18n |
| [IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md) | 复刻实现路线图（分阶段） |
| [MY_THOUGHTS.md](./MY_THOUGHTS.md) | 分析思路和推理过程 |
| [raw/](./raw/) | 原始资料：HTML、JS 源码、API 响应、CSS 变量、i18n 字典 |
