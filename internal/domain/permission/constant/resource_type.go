package constant

// ResourceType 定义权限资源类型常量
const (
	ResourceTypeAPIPath = "api_path"
	ResourceTypeMenu    = "menu"
	ResourceTypeButton  = "button"
	ResourceTypeField   = "field"
	ResourceTypeOther   = "other"
)

// AllResourceTypes 所有资源类型
var AllResourceTypes = []string{
	ResourceTypeAPIPath,
	ResourceTypeMenu,
	ResourceTypeButton,
	ResourceTypeField,
	ResourceTypeOther,
}
