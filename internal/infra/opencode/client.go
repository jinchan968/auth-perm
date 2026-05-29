package opencode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://opencode.ai/zen/go/v1"

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(apiKey string) *Client {
	if apiKey == "" {
		apiKey = "sk-no-key"
	}
	return &Client{
		baseURL: DefaultBaseURL,
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type chatRequest struct {
	Model           string        `json:"model"`
	Messages        []chatMessage `json:"messages"`
	ReasoningEffort string        `json:"reasoning_effort,omitempty"`
}

// ContentPart represents a single content part in a multimodal message.
type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL holds a data URL or HTTP URL for image content.
type ImageURL struct {
	URL string `json:"url"`
}

// MessageContent can serialize as either a plain string or a []ContentPart array,
// supporting both text-only and multimodal OpenAI-compatible requests.
type MessageContent struct {
	raw   string
	parts []ContentPart
}

// NewTextContent creates a plain-text message content (backward compatible).
func NewTextContent(text string) MessageContent {
	return MessageContent{raw: text}
}

// NewMultimodalContent creates a multimodal message content with text and/or images.
func NewMultimodalContent(parts ...ContentPart) MessageContent {
	return MessageContent{parts: parts}
}

func (mc MessageContent) MarshalJSON() ([]byte, error) {
	if mc.parts != nil {
		return json.Marshal(mc.parts)
	}
	return json.Marshal(mc.raw)
}

func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		mc.raw = s
		mc.parts = nil
		return nil
	}
	mc.raw = ""
	return json.Unmarshal(data, &mc.parts)
}

// Text extracts the plain text from either mode.
func (mc MessageContent) Text() string {
	if mc.raw != "" {
		return mc.raw
	}
	var sb strings.Builder
	for _, p := range mc.parts {
		if p.Type == "text" {
			sb.WriteString(p.Text)
		}
	}
	return sb.String()
}

type chatMessage struct {
	Role    string         `json:"role"`
	Content MessageContent `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type messagesRequest struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	System    string            `json:"system,omitempty"`
	Messages  []messagesContent `json:"messages"`
}

type messagesContent struct {
	Role    string         `json:"role"`
	Content MessageContent `json:"content"`
}

type messagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) Enabled() bool {
	return c.apiKey != "" && c.apiKey != "sk-no-key"
}

// Chat 发送聊天请求。useMessagesEndpoint 为 true 时走 /messages 端点，否则走 /chat/completions。
func (c *Client) Chat(ctx context.Context, modelID, systemPrompt, userPrompt string, useMessagesEndpoint bool, reasoningMode string) (string, error) {
	userMsg := NewTextContent(userPrompt)
	return c.doChatWithRetry(ctx, modelID, systemPrompt, userMsg, useMessagesEndpoint, reasoningMode)
}

// ChatMultimodal sends a multimodal chat request with text and optional images.
func (c *Client) ChatMultimodal(ctx context.Context, modelID, systemPrompt string, userParts []ContentPart, useMessagesEndpoint bool, reasoningMode string) (string, error) {
	userMsg := NewMultimodalContent(userParts...)
	return c.doChatWithRetry(ctx, modelID, systemPrompt, userMsg, useMessagesEndpoint, reasoningMode)
}

func (c *Client) doChatWithRetry(ctx context.Context, modelID, systemPrompt string, userContent MessageContent, useMessagesEndpoint bool, reasoningMode string) (string, error) {
	const maxAttempts = 2
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var content string
		var err error
		if useMessagesEndpoint {
			content, err = c.doMessagesRequest(ctx, modelID, systemPrompt, userContent)
		} else {
			content, err = c.doCompletionsRequest(ctx, modelID, systemPrompt, userContent, reasoningMode)
		}
		if err == nil {
			return content, nil
		}
		lastErr = err
		retry := strings.Contains(err.Error(), "api error 5")
		if !retry || attempt >= maxAttempts-1 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return "", lastErr
}

func (c *Client) doCompletionsRequest(ctx context.Context, modelID, systemPrompt string, userContent MessageContent, reasoningMode string) (string, error) {
	reqBody := chatRequest{
		Model: modelID,
		Messages: []chatMessage{
			{Role: "system", Content: NewTextContent(systemPrompt)},
			{Role: "user", Content: userContent},
		},
	}
	if reasoningMode != "" {
		reqBody.ReasoningEffort = reasoningMode
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("[%s] api error: %s", modelID, chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("[%s] no choices returned", modelID)
	}

	return chatResp.Choices[0].Message.Content.Text(), nil
}

func (c *Client) doMessagesRequest(ctx context.Context, modelID, systemPrompt string, userContent MessageContent) (string, error) {
	reqBody := messagesRequest{
		Model:     modelID,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages: []messagesContent{
			{Role: "user", Content: userContent},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	var msgResp messagesResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if msgResp.Error != nil {
		return "", fmt.Errorf("api error: %s", msgResp.Error.Message)
	}
	if len(msgResp.Content) == 0 {
		return "", fmt.Errorf("no content returned")
	}

	var sb strings.Builder
	for _, c := range msgResp.Content {
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	return sb.String(), nil
}
