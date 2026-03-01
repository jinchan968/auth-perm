package dm

import (
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// strPtr 将非空字符串转为指针，空字符串转为 nil（写入数据库时为 NULL）
func strPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// strVal 安全地从指针中取字符串值，nil 返回空字符串（内部使用）
func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StrVal 安全地从指针中取字符串值，nil 返回空字符串（导出供其他包使用）
func StrVal(s *string) string {
	return strVal(s)
}

// UserDO 用户领域对象
type UserDO struct {
	ID              string                  `gorm:"primaryKey;type:uuid"`
	Username        *string                 `gorm:"column:username;uniqueIndex:idx_users_username"`
	Nickname        *string                 `gorm:"column:nickname"`
	Avatar          *string                 `gorm:"column:avatar"`
	Phone           *string                 `gorm:"column:phone;uniqueIndex:idx_users_phone"`
	Email           *string                 `gorm:"column:email;uniqueIndex:idx_users_email"`
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
		Username:        strPtr(strings.ToLower(strings.TrimSpace(username))),
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
		Username:        strVal(u.Username),
		Nickname:        strVal(u.Nickname),
		Avatar:          strVal(u.Avatar),
		Phone:           strVal(u.Phone),
		Email:           strVal(u.Email),
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
// 空字符串字段会被转为 nil，GORM 写入数据库时存为 NULL，
// 从而避免 phone/email/username 的部分唯一索引冲突。
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
		Username:        strPtr(d.Username),
		Nickname:        strPtr(d.Nickname),
		Avatar:          strPtr(d.Avatar),
		Phone:           strPtr(d.Phone),
		Email:           strPtr(d.Email),
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
