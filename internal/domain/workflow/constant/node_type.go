package constant

const (
	NodeTypeTrigger   = "trigger"
	NodeTypeLLM       = "llm"
	NodeTypeCondition = "condition"
	NodeTypeTransform = "transform"
	NodeTypeMerge     = "merge"
	NodeTypeOutput    = "output"
)

var validNodeTypes = map[string]bool{
	NodeTypeTrigger:   true,
	NodeTypeLLM:       true,
	NodeTypeCondition: true,
	NodeTypeTransform: true,
	NodeTypeMerge:     true,
	NodeTypeOutput:    true,
}

func IsValidNodeType(t string) bool {
	return validNodeTypes[t]
}
