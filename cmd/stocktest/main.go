package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/service"
	"auth-perm/internal/infra/eastmoney"
	"auth-perm/internal/infra/sina"
	"auth-perm/internal/infra/tdx"
	"auth-perm/internal/infra/tencent"
	"auth-perm/internal/infra/tushare"
)

func main() {
	log.SetOutput(io.Discard) // 隐藏 provider 内部日志，只显示结果表格
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 构建各领域 provider 列表
	stockProviders := []dm.StockListProvider{
		sina.NewClient(),
		eastmoney.NewClient(),
		tdx.NewClient(),
	}
	klineProviders := []dm.KlineProvider{
		sina.NewClient(),
		tencent.NewClient(),
		eastmoney.NewClient(),
		tdx.NewClient(),
	}
	boardProviders := []dm.BoardProvider{
		eastmoney.NewConceptClient(),
		tdx.NewClient(),
	}

	hasTushare := os.Getenv("TUSHARE_TOKEN") != ""
	if hasTushare {
		token := os.Getenv("TUSHARE_TOKEN")
		stockProviders = append(stockProviders, tushare.NewClient(token))
		klineProviders = append(klineProviders, tushare.NewClient(token))
	}

	onProbe := func(e service.ProbeEvent) {
		fmt.Printf("  ⏳ testing %s / %s ...\n", e.Name, e.Interface)
	}
	result := service.CheckProviderHealth(ctx, stockProviders, klineProviders, boardProviders, onProbe)

	// 按 provider 聚合结果
	type row struct {
		name      string
		stocklist string
		kline     string
		board     string
	}
	rows := map[string]*row{}
	order := []string{"sina", "tencent", "eastmoney", "tdx", "tushare"}
	for _, name := range order {
		rows[name] = &row{name: name}
	}

	for _, r := range result.Results {
		cell := ""
		if r.OK {
			if r.Count > 1 {
				cell = fmt.Sprintf("✅ %d (%s)", r.Count, r.Latency)
			} else {
				cell = fmt.Sprintf("✅ ok (%s)", r.Latency)
			}
		} else {
			cell = fmt.Sprintf("❌ %s", shortErr(r.Error))
		}
		if _, ok := rows[r.Name]; !ok {
			rows[r.Name] = &row{name: r.Name}
			order = append(order, r.Name)
		}
		switch r.Interface {
		case "stocklist":
			rows[r.Name].stocklist = cell
		case "kline":
			rows[r.Name].kline = cell
		case "board":
			rows[r.Name].board = cell
		}
	}

	// 填充默认值
	for _, r := range rows {
		if r.stocklist == "" {
			r.stocklist = "-"
		}
		if r.kline == "" {
			r.kline = "-"
		}
		if r.board == "" {
			r.board = "-"
		}
	}

	// 输出表格
	fmt.Println()
	fmt.Printf("  Provider Health: %s\n", result.Summary)
	fmt.Println()
	fmt.Printf("  %-10s │ %-24s │ %-24s │ %-24s\n", "PROVIDER", "STOCKLIST", "KLINE", "BOARD")
	fmt.Printf("  %-10s─┼─%-24s─┼─%-24s─┼─%-24s\n", "──────────", "────────────────────────", "────────────────────────", "────────────────────────")
	for _, name := range order {
		r, ok := rows[name]
		if !ok {
			continue
		}
		fmt.Printf("  %-10s │ %-24s │ %-24s │ %-24s\n", r.name, padCell(r.stocklist, 24), padCell(r.kline, 24), padCell(r.board, 24))
	}
	fmt.Println()
}

// padCell 用空格补齐到指定宽度（emoji 占 2 列宽）
func padCell(s string, width int) string {
	w := displayWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// displayWidth 计算字符串的显示宽度（emoji/wide char 算 2）
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if isWide(r) {
			w += 2
		} else {
			w += 1
		}
	}
	return w
}

func isWide(r rune) bool {
	return (r >= 0x1100 && r <= 0x115f) ||
		(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
		(r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) ||
		(r >= 0xfe10 && r <= 0xfe6f) ||
		(r >= 0xff01 && r <= 0xff60) ||
		(r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x1f000 && r <= 0x1faff) || // emoji
		(r >= 0x20000 && r <= 0x2ffff)
}

// shortErr 截断错误信息，只保留关键部分
func shortErr(err string) string {
	if err == "" {
		return "unknown"
	}
	// 提取关键错误类型
	for _, kw := range []string{"EOF", "timeout", "connection refused", "no such host", "rate limit"} {
		if strings.Contains(err, kw) {
			return kw
		}
	}
	// 截断到 30 字符
	if len(err) > 30 {
		return err[:27] + "..."
	}
	return err
}
