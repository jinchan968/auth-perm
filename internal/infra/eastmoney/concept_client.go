package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/infra/httpclient"
)

const eastmoneyUT = "bd1d9ddb04089700cf9c27f6f7426281"

// ConceptClient 东方财富板块概念 API 客户端
type ConceptClient struct {
	http *http.Client
}

// NewConceptClient 创建板块概念客户端，内置 Chrome UA。
func NewConceptClient() *ConceptClient {
	return &ConceptClient{
		http: httpclient.NewWithTimeout(60 * time.Second),
	}
}

func (c *ConceptClient) Name() string { return "eastmoney" }

// boardListResponse 东方财富板块列表 API 响应
type boardListResponse struct {
	Data struct {
		Total int               `json:"total"`
		Diff  []json.RawMessage `json:"diff"`
	} `json:"data"`
}

// FetchBoards 获取指定类型的板块列表。
// boardType: 3=概念板块, 2=行业板块, 1=地域板块
func (c *ConceptClient) FetchBoards(ctx context.Context, boardType int) ([]dm.BoardInfo, error) {
	const pageSize = 1000
	var allBoards []dm.BoardInfo
	page := 1

	for {
		rawURL := fmt.Sprintf("%s?pn=%d&pz=%d&po=1&np=1&ut=%s&fltt=2&invt=2&fs=m:90+t:%d&fields=f12,f14",
			stockListURL, page, pageSize, eastmoneyUT, boardType)

		raw, err := c.doRequestWithRetry(ctx, rawURL)
		if err != nil {
			return nil, err
		}

		for _, item := range raw.Data.Diff {
			var m map[string]any
			if err := json.Unmarshal(item, &m); err != nil {
				continue
			}
			code := getString(m, "f12")
			name := getString(m, "f14")
			if code == "" || name == "" || code == "-" {
				continue
			}
			allBoards = append(allBoards, dm.BoardInfo{Code: code, Name: name})
		}

		if len(raw.Data.Diff) < pageSize || len(allBoards) >= raw.Data.Total {
			break
		}
		page++
	}

	return allBoards, nil
}

// FetchBoardStocks 获取指定板块内的所有股票。
// boardCode: 板块代码，如 BK0800
func (c *ConceptClient) FetchBoardStocks(ctx context.Context, boardCode string) ([]dm.BoardStockInfo, error) {
	const pageSize = 2000
	var allStocks []dm.BoardStockInfo
	page := 1

	for {
		rawURL := fmt.Sprintf("%s?pn=%d&pz=%d&po=1&np=1&ut=%s&fltt=2&invt=2&fs=b:%s&fields=f12,f13,f14",
			stockListURL, page, pageSize, eastmoneyUT, boardCode)

		raw, err := c.doRequestWithRetry(ctx, rawURL)
		if err != nil {
			return nil, err
		}

		for _, item := range raw.Data.Diff {
			info, err := parseBoardStock(item)
			if err != nil {
				continue
			}
			allStocks = append(allStocks, info)
		}

		if len(raw.Data.Diff) < pageSize || len(allStocks) >= raw.Data.Total {
			break
		}
		page++
	}

	return allStocks, nil
}

// doRequestWithRetry 发送 HTTP 请求并在遇到 EOF 时重试。
// 东方财富板块 API 偶发 EOF，通过轮换域名（HTTPS/HTTP）、添加浏览器头和重试机制缓解。
func (c *ConceptClient) doRequestWithRetry(ctx context.Context, rawURL string) (boardListResponse, error) {
	const maxRetries = 4
	// HTTPS 优先，HTTP 备选；多个 push 域名轮换
	candidates := []string{
		"https://push2.eastmoney.com",
		"http://80.push2.eastmoney.com",
		"http://push2.eastmoney.com",
		"http://push2ex.eastmoney.com",
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		tryURL := rawURL
		if attempt < len(candidates) {
			if u, err := url.Parse(rawURL); err == nil {
				u.Scheme = ""
				suffix := u.String() // 去掉 scheme 后的 //host/path
				tryURL = candidates[attempt] + suffix[len("//"+u.Host):]
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, tryURL, nil)
		if err != nil {
			return boardListResponse{}, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Referer", "https://data.eastmoney.com")
		req.Header.Set("User-Agent", httpclient.ChromeUA)

		resp, err := c.http.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return boardListResponse{}, fmt.Errorf("do request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return boardListResponse{}, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return boardListResponse{}, fmt.Errorf("board not found (404)")
		}
		if resp.StatusCode != http.StatusOK {
			return boardListResponse{}, fmt.Errorf("eastmoney api error %d: %s", resp.StatusCode, string(body))
		}

		var raw boardListResponse
		if err := json.Unmarshal(body, &raw); err != nil {
			return boardListResponse{}, fmt.Errorf("unmarshal: %w", err)
		}

		return raw, nil
	}

	return boardListResponse{}, fmt.Errorf("max retries exceeded")
}

// parseBoardStock 解析板块内股票数据
func parseBoardStock(raw json.RawMessage) (dm.BoardStockInfo, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return dm.BoardStockInfo{}, err
	}

	code := getString(m, "f12")
	name := getString(m, "f14")
	market := getInt(m, "f13")

	if code == "" || name == "" || code == "-" {
		return dm.BoardStockInfo{}, fmt.Errorf("invalid stock data")
	}

	prefix := "0"
	if market == 1 {
		prefix = "1"
	}
	symbol := prefix + "." + code

	return dm.BoardStockInfo{
		Symbol: symbol,
		Code:   code,
		Name:   name,
		Market: market,
	}, nil
}
