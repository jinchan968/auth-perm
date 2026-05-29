package constant

const (
	MultimodalModelID   = "mimo-v2.5"
	MultimodalModelName = "MiMo-V2.5 (多模态)"

	RecognizeSystemPrompt = "你是一个专业的图像分析助手。请仔细观察图片内容，用准确、详细的语言描述图片中的所有可见元素，包括文字、物体、场景、颜色、布局等。如果图片包含文字，请完整提取并翻译（如需）。分析应全面且有条理。"
	GenerateSystemPrompt  = "你是一个专业的 AI 绘图提示词（Prompt）专家。根据用户的简短描述，生成详细、高质量的英文绘图提示词，适用于 Midjourney、DALL-E、Stable Diffusion 等 AI 绘图工具。输出格式：先给出英文 prompt，然后附上中文翻译和使用建议。"

	MaxImageSizeMB = 10
	MaxImages      = 5
)
