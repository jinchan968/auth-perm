//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"auth-perm/internal/infra/httpclient"
)

func main() {
	client := httpclient.NewWithTimeout(60 * time.Second)
	const pageSize = 500
	var all []map[string]any
	page := 1

	for {
		url := fmt.Sprintf("http://80.push2.eastmoney.com/api/qt/clist/get?pn=%d&pz=%d&po=1&np=1&fltt=2&invt=2&fs=m:0+t:6,m:0+t:80,m:0+t:81&fields=f12,f13,f14", page, pageSize)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("page %d error: %v", page, err)
			break
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var raw struct {
			Data struct {
				Total int               `json:"total"`
				Diff  []json.RawMessage `json:"diff"`
			} `json:"data"`
		}
		json.Unmarshal(body, &raw)

		for _, item := range raw.Data.Diff {
			var m map[string]any
			json.Unmarshal(item, &m)
			all = append(all, m)
		}

		fmt.Printf("page %d: got %d (total=%d, accumulated=%d)\n", page, len(raw.Data.Diff), raw.Data.Total, len(all))

		if len(raw.Data.Diff) < pageSize || len(all) >= raw.Data.Total {
			break
		}
		page++
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("\nTotal fetched: %d\n", len(all))

	// 统计
	codePrefix := map[string]int{}
	st, pt, 退市, b股 := 0, 0, 0, 0

	for _, m := range all {
		code, _ := m["f12"].(string)
		name, _ := m["f14"].(string)
		market, _ := m["f13"].(float64)
		upper := strings.ToUpper(name)

		if strings.HasPrefix(upper, "ST") {
			st++
			continue
		}
		if strings.HasPrefix(upper, "PT") {
			pt++
			continue
		}
		if strings.Contains(upper, "退") {
			退市++
			continue
		}
		if strings.HasPrefix(code, "900") || strings.HasPrefix(code, "200") {
			b股++
			continue
		}

		prefix := code[:2]
		codePrefix[prefix]++
		_ = market
	}

	fmt.Printf("\n=== 过滤统计 ===\n")
	fmt.Printf("ST: %d\n", st)
	fmt.Printf("PT: %d\n", pt)
	fmt.Printf("含'退': %d\n", 退市)
	fmt.Printf("B股(900/200): %d\n", b股)
	fmt.Printf("过滤后有效: %d\n", len(all)-st-pt-退市-b股)

	fmt.Printf("\n=== 有效股票按代码前缀 ===\n")
	for _, p := range []string{"60", "68", "00", "30", "83", "87", "43", "11", "12", "15", "16", "18", "50", "51", "52", "56", "58"} {
		if c, ok := codePrefix[p]; ok {
			fmt.Printf("  %sxxxx: %d\n", p, c)
		}
	}
	// 其他前缀
	for p, c := range codePrefix {
		found := false
		for _, known := range []string{"60", "68", "00", "30", "83", "87", "43", "11", "12", "15", "16", "18", "50", "51", "52", "56", "58"} {
			if p == known {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  %sxxxx: %d (其他)\n", p, c)
		}
	}

	_ = strconv.Itoa(page)
}
