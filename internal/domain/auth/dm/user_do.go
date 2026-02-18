package dm

import (
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserDO 用户领域对象
type UserDO struct {
	ID              string                  `gorm:"primaryKey;type:uuid"`
	Username        string                  `gorm:"unique:not null"`
	Nickname        string                  `gorm:"column:nickname"`
	Avatar          string                  `gorm:"column:avatar"`
	Phone           string                  `gorm:"column:phone"`
	Email           string                  `gorm:"column:email"`
	PasswordHash    string                  `gorm:"column:password_hash"`
	IdentifierType  constant.IdentifierType `gorm:"column:identifier_type;not null;index"`
	IdentifierValue string                  `gorm:"column:identifier_value;not null;index"`
	Status          constant.UserStatus     `gorm:"column:status;not null;default:active"`
	Profile         dto.ProfileDTO          `gorm:"column:profile;type:jsonb;default:'{}'"`
	Preferences     dto.PreferencesDTO      `gorm:"column:preferences;type:jsonb;default:'{}'"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Accounts []AccountDO `gorm:"foreignKey:UserID"`
}

// NewUser 创建新用户（数据库操作）
func NewUser(username string, identifierType constant.IdentifierType, identifierValue string) *UserDO {
	now := time.Now()
	return &UserDO{
		ID:              uuid.New().String(),
		Username:        strings.ToLower(strings.TrimSpace(username)),
		IdentifierType:  identifierType,
		IdentifierValue: strings.TrimSpace(identifierValue),
		Status:          constant.UserStatusActive,
		Profile:         dto.ProfileDTO{},
		Preferences:     dto.PreferencesDTO{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// SetProfile 设置用户资料
func (u *UserDO) SetProfile(profile *dto.ProfileDTO) {
	u.Profile = *profile
}

// GetProfile 获取用户资料
func (u *UserDO) GetProfile() *dto.ProfileDTO {
	return &u.Profile
}

// SetPreferences 设置用户偏好
func (u *UserDO) SetPreferences(preferences *dto.PreferencesDTO) {
	u.Preferences = *preferences
}

// GetPreferences 获取用户偏好
func (u *UserDO) GetPreferences() *dto.PreferencesDTO {
	return &u.Preferences
}

// ToDTO 转换为DTO
func (u *UserDO) ToDTO() *dto.UserDTO {
	if u == nil {
		return nil
	}
	return &dto.UserDTO{
		ID:              u.ID,
		Username:        u.Username,
		Nickname:        u.Nickname,
		Avatar:          u.Avatar,
		Phone:           u.Phone,
		Email:           u.Email,
		PasswordHash:    u.PasswordHash,
		IdentifierType:  u.IdentifierType,
		IdentifierValue: u.IdentifierValue,
		Status:          u.Status,
		Profile:         u.GetProfile(),
		Preferences:     u.GetPreferences(),
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// UserFromDTO 从DTO创建UserDO
func UserFromDTO(d *dto.UserDTO) *UserDO {
	if d == nil {
		return nil
	}

	profile := dto.ProfileDTO{}
	if d.Profile != nil {
		profile = *d.Profile
	}

	preferences := dto.PreferencesDTO{}
	if d.Preferences != nil {
		preferences = *d.Preferences
	}

	return &UserDO{
		ID:              d.ID,
		Username:        d.Username,
		Nickname:        d.Nickname,
		Avatar:          d.Avatar,
		Phone:           d.Phone,
		Email:           d.Email,
		PasswordHash:    d.PasswordHash,
		IdentifierType:  d.IdentifierType,
		IdentifierValue: d.IdentifierValue,
		Status:          d.Status,
		Profile:         profile,
		Preferences:     preferences,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

// TableName 指定表名
func (u *UserDO) TableName() string {
	return "users"
}
