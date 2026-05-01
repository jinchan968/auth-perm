package repo

import (
	"context"
	errStd "errors"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/auth/dm"
)

// TOTPSecretRepo TOTP密钥仓库接口
type TOTPSecretRepo interface {
	// FindByAccountID 根据账户ID查找TOTP密钥
	FindByAccountID(accountID string) (*dm.TOTPSecretDO, error)

	// Save 保存TOTP密钥
	Save(totpSecret *dm.TOTPSecretDO) error

	// Delete 删除TOTP密钥
	Delete(accountID string) error

	// FindByID 根据ID查找TOTP密钥
	FindByID(id string) (*dm.TOTPSecretDO, error)
}

// TOTPSecretRepository TOTP密钥仓储实现
type TOTPSecretRepository struct {
	db *gorm.DB
}

// NewTOTPSecretRepository 创建TOTP密钥仓储
func NewTOTPSecretRepository(db *gorm.DB) TOTPSecretRepo {
	return &TOTPSecretRepository{
		db: db,
	}
}

// FindByAccountID 根据账户ID查找TOTP密钥
func (r *TOTPSecretRepository) FindByAccountID(accountID string) (*dm.TOTPSecretDO, error) {
	ctx := context.Background()
	var totpSecret dm.TOTPSecretDO
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		First(&totpSecret).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WrapBizError(err, "查找TOTP密钥失败")
	}

	return &totpSecret, nil
}

// Save 保存TOTP密钥
func (r *TOTPSecretRepository) Save(totpSecret *dm.TOTPSecretDO) error {
	ctx := context.Background()
	if totpSecret.ID == "" {
		return errors.NewBusinessError("TOTP密钥ID不能为空")
	}

	// 检查是否已存在
	existing, err := r.FindByAccountID(totpSecret.AccountID)
	if err != nil {
		return err
	}

	if existing != nil {
		// 更新现有记录
		totpSecret.ID = existing.ID
		totpSecret.CreatedAt = existing.CreatedAt
		totpSecret.UpdatedAt = time.Now()
		err := r.db.WithContext(ctx).Save(totpSecret).Error
		if err != nil {
			return errors.WrapBizError(err, "更新TOTP密钥失败")
		}
	} else {
		// 创建新记录
		totpSecret.CreatedAt = time.Now()
		totpSecret.UpdatedAt = time.Now()
		err := r.db.WithContext(ctx).Create(totpSecret).Error
		if err != nil {
			return errors.WrapBizError(err, "创建TOTP密钥失败")
		}
	}

	return nil
}

// Delete 删除TOTP密钥
func (r *TOTPSecretRepository) Delete(accountID string) error {
	ctx := context.Background()
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Delete(&dm.TOTPSecretDO{}).Error

	if err != nil {
		return errors.WrapBizError(err, "删除TOTP密钥失败")
	}

	return nil
}

// FindByID 根据ID查找TOTP密钥
func (r *TOTPSecretRepository) FindByID(id string) (*dm.TOTPSecretDO, error) {
	ctx := context.Background()
	var totpSecret dm.TOTPSecretDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&totpSecret).Error

	if err != nil {
		if errStd.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WrapBizError(err, "查找TOTP密钥失败")
	}

	return &totpSecret, nil
}
