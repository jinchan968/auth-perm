// Package tdx 通达信行情客户端，通过 TDX 协议获取 A 股数据。
//
// 实现接口：
//   - dm.StockListProvider（FetchAllStocks）— 通过 StockCount/StockList 分页获取沪深全量股票
//   - dm.KlineProvider（FetchKline）— 通过 StockKLine 获取日线 OHLCV（前复权）
//   - dm.BoardProvider（FetchBoards/FetchBoardStocks）— 通过 GetGroupedBlockFile 获取板块数据
//
// 作为新浪/腾讯/东方财富的备用数据源，通过 TDX 私有协议直连通达信服务器，
// 无需 HTTP、无需 token、无频率限制（TCP 长连接）。
// secid 约定：1.600519（沪市）/ 0.000001（深市）
package tdx

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bensema/gotdx"

	"auth-perm/internal/domain/newshock/dm"
)

// Client 通达信行情客户端，实现 dm.StockListProvider、dm.KlineProvider、dm.BoardProvider 三个接口。
type Client struct {
	mu     sync.Mutex
	client *gotdx.Client
}

// NewClient 创建通达信客户端
func NewClient() *Client {
	return &Client{}
}

func (c *Client) Name() string { return "tdx" }

// ensureConnected 确保 TDX 连接已建立（懒初始化）
func (c *Client) ensureConnected() error {
	if c.client != nil {
		return nil
	}

	mainHosts := gotdx.MainHostAddresses()
	if len(mainHosts) == 0 {
		return fmt.Errorf("no TDX hosts available")
	}

	opts := []gotdx.Option{
		gotdx.WithTCPAddress(mainHosts[0]),
		gotdx.WithAutoSelectFastest(true),
		gotdx.WithTimeoutSec(10),
	}
	if len(mainHosts) > 1 {
		opts = append(opts, gotdx.WithTCPAddressPool(mainHosts[1:]...))
	}

	client := gotdx.New(opts...)
	// 通过 StockCount 触发实际 TCP 连接
	_, err := client.StockCount(gotdx.MarketSH)
	if err != nil {
		client.Disconnect()
		return fmt.Errorf("tdx connect: %w", err)
	}
	c.client = client
	return nil
}

// FetchBoards 获取指定类型的板块列表。
// 通过下载通达信板块文件获取，boardType: 3=概念(block_gn.dat), 2=行业(block.dat)
func (c *Client) FetchBoards(ctx context.Context, boardType int) ([]dm.BoardInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	blockFile := mapBlockFile(boardType)
	groups, err := c.client.GetGroupedBlockFile(blockFile)
	if err != nil {
		// 连接可能断开，重置后重试一次
		c.client.Disconnect()
		c.client = nil
		if err2 := c.ensureConnected(); err2 != nil {
			return nil, fmt.Errorf("tdx reconnect: %w", err2)
		}
		groups, err = c.client.GetGroupedBlockFile(blockFile)
		if err != nil {
			return nil, fmt.Errorf("tdx GetGroupedBlockFile(%s): %w", blockFile, err)
		}
	}

	boards := make([]dm.BoardInfo, 0, len(groups))
	for _, g := range groups {
		if g.BlockName == "" {
			continue
		}
		boards = append(boards, dm.BoardInfo{
			Code: g.BlockName, // 通达信板块文件没有板块代码，用名称做标识
			Name: g.BlockName,
		})
	}

	log.Printf("[TDX] fetched %d boards from %s", len(boards), blockFile)
	return boards, nil
}

// FetchBoardStocks 获取板块内所有股票。
// boardCode 在通达信方案中实际是板块名称（来自 FetchBoards 返回的 Code）。
func (c *Client) FetchBoardStocks(ctx context.Context, boardCode string) ([]dm.BoardStockInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	// 遍历所有板块文件，找到匹配的板块
	blockFiles := []string{"block_gn.dat", "block.dat", "block_fg.dat"}
	for _, file := range blockFiles {
		groups, err := c.client.GetGroupedBlockFile(file)
		if err != nil {
			continue
		}

		for _, g := range groups {
			if g.BlockName != boardCode {
				continue
			}

			stocks := make([]dm.BoardStockInfo, 0, len(g.Codes))
			for _, code := range g.Codes {
				if code == "" {
					continue
				}
				market := guessMarket(code)
				prefix := "0"
				if market == 1 {
					prefix = "1"
				}
				stocks = append(stocks, dm.BoardStockInfo{
					Symbol: prefix + "." + code,
					Code:   code,
					Name:   "", // 板块文件不包含股票名称
					Market: market,
				})
			}
			return stocks, nil
		}
	}

	return nil, fmt.Errorf("board %q not found in any block file", boardCode)
}

// mapBlockFile 将统一 boardType 映射为通达信板块文件名
func mapBlockFile(boardType int) string {
	switch boardType {
	case 3: // 概念板块
		return "block_gn.dat"
	case 2: // 行业板块
		return "block.dat"
	default:
		return "block.dat"
	}
}

