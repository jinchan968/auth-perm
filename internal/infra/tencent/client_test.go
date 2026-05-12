package tencent

import (
	"context"
	"testing"
	"time"
)

func TestFetchKline(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实 API 调用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewClient()
	start := time.Now()
	bars, err := client.FetchKline(ctx, "1.600519", 10) // 茅台最近 10 天
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

	t.Logf("OK: %d bars, elapsed=%v, latest: %s O=%.2f H=%.2f L=%.2f C=%.2f V=%d",
		len(bars), elapsed, b.Date, b.Open, b.High, b.Low, b.Close, b.Volume)
}
