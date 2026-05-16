// QuoteClient 腾讯财经实时行情客户端，获取 PE/PB/市值/换手率/涨跌停等估值数据。
// API: https://qt.gtimg.cn/q=sh601006,sz000001,...
// 返回 GBK 编码，~ 分隔 88 个字段。
// 实现 dm.F10Provider 接口。
package tencent

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/infra/httpclient"
)

const quoteURL = "https://qt.gtimg.cn/q="

// QuoteClient 腾讯财经行情客户端，实现 dm.F10Provider。
type QuoteClient struct {
	http *http.Client
}

// NewQuoteClient 创建腾讯行情客户端。
func NewQuoteClient() *QuoteClient {
	return &QuoteClient{http: httpclient.New()}
}

func (c *QuoteClient) Name() string { return "tencent-quote" }

// FetchF10 批量获取股票估值数据。
// codes 为 secid 格式（1.600519 / 0.000001），自动转换为腾讯格式（sh601006 / sz000001）。
// 每批最多 50 只，避免 URL 过长。
func (c *QuoteClient) FetchF10(ctx context.Context, codes []string) ([]dm.TickerF10, error) {
	const batchSize = 50
	var all []dm.TickerF10

	for i := 0; i < len(codes); i += batchSize {
		end := min(i+batchSize, len(codes))
		batch := codes[i:end]

		tcCodes := make([]string, 0, len(batch))
		codeMap := make(map[string]string) // tcCode → secid
		for _, secid := range batch {
			tc := secidToTencent(secid)
			tcCodes = append(tcCodes, tc)
			codeMap[tc] = secid
		}

		url := quoteURL + strings.Join(tcCodes, ",")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}

		body, err := httpclient.DecodeGBKResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		parsed, err := c.parseQuoteResponse(body, codeMap)
		if err != nil {
			return nil, err
		}
		all = append(all, parsed...)
	}
	return all, nil
}

// parseQuoteResponse 解析腾讯行情响应。
// 每行格式：v_sh601006="1~大秦铁路~601006~...~"（88 个字段，~ 分隔）
func (c *QuoteClient) parseQuoteResponse(body string, codeMap map[string]string) ([]dm.TickerF10, error) {
	var results []dm.TickerF10

	for _, line := range strings.Split(body, ";") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "~") {
			continue
		}

		// 提取代码部分：v_sh601006="..." → sh601006
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimPrefix(parts[0], "v_")
		key = strings.TrimSpace(key)

		// 提取引号内的值
		valPart := parts[1]
		start := strings.Index(valPart, "\"")
		end := strings.LastIndex(valPart, "\"")
		if start < 0 || end <= start {
			continue
		}
		vals := strings.Split(valPart[start+1:end], "~")
		if len(vals) < 53 {
			continue
		}

		secid, ok := codeMap[key]
		if !ok {
			continue
		}

		f10 := dm.TickerF10{
			TickerID:     secid,
			PeTtm:        parseFloat(vals[39]),
			PeStatic:     parseFloat(vals[52]),
			Pb:           parseFloat(vals[46]),
			TotalMcap:    parseFloat(vals[44]),
			FloatMcap:    parseFloat(vals[45]),
			TurnoverRate: parseFloat(vals[38]),
			VolumeRatio:  parseFloat(vals[49]),
			LimitUp:      parseFloat(vals[47]),
			LimitDown:    parseFloat(vals[48]),
			Source:       "tencent",
		}

		// 跳过无效数据（所有估值字段为 0 说明数据不可用）
		if f10.PeTtm == 0 && f10.Pb == 0 && f10.TotalMcap == 0 {
			continue
		}

		results = append(results, f10)
	}
	return results, nil
}

// secidToTencent 将 secid 转为腾讯格式：1.600519 → sh600519, 0.000001 → sz000001
func secidToTencent(secid string) string {
	parts := strings.SplitN(secid, ".", 2)
	if len(parts) != 2 {
		return secid
	}
	switch parts[0] {
	case "1":
		return "sh" + parts[1]
	case "0":
		return "sz" + parts[1]
	case "2":
		return "bj" + parts[1]
	default:
		return secid
	}
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
