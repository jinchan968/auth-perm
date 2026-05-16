// F10Client 东方财富基本面数据客户端，获取行业/股本/EPS/BVPS/ROE 等财务指标。
// API: http://push2.eastmoney.com/api/qt/stock/get
// 实现 dm.F10Provider 接口，作为腾讯估值数据的补充。
package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/infra/httpclient"
)

const stockDetailURL = "http://push2.eastmoney.com/api/qt/stock/get"

// F10Client 东方财富 F10 客户端，实现 dm.F10Provider。
type F10Client struct {
	http *http.Client
}

// NewF10Client 创建东财 F10 客户端。
func NewF10Client() *F10Client {
	return &F10Client{http: httpclient.NewWithTimeout(60 * time.Second)}
}

func (c *F10Client) Name() string { return "eastmoney-f10" }

// FetchF10 批量获取股票财务指标。5 并发请求，返回行业/股本/EPS/BVPS/ROE。
func (c *F10Client) FetchF10(ctx context.Context, codes []string) ([]dm.TickerF10, error) {
	const concurrency = 5
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var results []dm.TickerF10
	var wg sync.WaitGroup

	for _, secid := range codes {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			f10, err := c.fetchOne(ctx, s)
			if err != nil {
				return
			}
			mu.Lock()
			results = append(results, f10)
			mu.Unlock()
		}(secid)
	}
	wg.Wait()
	return results, nil
}

// fetchOne 获取单只股票的 F10 数据。
// secid: 1.600519（沪市）/ 0.000001（深市）/ 2.830001（北交所，自动修正为 0 前缀）
func (c *F10Client) fetchOne(ctx context.Context, secid string) (dm.TickerF10, error) {
	// f12=代码, f14=名称, f100=行业, f116=总市值, f117=流通市值
	// f37=ROE, f40=营收, f45=净利润, f55=EPS, f162=PE, f167=PB
	// f115=总股本, f84=流通股本
	url := fmt.Sprintf("%s?secid=%s&fields=f12,f14,f100,f115,f84,f55,f48,f37",
		stockDetailURL, fixBseSecid(secid))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return dm.TickerF10{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return dm.TickerF10{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dm.TickerF10{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return dm.TickerF10{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var raw struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return dm.TickerF10{}, err
	}

	if raw.Data == nil || string(raw.Data) == "null" {
		return dm.TickerF10{}, fmt.Errorf("no data for %s", secid)
	}

	var m map[string]any
	if err := json.Unmarshal(raw.Data, &m); err != nil {
		return dm.TickerF10{}, err
	}

	return dm.TickerF10{
		TickerID:    secid,
		Industry:    getString(m, "f100"),
		TotalShares: getFloat(m, "f115") / 10000, // 股 → 万
		FloatShares: getFloat(m, "f84") / 10000,
		Eps:         getFloat(m, "f55"),
		Bvps:        getFloat(m, "f48"),
		Roe:         getFloat(m, "f37"),
		Source:      "eastmoney",
	}, nil
}
