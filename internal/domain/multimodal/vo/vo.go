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

// ImageGenerateRequest 生图请求
type ImageGenerateRequest struct {
	Prompt  string `json:"prompt" binding:"required,max=4000"`
	Size    string `json:"size"`
	Quality string `json:"quality"`
}

var validImageSizes = map[string]struct{}{
	"":          {},
	"auto":      {},
	"1024x1024": {},
	"1536x1024": {},
	"1024x1536": {},
}

var validImageQualities = map[string]struct{}{
	"":       {},
	"auto":   {},
	"low":    {},
	"medium": {},
	"high":   {},
}

func IsValidImageSize(size string) bool {
	_, ok := validImageSizes[size]
	return ok
}

func IsValidImageQuality(quality string) bool {
	_, ok := validImageQualities[quality]
	return ok
}

// MultimodalResponse 通用响应
type MultimodalResponse struct {
	Content    string `json:"content"`
	DurationMs int64  `json:"duration_ms"`
	ModelID    string `json:"model_id"`
}

// ImageGenerateResponse 生图响应
type ImageGenerateResponse struct {
	ImageDataURL  string `json:"image_data_url"`
	B64JSON       string `json:"b64_json"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
	DurationMs    int64  `json:"duration_ms"`
	ModelID       string `json:"model_id"`
}
