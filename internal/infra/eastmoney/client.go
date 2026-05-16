// Package eastmoney 东方财富 HTTP API 客户端。
//
// 实现接口：
//   - dm.StockListProvider（FetchAllStocks）— 全量 A 股列表（分页拉取）
//   - dm.KlineProvider（FetchKline）— 日线 OHLCV 数据
//
// 作为 sina 的备用数据源，由 FailoverStockListProvider / FailoverKlineProvider 自动切换。
// secid 约定：1.600519（沪市）/ 0.000001（深市）
package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/infra/httpclient"
)

const (
	stockListURL = "http://80.push2.eastmoney.com/api/qt/clist/get"
	klineURL     = "http://push2his.eastmoney.com/api/qt/stock/kline/get"
)

// Client 东方财富 API HTTP 客户端，实现 dm.StockListProvider 和 dm.KlineProvider 接口。
type Client struct {
	http *http.Client
}

// NewClient 创建东方财富客户端，超时 60 秒，内置 Chrome UA。
func NewClient() *Client {
	return &Client{
		http: httpclient.NewWithTimeout(60 * time.Second),
	}
}

// Name 返回数据源名称
func (c *Client) Name() string { return "eastmoney" }

// stockListResponse 东方财富股票列表 API 响应
type stockListResponse struct {
	Data struct {
		Total int               `json:"total"`
		Diff  []json.RawMessage `json:"diff"`
	} `json:"data"`
}

// klineResponse 东方财富 K 线 API 响应
type klineResponse struct {
	Data struct {
		Klines []string `json:"klines"`
	} `json:"data"`
}

// allSecuritiesFS 全品种 filter：A 股 + ETF + 可转债 + 指数
const allSecuritiesFS = "m:0+t:6,m:0+t:80,m:0+t:81,m:0+t:13,m:1+t:13,m:0+t:14,m:1+t:14,m:1+t:2,m:0+t:2"

// Probe 轻量健康检查，只拉 1 页 1 条数据验证 API 可用。
func (c *Client) Probe(ctx context.Context) error {
	url := fmt.Sprintf("%s?pn=1&pz=1&po=1&np=1&fltt=2&invt=2&fs=%s&fields=f12,f14",
		stockListURL, allSecuritiesFS)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

// FetchAllStocks 获取所有证券（股票/ETF/可转债/指数）的最新行情。
// 分页获取沪深北交所全品种，包含最新价、涨跌幅、成交量等。
func (c *Client) FetchAllStocks(ctx context.Context) ([]dm.StockInfo, error) {
	const pageSize = 2000
	var allStocks []dm.StockInfo
	page := 1

	for {
		url := fmt.Sprintf("%s?pn=%d&pz=%d&po=1&np=1&fltt=2&invt=2&fs=%s&fields=f2,f3,f5,f6,f8,f12,f13,f14",
			stockListURL, page, pageSize, allSecuritiesFS)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("eastmoney api error %d: %s", resp.StatusCode, string(body))
		}

		var raw stockListResponse
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		for _, item := range raw.Data.Diff {
			info, err := parseStockInfo(item)
			if err != nil {
				continue
			}
			allStocks = append(allStocks, info)
		}

		// 本页数据不足一页或已获取全部数据，结束分页
		if len(raw.Data.Diff) < pageSize || len(allStocks) >= raw.Data.Total {
			break
		}
		page++
	}

	// 全品种正常应有 7000+ 只（A 股 5000+ + ETF 900+ + 可转债 500+ + 指数 500+）
	if len(allStocks) < 2000 {
		return nil, fmt.Errorf("eastmoney returned only %d securities (expected 7000+)", len(allStocks))
	}
	return allStocks, nil
}

// parseStockInfo 解析单条证券数据。
// 东方财富用数字字段名：f2=价格, f3=涨跌幅, f5=成交量, f6=成交额, f8=换手率, f12=代码, f13=市场, f14=名称
func parseStockInfo(raw json.RawMessage) (dm.StockInfo, error) {
	// 用 map 解析，因为字段名是数字
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return dm.StockInfo{}, err
	}

	code := getString(m, "f12")
	name := getString(m, "f14")
	market := getInt(m, "f13")

	// 跳过无效数据
	if code == "" || name == "" || code == "-" {
		return dm.StockInfo{}, fmt.Errorf("invalid stock data")
	}

	// 推断品种类型
	secType := classifySecurity(code, market)

	// 构建 secid：沪市=1.code, 深市=0.code, 北交所=2.code
	// 东方财富 B 股 market 可能返回 0，通过代码前缀兜底识别
	prefix := "0"
	if market == 1 {
		prefix = "1"
	} else if market == 2 || strings.HasPrefix(code, "83") || strings.HasPrefix(code, "87") || strings.HasPrefix(code, "43") {
		prefix = "2"
		market = 2
	}
	symbol := prefix + "." + code

	return dm.StockInfo{
		Symbol:       symbol,
		Code:         code,
		Name:         name,
		Market:       market,
		SecurityType: secType,
	}, nil
}

