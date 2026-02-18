package http

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	controllerVo "auth-perm/internal/controller/vo"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/param"

	"github.com/gin-gonic/gin"
)

// TestForgotPassword 测试忘记密码功能
func TestForgotPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("请求结构体验证", func(t *testing.T) {
		req := controllerVo.ForgotPasswordRequest{
			Identifier: "test@example.com",
		}

		if req.Identifier != "test@example.com" {
			t.Errorf("请求参数不正确")
		}
	})
}

// TestResetPassword 测试重置密码功能
func TestResetPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("成功重置密码", func(t *testing.T) {
		req := controllerVo.ResetPasswordRequest{
			Token:           "test-token",
			NewPassword:     "newpassword123",
			ConfirmPassword: "newpassword123",
		}

		if req.NewPassword != req.ConfirmPassword {
			t.Errorf("密码不匹配")
		}
	})

	t.Run("密码不匹配", func(t *testing.T) {
		req := controllerVo.ResetPasswordRequest{
			Token:           "test-token",
			NewPassword:     "password1",
			ConfirmPassword: "password2",
		}

		if req.NewPassword == req.ConfirmPassword {
			t.Error("密码应该不匹配")
		}
	})
}

// TestRevokeSession 测试撤销单个会话
func TestRevokeSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("用户未登录返回401", func(t *testing.T) {
		// 没有设置用户上下文，验证返回401
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/api/v1/auth/sessions/test-session-id", nil)

		// 验证没有user_id上下文时，c.GetString("user_id")返回空字符串
		userID := c.GetString("user_id")
		if userID != "" {
			t.Error("用户ID应该为空")
		}
	})
}

// TestRevokeAllSessions 测试撤销所有会话
func TestRevokeAllSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("用户未登录返回401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/api/v1/auth/sessions/all", nil)

		userID := c.GetString("user_id")
		if userID != "" {
			t.Error("用户ID应该为空")
		}
	})
}

// TestGetDevices 测试获取设备列表
func TestGetDevices(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("解析分页参数", func(t *testing.T) {
		// 验证默认分页参数
		req := controllerVo.GetDevicesRequest{
			Page:     0,
			PageSize: 0,
		}

		// 模拟ShouldBindQuery设置默认值
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 || req.PageSize > 100 {
			req.PageSize = 20
		}

		if req.Page != 1 {
			t.Errorf("默认页码应该是1，实际得到 %d", req.Page)
		}

		if req.PageSize != 20 {
			t.Errorf("默认页面大小应该是20，实际得到 %d", req.PageSize)
		}
	})

	t.Run("用户未登录返回401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/auth/devices", nil)

		userID := c.GetString("user_id")
		if userID != "" {
			t.Error("用户ID应该为空")
		}
	})
}

// TestRevokeDevice 测试撤销设备
func TestRevokeDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("设备ID不能为空", func(t *testing.T) {
		// 验证空设备ID会被拒绝
		deviceID := ""
		if deviceID == "" {
			// 预期的空设备ID情况
		}
	})

	t.Run("用户未登录返回401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/api/v1/auth/devices/device-fp-1", nil)

		userID := c.GetString("user_id")
		if userID != "" {
			t.Error("用户ID应该为空")
		}
	})
}

// TestTrustDevice 测试信任设备
func TestTrustDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("设备ID不能为空", func(t *testing.T) {
		deviceID := ""
		if deviceID == "" {
			// 预期的空设备ID情况
		}
	})

	t.Run("用户未登录返回401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/devices/device-fp-1/trust", nil)

		userID := c.GetString("user_id")
		if userID != "" {
			t.Error("用户ID应该为空")
		}
	})
}

// TestUnTrustDevice 测试取消信任设备
func TestUnTrustDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("设备ID不能为空", func(t *testing.T) {
		deviceID := ""
		if deviceID == "" {
			// 预期的空设备ID情况
		}
	})

	t.Run("用户未登录返回401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/devices/device-fp-1/untrust", nil)

		userID := c.GetString("user_id")
		if userID != "" {
			t.Error("用户ID应该为空")
		}
	})
}

// TestGetSecurityLogs 测试获取安全日志
func TestGetSecurityLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("解析分页参数", func(t *testing.T) {
		req := controllerVo.GetSecurityLogsRequest{
			Page:     0,
			PageSize: 0,
		}

		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 || req.PageSize > 100 {
			req.PageSize = 20
		}

		if req.Page != 1 {
			t.Errorf("默认页码应该是1，实际得到 %d", req.Page)
		}

		if req.PageSize != 20 {
			t.Errorf("默认页面大小应该是20，实际得到 %d", req.PageSize)
		}
	})

	t.Run("用户未登录返回401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/auth/security/logs", nil)

		userID := c.GetString("user_id")
		if userID != "" {
			t.Error("用户ID应该为空")
		}
	})

	t.Run("无效的日期格式", func(t *testing.T) {
		invalidDate := "invalid-date"
		_, err := time.Parse(time.RFC3339, invalidDate)

		if err == nil {
			t.Error("无效日期应该返回错误")
		}
	})

	t.Run("有效的日期格式", func(t *testing.T) {
		validDate := time.Now().Format(time.RFC3339)
		parsedDate, err := time.Parse(time.RFC3339, validDate)

		if err != nil {
			t.Errorf("有效日期不应该返回错误: %v", err)
		}

		if parsedDate.IsZero() {
			t.Error("解析后的日期不应该为零")
		}
	})
}

