# Provider 抽象架构

A 股数据采集使用 **Provider 模式**，将不同数据源抽象为统一接口，通过 Failover/Merge 链实现自动容错与数据合并。

## 核心设计

```
调度器 (Scheduler)
  └─ 服务 (Service)
       └─ Provider 接口 (dm/)
            ├─ FailoverProvider (service/)  ← 按优先级容错切换
            └─ MergeProvider (service/)     ← 多源合并取值
                 ├─ Provider A (infra/)  ← 主源
                 ├─ Provider B (infra/)  ← 备用 1
                 └─ Provider C (infra/)  ← 备用 2
```

- **调度器**：定时触发，只调用 Service 方法
- **Service**：业务逻辑，只依赖 Provider 接口（不关心具体数据源）
- **Provider 接口**：定义在 `dm/` 包，每个数据领域一个接口
- **FailoverProvider**：定义在 `service/` 包，按优先级串联多个数据源，第一个返回充足数据的胜出
- **MergeProvider**：定义在 `service/` 包，多源合并，后源的非零字段覆盖前源
- **infra 客户端**：实现 Provider 接口，封装具体 API 调用

## 三个数据领域

### 1. 股票列表（StockListProvider）

| 数据源 | 包 | API | 特点 |
|--------|-----|-----|------|
| 新浪（主源） | `sina.Client` | 腾讯财经批量行情 `qt.gtimg.cn` | 按代码段遍历，批次 50 个，间隔 300ms |
| 东方财富（备用 1） | `eastmoney.Client` | 东方财富全量接口 `push2.eastmoney.com` | 分页拉取，每页 2000 |
| 通达信（备用 2） | `tdx.Client` | TDX 协议 `StockCount`/`StockList` | TCP 直连，无 HTTP 限制 |
| Tushare（备用 3） | `tushare.Client` | Tushare Pro API `stock_basic` | 需 token，HTTP POST |

**接口**：`dm.StockListProvider.FetchAllStocks(ctx) ([]dm.StockInfo, error)`

**Failover 链**：`sina → eastmoney → tdx → tushare（需 token）`

**调度器**：`StockListScheduler`（24h 间隔，5min 启动延迟）

### 2. 日线 K 线（KlineProvider）

| 数据源 | 包 | API | 特点 |
|--------|-----|-----|------|
| 新浪（主源） | `sina.Client` | 新浪 K 线 API `money.finance.sina.com.cn` | 最多 1023 天 |
| 腾讯（备用 1） | `tencent.Client` | 腾讯前复权日线 `proxy.finance.qq.com` | JSON 格式 |
| 东方财富（备用 2） | `eastmoney.Client` | 东方财富日线 `push2his.eastmoney.com` | CSV 格式 |
| 通达信（备用 3） | `tdx.Client` | TDX 协议 `StockKLine` | TCP 直连，前复权，含换手率 |
| Tushare（备用 4） | `tushare.Client` | Tushare Pro API `daily` | 需 token，含换手率 |

**接口**：`dm.KlineProvider.FetchKline(ctx, secid, days) ([]dm.KlineBar, error)`

**Failover 链**：`sina → tencent → eastmoney → tdx → tushare（需 token）`

**调度器**：`DailyDataScheduler`（4h 间隔，10min 启动延迟）

### 3. 板块概念（BoardProvider）

| 数据源 | 包 | API | 特点 |
|--------|-----|-----|------|
| 东方财富（主源） | `eastmoney.ConceptClient` | 东方财富板块 API | 返回 BK0800 等板块代码 |
| 通达信（备用） | `tdx.Client` | TDX 协议板块文件 | block_gn.dat 概念、block.dat 行业 |

**接口**：`dm.BoardProvider.FetchBoards(ctx, boardType) ([]dm.BoardInfo, error)` + `FetchBoardStocks(ctx, boardCode)`

**Failover 链**：`eastmoney → tdx`

**调度器**：`ConceptScheduler`（24h 间隔）

### 4. F10 基本面（F10Provider）— Merge 模式

| 数据源 | 包 | API | 提供字段 |
|--------|-----|-----|----------|
| 腾讯财经（估值） | `tencent.QuoteClient` | `qt.gtimg.cn` GBK 批量行情 | PE(TTM), PE(静), PB, 总市值, 流通市值, 换手率, 量比, 涨停/跌停价 |
| 东方财富（财务） | `eastmoney.F10Client` | `push2.eastmoney.com/api/qt/stock/get` | 行业, 总股本, 流通股本, EPS, BVPS, ROE |

**接口**：`dm.F10Provider.FetchF10(ctx, codes []string) ([]dm.TickerF10, error)`

