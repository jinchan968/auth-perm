package vo

// RecognizeImageRequest 识图请求
type RecognizeImageRequest struct {
	Prompt string   `json:"prompt"`
	Images []string `json:"images" binding:"required,min=1,max=5"`
}

// GenerateImageRequest 生成提示词请求
type GenerateImageRequest struct {
	Prompt string `json:"prompt" binding:"required,max=2000"`
	Style  string `json:"style"`
}

// MultimodalResponse 通用响应
type MultimodalResponse struct {
	Content    string `json:"content"`
	DurationMs int64  `json:"duration_ms"`
	ModelID    string `json:"model_id"`
}
