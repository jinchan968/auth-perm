package service

import (
	"context"
	"fmt"
	"time"

	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/auth/repo"
)

// DeviceService 设备管理服务接口
type DeviceService interface {
	// GetDevices 获取用户设备列表
	GetDevices(ctx context.Context, userID string, page, pageSize int) (*DevicesListDTO, error)
	// RevokeDevice 撤销指定设备的所有会话
	RevokeDevice(ctx context.Context, userID, deviceID string) (int, error)
	// TrustDevice 标记设备为信任
	TrustDevice(ctx context.Context, userID, deviceID, reason string) error
	// UnTrustDevice 取消设备信任
	UnTrustDevice(ctx context.Context, userID, deviceID, reason string) error
}

// DevicesListDTO 设备列表数据传输对象
type DevicesListDTO struct {
	Devices  []DeviceDTO `json:"devices"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// DeviceDTO 设备数据传输对象
type DeviceDTO struct {
	DeviceID   string `json:"device_id"`
	Platform   string `json:"platform"`
	Browser    string `json:"browser"`
	Device     string `json:"device"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	SessionID  string `json:"session_id"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
	LastActive string `json:"last_active"`
	ExpiresAt  string `json:"expires_at"`
	Trusted    bool   `json:"trusted"`
}

// deviceService 设备服务实现
type deviceService struct {
	sessionRepo     *repo.SessionRepo
	deviceTrustRepo repo.DeviceTrustRepo
}

// NewDeviceService 创建设备服务
func NewDeviceService(
	sessionRepo *repo.SessionRepo,
	deviceTrustRepo repo.DeviceTrustRepo,
) DeviceService {
	return &deviceService{
		sessionRepo:     sessionRepo,
		deviceTrustRepo: deviceTrustRepo,
	}
}

// GetDevices 获取用户设备列表
func (s *deviceService) GetDevices(ctx context.Context, userID string, page, pageSize int) (*DevicesListDTO, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 获取用户会话列表（设备信息从会话中获取）
	// 使用较大的分页大小来获取所有会话，然后本地去重
	pagination := &model.Pagination{
		Page:     1,
		PageSize: 1000,
		SortBy:   "last_active_at",
		SortDesc: true,
	}
	sessions, err := s.sessionRepo.FindByUserID(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("获取会话列表失败: %w", err)
	}

	// 将会话转换为设备信息（按设备指纹去重）
	seenDevices := make(map[string]bool)
	deviceMap := make(map[string]*DeviceDTO)

	// 第一次遍历：按设备指纹去重和聚合信息
	for _, session := range sessions {
		deviceInfo := session.GetDeviceInfo()
		if deviceInfo == nil || deviceInfo.Fingerprint == "" {
			continue
		}

		// 跳过已见过的设备
		if seenDevices[deviceInfo.Fingerprint] {
			continue
		}
		seenDevices[deviceInfo.Fingerprint] = true

		deviceMap[deviceInfo.Fingerprint] = &DeviceDTO{
			DeviceID:   deviceInfo.Fingerprint,
			Platform:   deviceInfo.Platform,
			Browser:    deviceInfo.Browser,
			Device:     deviceInfo.Device,
			IPAddress:  deviceInfo.IPAddress,
			UserAgent:  deviceInfo.UserAgent,
			SessionID:  session.ID,
			IsActive:   session.IsActive && session.ExpiresAt.After(time.Now()),
			CreatedAt:  session.CreatedAt.Format("2006-01-02T15:04:05Z"),
			LastActive: session.LastActiveAt.Format("2006-01-02T15:04:05Z"),
			ExpiresAt:  session.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	// 第二次遍历：获取设备信任状态
	devices := make([]DeviceDTO, 0, len(deviceMap))
	for fingerprint := range deviceMap {
		device := deviceMap[fingerprint]

		// 查询设备信任状态
		isTrusted, err := s.deviceTrustRepo.IsTrusted(ctx, userID, fingerprint)
		if err == nil {
			device.Trusted = isTrusted
		} else {
			// 如果查询失败，默认不信任
			device.Trusted = false
		}

		devices = append(devices, *device)
	}

	// 分页计算
	total := int64(len(devices))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= int(total) {
		start = 0
		end = 0
	} else if end > int(total) {
		end = int(total)
	}
	paginatedDevices := devices[start:end]

	return &DevicesListDTO{
		Devices:  paginatedDevices,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// RevokeDevice 撤销指定设备的所有会话
func (s *deviceService) RevokeDevice(ctx context.Context, userID, deviceID string) (int, error) {
	// 获取用户所有会话
	pagination := &model.Pagination{
		Page:     1,
		PageSize: 1000,
	}
	sessions, err := s.sessionRepo.FindByUserID(ctx, userID, pagination)
	if err != nil {
		return 0, fmt.Errorf("获取会话列表失败: %w", err)
	}

	// 查找并撤销该设备的会话
	revokedCount := 0
	for _, session := range sessions {
		deviceInfo := session.GetDeviceInfo()
		if deviceInfo != nil && deviceInfo.Fingerprint == deviceID {
			// 撤销这个会话
			err := s.sessionRepo.Delete(ctx, session.ID)
			if err == nil {
				revokedCount++
			}
		}
	}

	return revokedCount, nil
}

// TrustDevice 标记设备为信任
func (s *deviceService) TrustDevice(ctx context.Context, userID, deviceID, reason string) error {
	// 验证设备是否存在
	pagination := &model.Pagination{
		Page:     1,
		PageSize: 1000,
	}
	sessions, err := s.sessionRepo.FindByUserID(ctx, userID, pagination)
	if err != nil {
		return fmt.Errorf("获取会话列表失败: %w", err)
	}

	// 查找设备
	deviceFound := false
	for _, session := range sessions {
		deviceInfo := session.GetDeviceInfo()
		if deviceInfo != nil && deviceInfo.Fingerprint == deviceID {
			deviceFound = true
			break
		}
	}

	if !deviceFound {
		return fmt.Errorf("设备不存在: %s", deviceID)
	}

	// 获取设备信息用于保存
	var deviceInfoStr string
	for _, session := range sessions {
		deviceInfo := session.GetDeviceInfo()
		if deviceInfo != nil && deviceInfo.Fingerprint == deviceID {
			// 将设备信息序列化为JSON字符串
			deviceInfoStr = fmt.Sprintf("{\"platform\": \"%s\", \"browser\": \"%s\", \"device\": \"%s\", \"ip_address\": \"%s\"}",
				deviceInfo.Platform, deviceInfo.Browser, deviceInfo.Device, deviceInfo.IPAddress)
			break
		}
	}

	// 使用设备信任仓储保存信任状态
	return s.deviceTrustRepo.TrustDevice(ctx, userID, deviceID, deviceInfoStr, reason)
}

// UnTrustDevice 取消设备信任
func (s *deviceService) UnTrustDevice(ctx context.Context, userID, deviceID, reason string) error {
	// 验证设备是否存在
	pagination := &model.Pagination{
		Page:     1,
		PageSize: 1000,
	}
	sessions, err := s.sessionRepo.FindByUserID(ctx, userID, pagination)
	if err != nil {
		return fmt.Errorf("获取会话列表失败: %w", err)
	}

	// 查找设备
	deviceFound := false
	for _, session := range sessions {
		deviceInfo := session.GetDeviceInfo()
		if deviceInfo != nil && deviceInfo.Fingerprint == deviceID {
			deviceFound = true
			break
		}
	}

	if !deviceFound {
		return fmt.Errorf("设备不存在: %s", deviceID)
	}

	// 使用设备信任仓储取消信任状态
	return s.deviceTrustRepo.UnTrustDevice(ctx, userID, deviceID, reason)
}
