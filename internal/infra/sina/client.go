// Package sina 新浪财经 HTTP API 客户端。
//
// 实现接口：
//   - dm.StockListProvider（FetchAllStocks）— 腾讯财经批量行情 API，按代码段遍历
//   - dm.KlineProvider（FetchKline）— 新浪 K 线 API，返回最近 N 天日线 OHLCV
//
// secid 约定：1.600519（沪市）/ 0.000001（深市）
// 无需注册、无需 token，直接 HTTP GET。
package sina

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"auth-perm/internal/domain/newshock/dm"
)

const (
	// 腾讯股票列表 API（无需认证）
	stockListURL  = "https://qt.gtimg.cn/q=sh601006"
	stockListBase = "https://qt.gtimg.cn/q="
	// 新浪 K 线 API
	klineURL = "https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData"
)

// Client 新浪财经 API 客户端，实现 dm.StockListProvider 和 dm.KlineProvider 接口。
type Client struct {
	http *http.Client
}

// NewClient 创建客户端
func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 60 * time.Second},
	}
}

// Name 返回数据源名称
func (c *Client) Name() string { return "sina" }

// Probe 轻量健康检查，只查 1 批（50 只股票）验证 API 可用。
func (c *Client) Probe(ctx context.Context) error {
	codes := make([]string, 50)
	for i := 0; i < 50; i++ {
		codes[i] = fmt.Sprintf("sh600%03d", i)
	}
	query := strings.Join(codes, ",")
	url := stockListBase + query
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

// FetchAllStocks 获取所有 A 股股票列表。
// 使用腾讯财经批量行情接口，按代码段分批获取。
func (c *Client) FetchAllStocks(ctx context.Context) ([]dm.StockInfo, error) {
	// A 股代码范围：
	// 沪市：600000-609999, 688000-688999（科创板）
	// 深市：000001-009999, 300000-309999（创业板）
	// A 股代码范围（每批 50 个代码查询，间隔 300ms）
	codeRanges := []struct {
		prefix string
		start  int
		end    int
		market int
	}{
		{"sh", 600000, 605000, 1}, // 沪市主板
		{"sh", 605000, 610000, 1}, // 沪市主板
		{"sh", 688000, 689000, 1}, // 科创板
		{"sz", 0, 5000, 0},        // 深市主板
		{"sz", 5000, 10000, 0},    // 深市主板
		{"sz", 300000, 305000, 0}, // 创业板
		{"sz", 305000, 310000, 0}, // 创业板
	}

	var allStocks []dm.StockInfo
	for _, r := range codeRanges {
		before := len(allStocks)
		stocks, err := c.fetchStockRange(ctx, r.prefix, r.start, r.end, r.market)
		if err != nil {
			log.Printf("[Sina] fetch %s%06d-%06d error: %v", r.prefix, r.start, r.end, err)
			continue
		}
		allStocks = append(allStocks, stocks...)
		added := len(allStocks) - before
		log.Printf("[Sina] %s%06d-%06d: %d stocks (total %d)", r.prefix, r.start, r.end, added, len(allStocks))
	}
	// A 股正常应有 5000+ 只，低于 2000 只说明大部分请求失败（如限流），应触发 failover
	if len(allStocks) < 2000 {
		return nil, fmt.Errorf("sina returned only %d stocks (expected 5000+), likely rate limited", len(allStocks))
	}
	log.Printf("[Sina] total fetched: %d stocks", len(allStocks))
	return allStocks, nil
}

// fetchStockRange 获取指定代码范围内的有效股票。
// 每批 50 个代码，间隔 300ms，遇限流自动重试。
func (c *Client) fetchStockRange(ctx context.Context, prefix string, start, end, market int) ([]dm.StockInfo, error) {
	const batchSize = 50
	var stocks []dm.StockInfo

	for i := start; i < end; i += batchSize {
		if ctx.Err() != nil {
			return stocks, ctx.Err()
		}

		batchEnd := i + batchSize
		if batchEnd > end {
			batchEnd = end
		}

		// 构建批量查询字符串：sh600000,sh600001,...
		var codes []string
		for code := i; code < batchEnd; code++ {
			codes = append(codes, fmt.Sprintf("%s%06d", prefix, code))
		}
		query := strings.Join(codes, ",")

		// 重试最多 3 次（应对 HTTP/2 ENHANCE_YOUR_CALM 限流）
		var body []byte
		var lastErr error
		for attempt := 0; attempt < 3; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * time.Second)
			}

			url := stockListBase + query
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				lastErr = err
				continue
			}

			resp, err := c.http.Do(req)
			if err != nil {
				lastErr = err
				log.Printf("[Sina] batch %s%06d attempt %d error: %v", prefix, i, attempt+1, err)
				continue
			}

			body, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = err
				continue
			}
			lastErr = nil
			break
		}
		if lastErr != nil {
			log.Printf("[Sina] batch %s%06d failed after 3 attempts: %v", prefix, i, lastErr)
			time.Sleep(2 * time.Second) // 限流后额外等待
			continue
		}

		if len(body) == 0 {
			continue
		}

		// 腾讯 API 返回 GBK 编码，转 UTF-8
		utf8Body, _, err := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), body)
		if err != nil {
			utf8Body = body // 转换失败用原始数据
		}

		parsed := c.parseStockListResponse(string(utf8Body), market)
		stocks = append(stocks, parsed...)

		time.Sleep(300 * time.Millisecond)
	}
	return stocks, nil
}

