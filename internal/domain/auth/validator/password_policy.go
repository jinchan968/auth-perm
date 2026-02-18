package validator

import (
	"fmt"
	"regexp"
	"strings"

	"auth-perm/internal/common/errors"
)

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength      int      `json:"min_length"`
	RequireUpper   bool     `json:"require_upper"`
	RequireLower   bool     `json:"require_lower"`
	RequireDigit   bool     `json:"require_digit"`
	RequireSpecial bool     `json:"require_special"`
	MaxLength      int      `json:"max_length"`
	DisallowWords  []string `json:"disallow_words"`
}

// DefaultPasswordPolicy 获取默认密码策略
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:      8,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		DisallowWords:  []string{"password", "123456", "admin", "root"},
	}
}

// NewPasswordPolicy 创建密码策略
func NewPasswordPolicy(minLength int, requireUpper, requireLower, requireDigit, requireSpecial bool, maxLength int, disallowWords []string) *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:      minLength,
		RequireUpper:   requireUpper,
		RequireLower:   requireLower,
		RequireDigit:   requireDigit,
		RequireSpecial: requireSpecial,
		MaxLength:      maxLength,
		DisallowWords:  disallowWords,
	}
}

// ValidatePassword 验证密码强度
func (p *PasswordPolicy) ValidatePassword(password, username, email string) error {
	if len(password) < p.MinLength {
		return errors.NewValidationError(fmt.Sprintf("密码长度不能少于%d位", p.MinLength))
	}

	if len(password) > p.MaxLength {
		return errors.NewValidationError(fmt.Sprintf("密码长度不能超过%d位", p.MaxLength))
	}

	// 检查是否包含大写字母
	if p.RequireUpper && !containsUpper(password) {
		return errors.NewValidationError("密码必须包含大写字母")
	}

	// 检查是否包含小写字母
	if p.RequireLower && !containsLower(password) {
		return errors.NewValidationError("密码必须包含小写字母")
	}

	// 检查是否包含数字
	if p.RequireDigit && !containsDigit(password) {
		return errors.NewValidationError("密码必须包含数字")
	}

	// 检查是否包含特殊字符
	if p.RequireSpecial && !containsSpecial(password) {
		return errors.NewValidationError("密码必须包含特殊字符")
	}

	// 检查是否包含禁用词
	for _, word := range p.DisallowWords {
		if strings.Contains(strings.ToLower(password), strings.ToLower(word)) {
			return errors.NewValidationError(fmt.Sprintf("密码不能包含禁用词：%s", word))
		}
	}

	// 检查是否包含用户名
	if username != "" && strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
		return errors.NewValidationError("密码不能包含用户名")
	}

	// 检查是否包含邮箱用户名部分
	if email != "" {
		emailParts := strings.Split(email, "@")
		if len(emailParts) > 0 && strings.Contains(strings.ToLower(password), strings.ToLower(emailParts[0])) {
			return errors.NewValidationError("密码不能包含邮箱用户名")
		}
	}

	return nil
}

// CheckPasswordStrength 检查密码强度并返回评分
func (p *PasswordPolicy) CheckPasswordStrength(password, username, email string) (int, []string) {
	score := 0
	var warnings []string

	// 长度评分
	if len(password) >= p.MinLength {
		score += 20
	} else {
		warnings = append(warnings, fmt.Sprintf("密码长度应至少%d位", p.MinLength))
	}

	if len(password) >= 12 {
		score += 10
	}

	// 字符类型评分
	if containsUpper(password) {
		score += 10
	} else {
		warnings = append(warnings, "密码应包含大写字母")
	}

	if containsLower(password) {
		score += 10
	} else {
		warnings = append(warnings, "密码应包含小写字母")
	}

	if containsDigit(password) {
		score += 10
	} else {
		warnings = append(warnings, "密码应包含数字")
	}

	if containsSpecial(password) {
		score += 10
	} else {
		warnings = append(warnings, "密码应包含特殊字符")
	}

	// 长度额外加分
	if len(password) >= 16 {
		score += 10
	}

	// 字符多样性加分
	uniqueChars := uniqueCharCount(password)
	if uniqueChars >= 10 {
		score += 10
	} else if uniqueChars >= 6 {
		score += 5
	}

	// 常见密码检查
	commonPasswords := []string{"123456", "password", "123456789", "12345678", "12345"}
	for _, common := range commonPasswords {
		if password == common {
			score = 0
			warnings = append(warnings, "不能使用常见密码")
			break
		}
	}

	return score, warnings
}

// containsUpper 检查是否包含大写字母
func containsUpper(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

// containsLower 检查是否包含小写字母
func containsLower(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return true
		}
	}
	return false
}

// containsDigit 检查是否包含数字
func containsDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// containsSpecial 检查是否包含特殊字符
func containsSpecial(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?/~`"
	for _, c := range s {
		if strings.ContainsRune(specialChars, c) {
			return true
		}
	}
	return false
}

// uniqueCharCount FUTURE: 唯一字符计数 - 在实现密码强度检查时使用
func uniqueCharCount(s string) int {
	charSet := make(map[rune]bool)
	for _, c := range s {
		charSet[c] = true
	}
	return len(charSet)
}

// IsProxyIP FUTURE: 代理IP检查 - 在实现IP检查时使用
func IsProxyIP(ip string) bool {
	// 这里可以添加更复杂的代理IP检测逻辑
	// 例如查询已知的代理IP段
	proxyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^10\.`),          // 私有网络
		regexp.MustCompile(`^192\.168\.`),    // 私有网络
		regexp.MustCompile(`^172\.1[6-9]\.`), // 私有网络
		regexp.MustCompile(`^172\.2[0-9]\.`), // 私有网络
		regexp.MustCompile(`^172\.3[0-1]\.`), // 私有网络
	}

	for _, pattern := range proxyPatterns {
		if pattern.MatchString(ip) {
			return true
		}
	}

	return false
}
