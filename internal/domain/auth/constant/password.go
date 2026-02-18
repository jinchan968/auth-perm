package constant

// 密码相关常量
const (
	// bcrypt成本因子 - 统一使用固定值避免版本差异
	PasswordCost = 12

	// 密码最小长度
	PasswordMinLength = 6

	// 密码最大长度
	PasswordMaxLength = 128
)