**Merge 链**：`tencent → eastmoney`（后源的非零字段覆盖前源）

**设计决策**：F10 使用 Merge 而非 Failover，因为两个源各有所长——腾讯有估值数据（PE/PB/市值），东财有财务数据（EPS/ROE/行业）。合并后得到完整的基本面快照。

**批量策略**：F10Client 内部 5 并发 goroutine + 信号量限流；AStockService 外部分批 50 只/批，连续失败 3 次熔断退出。

**调度器**：`F10DataScheduler`（1h 间隔，3x 启动延迟）

### 5. 个股新闻（NewsProvider）— Failover 模式

| 数据源 | 包 | API | 特点 |
|--------|-----|-----|------|
| 东方财富（主源） | `eastmoney.NewsClient` | `np-listapi.eastmoney.com/comm/web/getNewsByStock` | 个股新闻，HTML 标签自动清理 |

**接口**：`dm.NewsProvider.FetchNews(ctx, secid string, limit int) ([]dm.TickerNews, error)`

**Failover 链**：`eastmoney`（单源，预留扩展位）

**采集策略**：按 `hot_score DESC` 取前 N 只（默认 200）CN 股票，每只拉 20 条新闻。5 并发 goroutine + 信号量限流。

**调度器**：`StockNewsScheduler`（1h 间隔，3x 启动延迟）

## 文件结构

```
internal/domain/newshock/dm/
  ├── kline_provider.go    # KlineBar 类型 + KlineProvider 接口
  ├── stock_provider.go    # StockInfo 类型 + StockListProvider 接口
  ├── board_provider.go    # BoardInfo/BoardStockInfo + BoardProvider 接口
  ├── f10_provider.go      # F10Provider 接口
  ├── news_provider.go     # NewsProvider 接口
  ├── ticker_f10_do.go     # TickerF10 领域模型
  └── ticker_news_do.go    # TickerNews 领域模型

internal/domain/newshock/service/
  ├── kline_provider.go    # FailoverKlineProvider 实现（含数据充足性判断）
  ├── stock_provider.go    # FailoverStockListProvider 实现
  ├── board_provider.go    # FailoverBoardProvider 实现
  ├── f10_provider.go      # MergeF10Provider 实现（多源合并非零字段）
  ├── news_provider.go     # FailoverNewsProvider 实现
  ├── provider_health.go   # CheckProviderHealth 统一健康检查
  ├── astock_service.go    # AStockService（SyncStockList + SyncDailyData + SyncF10Data + SyncStockNews）
  └── concept_service.go   # ConceptService（依赖 BoardProvider）

internal/domain/newshock/repo/
  ├── ticker_f10_repo.go   # F10 仓储（UpsertBatch / GetByTickerID）
  └── ticker_news_repo.go  # 新闻仓储（UpsertBatch / GetByTickerID）

cmd/stocktest/main.go      # CLI 健康检查入口

internal/infra/
  ├── httpclient/          # 共享 HTTP 客户端（Chrome UA + GBK 解码 + 指数退避重试）
  ├── sina/client.go       # 实现 StockListProvider + KlineProvider
  ├── tencent/client.go    # 实现 KlineProvider
  ├── tencent/quote_client.go  # 实现 F10Provider（估值数据）
  ├── eastmoney/client.go  # 实现 StockListProvider + KlineProvider
  ├── eastmoney/concept_client.go  # 实现 BoardProvider
  ├── eastmoney/f10_client.go      # 实现 F10Provider（财务指标）
  ├── eastmoney/news_client.go     # 实现 NewsProvider（个股新闻）
  ├── tdx/client.go        # 实现 StockListProvider + KlineProvider + BoardProvider
  └── tushare/client.go    # 实现 StockListProvider + KlineProvider（需 token）
```

## 扩展新数据源

1. 在 `internal/infra/<source>/` 创建客户端，实现对应的 `dm.XxxProvider` 接口
2. 在 `internal/domain/newshock/module.go` 的 Failover 链中追加新客户端
3. 不需要修改任何 Service 代码

示例：添加 akshare 作为 K 线数据源：
```go
// module.go
if err := container.Provide(func() dm.KlineProvider {
    return service.NewFailoverKlineProvider(
        sina.NewClient(),
        tencent.NewClient(),
        eastmoney.NewClient(),
        tdx.NewClient(),
        akshare.NewClient(),  // 新增
    )
}); err != nil {
    return err
}
```

## Provider 能力矩阵