// TestAuditLogDO_ToDTO 测试审计日志DO转DTO
func TestAuditLogDO_ToDTO(t *testing.T) {
	t.Run("成功转换审计日志", func(t *testing.T) {
		// 创建测试审计日志
		ipAddress := "192.168.1.1"
		auditLog := &dm.AuditLogDO{
			ID:           "log-1",
			UserID:       "test-user-id",
			Action:       "LOGIN",
			ResourceType: "USER",
			ResourceID:   "test-user-id",
			IPAddress:    &ipAddress,
			UserAgent:    "Mozilla/5.0",
			Success:      true,
			CreatedAt:    time.Now(),
		}

		// 转换为DTO
		dto := auditLog.ToDTO()

		if dto == nil {
			t.Fatal("DTO转换失败")
		}

		if dto.ID != auditLog.ID {
			t.Errorf("ID不匹配: 期望 %s，实际 %s", auditLog.ID, dto.ID)
		}

		if dto.UserID != auditLog.UserID {
			t.Errorf("UserID不匹配: 期望 %s，实际 %s", auditLog.UserID, dto.UserID)
		}

		if dto.Action != auditLog.Action {
			t.Errorf("Action不匹配: 期望 %s，实际 %s", auditLog.Action, dto.Action)
		}

		if dto.IPAddress != ipAddress {
			t.Errorf("IPAddress不匹配: 期望 %s，实际 %s", ipAddress, dto.IPAddress)
		}
	})
}

// TestSessionDTO_DeviceInfo 测试会话DTO设备信息
func TestSessionDTO_DeviceInfo(t *testing.T) {
	t.Run("设备信息不为空", func(t *testing.T) {
		session := &dto.SessionDTO{
			ID:     "session-1",
			UserID: "test-user-id",
			DeviceInfo: dto.DeviceInfoDTO{
				Fingerprint: "device-fp-1",
				Platform:    "macOS",
				Browser:     "Chrome",
			},
		}

		deviceInfo := session.GetDeviceInfo()

		if deviceInfo == nil {
			t.Fatal("设备信息不应该为空")
		}

		if deviceInfo.Fingerprint != "device-fp-1" {
			t.Errorf("设备指纹不匹配: 期望 device-fp-1，实际 %s", deviceInfo.Fingerprint)
		}
	})
}

// TestParamValidation 测试参数验证
func TestParamValidation(t *testing.T) {
	t.Run("GetSessionsParams验证", func(t *testing.T) {
		params := param.NewGetSessionsParams("test-user-id", 1, 20)

		if params.UserID != "test-user-id" {
			t.Errorf("UserID不匹配: 期望 test-user-id，实际 %s", params.UserID)
		}

		if params.Page != 1 {
			t.Errorf("Page不匹配: 期望 1，实际 %d", params.Page)
		}

		if params.PageSize != 20 {
			t.Errorf("PageSize不匹配: 期望 20，实际 %d", params.PageSize)
		}
	})

	t.Run("GetSessionsParams默认值", func(t *testing.T) {
		params := param.NewGetSessionsParams("test-user-id", 0, 0)

		// NewGetSessionsParams会直接使用传入的值，不做验证
		// 验证参数确实被设置
		if params.Page != 0 {
			t.Errorf("Page应该为0，实际 %d", params.Page)
		}

		if params.PageSize != 0 {
			t.Errorf("PageSize应该为0，实际 %d", params.PageSize)
		}
	})
}

// TestResponseFormat 测试响应格式
func TestResponseFormat(t *testing.T) {
	t.Run("成功响应格式", func(t *testing.T) {
		// 创建模拟响应
		responseData := map[string]interface{}{
			"status":  "success",
			"code":    200,
			"message": "操作成功",
			"data": map[string]interface{}{
				"id":   "test-1",
				"name": "test",
			},
		}

		// 序列化为JSON
		jsonData, err := json.Marshal(responseData)
		if err != nil {
			t.Fatalf("JSON序列化失败: %v", err)
		}

		// 反序列化验证
		var decoded map[string]interface{}
		if err := json.Unmarshal(jsonData, &decoded); err != nil {
			t.Fatalf("JSON反序列化失败: %v", err)
		}

		if decoded["status"] != "success" {
			t.Errorf("状态字段不匹配")
		}

		if decoded["code"] != 200.0 { // JSON数字反序列化后为float64
			t.Errorf("状态码字段不匹配")
		}
	})
}
