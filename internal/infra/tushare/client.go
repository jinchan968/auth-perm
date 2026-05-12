// Package tushare 封装 Tushare Pro HTTP API 客户端。
// 用于获取 A 股股票列表和 K 线历史数据。
//
// API 文档：https://tushare.pro/document/2
// 股票列表：api_name=stock_basic
// 日线行情：api_name=daily
//
// secid 约定（与 eastmoney 一致）：1.600519（沪市）/ 0.000001（深市）
// Tushare 原始格式：600519.SH / 000001.SZ
package tushare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"auth-perm/internal/domain/newshock/dm"
)

const apiURL = "https://api.tushare.pro"

// Client Tushare Pro API 客户端，实现 dm.StockListProvider 和 dm.KlineProvider 接口。
type Client struct {
	http  *http.Client
	token string
}

// NewClient 创建 Tushare 客户端
func NewClient(token string) *Client {
	return &Client{
		http:  &http.Client{Timeout: 60 * time.Second},
		token: token,
	}
}

// Name 返回数据源名称
func (c *Client) Name() string { return "tushare" }

// tushareRequest Tushare API 请求体
type tushareRequest struct {
	APIName string            `json:"api_name"`
	Token   string            `json:"token"`
	Params  map[string]string `json:"params"`
	Fields  string            `json:"fields,omitempty"`
}