| Provider | StockList | Kline | Board | F10 | News | 北交所 |
|----------|-----------|-------|-------|-----|------|--------|
| sina     | ✅        | ✅    | -     | -   | -    | ✅     |
| tencent  | -         | ✅    | -     | ✅  | -    | ✅     |
| eastmoney.Client | ✅ | ✅    | -     | -   | -    | ✅     |
| eastmoney.F10Client | - | -   | -     | ✅  | -    | ✅     |
| eastmoney.NewsClient | - | -  | -     | -   | ✅   | ✅     |
| eastmoney.ConceptClient | - | - | ✅   | -   | -    | -      |
| tushare  | ✅        | ✅    | -     | -   | -    | ✅     |
| tdx      | ✅        | ✅    | ✅    | -   | -    | -      |

## Provider 健康检查

### CLI 方式

```bash
go run ./cmd/stocktest/
```

逐个探测所有 provider 的 stocklist/kline/board 能力，输出表格：

```
  Provider Health: 8/9 healthy

  PROVIDER   │ STOCKLIST              │ KLINE                  │ BOARD
  ───────────┼────────────────────────┼────────────────────────┼────────────────────────
  sina       │ ✅ 5500 (1.2s)         │ ✅ 10 (0.3s)           │ -
  tencent    │ -                      │ ✅ 10 (0.2s)           │ -
  eastmoney  │ ✅ 5500 (2.1s)         │ ✅ 10 (0.5s)           │ ✅ 120 (0.8s)
  tdx        │ ✅ ok (0.8s)           │ ✅ 10 (0.4s)           │ ✅ 85 (0.6s)
  tushare    │ -                      │ -                      │ -
```

无 token 时自动跳过 tushare。设 `TUSHARE_TOKEN` 环境变量可启用。

### API 方式

```
GET /api/v1/newshock/providers
```

返回 JSON 格式的探测结果，结构与 CLI 相同。需要认证 + `newshock.read` 权限。

### Probe 接口

`StockListProvider` 接口包含 `Probe(ctx) error` 方法，用于轻量健康检查（单次请求验证 API 可达），不执行全量拉取。

`KlineProvider` 和 `BoardProvider` 没有 Probe 方法，健康检查使用完整方法 + 短超时（10s）。

### 实现文件

```
internal/domain/newshock/service/provider_health.go   # CheckProviderHealth + probeProviderLight + probeProvider
cmd/stocktest/main.go                                 # CLI 入口，调用 CheckProviderHealth
```

## 共享类型

所有 infra 包不再各自定义 `KlineBar`、`StockInfo`，统一使用 `dm` 包的类型：

- `dm.KlineBar`：Date, Open, Close, High, Low, Volume, Amount, ChangePct, Turnover
- `dm.StockInfo`：Symbol（secid 格式）, Code, Name, Market（0=深市, 1=沪市, 2=北交所）
- `dm.BoardInfo`：Code, Name
- `dm.BoardStockInfo`：Symbol, Code, Name, Market
- `dm.TickerF10`：PeTtm, PeStatic, Pb, TotalMcap, FloatMcap, TurnoverRate, VolumeRatio, LimitUp, LimitDown, Industry, Eps, Bvps, Roe, Source
- `dm.TickerNews`：Title, Content, Source, PublishTime, URL

## secid 约定

所有数据源统一使用 secid 格式标识股票：
- 沪市：`1.600519`（前缀 1）
- 深市：`0.000001`（前缀 0）
- 北交所：`2.830001`（前缀 2）

各 infra 客户端内部转换为自己的格式（如 sina: `sh600519`，tushare: `600519.SH`，北交所: `bj830001`/`830001.BJ`）。

## 覆盖板块

| 板块 | 代码范围 | Sina | Tencent | Eastmoney | TDX | Tushare |
|------|----------|------|---------|-----------|-----|---------|
| 沪市主板 | 600xxx-609xxx | ✅ | - | ✅ | ✅ | ✅ |
| 科创板 | 688xxx | ✅ | - | ✅ | ✅ | ✅ |
| 深市主板 | 000xxx-001xxx | ✅ | - | ✅ | ✅ | ✅ |
| 创业板 | 300xxx-305xxx | ✅ | - | ✅ | ✅ | ✅ |
| 北交所 | 830xxx/870xxx/430xxx | ✅ | - | ✅ | - | ✅ |

注：Tencent 仅实现 KlineProvider，不提供 StockListProvider。TDX 不支持北交所协议。

---

## 编码实践与设计决策

### 1. 反爬虫：httpclient 统一收口

所有 infra 包共享 `internal/infra/httpclient` 提供的 HTTP 客户端，避免在各包中重复处理反爬虫逻辑。

