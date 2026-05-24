package constant

type AIModelType string

const (
	EndpointChat     AIModelType = "chat"
	EndpointMessages AIModelType = "messages"
)

type AIModel struct {
	ID           string
	Name         string
	EndpointType AIModelType
}

var Models = []AIModel{
	{ID: "glm-5.1", Name: "GLM-5.1", EndpointType: EndpointChat},
	{ID: "kimi-k2.6", Name: "Kimi K2.6", EndpointType: EndpointChat},
	{ID: "mimo-v2.5-pro", Name: "MiMo-V2.5-Pro", EndpointType: EndpointChat},
	{ID: "deepseek-v4-pro", Name: "DeepSeek V4 Pro", EndpointType: EndpointChat},
	{ID: "deepseek-v4-flash", Name: "DeepSeek V4 Flash", EndpointType: EndpointChat},
	{ID: "qwen3.6-plus", Name: "Qwen3.6 Plus", EndpointType: EndpointChat},
	{ID: "minimax-m2.7", Name: "MiniMax M2.7", EndpointType: EndpointMessages},
	{ID: "minimax-m2.5", Name: "MiniMax M2.5", EndpointType: EndpointMessages},
	{ID: "glm-5", Name: "GLM-5", EndpointType: EndpointChat},
	{ID: "kimi-k2.5", Name: "Kimi K2.5", EndpointType: EndpointChat},
}

var DefaultModelIDs = []string{
	"glm-5.1",
	"kimi-k2.6",
	"mimo-v2.5-pro",
	"deepseek-v4-pro",
	"qwen3.6-plus",
}

var ReplaceableModelIDs = []string{
	"qwen3.6-plus",
	"minimax-m2.7",
	"minimax-m2.5",
	"deepseek-v4-flash",
	"glm-5",
	"kimi-k2.5",
}

func GetModelName(modelID string) string {
	for _, m := range Models {
		if m.ID == modelID {
			return m.Name
		}
	}
	return modelID
}

func IsValidModel(modelID string) bool {
	for _, m := range Models {
		if m.ID == modelID {
			return true
		}
	}
	return false
}

func IsMiniMax(modelID string) bool {
	for _, m := range Models {
		if m.ID == modelID && m.EndpointType == EndpointMessages {
			return true
		}
	}
	return false
}

const DefaultSystemPrompt = "你是一个有帮助的助手。"

const DailyCallLimitPerModel = 10