// FetchKline 获取指定股票的日线 K 线数据。
// secid: 1.600519（沪市）/ 0.000001（深市）
// days: 获取最近 N 天的数据
func (c *Client) FetchKline(ctx context.Context, secid string, days int) ([]dm.KlineBar, error) {
	if days <= 0 {
		days = 30
	}

	url := fmt.Sprintf("%s?secid=%s&fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&klt=101&fqt=1&beg=0&end=20500101&lmt=%d",
		klineURL, fixKlineSecid(secid), days)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eastmoney kline api error %d: %s", resp.StatusCode, string(body))
	}

	var raw klineResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// klines 格式：["2026-05-09,开,收,高,低,量,额,振幅,涨跌幅,涨跌额,换手率", ...]
	var bars []dm.KlineBar
	for _, line := range raw.Data.Klines {
		bar, err := parseKlineBar(line)
		if err != nil {
			continue
		}
		bars = append(bars, bar)
	}
	return bars, nil
}

// parseKlineBar 解析单条 K 线数据。
// 格式：日期,开盘,收盘,最高,最低,成交量,成交额,振幅,涨跌幅,涨跌额,换手率
func parseKlineBar(line string) (dm.KlineBar, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 11 {
		return dm.KlineBar{}, fmt.Errorf("invalid kline format: %s", line)
	}

	open, _ := strconv.ParseFloat(parts[1], 64)
	close, _ := strconv.ParseFloat(parts[2], 64)
	high, _ := strconv.ParseFloat(parts[3], 64)
	low, _ := strconv.ParseFloat(parts[4], 64)
	volume, _ := strconv.ParseInt(parts[5], 10, 64)
	amount, _ := strconv.ParseFloat(parts[6], 64)
	changePct, _ := strconv.ParseFloat(parts[8], 64)
	turnover, _ := strconv.ParseFloat(parts[10], 64)

	return dm.KlineBar{
		Date:      parts[0],
		Open:      open,
		Close:     close,
		High:      high,
		Low:       low,
		Volume:    volume,
		Amount:    amount,
		ChangePct: changePct,
		Turnover:  turnover,
	}, nil
}

// classifySecurity 根据证券代码前缀和市场推断品种类型。
// ETF: 沪市 51xxxx/58xxxx, 深市 15xxxx/159xxx
// 可转债: 沪市 11xxxx, 深市 12xxxx
// 指数: 沪市 000xxx(上证指数, market=1), 深市 399xxx(深证指数), 中证 930xxx/950xxx
// 000xxx + market=0 是深市主板股票，不是指数
func classifySecurity(code string, market int) string {
	if len(code) != 6 {
		return "stock"
	}
	prefix2 := code[:2]
	switch prefix2 {
	case "51", "58": // 沪市 ETF
		if market == 1 {
			return "etf"
		}
	case "15": // 深市 ETF
		if market == 0 {
			return "etf"
		}
	case "11": // 沪市可转债
		if market == 1 {
			return "bond"
		}
	case "12": // 深市可转债
		if market == 0 {
			return "bond"
		}
	}
	prefix3 := code[:3]
	switch prefix3 {
	case "000": // 000xxx: 沪市是上证指数，深市是主板股票
		if market == 1 {
			return "index"
		}
	case "399": // 深证指数
		return "index"
	case "930", "950": // 中证指数
		return "index"
	}
	return "stock"
}

var bseSecidLogOnce sync.Once

// fixBseSecid 修正北交所 secid 前缀。
// 东方财富 stock list 接口对北交所返回 market=2（secid 前缀 2），但 push2 系列 API
// （K 线 / F10）只接受前缀 0。2.830001 → 0.830001。
func fixBseSecid(secid string) string {
	parts := strings.SplitN(secid, ".", 2)
	if len(parts) == 2 && parts[0] == "2" {
		rewritten := "0." + parts[1]
		bseSecidLogOnce.Do(func() {
			log.Printf("[Eastmoney] rewriting BSE secid prefix 2.→0. for push2 API (例: %s → %s)", secid, rewritten)
		})
		return rewritten
	}
	return secid
}

// fixKlineSecid 修正 K 线 API 的 secid 前缀，委托给 fixBseSecid。
func fixKlineSecid(secid string) string {
	return fixBseSecid(secid)
}
