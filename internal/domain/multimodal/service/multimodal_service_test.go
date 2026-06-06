package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"auth-perm/config"
	"auth-perm/internal/domain/multimodal/vo"
	"auth-perm/internal/infra/opencode"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestGenerateActualImageRequiresAPIKey(t *testing.T) {
	svc := NewMultimodalService(opencode.NewClient(""), &config.ImageGenerationConfig{
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-image-2",
	})

	_, err := svc.GenerateActualImage(context.Background(), "default", &vo.ImageGenerateRequest{Prompt: "a cat"})
	if err == nil {
		t.Fatal("expected missing API key error")
	}
	if err.Error() != "图片生成服务未配置" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateActualImageUpstreamError(t *testing.T) {
	svc := NewMultimodalService(opencode.NewClient(""), &config.ImageGenerationConfig{
		BaseURL: "https://example.test/v1",
		APIKey:  "test-key",
		Model:   "gpt-image-2",
	})
	svc.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString(`{"error":{"message":"invalid prompt"}}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := svc.GenerateActualImage(context.Background(), "default", &vo.ImageGenerateRequest{Prompt: "a cat"})
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if err.Error() != "图片生成服务错误: invalid prompt" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateActualImageSuccess(t *testing.T) {
	var gotPath string
	var gotAuth string
	svc := NewMultimodalService(opencode.NewClient(""), &config.ImageGenerationConfig{
		BaseURL: "https://example.test/v1",
		APIKey:  "test-key",
		Model:   "gpt-image-2",
	})
	svc.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotPath = req.URL.Path
		gotAuth = req.Header.Get("Authorization")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"data":[{"b64_json":"YWJj","revised_prompt":"a vivid cat"}]}`)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := svc.GenerateActualImage(context.Background(), "default", &vo.ImageGenerateRequest{
		Prompt:  "a cat",
		Size:    "1024x1024",
		Quality: "high",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v1/images/generations" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("unexpected auth header: %s", gotAuth)
	}
	if result.B64JSON != "YWJj" {
		t.Fatalf("unexpected b64: %s", result.B64JSON)
	}
	if result.ImageDataURL != "data:image/png;base64,YWJj" {
		t.Fatalf("unexpected data url: %s", result.ImageDataURL)
	}
	if result.RevisedPrompt != "a vivid cat" {
		t.Fatalf("unexpected revised prompt: %s", result.RevisedPrompt)
	}
	if result.ModelID != "gpt-image-2" {
		t.Fatalf("unexpected model id: %s", result.ModelID)
	}
}
