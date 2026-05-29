package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auth-perm/internal/domain/multimodal/constant"
	"auth-perm/internal/domain/multimodal/vo"
	"auth-perm/internal/infra/opencode"
)

type MultimodalService struct {
	openCodeClient *opencode.Client
}

func NewMultimodalService(oc *opencode.Client) *MultimodalService {
	return &MultimodalService{openCodeClient: oc}
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
