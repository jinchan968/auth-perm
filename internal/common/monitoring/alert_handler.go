package monitoring

import (
	"fmt"
	"log"
	"time"
)

// DefaultAlertHandler 默认告警处理器
type DefaultAlertHandler struct {
	webhookURL string
	email      string
}

// NewDefaultAlertHandler FUTURE: 默认告警处理器 - 在实现告警时使用
func NewDefaultAlertHandler(webhookURL, email string) *DefaultAlertHandler {
	return &DefaultAlertHandler{
		webhookURL: webhookURL,
		email:      email,
	}
}

// HandleAlert 处理告警
func (h *DefaultAlertHandler) HandleAlert(alert *SecurityAlert) error {
	// 根据严重级别处理告警
	switch alert.Severity {
	case SeverityCritical:
		return h.handleCriticalAlert(alert)
	case SeverityWarning:
		return h.handleWarningAlert(alert)
	case SeverityInfo:
		return h.handleInfoAlert(alert)
	default:
		return nil
	}
}

// handleCriticalAlert 处理严重告警
func (h *DefaultAlertHandler) handleCriticalAlert(alert *SecurityAlert) error {
	message := fmt.Sprintf("[CRITICAL] %s\n账户: %s\nIP: %s\n时间: %s\n元数据: %v",
		alert.Message, alert.AccountID, alert.IPAddress, alert.Timestamp.Format(time.RFC3339), alert.Metadata)

	log.Println(message)

	// 这里可以发送邮件、短信或调用webhook
	// 例如：
	// if h.webhookURL != "" {
	//     sendWebhook(h.webhookURL, alert)
	// }
	// if h.email != "" {
	//     sendEmail(h.email, "安全告警", message)
	// }

	return nil
}

// handleWarningAlert 处理警告告警
func (h *DefaultAlertHandler) handleWarningAlert(alert *SecurityAlert) error {
	message := fmt.Sprintf("[WARNING] %s\n账户: %s\nIP: %s\n时间: %s",
		alert.Message, alert.AccountID, alert.IPAddress, alert.Timestamp.Format(time.RFC3339))

	log.Println(message)

	return nil
}

// handleInfoAlert 处理信息告警
func (h *DefaultAlertHandler) handleInfoAlert(alert *SecurityAlert) error {
	message := fmt.Sprintf("[INFO] %s\n账户: %s\nIP: %s\n时间: %s",
		alert.Message, alert.AccountID, alert.IPAddress, alert.Timestamp.Format(time.RFC3339))

	log.Println(message)

	return nil
}
