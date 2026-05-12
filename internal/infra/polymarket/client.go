// polymarket 包封装 Polymarket Gamma API 的 HTTP 客户端。
// 用于获取活跃的预测市场数据（标题、概率、交易量等）。
//
// API 文档：https://docs.polymarket.com/
// 使用的端点：GET /markets?active=true&closed=false&limit=N
//
// 响应中的 outcomes 和 outcomePrices 是 JSON 数组字符串，需要二次解析。
// volume 是字符串格式的数字，需要转换为 float64。
package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// gammaAPIBase Polymarket Gamma API 基础地址
const gammaAPIBase = "https://gamma-api.polymarket.com"

// Client Polymarket API HTTP 客户端
type Client struct {
	http *http.Client
}

// NewClient 创建 Polymarket 客户端，超时 30 秒
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

// Market 解析后的 Polymarket 市场数据（内部使用，对外返回的干净结构）
type Market struct {
	ConditionID string  `json:"condition_id"` // 市场唯一标识（Polymarket condition ID）
	Title       string  `json:"title"`        // 市场标题（如 "Will X happen by Y?"）
	Description string  `json:"description"`  // 市场详细描述
	Outcome     string  `json:"outcome"`      // 第一个选项的名称（通常是 "Yes"）
	Probability float64 `json:"probability"`  // 第一个选项的概率（0-1）
	Volume      float64 `json:"volume"`       // 总交易量（美元）
	Active      bool    `json:"active"`       // 是否活跃
	Closed      bool    `json:"closed"`       // 是否已关闭
}

// marketResponse Polymarket API 原始响应结构（字段格式与内部 Market 不同）
// outcomes/outcomePrices 是 JSON 数组字符串，volume 是数字字符串
type marketResponse struct {
	ConditionID   string `json:"conditionId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Outcomes      string `json:"outcomes"`      // JSON 数组字符串，如 '["Yes","No"]'
	OutcomePrices string `json:"outcomePrices"` // JSON 数组字符串，如 '[0.65,0.35]'
	Volume        string `json:"volume"`        // 数字字符串，如 "1234567.89"
	Active        bool   `json:"active"`
	Closed        bool   `json:"closed"`
}

// FetchActiveMarkets 从 Polymarket API 获取活跃的预测市场（最多 limit 个）。
// 返回解析后的 Market 列表，包含标题、概率、交易量等字段。
// 只返回 active=true 且 closed=false 的市场。
func (c *Client) FetchActiveMarkets(ctx context.Context, limit int) ([]Market, error) {
	url := fmt.Sprintf("%s/markets?limit=%d&active=true&closed=false", gammaAPIBase, limit)
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
		return nil, fmt.Errorf("polymarket api error %d: %s", resp.StatusCode, string(body))
	}

	var raw []marketResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	var markets []Market
	for _, r := range raw {
		m := Market{
			ConditionID: r.ConditionID,
			Title:       r.Title,
			Description: r.Description,
			Active:      r.Active,
			Closed:      r.Closed,
		}
		// 解析 outcomes 和 outcomePrices
		var outcomes []string
		json.Unmarshal([]byte(r.Outcomes), &outcomes)
		var prices []float64
		json.Unmarshal([]byte(r.OutcomePrices), &prices)

		if len(outcomes) > 0 {
			m.Outcome = outcomes[0]
		}
		if len(prices) > 0 {
			m.Probability = prices[0]
		}
		// volume 是字符串
		var vol float64
		fmt.Sscanf(r.Volume, "%f", &vol)
		m.Volume = vol

		markets = append(markets, m)
	}
	return markets, nil
}
