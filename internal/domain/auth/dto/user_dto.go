package dto

import (
	commonConstant "auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/auth/constant"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserDTO 用户数据传输对象（包含业务逻辑）
type UserDTO struct {
	// 基本信息
	ID           string `json:"id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`

	// 标识信息
	IdentifierType  constant.IdentifierType `json:"identifier_type"`
	IdentifierValue string                  `json:"identifier_value"`

	// 状态信息
	Status constant.UserStatus `json:"status"`

	// 扩展信息
	Profile     *ProfileDTO     `json:"profile"`
	Preferences *PreferencesDTO `json:"preferences"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *UserDTO) WithNewID(id string) *UserDTO {
	u.ID = id
	return u
}

// NewUserDTO 创建用户DTO
func NewUserDTO(username string, identifierType constant.IdentifierType, identifierValue string) (*UserDTO, error) {
	// 验证用户名
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	// 清理标识符值
	identifierValue = strings.TrimSpace(identifierValue)
	if identifierValue == "" {
		return nil, errors.NewValidationError("标识符值不能为空")
	}

	now := time.Now()
	user := &UserDTO{
		ID:              uuid.New().String(),
		Username:        strings.ToLower(strings.TrimSpace(username)),
		IdentifierType:  identifierType,
		IdentifierValue: identifierValue,
		Status:          constant.UserStatusActive,
		Profile:         &ProfileDTO{},
		Preferences:     &PreferencesDTO{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 根据标识符类型设置对应的字段
	if err := user.SetIdentifier(identifierType, identifierValue); err != nil {
		return nil, err
	}

	return user, nil
}

// ==================== Getter 方法（只读访问） ====================

// GetID 获取用户ID
func (u *UserDTO) GetID() string {
	return u.ID
}

// GetUsername 获取用户名
func (u *UserDTO) GetUsername() string {
	return u.Username
}

// GetNickname 获取昵称
func (u *UserDTO) GetNickname() string {
	return u.Nickname
}

// GetAvatar 获取头像
func (u *UserDTO) GetAvatar() string {
	return u.Avatar
}

// GetPhone 获取手机号
func (u *UserDTO) GetPhone() string {
	return u.Phone
}

// GetEmail 获取邮箱
func (u *UserDTO) GetEmail() string {
	return u.Email
}

// GetIdentifierType 获取标识类型
func (u *UserDTO) GetIdentifierType() constant.IdentifierType {
	return u.IdentifierType
}

// GetIdentifierValue 获取标识值
func (u *UserDTO) GetIdentifierValue() string {
	return u.IdentifierValue
}

// GetStatus 获取用户状态
func (u *UserDTO) GetStatus() constant.UserStatus {
	return u.Status
}

func (u *UserDTO) GetDisplayStatus() string {
	return u.Status.String()
}

// GetCreatedAt 获取创建时间
func (u *UserDTO) GetCreatedAt() time.Time {
	return u.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (u *UserDTO) GetUpdatedAt() time.Time {
	return u.UpdatedAt
}

// GetProfile 获取用户配置
func (u *UserDTO) GetProfile() *ProfileDTO {
	return u.Profile
}

// GetPreferences 获取用户偏好
func (u *UserDTO) GetPreferences() *PreferencesDTO {
	return u.Preferences
}

// ==================== Setter 方法（可控写入） ====================

// SetNickname 设置昵称
func (u *UserDTO) SetNickname(nickname string) error {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return errors.NewValidationError("昵称不能为空")
	}
	if len(nickname) > commonConstant.MaxNicknameLength {
		return errors.NewValidationError(fmt.Sprintf("昵称过长，最大%d个字符", commonConstant.MaxNicknameLength))
	}
	u.Nickname = nickname
	u.UpdatedAt = time.Now()
	return nil
}

// SetAvatar 设置头像
func (u *UserDTO) SetAvatar(avatar string) error {
	avatar = strings.TrimSpace(avatar)
	if avatar != "" && len(avatar) > 500 {
		return errors.NewValidationError("头像URL过长，最大500个字符")
	}
	u.Avatar = avatar
	u.UpdatedAt = time.Now()
	return nil
}

// SetPhone 设置手机号
func (u *UserDTO) SetPhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		u.Phone = phone
		u.UpdatedAt = time.Now()
		return nil
	}
	if !validatePhone(phone) {
		return errors.NewValidationError("手机号格式无效")
	}
	u.Phone = phone
	u.UpdatedAt = time.Now()
	return nil
}

// SetEmail 设置邮箱
func (u *UserDTO) SetEmail(email string) error {
	email = strings.TrimSpace(email)
	if email != "" {
		email = strings.ToLower(email)
	}
	if email == "" {
		u.Email = email
		u.UpdatedAt = time.Now()
		return nil
	}
	if !validateEmail(email) {
		return errors.NewValidationError("邮箱格式无效")
	}
	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// SetIdentifier 设置用户标识
func (u *UserDTO) SetIdentifier(identifierType constant.IdentifierType, identifierValue string) error {
	identifierValue = strings.TrimSpace(identifierValue)
	if identifierValue == "" {
		return errors.NewValidationError("标识符值不能为空")
	}

	// 总是更新标识符类型和值
	u.IdentifierType = identifierType
	u.IdentifierValue = identifierValue

	// 同步更新对应的phone或email字段
	if identifierType == constant.IdentifierTypePhone {
		u.Phone = identifierValue
		u.Email = ""
	} else if identifierType == constant.IdentifierTypeEmail {
		u.Email = identifierValue
		u.Phone = ""
	}
	u.UpdatedAt = time.Now()

	return nil
}

// SetPassword 设置密码
func (u *UserDTO) SetPassword(password string) error {
	// 清理密码前后空格
	password = strings.TrimSpace(password)

	if password == "" {
		return errors.NewValidationError("密码不能为空")
	}
	if len(password) < constant.PasswordMinLength {
		return errors.NewValidationErrorF("密码长度不能少于%d个字符", constant.PasswordMinLength)
	}
	if len(password) > constant.PasswordMaxLength {
		return errors.NewValidationErrorF("密码长度不能超过%d个字符", constant.PasswordMaxLength)
	}

	// 使用固定成本因子的bcrypt哈希密码（避免版本差异）
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), constant.PasswordCost)
	if err != nil {
		return errors.WrapBizError(err, "密码哈希失败")
	}

	u.PasswordHash = string(hashedBytes)
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateProfile 更新用户配置
func (u *UserDTO) UpdateProfile(profile *ProfileDTO) {
	u.Profile = profile
	u.UpdatedAt = time.Now()
}

// UpdatePreferences 更新用户偏好
func (u *UserDTO) UpdatePreferences(preferences *PreferencesDTO) {
	u.Preferences = preferences
	u.UpdatedAt = time.Now()
}

// IsActive 检查用户是否活跃
func (u *UserDTO) IsActive() bool {
	return u.Status.IsActive()
}

// GetIdentifier 获取用户标识
func (u *UserDTO) GetIdentifier() (constant.IdentifierType, string) {
	return u.IdentifierType, u.IdentifierValue
}

// 私有方法

// validateUsername 验证用户名（独立函数，供NewUser使用）
func validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return errors.NewValidationError("用户名不能为空")
	}

	if len(username) < commonConstant.MinUsernameLength {
		return errors.NewValidationError(fmt.Sprintf("用户名至少需要%d个字符", commonConstant.MinUsernameLength))
	}

	if len(username) > commonConstant.MaxUsernameLength {
		return errors.NewValidationError(fmt.Sprintf("用户名过长，最大%d个字符", commonConstant.MaxUsernameLength))
	}

	// 检查用户名格式（只允许字母、数字、下划线、连字符）
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '-') {
			return errors.NewValidationError("用户名只能包含字母、数字、下划线和连字符")
		}
	}

	// 检查是否为保留的用户名
	reservedNames := []string{"admin", "root", "system", "api", "www", "mail", "ftp"}
	lowerUsername := strings.ToLower(username)

	for _, reserved := range reservedNames {
		if lowerUsername == reserved {
			return errors.NewValidationError("用户名是保留字，不可使用")
		}
	}

	return nil
}

// validatePhone 验证手机号格式
func validatePhone(phone string) bool {
	// 简单的手机号验证，可以根据需要调整
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}

	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// validateEmail 验证邮箱格式
func validateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
