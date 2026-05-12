package main

import (
	"context"
	"fmt"

	"auth-perm/internal/infra/tencent"
)

func main() {
	c := tencent.NewClient()
	ctx := context.Background()

	symbols := []string{"1.600519", "0.000001", "0.300750"}
	for _, s := range symbols {
		bars, err := c.FetchKline(ctx, s, 5)
		if err != nil {
			fmt.Printf("%s error: %v\n", s, err)
			continue
		}
		fmt.Printf("%s: %d bars\n", s, len(bars))
		for _, b := range bars {
			fmt.Printf("  %s O=%.2f C=%.2f H=%.2f L=%.2f V=%d Turnover=%.2f%% Amount=%.0f\n",
				b.Date, b.Open, b.Close, b.High, b.Low, b.Volume, b.Turnover, b.Amount)
		}
	}
}