// parseStockListResponse 解析腾讯财经批量行情响应
// 格式：v_sh600519="1~贵州茅台~600519~1500.00~..."
// 字段：0=未知 1=名称 2=代码 3=最新价 ... 32=市场(1沪/0深)
func (c *Client) parseStockListResponse(body string, market int) []dm.StockInfo {
	lines := strings.Split(body, ";")
	var stocks []dm.StockInfo
	re := regexp.MustCompile(`v_\w+="(.*)"`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		fields := strings.Split(matches[1], "~")
		if len(fields) < 5 {
			continue
		}
		name := fields[1]
		code := fields[2]
		price := fields[3]
		if name == "" || code == "" || price == "" || price == "0.000" {
			continue
		}

		// 从行中提取前缀判断市场
		prefixPart := strings.Split(strings.TrimPrefix(line, "v_"), "~")[0]
		m := market
		if strings.HasPrefix(prefixPart, "sh") {
			m = 1
		} else if strings.HasPrefix(prefixPart, "sz") {
			m = 0
		}

		secid := fmt.Sprintf("%d.%s", m, code)
		stocks = append(stocks, dm.StockInfo{
			Symbol: secid,
			Code:   code,
			Name:   name,
			Market: m,
		})
	}
	return stocks
}

// FetchKline 获取指定股票的日线 K 线数据。
// secid: 1.600519（沪市）/ 0.000001（深市）
// days: 获取最近 N 天的数据（新浪最多返回 1023 天）
func (c *Client) FetchKline(ctx context.Context, secid string, days int) ([]dm.KlineBar, error) {
	if days <= 0 {
		days = 30
	}
	if days > 1023 {
		days = 1023
	}

	// secid 转新浪格式：1.600519 -> sh600519, 0.000001 -> sz000001
	symbol, err := secidToSinaSymbol(secid)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s?symbol=%s&scale=240&ma=no&datalen=%d", klineURL, symbol, days)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn")

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
		return nil, fmt.Errorf("sina kline api error %d: %s", resp.StatusCode, string(body))
	}

	// 响应是 JSON 数组
	type sinaKline struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}

	var raw []sinaKline
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal: %w (body: %s)", err, string(body[:min(len(body), 200)]))
	}

	var bars []dm.KlineBar
	for _, r := range raw {
		open, _ := strconv.ParseFloat(r.Open, 64)
		close_, _ := strconv.ParseFloat(r.Close, 64)
		high, _ := strconv.ParseFloat(r.High, 64)
		low, _ := strconv.ParseFloat(r.Low, 64)
		volume, _ := strconv.ParseInt(r.Volume, 10, 64)

		// 计算涨跌幅（新浪不直接返回）
		var changePct float64
		if open > 0 {
			changePct = (close_ - open) / open * 100
		}

		bars = append(bars, dm.KlineBar{
			Date:      r.Day,
			Open:      open,
			Close:     close_,
			High:      high,
			Low:       low,
			Volume:    volume,
			ChangePct: changePct,
		})
	}
	return bars, nil
}

// secidToSinaSymbol 将 secid 格式转为新浪 symbol 格式
// 1.600519 -> sh600519, 0.000001 -> sz000001
func secidToSinaSymbol(secid string) (string, error) {
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
