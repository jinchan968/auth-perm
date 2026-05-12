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
	"net/http"
	"strconv"
	"strings"
	"time"

	"auth-perm/internal/domain/newshock/dm"
)

const (
	stockListURL = "http://80.push2.eastmoney.com/api/qt/clist/get"
	klineURL     = "http://push2his.eastmoney.com/api/qt/stock/kline/get"
)

// Client 东方财富 API HTTP 客户端，实现 dm.StockListProvider 和 dm.KlineProvider 接口。
type Client struct {
	http *http.Client
}

// NewClient 创建东方财富客户端，超时 30 秒
func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 60 * time.Second},
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

// Probe 轻量健康检查，只拉 1 页 1 条数据验证 API 可用。
func (c *Client) Probe(ctx context.Context) error {
	url := fmt.Sprintf("%s?pn=1&pz=1&po=1&np=1&fltt=2&invt=2&fs=m:0+t:6,m:0+t:80&fields=f12,f14",
		stockListURL)
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

// FetchAllStocks 获取所有 A 股股票的最新行情。
// 分页获取沪深两市全部股票，包含最新价、涨跌幅、成交量等。
func (c *Client) FetchAllStocks(ctx context.Context) ([]dm.StockInfo, error) {
	const pageSize = 2000
	var allStocks []dm.StockInfo
	page := 1

	for {
		// fs=m:0+t:6,m:0+t:80 表示沪深 A 股
		// fields=f2,f3,f5,f6,f8,f12,f13,f14 表示价格/涨跌幅/成交量/成交额/换手率/代码/市场/名称
		url := fmt.Sprintf("%s?pn=%d&pz=%d&po=1&np=1&fltt=2&invt=2&fs=m:0+t:6,m:0+t:80&fields=f2,f3,f5,f6,f8,f12,f13,f14",
			stockListURL, page, pageSize)

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

	// A 股正常应有 5000+ 只，低于 2000 只说明数据异常
	if len(allStocks) < 2000 {
		return nil, fmt.Errorf("eastmoney returned only %d stocks (expected 5000+)", len(allStocks))
	}
	return allStocks, nil
}

// parseStockInfo 解析单条股票数据。
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

	// 构建 secid：沪市=1.code, 深市=0.code
	prefix := "0"
	if market == 1 {
		prefix = "1"
	}
	symbol := prefix + "." + code

	return dm.StockInfo{
		Symbol: symbol,
		Code:   code,
		Name:   name,
		Market: market,
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
		klineURL, secid, days)

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

// 辅助函数：从 map 中安全提取值
func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func getInt(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}