// guessMarket 根据股票代码猜测市场。
// 沪市：60x/688x/689x；深市：00x/002x/300x/301x
func guessMarket(code string) int {
	if strings.HasPrefix(code, "6") {
		return 1 // 沪市
	}
	return 0 // 深市
}

// ── StockListProvider 实现 ────────────────────────────────────────────

// Probe 轻量健康检查，只调用 StockCount 验证 TDX 连接可用。
func (c *Client) Probe(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureConnected(); err != nil {
		return err
	}
	_, err := c.client.StockCount(gotdx.MarketSH)
	return err
}

// FetchAllStocks 通过 TDX 协议分页获取沪深两市全部股票。
// 使用 StockCount 获取总数，StockList 分页拉取，每页 1600 条。
func (c *Client) FetchAllStocks(ctx context.Context) ([]dm.StockInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	var allStocks []dm.StockInfo

	// 遍历沪深两市
	for _, market := range []uint8{gotdx.MarketSH, gotdx.MarketSZ} {
		total, err := c.client.StockCount(market)
		if err != nil {
			log.Printf("[TDX] StockCount(market=%d) error: %v", market, err)
			continue
		}

		const pageSize = 1600 // gotdx.DefaultSecurityListCount
		for start := uint32(0); start < uint32(total); start += pageSize {
			if ctx.Err() != nil {
				return allStocks, ctx.Err()
			}

			items, err := c.client.StockList(market, start, pageSize)
			if err != nil {
				log.Printf("[TDX] StockList(market=%d, start=%d) error: %v", market, start, err)
				break
			}

			for _, sec := range items {
				if sec.Code == "" || sec.Name == "" {
					continue
				}
				prefix := "0"
				m := 0
				if market == gotdx.MarketSH {
					prefix = "1"
					m = 1
				}
				allStocks = append(allStocks, dm.StockInfo{
					Symbol: prefix + "." + sec.Code,
					Code:   sec.Code,
					Name:   sec.Name,
					Market: m,
				})
			}
		}
	}

	log.Printf("[TDX] fetched %d stocks", len(allStocks))
	if len(allStocks) < 2000 {
		return nil, fmt.Errorf("tdx returned only %d stocks (expected 5000+)", len(allStocks))
	}
	return allStocks, nil
}

// ── KlineProvider 实现 ────────────────────────────────────────────────

// FetchKline 通过 TDX 协议获取指定股票的日线 K 线数据（前复权）。
// secid: 1.600519（沪市）/ 0.000001（深市）
// days: 获取最近 N 天的数据
func (c *Client) FetchKline(ctx context.Context, secid string, days int) ([]dm.KlineBar, error) {
	if days <= 0 {
		days = 30
	}

	market, code, err := parseSecid(secid)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	// StockKLine(category, market, code, start, count, times, adjust)
	// category=4(KLINE_TYPE_DAILY), start=0(最新), adjust=1(前复权)
	bars, err := c.client.StockKLine(gotdx.KLINE_TYPE_DAILY, market, code, 0, uint16(days), 1, gotdx.AdjustQFQ)
	if err != nil {
		// 连接可能断开，重置后重试一次
		c.client.Disconnect()
		c.client = nil
		if err2 := c.ensureConnected(); err2 != nil {
			return nil, fmt.Errorf("tdx reconnect: %w", err2)
		}
		bars, err = c.client.StockKLine(gotdx.KLINE_TYPE_DAILY, market, code, 0, uint16(days), 1, gotdx.AdjustQFQ)
		if err != nil {
			return nil, fmt.Errorf("tdx StockKLine(%s): %w", secid, err)
		}
	}

	result := make([]dm.KlineBar, 0, len(bars))
	for _, bar := range bars {
		// DateTime 格式 "2006-01-02 15:04:05"，取前 10 位
		date := bar.DateTime
		if len(date) > 10 {
			date = date[:10]
		}
		result = append(result, dm.KlineBar{
			Date:      date,
			Open:      bar.Open,
			Close:     bar.Close,
			High:      bar.High,
			Low:       bar.Low,
			Volume:    int64(bar.Vol),
			Amount:    bar.Amount,
			ChangePct: bar.RiseRate,
			Turnover:  bar.Turnover,
		})
	}
	return result, nil
}

// parseSecid 解析 secid 格式为 TDX market 和 code。
// 1.600519 -> (MarketSH=1, "600519"), 0.000001 -> (MarketSZ=0, "000001")
func parseSecid(secid string) (uint8, string, error) {
	parts := strings.SplitN(secid, ".", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid secid format: %s", secid)
	}
	switch parts[0] {
	case "1":
		return gotdx.MarketSH, parts[1], nil
	case "0":
		return gotdx.MarketSZ, parts[1], nil
	default:
		return 0, "", fmt.Errorf("unknown market prefix: %s", parts[0])
	}
}
