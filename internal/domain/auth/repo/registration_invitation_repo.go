package repo

import (
	"context"
	stdErr "errors"
	"time"

	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegistrationInvitationRepo 注册邀请码仓储
type RegistrationInvitationRepo struct {
	db *gorm.DB
}

func NewRegistrationInvitationRepo(db *gorm.DB) *RegistrationInvitationRepo {
	return &RegistrationInvitationRepo{db: db}
}

func (r *RegistrationInvitationRepo) Create(ctx context.Context, invitation *dm.RegistrationInvitationDO) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

func (r *RegistrationInvitationRepo) List(ctx context.Context, tenantID, status string, page, pageSize int) ([]*dm.RegistrationInvitationDO, int64, error) {
	var items []*dm.RegistrationInvitationDO
	var total int64

	query := r.db.WithContext(ctx).Model(&dm.RegistrationInvitationDO{})
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if status != "" {
		switch status {
		case authConstant.InvitationStatusExpired:
			query = query.Where("status = ? AND expires_at <= ?", authConstant.InvitationStatusActive, time.Now())
		case authConstant.InvitationStatusActive:
			query = query.Where("status = ? AND expires_at > ?", status, time.Now())
		default:
			query = query.Where("status = ?", status)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计邀请码失败")
	}

	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询邀请码失败")
	}
	return items, total, nil
}

func (r *RegistrationInvitationRepo) FindByID(ctx context.Context, id string) (*dm.RegistrationInvitationDO, error) {
	var invitation dm.RegistrationInvitationDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&invitation).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundError("邀请码不存在")
		}
		return nil, errors.WrapBizError(err, "查询邀请码失败")
	}
	return &invitation, nil
}

func (r *RegistrationInvitationRepo) FindByCodeHashForUpdate(ctx context.Context, tx *gorm.DB, codeHash string) (*dm.RegistrationInvitationDO, error) {
	var invitation dm.RegistrationInvitationDO
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("code_hash = ?", codeHash).
		First(&invitation).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundError("邀请码不存在")
		}
		return nil, errors.WrapBizError(err, "查询邀请码失败")
	}
	return &invitation, nil
}

func (r *RegistrationInvitationRepo) MarkUsedWithTx(ctx context.Context, tx *gorm.DB, id, accountID string) error {
	now := time.Now()
	result := tx.WithContext(ctx).Model(&dm.RegistrationInvitationDO{}).
		Where("id = ? AND status = ?", id, authConstant.InvitationStatusActive).
		Updates(map[string]interface{}{
			"status":             authConstant.InvitationStatusUsed,
			"used_at":            now,
			"used_by_account_id": accountID,
			"updated_at":         now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewBusinessError("邀请码不可用")
	}
	return nil
}

func (r *RegistrationInvitationRepo) Invalidate(ctx context.Context, id, accountID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&dm.RegistrationInvitationDO{}).
		Where("id = ? AND status = ?", id, authConstant.InvitationStatusActive).
		Updates(map[string]interface{}{
			"status":                    authConstant.InvitationStatusInvalidated,
			"invalidated_at":            now,
			"invalidated_by_account_id": accountID,
			"updated_at":                now,
		})
	if result.Error != nil {
		return errors.WrapBizError(result.Error, "失效邀请码失败")
	}
	if result.RowsAffected == 0 {
		return errors.NewBusinessError("邀请码不可失效")
	}
	return nil
}

func (r *RegistrationInvitationRepo) GetDB() *gorm.DB {
	return r.db
}
