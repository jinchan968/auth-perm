package repo

import (
	"context"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/auth/dm"
)

// DeviceTrustRepository 设备信任仓储实现
type DeviceTrustRepository struct {
	db *gorm.DB
}

// NewDeviceTrustRepo 创建设备信任仓储
func NewDeviceTrustRepo(db *gorm.DB) *DeviceTrustRepository {
	return &DeviceTrustRepository{
		db: db,
	}
}

// Save 保存设备信任记录
func (r *DeviceTrustRepository) Save(ctx context.Context, deviceTrust *dm.DeviceTrustDO) error {
	return r.db.WithContext(ctx).Create(deviceTrust).Error
}

// Update 更新设备信任记录
func (r *DeviceTrustRepository) Update(ctx context.Context, deviceTrust *dm.DeviceTrustDO) error {
	return r.db.WithContext(ctx).Save(deviceTrust).Error
}

// Delete 删除设备信任记录
func (r *DeviceTrustRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&dm.DeviceTrustDO{}, "id = ?", id).Error
}

// FindByID 根据ID查找
func (r *DeviceTrustRepository) FindByID(ctx context.Context, id string) (*dm.DeviceTrustDO, error) {
	var deviceTrust dm.DeviceTrustDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&deviceTrust).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("设备信任记录不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找设备信任记录失败")
	}

	return &deviceTrust, nil
}

// FindByUserIDAndDeviceID 根据用户ID和设备ID查找
func (r *DeviceTrustRepository) FindByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*dm.DeviceTrustDO, error) {
	var deviceTrust dm.DeviceTrustDO
	err := r.db.WithContext(ctx).Where("user_id = ? AND device_id = ?", userID, deviceID).First(&deviceTrust).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("设备信任记录不存在: user=%s, device=%s", userID, deviceID)
		}
		return nil, errors.WrapBizError(err, "查找设备信任记录失败")
	}

	return &deviceTrust, nil
}

// FindByUserID 根据用户ID查找所有设备信任记录
func (r *DeviceTrustRepository) FindByUserID(ctx context.Context, userID string) ([]*dm.DeviceTrustDO, error) {
	var deviceTrusts []*dm.DeviceTrustDO
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&deviceTrusts).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "查找用户设备信任记录失败")
	}

	return deviceTrusts, nil
}

// IsTrusted 检查设备是否被信任
func (r *DeviceTrustRepository) IsTrusted(ctx context.Context, userID, deviceID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.DeviceTrustDO{}).Where("user_id = ? AND device_id = ? AND trusted = ?", userID, deviceID, true).Count(&count).Error

	if err != nil {
		return false, errors.WrapBizError(err, "检查设备信任状态失败")
	}

	return count > 0, nil
}

// TrustDevice 信任设备
func (r *DeviceTrustRepository) TrustDevice(ctx context.Context, userID, deviceID, deviceInfo, reason string) error {
	// 查找是否已存在记录
	existing, err := r.FindByUserIDAndDeviceID(ctx, userID, deviceID)

	// 如果记录不存在，创建新记录
	if err != nil {
		deviceTrust := dm.NewDeviceTrust("", userID, deviceID, deviceInfo)
		deviceTrust.SetTrustReason(reason)
		return r.Save(ctx, deviceTrust)
	}

	// 记录存在，更新为信任状态
	existing.SetTrustReason(reason)
	existing.SetTrusted(true)
	return r.Update(ctx, existing)
}

// UnTrustDevice 取消信任设备
func (r *DeviceTrustRepository) UnTrustDevice(ctx context.Context, userID, deviceID, reason string) error {
	// 查找记录
	existing, err := r.FindByUserIDAndDeviceID(ctx, userID, deviceID)

	// 如果记录不存在，创建一个不信任的记录
	if err != nil {
		deviceTrust := dm.NewDeviceTrust("", userID, deviceID, "{}")
		deviceTrust.SetTrustReason(reason)
		deviceTrust.SetTrusted(false)
		return r.Save(ctx, deviceTrust)
	}

	// 更新为不信任状态
	existing.SetTrustReason(reason)
	existing.SetTrusted(false)
	return r.Update(ctx, existing)
}