```
httpclient.New()              ← 默认 30s 超时 + Chrome UA + 指数退避重试
httpclient.NewWithTimeout(60) ← 自定义超时，其余同上
```

三层机制：
- **Chrome UA**：`uaTransport` 自动注入 `User-Agent`，无需手动设置
- **GBK 解码**：`DecodeGBK(body)` / `DecodeGBKResponse(resp)` 处理腾讯/新浪等 GBK 编码响应
- **指数退避重试**：`retryTransport` 对 5xx/网络错误自动重试（1s/2s/4s），最多 3 次

**踩坑**：各数据源的 `NewClient()` 统一使用 `httpclient.New*()`，不再自行构建 `http.Client`。

### 2. Failover 数据充足性判断

`FailoverKlineProvider` 不再简单取"第一个非空结果"，而是判断数据是否充足：

```go
minBars := max(days/3, 1)  // 请求 30 天至少需要 10 条，请求 2 天至少 1 条

if len(bars) >= minBars {
    return bars, nil  // 数据充足，直接返回
}
// 数据不足但非空，暂存，继续尝试其他 provider
```

这解决了某些 provider 返回部分数据（如只返回最近 5 天）时被误判为"成功"的问题。

### 3. Merge vs Failover 选择标准

| 场景 | 模式 | 原因 |
|------|------|------|
| 同质数据源（K线、股票列表） | Failover | 数据相同，选最快的 |
| 互补数据源（F10 基本面） | Merge | 各源各有所长，合并后更完整 |
| 单源数据（新闻） | Failover（预留） | 当前只有东财，接口保留扩展位 |

**MergeF10Provider 合并规则**：后源的非零字段覆盖前源。腾讯提供估值（PE/PB/市值），东财提供财务（EPS/ROE/行业），合并得到完整 F10。

### 4. TenantID 简化：配置注入 vs 数据库查询

**旧方案**：`DistinctTenantIDs()` 每次调度 tick 查询数据库获取租户列表。

**新方案**：从 `config.RSS.TenantID` 直接注入，消除无谓的 DB 查询。

适用于单租户部署场景。如果需要多租户支持，可改回 DB 查询方式。

### 5. SyncDailyData 性能优化：批量 COUNT

**旧方案**：逐 ticker 调用 `GetLatestByTickerID` 判断是否有数据，O(n) 次 DB 查询。

**新方案**：`CountAllByTenant` 一次 `GROUP BY ticker_id` 查询拿到所有 count，O(1) 次 DB 查询。

```go
countMap, err := s.dailyRepo.CountAllByTenant(ctx, s.tenantID)
// countMap["ticker-xxx"] = 30

if countMap[ticker.ID] < minHealthyRecords {
    days = 90  // 数据不足，回溯 90 天
}
```

`minHealthyRecords = 20`：低于此阈值视为数据不足，触发全量回溯。

### 6. 熔断模式：连续失败提前退出

`SyncF10Data` 使用连续失败计数器实现简单熔断：

```go
const maxConsecutiveFails = 3
consecutiveFails := 0

for each batch {
    if err != nil {
        consecutiveFails++
        if consecutiveFails >= maxConsecutiveFails {
            break  // 提前退出，避免无意义的 API 调用
        }
        continue
    }
    consecutiveFails = 0  // 成功则重置
}
```

避免在系统性故障（如 API 全面不可用）时刷屏日志。

### 7. 并发采集模式

`SyncDailyData` 和 `SyncStockNews` 共用同一并发模式：

```go
sem := make(chan struct{}, 5)  // 最多 5 个并发
var wg sync.WaitGroup

for _, ticker := range tickers {
    wg.Add(1)
    go func(t dm.Ticker) {
        defer wg.Done()
        sem <- struct{}{}        // 获取令牌
        defer func() { <-sem }() // 释放令牌
        // ... 实际采集逻辑
    }(ticker)
}
wg.Wait()
```

**注意**：不需要在 goroutine 内额外 `time.Sleep`，信号量已经限制了并发数。

### 8. 前端 F10/新闻展示

Ticker 详情页通过独立的 `useQuery` 请求 F10 和新闻数据，与主详情数据解耦：

```tsx
const { data: ticker } = useQuery({ queryKey: ['ticker', symbol], ... });
const { data: f10 } = useQuery({ queryKey: ['ticker-f10', symbol], ... });
const { data: news } = useQuery({ queryKey: ['ticker-news', symbol], ... });
```

- F10 数据可能尚未同步（首次启动），后端返回 `null` 而非 404
- 前端用 `{f10 && (...)}` 条件渲染，null 时不显示卡片
- 新闻卡片在右栏，使用 `target="_blank" rel="noopener noreferrer"` 处理外链
