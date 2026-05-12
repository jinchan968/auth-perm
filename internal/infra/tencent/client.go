// Package tencent 腾讯财经 HTTP API 客户端。
//
// 实现接口：
//   - dm.KlineProvider（FetchKline）— 前复权日线 K 线数据
//
// 作为新浪财经的 K 线备用数据源，当 sina API 失败时由 FailoverKlineProvider 自动切换。
// API：https://proxy.finance.qq.com/ifzqgtimg/appstock/app/newfqkline/get
// secid 约定：1.600519（沪市）/ 0.000001（深市）
package tencent

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

// Client 腾讯财经 HTTP 客户端
type Client struct {
	http *http.Client
}

// NewClient 创建客户端，超时 30 秒
func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 15 * time.Second,
			},
		},
	}
}

// Name 返回数据源名称
func (c *Client) Name() string { return "tencent" }

// FetchKline 获取指定股票的日线 K 线数据。
// symbol: secid 格式，1.600519（沪市）/ 0.000001（深市）
// days: 获取最近 N 天的数据
func (c *Client) FetchKline(ctx context.Context, symbol string, days int) ([]dm.KlineBar, error) {
	if days <= 0 {
		days = 30
	}

	// secid 转腾讯格式：1.600519 -> sh600519, 0.000001 -> sz000001
	qqSymbol, err := secidToQQSymbol(symbol)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://proxy.finance.qq.com/ifzqgtimg/appstock/app/newfqkline/get?param=%s,day,,,%d,qfq",
		qqSymbol, days)

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
		return nil, fmt.Errorf("tencent kline api error %d: %s", resp.StatusCode, string(body))
	}

	return parseKlineResponse(body, qqSymbol)
}

// klineResponse 腾讯 K 线 API 响应结构
type klineResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

// parseKlineResponse 解析腾讯 K 线 JSON 响应。
// 响应格式：{"code":0,"data":{"sh600519":{"qfqday":[["2026-05-08","1375.000","1371.050","1388.000","1370.010","40461.000",{},"0.27","477251.37",""],...]}}}
// 字段：[0]=日期, [1]=开盘, [2]=收盘, [3]=最高, [4]=最低, [5]=成交量, [7]=换手率, [8]=成交额
func parseKlineResponse(body []byte, qqSymbol string) ([]dm.KlineBar, error) {
	var resp klineResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// 解析 data 中对应股票的 K 线数组
	var dataMap map[string]json.RawMessage
	if err := json.Unmarshal(resp.Data, &dataMap); err != nil {
		return nil, fmt.Errorf("unmarshal data: %w", err)
	}

	stockData, ok := dataMap[qqSymbol]
	if !ok {
		return nil, fmt.Errorf("symbol %s not found in response", qqSymbol)
	}

	var stockMap map[string]json.RawMessage
	if err := json.Unmarshal(stockData, &stockMap); err != nil {
		return nil, fmt.Errorf("unmarshal stock data: %w", err)
	}

	// 优先取 qfqday（前复权日线），其次 day
	var rawLines []json.RawMessage
	for _, key := range []string{"qfqday", "day"} {
		if v, ok := stockMap[key]; ok {
			if err := json.Unmarshal(v, &rawLines); err == nil && len(rawLines) > 0 {
				break
			}
		}
	}

	if len(rawLines) == 0 {
		return nil, fmt.Errorf("no kline data for %s", qqSymbol)
	}

	bars := make([]dm.KlineBar, 0, len(rawLines))
	for _, raw := range rawLines {
		var fields []json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil || len(fields) < 6 {
			continue
		}

		date := parseJSONString(fields[0])
		open := parseJSONFloat(fields[1])
		close_ := parseJSONFloat(fields[2])
		high := parseJSONFloat(fields[3])
		low := parseJSONFloat(fields[4])
		volume := parseJSONInt64(fields[5])

		// 换手率和成交额可能不存在
		var turnover, amount float64
		if len(fields) > 7 {
			turnover = parseJSONFloat(fields[7])
		}
		if len(fields) > 8 {
			amount = parseJSONFloat(fields[8])
		}

		// 计算涨跌幅（基于开盘价近似）
		var changePct float64
		if open > 0 {
			changePct = (close_ - open) / open * 100
		}

		bars = append(bars, dm.KlineBar{
			Date:      date,
			Open:      open,
			Close:     close_,
			High:      high,
			Low:       low,
			Volume:    volume,
			Amount:    amount,
			ChangePct: changePct,
			Turnover:  turnover,
		})
	}

	return bars, nil
}

// secidToQQSymbol 将 secid 格式转为腾讯 symbol 格式
func secidToQQSymbol(secid string) (string, error) {
	parts := strings.SplitN(secid, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid secid format: %s", secid)
	}
	code := parts[1]
	switch parts[0] {
	case "1":
		return "sh" + code, nil
	case "0":
		return "sz" + code, nil
	default:
		return "", fmt.Errorf("unknown market prefix: %s", parts[0])
	}
}

func parseJSONString(v json.RawMessage) string {
	var s string
	json.Unmarshal(v, &s)
	return s
}

func parseJSONFloat(v json.RawMessage) float64 {
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	var f float64
	json.Unmarshal(v, &f)
	return f
}

func parseJSONInt64(v json.RawMessage) int64 {
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		i, _ := strconv.ParseInt(s, 10, 64)
		return i
	}
	var f float64
	json.Unmarshal(v, &f)
	return int64(f)
}
