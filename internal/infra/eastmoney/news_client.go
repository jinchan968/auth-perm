// NewsClient 东方财富个股新闻客户端。
// API: https://np-listapi.eastmoney.com/comm/web/getNewsByStock
// 实现 dm.NewsProvider 接口。
package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/infra/httpclient"
)

const stockNewsURL = "https://np-listapi.eastmoney.com/comm/web/getNewsByStock"

// NewsClient 东财个股新闻客户端，实现 dm.NewsProvider。
type NewsClient struct {
	http *http.Client
}

// NewNewsClient 创建东财新闻客户端。
func NewNewsClient() *NewsClient {
	return &NewsClient{http: httpclient.NewWithTimeout(30 * time.Second)}
}

func (c *NewsClient) Name() string { return "eastmoney-news" }

// newsResponse 东财新闻 API 响应
type newsResponse struct {
	Data struct {
		List []struct {
			Title       string `json:"Art_Title"`
			Content     string `json:"Art_Content"`
			ArticleURL  string `json:"Art_Url"` // 原文链接
			PublishTime string `json:"Art_ShowTime"`
			Code        string `json:"Art_Code"`
		} `json:"list"`
	} `json:"data"`
}

// FetchNews 获取指定股票的新闻列表。
// secid: 1.600519（沪市）/ 0.000001（深市）
func (c *NewsClient) FetchNews(ctx context.Context, secid string, limit int) ([]dm.TickerNews, error) {
	if limit <= 0 {
		limit = 20
	}

	code := secidToCode(secid)
	pageIndex := 1
	pageSize := min(limit, 50)

	url := fmt.Sprintf("%s?code=%s&page_index=%d&page_size=%d&source=web&client=web",
		stockNewsURL, code, pageIndex, pageSize)

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
		return nil, fmt.Errorf("eastmoney news api error %d", resp.StatusCode)
	}

	var raw newsResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	var results []dm.TickerNews
	for _, item := range raw.Data.List {
		title := cleanHTML(item.Title)
		if title == "" {
			continue
		}

		pubTime, _ := time.Parse("2006-01-02 15:04:05", item.PublishTime)
		if pubTime.IsZero() {
			pubTime, _ = time.Parse("2006-01-02", item.PublishTime)
		}

		results = append(results, dm.TickerNews{
			TickerID:    secid,
			Title:       title,
			Content:     truncate(cleanHTML(item.Content), 1000),
			Source:      "eastmoney",
			PublishTime: pubTime,
			URL:         item.ArticleURL,
		})
	}
	return results, nil
}

// cleanHTML 去除 HTML 标签，按 rune 遍历避免多字节字符截断。
func cleanHTML(s string) string {
	out := make([]rune, 0, len(s))
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			out = append(out, r)
		}
	}
	return strings.TrimSpace(string(out))
}

// truncate 截断字符串到指定 rune 数，避免按字节截断导致多字节字符显示乱码。
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
