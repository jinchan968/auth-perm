package repo

import (
	"context"

	"auth-perm/internal/domain/auth/dm"
)

// DeviceTrustRepo 设备信任仓储接口
type DeviceTrustRepo interface {
	// Save 保存设备信任记录
	Save(ctx context.Context, deviceTrust *dm.DeviceTrustDO) error

	// Update 更新设备信任记录
	Update(ctx context.Context, deviceTrust *dm.DeviceTrustDO) error

	// Delete 删除设备信任记录
	Delete(ctx context.Context, id string) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id string) (*dm.DeviceTrustDO, error)

	// FindByUserIDAndDeviceID 根据用户ID和设备ID查找
	FindByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*dm.DeviceTrustDO, error)

	// FindByUserID 根据用户ID查找所有设备信任记录
	FindByUserID(ctx context.Context, userID string) ([]*dm.DeviceTrustDO, error)

	// IsTrusted 检查设备是否被信任
	IsTrusted(ctx context.Context, userID, deviceID string) (bool, error)

	// TrustDevice 信任设备
	TrustDevice(ctx context.Context, userID, deviceID, deviceInfo, reason string) error

	// UnTrustDevice 取消信任设备
	UnTrustDevice(ctx context.Context, userID, deviceID, reason string) error
}
