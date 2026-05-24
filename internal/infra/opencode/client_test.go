package opencode

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestChat_GLM51(t *testing.T) {
	apiKey := os.Getenv("OPENCODE_API_KEY")
	if apiKey == "" {
		t.Skip("OPENCODE_API_KEY not set")
	}

	client := NewClient(apiKey)
	if !client.Enabled() {
		t.Fatal("client not enabled — key is empty or invalid")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	modelID := "glm-5.1"
	systemPrompt := "你是一个有帮助的助手。"
	userPrompt := "hi"

	for attempt := 1; attempt <= 3; attempt++ {
		content, err := client.Chat(ctx, modelID, systemPrompt, userPrompt, false, "low")
		if err == nil {
			if content == "" {
				t.Fatalf("attempt %d: got empty content", attempt)
			}
			t.Logf("attempt %d: success, content length=%d", attempt, len(content))
			return
		}
		t.Logf("attempt %d: error = %v", attempt, err)
		if !strings.Contains(err.Error(), "Missing API key") {
			t.Fatalf("unexpected error (attempt %d): %v", attempt, err)
		}
		if attempt < 3 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	t.Fatal("all 3 attempts failed with Missing API key")
}

func TestChat_ConcurrentAModels(t *testing.T) {
	apiKey := os.Getenv("OPENCODE_API_KEY")
	if apiKey == "" {
		t.Skip("OPENCODE_API_KEY not set")
	}

	client := NewClient(apiKey)
	if !client.Enabled() {
		t.Fatal("client not enabled — key is empty or invalid")
	}

	models := []struct {
		id       string
		useMsgEP bool
	}{
		{id: "glm-5.1", useMsgEP: false},
		{id: "kimi-k2.6", useMsgEP: false},
		{id: "mimo-v2.5-pro", useMsgEP: false},
		{id: "deepseek-v4-pro", useMsgEP: false},
		{id: "qwen3.6-plus", useMsgEP: false},
	}

	systemPrompt := "你是一个有帮助的助手。"
	userPrompt := "hi"

	var wg sync.WaitGroup
	results := make([]struct {
		model   string
		content string
		err     error
		elapsed time.Duration
	}, len(models))

	for i, m := range models {
		wg.Add(1)
		go func(idx int, modelID string, useMsg bool) {
			defer wg.Done()
			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			content, err := client.Chat(ctx, modelID, systemPrompt, userPrompt, useMsg, "low")
			results[idx] = struct {
				model   string
				content string
				err     error
				elapsed time.Duration
			}{model: modelID, content: content, err: err, elapsed: time.Since(start)}
		}(i, m.id, m.useMsgEP)
	}
	wg.Wait()

	failed := 0
	for _, r := range results {
		if r.err != nil {
			t.Logf("%s: ERROR (%v) — %v", r.model, r.elapsed, r.err)
			if strings.Contains(r.err.Error(), "Missing API key") {
				failed++
			}
		} else {
			t.Logf("%s: OK (%v) len=%d", r.model, r.elapsed, len(r.content))
		}
	}
	if failed > 0 {
		t.Errorf("%d model(s) failed with Missing API key", failed)
	}
}
