# Newshock UI 组件分析

---

## 1. 整体布局 (AppLayout)

```
┌─────────────────────────────────────────────────────────┐
│                    Topbar (顶部栏)                        │
│  ┌─────────────────────────────────────────────────────┐│
│  │ [🔍 搜索框 (Cmd+K)]          [EN] [🌙] [⋯]        ││
│  └─────────────────────────────────────────────────────┘│
├────────┬────────────────────────────────────────────────┤
│        │                                                │
│ Side   │              Main Content                      │
│ bar    │              (主内容区)                          │
│        │                                                │
│ ┌────┐ │                                                │
│ │Logo│ │                                                │
│ ├────┤ │                                                │
│ │雷达│ │                                                │
│ │主题│ │                                                │
│ │股票│ │                                                │
│ │盘口│ │                                                │
│ │边缘│ │                                                │
│ │事件│ │                                                │
│ ├────┤ │                                                │
│ │About│ │                                                │
│ └────┘ │                                                │
│        │                                                │
├────────┴────────────────────────────────────────────────┤
│              Mobile Tabbar (移动端底部导航)                │
│  [雷达] [主题] [股票] [盘口] [边缘] [事件]                  │
└─────────────────────────────────────────────────────────┘
```

### 侧边栏 (Sidebar)

- Logo: `/newshock-logo.png` (32x32)
- 品牌名: News*hock* (em 标签斜体)
- 导航项 6 个，带 SVG 图标
- 活跃指示器: `.sidebar-nav-indicator` (CSS 动画滑动)
- 底部: About 链接
- 路由匹配: 精确匹配或前缀匹配 (`/themes` 匹配 `/themes/13`)

### 顶部栏 (Topbar)

- 搜索框: 只读，点击打开全局搜索弹窗，显示 `⌘K` 快捷键提示
- 语言切换: 按钮显示 "EN" 或 "中"
- 主题切换: 太阳/月亮图标
- 更多菜单 (Dropdown):
  - 📝 更新日志 → `/changelog`
  - 📖 API 文档 → `/api/docs`
  - 🔐 管理后台 → `/admin/login`
  - About Newshock

### 移动端适配

- **Topbar Mobile**: 顶部栏简化版，只显示 Logo + 操作按钮
- **Mobile Tabbar**: 底部固定导航栏，6 个图标 + 文字
- 响应式断点: 使用 antd 的 `xs`/`sm`/`lg` 栅格系统

---

## 2. Radar 首页 (Radar Page)

### 布局结构

```
┌─────────────────────────────────────────────────────────┐
│ "实时信号 · 你的主题"                                      │
│ 📡 数据更新：新闻 1分钟前 · 事件 1分钟前 · 评分 8小时前      │
├─────────────────────────────────────────────────────────┤
│ ⭐ 我的关注 (3)                                           │
│ [中东地缘与能源 强度100] [AI算力 强度97] [央行货币 强度94]   │
├──────────────────────────────┬──────────────────────────┤
│                              │                          │
│  ┌──────┐ ┌──────┐ ┌──────┐ │  ┌──────────────────────┐ │
│  │主题数│ │股票数│ │事件数│ │  │ 市场环境              │ │
│  │ 517  │ │ 6702 │ │32372│ │  │ [风险偏好]            │ │
│  └──────┘ └──────┘ └──────┘ │  │ 纳指创历史新高...      │ │
│                              │  └──────────────────────┘ │
│  ┌─────────────────────────┐ │                          │
│  │ 事件流           查看全部 │ │  ┌──────────────────────┐ │
│  │ ─────────────────────── │ │  │ 近一周事件            │ │
│  │ 15:06 · 金融快报 · ⬤⬤⬤○○│ │  │ 15:06 ⬤⬤⬤○○         │ │
│  │ 瑞杰金融上调Arm目标价     │ │  │ 瑞杰金融上调Arm目标价  │ │
│  │ ...                     │ │  │ ...                  │ │
│  │ (共 15 条，可滚动)       │ │  │ (共 8 条)            │ │
│  └─────────────────────────┘ │  └──────────────────────┘ │
│                              │                          │
│  ┌─────────────────────────┐ │  ┌──────────────────────┐ │
│  │ 主题领涨榜  [主题榜|股票榜]│ │  │ Stocks        查看全部│ │
│  │ ─────────────────────── │ │  │ AAPL    苹果    US   │ │
│  │ 中东地缘与能源  ████ 100 │ │  │ GOOGL   谷歌    US   │ │
│  │ AI算力与半导体  ████ 97  │ │  │ ...                  │ │
│  │ ...                     │ │  │ (共 8 条)            │ │
│  └─────────────────────────┘ │  └──────────────────────┘ │
└──────────────────────────────┴──────────────────────────┘
```

### 关键组件

**统计卡片** (3 个):
- 主题数、股票数、事件数
- 大号数字 (28px, bold) + 小写标签

**事件流** (Event Stream):
- 最多显示 15 条
- 每条: 时间 · 频道 · 主题 Tag + 重要性圆点 (1-5)
- 标题 + 摘要 (截断 130 字) + 关联股票代码
- 可滚动容器 (maxHeight: 420px)
- 点击主题 Tag 跳转主题详情

