package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"auth-perm/config"
	"auth-perm/internal/domain/multimodal/constant"
	"auth-perm/internal/domain/multimodal/vo"
	"auth-perm/internal/infra/opencode"
)

type MultimodalService struct {
	openCodeClient *opencode.Client
	imageGen       *config.ImageGenerationConfig
	httpClient     *http.Client
}

func NewMultimodalService(oc *opencode.Client, imageGen *config.ImageGenerationConfig) *MultimodalService {
	return &MultimodalService{
		openCodeClient: oc,
		imageGen:       imageGen,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// RecognizeImage sends images to the model for analysis.
func (s *MultimodalService) RecognizeImage(ctx context.Context, tenantID, prompt string, images []string) (*vo.MultimodalResponse, error) {
	if !s.openCodeClient.Enabled() {
		return nil, fmt.Errorf("OpenCode 客户端未配置")
	}

	parts := make([]opencode.ContentPart, 0, 1+len(images))

	text := prompt
	if text == "" {
		text = "请分析这张图片的内容。"
	}
	parts = append(parts, opencode.ContentPart{Type: "text", Text: text})

	for _, img := range images {
		parts = append(parts, opencode.ContentPart{
			Type:     "image_url",
			ImageURL: &opencode.ImageURL{URL: img},
		})
	}

	start := time.Now()
	content, err := s.openCodeClient.ChatMultimodal(
		ctx,
		constant.MultimodalModelID,
		constant.RecognizeSystemPrompt,
		parts,
		false,
		"low",
	)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("识图失败: %w", err)
	}

	return &vo.MultimodalResponse{
		Content:    content,
		DurationMs: duration,
		ModelID:    constant.MultimodalModelID,
	}, nil
}

// GenerateImage generates an AI drawing prompt based on user description.
func (s *MultimodalService) GenerateImage(ctx context.Context, tenantID, prompt, style string) (*vo.MultimodalResponse, error) {
	if !s.openCodeClient.Enabled() {
		return nil, fmt.Errorf("OpenCode 客户端未配置")
	}

	var sb strings.Builder
	sb.WriteString(prompt)
	if style != "" {
		sb.WriteString(fmt.Sprintf("\n\n风格偏好: %s", style))
	}

	start := time.Now()
	content, err := s.openCodeClient.Chat(
		ctx,
		constant.MultimodalModelID,
		constant.GenerateSystemPrompt,
		sb.String(),
		false,
		"low",
	)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("生成提示词失败: %w", err)
	}

	return &vo.MultimodalResponse{
		Content:    content,
		DurationMs: duration,
		ModelID:    constant.MultimodalModelID,
	}, nil
}

type imageGenerationRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Size    string `json:"size,omitempty"`
	Quality string `json:"quality,omitempty"`
}

type imageGenerationResponse struct {
	Data []struct {
		B64JSON       string `json:"b64_json"`
		RevisedPrompt string `json:"revised_prompt"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

// GenerateActualImage calls an OpenAI Images-compatible endpoint to generate an image.
func (s *MultimodalService) GenerateActualImage(ctx context.Context, tenantID string, req *vo.ImageGenerateRequest) (*vo.ImageGenerateResponse, error) {
	cfg := s.imageGen
	if cfg == nil || strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("图片生成服务未配置")
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "gpt-image-2"
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	requestBody := imageGenerationRequest{
		Model:  model,
		Prompt: strings.TrimSpace(req.Prompt),
	}
	if req.Size != "" && req.Size != "auto" {
		requestBody.Size = req.Size
	}
	if req.Quality != "" && req.Quality != "auto" {
		requestBody.Quality = req.Quality
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化生图请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/images/generations", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建生图请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	start := time.Now()
	resp, err := s.httpClient.Do(httpReq)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		return nil, fmt.Errorf("请求图片生成服务失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取生图响应失败: %w", err)
	}

	var parsed imageGenerationResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("解析生图响应失败: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		msg := strings.TrimSpace(parsedErrorMessage(parsed.Error))
		if msg == "" {
			msg = fmt.Sprintf("图片生成服务返回状态码 %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("图片生成服务错误: %s", msg)
	}

	if len(parsed.Data) == 0 || strings.TrimSpace(parsed.Data[0].B64JSON) == "" {
		return nil, fmt.Errorf("图片生成失败: 响应中没有图片数据")
	}

	b64 := strings.TrimSpace(parsed.Data[0].B64JSON)
	return &vo.ImageGenerateResponse{
		ImageDataURL:  "data:image/png;base64," + b64,
		B64JSON:       b64,
		RevisedPrompt: parsed.Data[0].RevisedPrompt,
		DurationMs:    duration,
		ModelID:       model,
	}, nil
}

func parsedErrorMessage(apiErr *struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    any    `json:"code"`
}) string {
	if apiErr == nil {
		return ""
	}
	return apiErr.Message
}
