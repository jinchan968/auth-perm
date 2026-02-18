package vo

import (
	"time"

	"auth-perm/internal/domain/auth/service"
)

// PaginatedResponse 分页响应基类
type PaginatedResponse struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// SessionsListResponse 会话列表响应
type SessionsListResponse struct {
	PaginatedResponse
	Sessions []SessionResponse `json:"sessions"`
}

// DevicesListResponse 设备列表响应
type DevicesListResponse struct {
	PaginatedResponse
	Devices []DeviceResponse `json:"devices"`
}

// FromDevicesListDTO 从DTO转换为设备列表响应
func (r *DevicesListResponse) FromDevicesListDTO(dto *service.DevicesListDTO) {
	devices := make([]DeviceResponse, 0, len(dto.Devices))
	for _, device := range dto.Devices {
		createdAt, _ := time.Parse("2006-01-02T15:04:05Z", device.CreatedAt)
		lastActive, _ := time.Parse("2006-01-02T15:04:05Z", device.LastActive)
		expiresAt, _ := time.Parse("2006-01-02T15:04:05Z", device.ExpiresAt)

		devices = append(devices, DeviceResponse{
			DeviceID:   device.DeviceID,
			Platform:   device.Platform,
			Browser:    device.Browser,
			Device:     device.Device,
			IPAddress:  device.IPAddress,
			UserAgent:  device.UserAgent,
			SessionID:  device.SessionID,
			IsActive:   device.IsActive,
			CreatedAt:  createdAt,
			LastActive: lastActive,
			ExpiresAt:  expiresAt,
			Trusted:    device.Trusted,
		})
	}

	r.PaginatedResponse = PaginatedResponse{
		Total:    dto.Total,
		Page:     dto.Page,
		PageSize: dto.PageSize,
	}
	r.Devices = devices
}

// SecurityLogsListResponse 安全日志列表响应
type SecurityLogsListResponse struct {
	PaginatedResponse
	Logs []SecurityLogResponse `json:"logs"`
}

// FromSecurityLogsListDTO 从DTO转换为安全日志列表响应
func (r *SecurityLogsListResponse) FromSecurityLogsListDTO(dto *service.SecurityLogsListDTO) {
	logs := make([]SecurityLogResponse, 0, len(dto.Logs))
	for _, log := range dto.Logs {
		logs = append(logs, SecurityLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			IPAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			Success:      log.Success,
			ErrorMessage: log.ErrorMessage,
			CreatedAt:    log.CreatedAt,
		})
	}

	r.PaginatedResponse = PaginatedResponse{
		Total:    dto.Total,
		Page:     dto.Page,
		PageSize: dto.PageSize,
	}
	r.Logs = logs
}