**主题领涨榜** (Thematic Leaders):
- Tab 切换: 主题榜 / 股票榜
- 主题榜: 名称 + 分类 Tag + 股票数·事件数 + 强度条
- 股票榜: 代码 + 名称 + 市场 Tag + 热度分数

**市场环境** (Market Regime):
- Badge 显示 regime_type
- 摘要文字

**我的关注** (Watchlist):
- 本地存储的 theme ID 列表
- 过滤 themes 获取完整数据
- 显示为 Tag 列表，点击跳转详情

---

## 3. 主题列表页 (Themes Page)

- 搜索框: 搜索主题名称或描述
- 列表: 按 strength 降序排列
- 每行: 名称 + 分类 + ticker_count + event_count + 强度条
- 点击跳转 `/themes/:id`

---

## 4. 股票列表页 (Tickers Page)

- 搜索框: 搜索代码或名称
- 市场筛选: All / US / CN / HK
- 列表: symbol + name + market Tag + hot_score
- 点击跳转 `/tickers/:symbol`

---

## 5. 事件列表页 (Events Page)

- 完整事件列表
- 每条: 时间 + 频道 + 重要性 + 标题 + 摘要 + 关联股票

---

## 6. 全局搜索弹窗 (Search Modal)

```
┌──────────────────────────────────────┐
│ 🔍 搜索主题、股票、事件…              │
├──────────────────────────────────────┤
│                                      │
│ THEMES                               │
│ 中东地缘与能源                  100  │
│ AI算力与半导体                   97  │
│                                      │
│ TICKERS                              │
│ AAPL    苹果                    473  │
│ NVDA    英伟达                  329  │
│                                      │
│ EVENTS                               │
│ 埃森哲以12亿美元收购Ookla  05-05     │
│                                      │
├──────────────────────────────────────┤
│ ESC to close                         │
└──────────────────────────────────────┘
```

**交互**:
- `Cmd+K` / `Ctrl+K` 打开
- `ESC` 关闭
- 点击背景关闭
- 输入 debounce 200ms 后调用 `/api/search?q=`
- 每类最多显示 5 条结果
- 点击跳转对应详情页

---

## 7. CSS 变量系统

### Newshock 自定义变量 (--nshock-*)

从 CSS 文件中提取的 28 个自定义属性:

```css
--nshock-bg              /* 背景色 */
--nshock-bg-card         /* 卡片背景 */
--nshock-bg-elevated     /* 悬浮背景 */
--nshock-text            /* 主文字色 */
--nshock-text-muted      /* 次要文字色 */
--nshock-border          /* 边框色 */
--nshock-primary         /* 主题色 */
--nshock-danger          /* 危险色 */
--nshock-shadow          /* 阴影 */
```

### 深色主题 Token

```javascript
{
  colorPrimary: "#6366f1",           // 靛蓝色
  colorBgContainer: "#14141f",       // 深蓝黑背景
  colorBgElevated: "#1a1a2e",        // 稍亮背景
  colorBorderSecondary: "rgba(255,255,255,0.06)",
  colorText: "#e4e4ed",              // 浅灰文字
  colorTextSecondary: "#8888a0",     // 暗灰文字
  borderRadius: 8,
  fontFamily: "-apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', sans-serif"
}
```

### 浅色主题 Token

```javascript
{
  colorPrimary: "#4f46e5",           // 深靛蓝
  borderRadius: 8,
  fontFamily: "-apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', sans-serif"
}
```

---

## 8. 国际化 (i18n) 字典

共 67 个翻译键，完整字典见 `raw/i18n-dictionary.json`。

**核心键**:

| Key | 中文 | 英文 |
|-----|------|------|
| radar | 雷达 | Radar |
| themes | 主题 | Themes |
| tickers | 股票 | Tickers |
| events | 事件 | Events |
| edge | 边缘信号 | Edge |
| markets | 盘口 | Markets |
| themeCount | 主题数 | Themes |
| tickerCount | 股票数 | Tickers |
| eventCount | 事件数 | Events |
| marketRegime | 市场环境 | Market Regime |
| risk_on | 风险偏好 | Risk On |
| risk_off | 风险规避 | Risk Off |
| neutral | 中性 | Neutral |
| eventStream | 事件流 | Event Stream |
| thematicLeaders | 主题领涨榜 | Thematic Leaders |
| hotScore | 热度 | Hot Score |
| searchGlobal | 搜索主题、股票、事件… | Search themes, tickers, events… |
| strength | 强度 | Strength |
| rising | ▲ 上升 | ▲ rising |
| stable | — 稳定 | — stable |
| declining | ▼ 下降 | ▼ declining |

---

## 9. 交互特性

| 特性 | 实现方式 |
|------|----------|
| 路由高亮 | `usePathname()` + 精确/前缀匹配 |
| 搜索快捷键 | `window.addEventListener('keydown', ...)` 监听 Cmd+K |
| 主题持久化 | `localStorage('theme-mode')` + cookie |
| 语言持久化 | `localStorage('newshock-lang')` + cookie |
| Watchlist | `localStorage('newshock-watchlist')` JSON 数组 |
| 时间显示 | 自实现 `v()` 函数: "刚刚" / "N 分钟前" / "N 小时前" / "N 天前" |
| 强度条 | `.strength-bar > .fill` CSS 宽度百分比 |
| 重要性圆点 | `.importance > .dot.filled` CSS 类切换 |
