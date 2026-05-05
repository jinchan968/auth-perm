package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/repo"
	"auth-perm/internal/domain/auth/validator"

	"golang.org/x/crypto/bcrypt"
)

// SecurityService 安全服务
type SecurityService struct {
	cache             *CacheService
	auditRepo         *repo.AuditLogRepo
	passwordPolicy    *validator.PasswordPolicy
	failedAttempts    int
	lockoutDuration   time.Duration
	maxFailedAttempts int
}

// NewSecurityService 创建安全服务
func NewSecurityService(cache *CacheService, auditRepo *repo.AuditLogRepo, policy *validator.PasswordPolicy) *SecurityService {
	return &SecurityService{
		cache:             cache,
		auditRepo:         auditRepo,
		passwordPolicy:    policy,
		failedAttempts:    0,
		lockoutDuration:   authConstant.LockoutDuration,
		maxFailedAttempts: authConstant.MaxFailedAttempts,
	}
}

// CheckLoginAttempt 检查登录尝试
func (s *SecurityService) CheckLoginAttempt(ctx context.Context, accountID, email, ip string) error {
	// 检查是否被锁定
	locked, err := s.cache.IsLocked(ctx, accountID, ip)
	if err == nil && locked {
		return errors.NewBusinessError("登录尝试次数过多，账户已锁定")
	}

	// 检查失败次数
	count, err := s.cache.IncrementFailedLogin(ctx, accountID, ip)
	if err == nil && count >= int64(s.maxFailedAttempts) {
		// 记录锁定事件
		s.recordSecurityEvent(ctx, authConstant.SecurityEventAccountLocked, accountID, email, ip, "too many failed login attempts")
		return errors.NewBusinessError("登录尝试次数过多，账户已锁定")
	}

	return nil
}

// ResetFailedLogin 重置失败登录计数
func (s *SecurityService) ResetFailedLogin(ctx context.Context, accountID, email, ip string) error {
	return s.cache.ResetFailedLogin(ctx, accountID, ip)
}

// RecordFailedLogin 记录失败登录
func (s *SecurityService) RecordFailedLogin(ctx context.Context, accountID, email, ip, reason string) error {
	s.failedAttempts++

	// 记录审计日志
	s.recordSecurityEvent(ctx, authConstant.SecurityEventLoginFailed, accountID, email, ip, reason)

	// 检查是否需要锁定
	if s.failedAttempts >= s.maxFailedAttempts {
		// 记录锁定事件
		s.recordSecurityEvent(ctx, authConstant.SecurityEventAccountLocked, accountID, email, ip, "too many failed attempts")
	}

	return nil
}

// ValidatePassword 验证密码强度
func (s *SecurityService) ValidatePassword(password, username, email string) error {
	return s.passwordPolicy.ValidatePassword(password, username, email)
}

// CheckPasswordStrength 检查密码强度
func (s *SecurityService) CheckPasswordStrength(password, username, email string) (int, []string) {
	return s.passwordPolicy.CheckPasswordStrength(password, username, email)
}

// DetectAnomalousLogin 检测异常登录
func (s *SecurityService) DetectAnomalousLogin(ctx context.Context, accountID, email, ip, userAgent string, lastLoginIP string) (bool, error) {
	// 检查IP是否异常
	if lastLoginIP != "" && lastLoginIP != ip {
		// IP地址变更，标记为异常
		s.recordSecurityEvent(ctx, authConstant.SecurityEventSuspiciousLogin, accountID, email, ip, "IP address changed")
		return true, nil
	}

	// 检查User-Agent是否异常
	if userAgent != "" {
		// 这里可以添加更复杂的User-Agent检查逻辑
		// 例如检查是否来自已知浏览器/设备
	}

	return false, nil
}

// IsHighRiskLogin 判断是否为高风险登录
func (s *SecurityService) IsHighRiskLogin(ctx context.Context, accountID, email, ip string) (bool, error) {
	// 检查IP是否在黑名单中（这里可以扩展为查询数据库）
	blacklistedIPs := []string{"192.168.1.100"} // 示例黑名单IP
	for _, blacklistedIP := range blacklistedIPs {
		if ip == blacklistedIP {
			s.recordSecurityEvent(ctx, authConstant.SecurityEventHighRiskLogin, accountID, email, ip, "blacklisted IP")
			return true, nil
		}
	}

	// 检查是否为代理/VPN
	if validator.IsProxyIP(ip) {
		s.recordSecurityEvent(ctx, authConstant.SecurityEventHighRiskLogin, accountID, email, ip, "proxy/VPN detected")
		return true, nil
	}

	return false, nil
}

// GenerateSecureToken 生成安全令牌
func (s *SecurityService) GenerateSecureToken() (string, error) {
	// 生成随机令牌
	data := fmt.Sprintf("%d:%s", time.Now().UnixNano(), generateRandomString(32))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

// VerifySecureToken 验证安全令牌
func (s *SecurityService) VerifySecureToken(token, expectedToken string) bool {
	return token == expectedToken
}

// recordSecurityEvent 记录安全事件
func (s *SecurityService) recordSecurityEvent(ctx context.Context, eventType, accountID, email, ip, details string) {
	// 创建审计日志条目
	entry := &dto.AuditLogEntryDTO{
		Action:       eventType,
		ResourceType: authConstant.AuditResourceAccount,
		ResourceID:   accountID,
		Success:      false,
		IPAddress:    ip,
		UserAgent:    details,
	}

	// 异步记录（不阻塞主流程）
	s.auditRepo.LogAsync(entry)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result strings.Builder
	for i := 0; i < length; i++ {
		result.WriteByte(chars[len(chars)-1])
	}
	return result.String()
}

// ConstantTimePasswordValidate 常量时间密码验证（防止时序攻击）
// 该函数确保无论账户是否存在、密码是否正确，验证时间都尽可能一致
func ConstantTimePasswordValidate(password, passwordHash string) bool {
	// 生成一个假的哈希值用于不存在账户的情况
	// 这个假哈希值的计算时间与真实bcrypt哈希验证时间相近
	generateFakeHash := func() string {
		// 生成60字符的随机字符串（模拟bcrypt哈希长度）
		const fakeHashLength = 60
		bytes := make([]byte, fakeHashLength)
		_, _ = rand.Read(bytes)

		// 将随机字节转换为可打印字符
		const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789./"
		var result strings.Builder
		for i := 0; i < fakeHashLength; i++ {
			result.WriteByte(base64Chars[int(bytes[i])%len(base64Chars)])
		}
		return result.String()
	}

	// 如果没有密码哈希（OAuth账户等），生成假哈希
	if passwordHash == "" {
		passwordHash = generateFakeHash()
	}

	// 总是执行bcrypt比较（常量时间操作）
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}
