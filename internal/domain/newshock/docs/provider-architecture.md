# Provider 抽象架构

A 股数据采集使用 **Provider 模式**，将不同数据源抽象为统一接口，通过 Failover 链实现自动容错切换。

## 核心设计

```
调度器 (Scheduler)
  └─ 服务 (Service)
       └─ Provider 接口 (dm/)
            └─ FailoverProvider (service/)
                 ├─ Provider A (infra/)  ← 主源
                 ├─ Provider B (infra/)  ← 备用 1
                 └─ Provider C (infra/)  ← 备用 2
```

- **调度器**：定时触发，只调用 Service 方法
- **Service**：业务逻辑，只依赖 Provider 接口（不关心具体数据源）
- **Provider 接口**：定义在 `dm/` 包，每个数据领域一个接口
- **FailoverProvider**：定义在 `service/` 包，按优先级串联多个数据源
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

## 文件结构

```
internal/domain/newshock/dm/
  ├── kline_provider.go    # KlineBar 类型 + KlineProvider 接口
  ├── stock_provider.go    # StockInfo 类型 + StockListProvider 接口
  └── board_provider.go    # BoardInfo/BoardStockInfo + BoardProvider 接口

internal/domain/newshock/service/
  ├── kline_provider.go    # FailoverKlineProvider 实现
  ├── stock_provider.go    # FailoverStockListProvider 实现
  ├── board_provider.go    # FailoverBoardProvider 实现
  ├── provider_health.go   # CheckProviderHealth 统一健康检查
  ├── astock_service.go    # AStockService（依赖 StockListProvider + KlineProvider）
  └── concept_service.go   # ConceptService（依赖 BoardProvider）

cmd/stocktest/main.go      # CLI 健康检查入口

internal/infra/
  ├── sina/client.go       # 实现 StockListProvider + KlineProvider
  ├── tencent/client.go    # 实现 KlineProvider
  ├── eastmoney/client.go  # 实现 StockListProvider + KlineProvider
  ├── eastmoney/concept_client.go  # 实现 BoardProvider
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

| Provider | StockList | Kline | Board |
|----------|-----------|-------|-------|
| sina     | ✅        | ✅    | -     |
| tencent  | -         | ✅    | -     |
| eastmoney.Client | ✅ | ✅    | -     |
| eastmoney.ConceptClient | - | - | ✅ |
| tushare  | ✅        | ✅    | -     |
| tdx      | ✅        | ✅    | ✅    |

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
- `dm.StockInfo`：Symbol（secid 格式）, Code, Name, Market（0=深市, 1=沪市）
- `dm.BoardInfo`：Code, Name
- `dm.BoardStockInfo`：Symbol, Code, Name, Market

## secid 约定

所有数据源统一使用 secid 格式标识股票：
- 沪市：`1.600519`（前缀 1）
- 深市：`0.000001`（前缀 0）

各 infra 客户端内部转换为自己的格式（如 sina: `sh600519`，tushare: `600519.SH`）。
