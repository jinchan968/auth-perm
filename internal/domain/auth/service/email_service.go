package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/auth/dto"
)

// EmailService 邮件服务
type EmailService struct {
	config dto.EmailConfig
}

// NewEmailService 创建邮件服务
func NewEmailService(config dto.EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// SendVerificationEmail 发送邮箱验证邮件
func (s *EmailService) SendVerificationEmail(ctx context.Context, email, username, token string) error {
	// 生成验证链接
	verificationURL := constant.FullURL(constant.VerificationEmailPath + "?token=" + token)

	// 构建邮件内容
	subject := "验证您的邮箱地址"
	body := fmt.Sprintf(`
		亲爱的 %s，

		感谢您注册我们的服务！

		请点击以下链接验证您的邮箱地址：
		%s

		此链接将在24小时后过期。

		如果您没有注册此账户，请忽略此邮件。

		谢谢！
	`, username, verificationURL)

	// 发送邮件（这里使用模拟实现）
	// 实际应用中应该使用SMTP服务或第三方邮件服务（如SendGrid、阿里云邮件推送等）
	return s.sendEmail(ctx, email, subject, body)
}

// sendEmail 发送邮件
func (s *EmailService) sendEmail(ctx context.Context, to, subject, body string) error {
	// 实现真实的SMTP发送
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// 设置认证信息
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = s.config.FromName + " <" + s.config.From + ">"
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=UTF-8"
	headers["Content-Transfer-Encoding"] = "8bit"

	// 构建邮件内容
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(k + ": " + v + "\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// 发送邮件
	err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg.String()))
	if err != nil {
		return errors.WrapBizError(err, "发送邮件失败")
	}

	return nil
}

// ValidateEmailFormat 验证邮箱格式
func (s *EmailService) ValidateEmailFormat(email string) error {
	if email == "" {
		return errors.NewValidationError("邮箱不能为空")
	}

	// 简单的邮箱格式验证
	if !contains(email, "@") {
		return errors.NewValidationError("邮箱格式不正确")
	}

	// 检查@符号位置
	atIndex := findLastIndex(email, "@")
	if atIndex == 0 || atIndex == len(email)-1 {
		return errors.NewValidationError("邮箱格式不正确")
	}

	// 检查域名部分
	domain := email[atIndex+1:]
	if len(domain) < 3 || !contains(domain, ".") {
		return errors.NewValidationError("邮箱格式不正确")
	}

	return nil
}

// GenerateVerificationToken 生成邮箱验证令牌
func (s *EmailService) GenerateVerificationToken(email, timestamp string) string {
	// 使用SHA256哈希生成令牌
	data := email + timestamp + s.config.Password
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VerifyToken 验证令牌
func (s *EmailService) VerifyToken(email, token, timestamp string) (bool, error) {
	// 检查时间戳是否过期（24小时）
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false, errors.NewValidationError("无效的时间戳")
	}

	tokenTime := time.Unix(timestampInt, 0)
	if time.Since(tokenTime) > constant.TokenExpiryDefault {
		return false, errors.NewBusinessError("验证链接已过期")
	}

	// 重新生成令牌并比较
	expectedToken := s.GenerateVerificationToken(email, timestamp)

	return token == expectedToken, nil
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring FUTURE: 子字符串查找 - 在实现字符串处理时使用
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// findLastIndex FUTURE: 最后索引查找 - 在实现字符串处理时使用
func findLastIndex(s, substr string) int {
	lastIndex := -1
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			lastIndex = i
		}
	}
	return lastIndex
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, username, token string) error {
	// 生成重置链接
	resetURL := constant.FullURL(constant.ResetPasswordPath + "?token=" + token)

	// 构建邮件内容
	subject := "重置您的密码"
	body := fmt.Sprintf(`
		亲爱的 %s，

		我们收到了重置您账户密码的请求。

		请点击以下链接重置您的密码：
		%s

		此链接将在1小时后过期。

		如果您没有请求重置密码，请忽略此邮件。

		谢谢！
	`, username, resetURL)

	// 发送邮件
	return s.sendEmail(ctx, email, subject, body)
}

// SendWelcomeEmail 发送欢迎邮件
func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, username string) error {
	// 构建邮件内容
	subject := "欢迎加入我们！"
	body := fmt.Sprintf(`
		亲爱的 %s，

		欢迎加入我们的平台！

		您的账户已经成功创建。

		请点击以下链接验证您的邮箱地址以获得完整体验：
		`+constant.FullURL(constant.SendVerificationEmailPath)+`

		谢谢！
	`, username)

	// 发送邮件
	return s.sendEmail(ctx, email, subject, body)
}
