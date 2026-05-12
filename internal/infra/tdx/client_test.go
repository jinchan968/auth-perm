package tdx

import (
	"context"
	"testing"
	"time"
)

func TestFetchAllStocks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实 API 调用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := NewClient()
	start := time.Now()
	stocks, err := client.FetchAllStocks(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("FetchAllStocks error: %v", err)
	}
	if len(stocks) < 2000 {
		t.Errorf("期望 >2000 只股票，实际 %d", len(stocks))
	}

	s := stocks[0]
	if s.Symbol == "" || s.Code == "" || s.Name == "" {
		t.Errorf("字段为空: %+v", s)
	}

	sh, sz := 0, 0
	for _, st := range stocks {
		if st.Market == 1 {
			sh++
		} else {
			sz++
		}
	}

	t.Logf("OK: %d stocks (SH=%d, SZ=%d), elapsed=%v, sample: %s %s %s",
		len(stocks), sh, sz, elapsed, s.Symbol, s.Code, s.Name)
}

func TestFetchKline(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实 API 调用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewClient()
	start := time.Now()
	bars, err := client.FetchKline(ctx, "1.600519", 10) // 茅台
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("FetchKline error: %v", err)
	}
	if len(bars) == 0 {
		t.Fatal("返回 0 条 K 线数据")
	}

	b := bars[0]
	if b.Date == "" {
		t.Error("日期为空")
	}
	if b.Open <= 0 || b.Close <= 0 {
		t.Errorf("价格异常: Open=%.2f, Close=%.2f", b.Open, b.Close)
	}

	t.Logf("OK: %d bars, elapsed=%v, latest: %s O=%.2f H=%.2f L=%.2f C=%.2f V=%d T=%.2f%%",
		len(bars), elapsed, b.Date, b.Open, b.High, b.Low, b.Close, b.Volume, b.Turnover)
}

func TestFetchBoards(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实 API 调用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewClient()

	// 测试概念板块
	start := time.Now()
	boards, err := client.FetchBoards(ctx, 3) // 3=概念
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("FetchBoards(concept) error: %v", err)
	}
	if len(boards) == 0 {
		t.Fatal("返回 0 个概念板块")
	}

	t.Logf("OK: %d concept boards, elapsed=%v, sample: %s", len(boards), elapsed, boards[0].Name)
}
