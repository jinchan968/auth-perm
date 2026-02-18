package constant

// ==================== 类型别名定义 ====================

// UserStatus 用户状态类型
type UserStatus string

// IdentifierType 用户标识类型
type IdentifierType string

// AccountType 账户类型
type AccountType string

// AccountStatus 账户状态类型
type AccountStatus string

// SessionStatus 会话状态类型
type SessionStatus string

// ==================== 用户相关常量 ====================

// UserStatus 常量
const (
	UserStatusActive   UserStatus = "active"   // 活跃
	UserStatusInactive UserStatus = "inactive" // 非活跃
)

// IsActive 检查用户状态是否为活跃
func (s UserStatus) IsActive() bool {
	return s == UserStatusActive
}

// IsInactive 检查用户状态是否为非活跃
func (s UserStatus) IsInactive() bool {
	return s == UserStatusInactive
}

// String 转换为字符串
func (s UserStatus) String() string {
	return string(s)
}

// IdentifierType 常量
const (
	IdentifierTypePhone  IdentifierType = "phone"  // 手机号
	IdentifierTypeEmail  IdentifierType = "email"  // 邮箱
	IdentifierTypeGitHub IdentifierType = "github" // GitHub OAuth
	IdentifierTypeGoogle IdentifierType = "google" // Google OAuth
	IdentifierTypeWeChat IdentifierType = "wechat" // 微信OAuth
)

// IsPhone 检查是否为手机号标识
func (t IdentifierType) IsPhone() bool {
	return t == IdentifierTypePhone
}

// IsEmail 检查是否为邮箱标识
func (t IdentifierType) IsEmail() bool {
	return t == IdentifierTypeEmail
}

// IsOAuth 检查是否为OAuth标识
func (t IdentifierType) IsOAuth() bool {
	return t == IdentifierTypeGitHub || t == IdentifierTypeGoogle || t == IdentifierTypeWeChat
}

// IsGitHub 检查是否为GitHub OAuth
func (t IdentifierType) IsGitHub() bool {
	return t == IdentifierTypeGitHub
}

// IsGoogle 检查是否为Google OAuth
func (t IdentifierType) IsGoogle() bool {
	return t == IdentifierTypeGoogle
}

// IsWeChat 检查是否为微信OAuth
func (t IdentifierType) IsWeChat() bool {
	return t == IdentifierTypeWeChat
}

// IsValid 检查是否为有效的标识类型
func (t IdentifierType) IsValid() bool {
	return t == IdentifierTypePhone || t == IdentifierTypeEmail || t.IsOAuth()
}

// String 转换为字符串
func (t IdentifierType) String() string {
	return string(t)
}

// ToAccountType 转换为账户类型
func (t IdentifierType) ToAccountType() AccountType {
	switch t {
	case IdentifierTypeEmail:
		return AccountTypeEmail
	case IdentifierTypePhone:
		return AccountTypePhone
	case IdentifierTypeGitHub:
		return AccountTypeGitHub
	case IdentifierTypeGoogle:
		return AccountTypeGoogle
	case IdentifierTypeWeChat:
		return AccountTypeWeChat
	default:
		// 默认为邮箱类型
		return AccountTypeEmail
	}
}

// AccountType 常量
const (
	AccountTypeEmail  AccountType = "email"  // 邮箱账户
	AccountTypeGitHub AccountType = "github" // GitHub OAuth
	AccountTypeGoogle AccountType = "google" // Google OAuth
	AccountTypeWeChat AccountType = "wechat" // 微信OAuth
	AccountTypePhone  AccountType = "phone"  // 手机号账户
)

// IsEmail 检查是否为邮箱账户
func (t AccountType) IsEmail() bool {
	return t == AccountTypeEmail
}

// IsOAuth 检查是否为OAuth账户
func (t AccountType) IsOAuth() bool {
	return t == AccountTypeGitHub || t == AccountTypeGoogle || t == AccountTypeWeChat
}

// IsPhone 检查是否为手机号账户
func (t AccountType) IsPhone() bool {
	return t == AccountTypePhone
}

// IsSocial 检查是否为社交登录（OAuth）
func (t AccountType) IsSocial() bool {
	return t.IsOAuth()
}

// String 转换为字符串
func (t AccountType) String() string {
	return string(t)
}

// AccountStatus 常量
const (
	AccountStatusActive    AccountStatus = "active"    // 活跃
	AccountStatusInactive  AccountStatus = "inactive"  // 非活跃
	AccountStatusSuspended AccountStatus = "suspended" // 暂停
)

// IsActive 检查账户状态是否为活跃
func (s AccountStatus) IsActive() bool {
	return s == AccountStatusActive
}

// IsInactive 检查账户状态是否为非活跃
func (s AccountStatus) IsInactive() bool {
	return s == AccountStatusInactive
}

// IsSuspended 检查账户状态是否为暂停
func (s AccountStatus) IsSuspended() bool {
	return s == AccountStatusSuspended
}

// String 转换为字符串
func (s AccountStatus) String() string {
	return string(s)
}

// SessionStatus 常量
const (
	SessionStatusActive   SessionStatus = "active"   // 活跃
	SessionStatusInactive SessionStatus = "inactive" // 非活跃
)

// IsActive 检查会话状态是否为活跃
func (s SessionStatus) IsActive() bool {
	return s == SessionStatusActive
}

// IsInactive 检查会话状态是否为非活跃
func (s SessionStatus) IsInactive() bool {
	return s == SessionStatusInactive
}

// String 转换为字符串
func (s SessionStatus) String() string {
	return string(s)
}