// tushareResponse Tushare API 通用响应体
type tushareResponse struct {
	RequestID string `json:"request_id"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      struct {
		Fields []string          `json:"fields"`
		Items  []json.RawMessage `json:"items"`
	} `json:"data"`
}

// Probe 轻量健康检查，只拉 1 条数据验证 API 可用。
func (c *Client) Probe(ctx context.Context) error {
	body := tushareRequest{
		APIName: "stock_basic",
		Token:   c.token,
		Params:  map[string]string{"list_status": "L"},
		Fields:  "ts_code",
	}
	var resp tushareResponse
	if err := c.doRequest(ctx, body, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("tushare error %d: %s", resp.Code, resp.Msg)
	}
	return nil
}

// FetchAllStocks 获取所有上市 A 股股票列表。
// 返回的 Symbol 使用 secid 格式：1.600519（沪市）/ 0.000001（深市）。
func (c *Client) FetchAllStocks(ctx context.Context) ([]dm.StockInfo, error) {
	body := tushareRequest{
		APIName: "stock_basic",
		Token:   c.token,
		Params: map[string]string{
			"list_status": "L",
		},
		Fields: "ts_code,symbol,name,market,list_date",
	}

	var resp tushareResponse
	if err := c.doRequest(ctx, body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("tushare error %d: %s", resp.Code, resp.Msg)
	}

	// stock_basic 返回的是对象数组：[{"ts_code":"000001.SZ","name":"平安银行",...}, ...]
	var stocks []dm.StockInfo
	for _, item := range resp.Data.Items {
		var m map[string]any
		if err := json.Unmarshal(item, &m); err != nil {
			continue
		}
		tsCode := getString(m, "ts_code")
		name := getString(m, "name")
		symbol := getString(m, "symbol")
		marketStr := getString(m, "market")

		if tsCode == "" || name == "" {
			continue
		}

		// 转换为 secid 格式：600519.SH -> 1.600519, 000001.SZ -> 0.000001
		code := strings.Split(tsCode, ".")[0]
		market := 0 // 深市
		if marketStr == "主板" && strings.HasSuffix(tsCode, ".SH") {
			market = 1
		} else if strings.HasSuffix(tsCode, ".SH") {
			market = 1
		}
		secid := fmt.Sprintf("%d.%s", market, code)

		stocks = append(stocks, dm.StockInfo{
			Symbol: secid,
			Code:   symbol,
			Name:   name,
			Market: market,
		})
	}
	return stocks, nil
}

// FetchKline 获取指定股票的日线 K 线数据。
// secid: 1.600519（沪市）/ 0.000001（深市）
// days: 获取最近 N 天的数据
func (c *Client) FetchKline(ctx context.Context, secid string, days int) ([]dm.KlineBar, error) {
	if days <= 0 {
		days = 30
	}

	// secid 转 ts_code：1.600519 -> 600519.SH, 0.000001 -> 000001.SZ
	tsCode, err := secidToTsCode(secid)
	if err != nil {
		return nil, err
	}

	// 计算起始日期
	startDate := time.Now().AddDate(0, 0, -days).Format("20060102")

	body := tushareRequest{
		APIName: "daily",
		Token:   c.token,
		Params: map[string]string{
			"ts_code":    tsCode,
			"start_date": startDate,
		},
		Fields: "ts_code,trade_date,open,high,low,close,vol,amount,pct_chg",
	}

	var resp tushareResponse
	if err := c.doRequest(ctx, body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("tushare error %d: %s", resp.Code, resp.Msg)
	}

	// daily 返回的是数组数组：[["600519.SH","20260509",1500.00,...], ...]
	// 需要用 fields 做列名映射
	fieldIndex := make(map[string]int)
	for i, f := range resp.Data.Fields {
		fieldIndex[f] = i
	}

	var bars []dm.KlineBar
	for _, item := range resp.Data.Items {
		var row []any
		if err := json.Unmarshal(item, &row); err != nil {
			continue
		}
		if len(row) < len(resp.Data.Fields) {
			continue
		}

		tradeDate := getStringFromRow(row, fieldIndex, "trade_date")
		if len(tradeDate) == 8 {
			// 20260509 -> 2026-05-09
			tradeDate = tradeDate[:4] + "-" + tradeDate[4:6] + "-" + tradeDate[6:8]
		}

		bars = append(bars, dm.KlineBar{
			Date:      tradeDate,
			Open:      getFloatFromRow(row, fieldIndex, "open"),
			High:      getFloatFromRow(row, fieldIndex, "high"),
			Low:       getFloatFromRow(row, fieldIndex, "low"),
			Close:     getFloatFromRow(row, fieldIndex, "close"),
			Volume:    getInt64FromRow(row, fieldIndex, "vol"),
			Amount:    getFloatFromRow(row, fieldIndex, "amount"),
			ChangePct: getFloatFromRow(row, fieldIndex, "pct_chg"),
		})
	}

	// 获取换手率并合并
	turnoverMap, _ := c.fetchTurnoverMap(ctx, tsCode, startDate)
	for i := range bars {
		if t, ok := turnoverMap[bars[i].Date]; ok {
			bars[i].Turnover = t
		}
	}

	return bars, nil
}

// fetchTurnoverMap 从 daily_basic 接口获取换手率，返回 date(YYYY-MM-DD) -> turnover_rate 的映射。
func (c *Client) fetchTurnoverMap(ctx context.Context, tsCode, startDate string) (map[string]float64, error) {
	body := tushareRequest{
		APIName: "daily_basic",
		Token:   c.token,
		Params: map[string]string{
			"ts_code":    tsCode,
			"start_date": startDate,
		},
		Fields: "trade_date,turnover_rate",
	}

	var resp tushareResponse
	if err := c.doRequest(ctx, body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("tushare error %d: %s", resp.Code, resp.Msg)
	}

	fieldIndex := make(map[string]int)
	for i, f := range resp.Data.Fields {
		fieldIndex[f] = i
	}

	m := make(map[string]float64, len(resp.Data.Items))
	for _, item := range resp.Data.Items {
		var row []any
		if err := json.Unmarshal(item, &row); err != nil || len(row) < len(resp.Data.Fields) {
			continue
		}
		date := getStringFromRow(row, fieldIndex, "trade_date")
		if len(date) == 8 {
			date = date[:4] + "-" + date[4:6] + "-" + date[6:8]
		}
		m[date] = getFloatFromRow(row, fieldIndex, "turnover_rate")
	}
	return m, nil
}

// doRequest 执行 Tushare API 请求
func (c *Client) doRequest(ctx context.Context, reqBody tushareRequest, result *tushareResponse) error {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tushare api error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}

// secidToTsCode 将 secid 格式转为 Tushare ts_code 格式
// 1.600519 -> 600519.SH, 0.000001 -> 000001.SZ
func secidToTsCode(secid string) (string, error) {
	parts := strings.SplitN(secid, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid secid format: %s", secid)
	}
	code := parts[1]
	switch parts[0] {
	case "1":
		return code + ".SH", nil
	case "0":
		return code + ".SZ", nil
	default:
		return "", fmt.Errorf("unknown market prefix: %s", parts[0])
	}
}

// 辅助函数
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

func getStringFromRow(row []any, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i >= len(row) {
		return ""
	}
	s, ok := row[i].(string)
	if ok {
		return s
	}
	return fmt.Sprintf("%v", row[i])
}

func getFloatFromRow(row []any, idx map[string]int, key string) float64 {
	i, ok := idx[key]
	if !ok || i >= len(row) {
		return 0
	}
	switch v := row[i].(type) {
	case float64:
		return v
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	default:
		return 0
	}
}

func getInt64FromRow(row []any, idx map[string]int, key string) int64 {
	i, ok := idx[key]
	if !ok || i >= len(row) {
		return 0
	}
	switch v := row[i].(type) {
	case float64:
		return int64(v)
	case string:
		var n int64
		fmt.Sscanf(v, "%d", &n)
		return n
	default:
		return 0
	}
}
