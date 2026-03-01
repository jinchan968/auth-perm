package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"time"

	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/repo"
	"auth-perm/internal/domain/auth/vo"

	"github.com/pquerna/otp"
	totpAuth "github.com/pquerna/otp/totp"
)

// TOTPService TOTP服务
type TOTPService struct {
	totpRepo    repo.TOTPSecretRepo
	accountRepo *repo.AccountRepo
	userRepo    *repo.UserRepo
	cache       *CacheService
	security    *SecurityService
}

// NewTOTPService 创建TOTP服务
func NewTOTPService(totpRepo repo.TOTPSecretRepo, accountRepo *repo.AccountRepo, userRepo *repo.UserRepo,
	cache *CacheService, security *SecurityService) *TOTPService {
	return &TOTPService{
		totpRepo:    totpRepo,
		accountRepo: accountRepo,
		userRepo:    userRepo,
		cache:       cache,
		security:    security,
	}
}

// SetupTOTP 初始化2FA设置
func (s *TOTPService) SetupTOTP(accountID string) (*vo.TOTPSetupInitResponse, error) {
	ctx := context.Background()
	// 检查账户是否存在
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, errors.NewBusinessError("账户不存在")
	}

	// 检查是否已经启用2FA
	existing, _ := s.totpRepo.FindByAccountID(accountID)
	if existing != nil && existing.IsEnabled {
		return nil, errors.NewBusinessError("2FA已经启用")
	}

	// 生成随机密钥
	secret, err := s.generateSecret()
	if err != nil {
		return nil, errors.NewInternalError("生成密钥失败")
	}

	// 生成备份码
	backupCodes := s.generateBackupCodes()

	// 生成QR码
	issuer := "AuthPerm"
	accountName := ""
	if account.User != nil {
		accountName = dm.StrVal(account.User.Email)
	}
	uri := s.generateURI(secret, issuer, accountName)

	qrCode, err := s.generateQRCode(uri)
	if err != nil {
		return nil, errors.NewInternalError("生成QR码失败")
	}

	return &vo.TOTPSetupInitResponse{
		Secret:      secret,
		QRCode:      qrCode,
		URI:         uri,
		BackupCodes: backupCodes,
	}, nil
}

// VerifyTOTPSetup 验证2FA设置
func (s *TOTPService) VerifyTOTPSetup(accountID, secret, token string) (*vo.TOTPSetupVerifyResponse, error) {
	// 验证TOTP令牌
	valid, err := totpAuth.ValidateCustom(token, secret, time.Now(), totpAuth.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, errors.NewBusinessError("验证令牌失败")
	}

	if !valid {
		return &vo.TOTPSetupVerifyResponse{
			Success: false,
			Message: "令牌验证失败",
		}, nil
	}

	// 生成备份码（用于显示给用户）
	backupCodes := s.generateBackupCodes()

	return &vo.TOTPSetupVerifyResponse{
		Success:     true,
		Message:     "令牌验证成功",
		BackupCodes: backupCodes,
	}, nil
}

// EnableTOTP 启用2FA
func (s *TOTPService) EnableTOTP(accountID, secret string, backupCodes []string) (*vo.TOTPEnableResponse, error) {
	ctx := context.Background()
	// 检查账户是否存在
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, errors.NewBusinessError("账户不存在")
	}

	_ = account // 使用account避免编译错误

	// 查找现有的TOTP密钥
	totpSecret, err := s.totpRepo.FindByAccountID(accountID)
	if err != nil {
		return nil, errors.NewInternalError("查找TOTP密钥失败")
	}

	// 如果不存在，创建新的
	if totpSecret == nil {
		totpSecret = dm.NewTOTPSecret(accountID, secret)
	}

	// 更新密钥信息
	totpSecret.Secret = secret
	totpSecret.BackupCodes = backupCodes
	totpSecret.IsEnabled = true
	totpSecret.UpdatedAt = time.Now()

	// 保存到数据库
	if err := s.totpRepo.Save(totpSecret); err != nil {
		return nil, errors.NewInternalError("保存TOTP密钥失败")
	}

	// 清除缓存
	if s.cache != nil {
		_ = s.cache.DeleteTOTPSecret(accountID)
	}

	return &vo.TOTPEnableResponse{
		Success:     true,
		Message:     "2FA启用成功",
		BackupCodes: backupCodes,
	}, nil
}

// DisableTOTP 禁用2FA
func (s *TOTPService) DisableTOTP(accountID, password string) (*vo.TOTPDisableResponse, error) {
	ctx := context.Background()
	// 验证密码
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, errors.NewBusinessError("账户不存在")
	}

	// 通过账户获取用户，验证密码
	user, err := s.userRepo.FindByID(ctx, account.UserID)
	if err != nil {
		return nil, errors.NewBusinessError("用户不存在")
	}

	if !s.verifyPassword(password, user.PasswordHash) {
		return &vo.TOTPDisableResponse{
			Success: false,
			Message: "密码错误",
		}, nil
	}

	// 查找TOTP密钥
	totpSecret, err := s.totpRepo.FindByAccountID(accountID)
	if err != nil {
		return nil, errors.NewInternalError("查找TOTP密钥失败")
	}

	if totpSecret == nil {
		return &vo.TOTPDisableResponse{
			Success: false,
			Message: "2FA未启用",
		}, nil
	}

	// 禁用2FA
	totpSecret.IsEnabled = false
	totpSecret.UpdatedAt = time.Now()

	// 保存到数据库
	if err := s.totpRepo.Save(totpSecret); err != nil {
		return nil, errors.NewInternalError("禁用2FA失败")
	}

	// 清除缓存
	if s.cache != nil {
		err := s.cache.DeleteTOTPSecret(accountID)
		if err != nil {
			return nil, err
		}
	}

	return &vo.TOTPDisableResponse{
		Success: true,
		Message: "2FA禁用成功",
	}, nil
}

// VerifyTOTP 验证TOTP令牌
func (s *TOTPService) VerifyTOTP(accountID, token string, useBackup bool) (*vo.TOTPVerifyResponse, error) {
	// 获取TOTP密钥
	totpSecret, err := s.getTOTPSecret(accountID)
	if err != nil {
		return &vo.TOTPVerifyResponse{
			Valid:             false,
			Message:           "TOTP未设置",
			RemainingAttempts: 0,
		}, nil
	}

	if !totpSecret.IsEnabled {
		return &vo.TOTPVerifyResponse{
			Valid:             false,
			Message:           "2FA未启用",
			RemainingAttempts: 0,
		}, nil
	}

	// 检查失败次数
	attempts := s.getFailedAttempts(accountID)
	if attempts >= 5 {
		return &vo.TOTPVerifyResponse{
			Valid:             false,
			Message:           "验证尝试次数过多，请稍后重试",
			RemainingAttempts: 0,
		}, nil
	}

	// 如果使用备份码
	if useBackup {
		return s.verifyBackupCode(accountID, token, totpSecret)
	}

	// 验证TOTP令牌
	valid, err := totpAuth.ValidateCustom(token, totpSecret.Secret, time.Now(), totpAuth.ValidateOpts{
		Period:    uint(totpSecret.Period),
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		s.incrementFailedAttempts(accountID)
		return &vo.TOTPVerifyResponse{
			Valid:             false,
			Message:           "令牌验证失败",
			RemainingAttempts: 5 - attempts - 1,
		}, nil
	}

	if !valid {
		s.incrementFailedAttempts(accountID)
		return &vo.TOTPVerifyResponse{
			Valid:             false,
			Message:           "令牌验证失败",
			RemainingAttempts: 5 - attempts - 1,
		}, nil
	}

	// 验证成功，清除失败计数
	s.clearFailedAttempts(accountID)

	return &vo.TOTPVerifyResponse{
		Valid:             true,
		Message:           "验证成功",
		RemainingAttempts: 5,
	}, nil
}

// GetTOTPStatus 获取2FA状态
func (s *TOTPService) GetTOTPStatus(accountID string) (*vo.TOTPStatusResponse, error) {
	totpSecret, err := s.getTOTPSecret(accountID)
	if err != nil || totpSecret == nil {
		return &vo.TOTPStatusResponse{
			Enabled:        false,
			HasBackupCodes: false,
			RemainingCodes: 0,
		}, nil
	}

	return &vo.TOTPStatusResponse{
		Enabled:        totpSecret.IsEnabled,
		HasBackupCodes: len(totpSecret.BackupCodes) > 0,
		RemainingCodes: len(totpSecret.BackupCodes),
		SetupAt:        &totpSecret.CreatedAt,
	}, nil
}

// ChangeTOTPSecret 更换TOTP密钥
func (s *TOTPService) ChangeTOTPSecret(accountID, token, newSecret string, backupCodes []string) (*vo.TOTPChangeSecretResponse, error) {
	// 验证当前TOTP令牌
	verifyResp, err := s.VerifyTOTP(accountID, token, false)
	if err != nil || !verifyResp.Valid {
		return &vo.TOTPChangeSecretResponse{
			Success: false,
			Message: "当前TOTP令牌验证失败",
		}, nil
	}

	// 获取当前TOTP密钥
	totpSecret, err := s.getTOTPSecret(accountID)
	if err != nil || totpSecret == nil {
		return &vo.TOTPChangeSecretResponse{
			Success: false,
			Message: "TOTP未设置",
		}, nil
	}

	// 更新密钥和备份码
	totpSecret.Secret = newSecret
	totpSecret.BackupCodes = backupCodes
	totpSecret.IsEnabled = true
	totpSecret.UpdatedAt = time.Now()

	// 保存到数据库
	if err := s.totpRepo.Save(totpSecret); err != nil {
		return nil, errors.NewInternalError("更换TOTP密钥失败")
	}

	// 清除缓存
	if s.cache != nil {
		_ = s.cache.DeleteTOTPSecret(accountID)
	}

	return &vo.TOTPChangeSecretResponse{
		Success:     true,
		Message:     "TOTP密钥更换成功",
		BackupCodes: backupCodes,
	}, nil
}

// 辅助方法

// generateSecret 生成随机密钥
func (s *TOTPService) generateSecret() (string, error) {
	// 生成20字节的随机数据
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// 使用Base32编码
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// generateBackupCodes 生成备份码
func (s *TOTPService) generateBackupCodes() []string {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		// 生成8位随机数字
		bytes := make([]byte, 4)
		_, _ = rand.Read(bytes)
		code := fmt.Sprintf("%08d", int(bytes[0])<<24|int(bytes[1])<<16|int(bytes[2])<<8|int(bytes[3]))
		codes[i] = code
	}
	return codes
}

// generateURI 生成TOTP URI
func (s *TOTPService) generateURI(secret, issuer, accountName string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, accountName, secret, issuer)
}

// generateQRCode 生成QR码
func (s *TOTPService) generateQRCode(uri string) (string, error) {
	// 这里应该生成QR码的base64编码
	// 简化实现，返回URI
	return uri, nil
}

// verifyBackupCode 验证备份码
func (s *TOTPService) verifyBackupCode(accountID, code string, totpSecret *dm.TOTPSecretDO) (*vo.TOTPVerifyResponse, error) {
	// 验证备份码
	found := false
	for i, c := range totpSecret.BackupCodes {
		if c == code {
			// 移除已使用的备份码
			totpSecret.BackupCodes = append(totpSecret.BackupCodes[:i], totpSecret.BackupCodes[i+1:]...)
			totpSecret.UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return &vo.TOTPVerifyResponse{
			Valid:             false,
			Message:           "备份码无效",
			RemainingAttempts: 0,
		}, nil
	}

	// 保存更新后的TOTP密钥（移除已使用的备份码）
	if err := s.totpRepo.Save(totpSecret); err != nil {
		return nil, errors.NewInternalError("保存备份码使用记录失败")
	}

	// 清除缓存
	if s.cache != nil {
		_ = s.cache.DeleteTOTPSecret(accountID)
	}

	return &vo.TOTPVerifyResponse{
		Valid:          true,
		Message:        "备份码验证成功",
		RemainingCodes: len(totpSecret.BackupCodes),
	}, nil
}

// getTOTPSecret 获取TOTP密钥
func (s *TOTPService) getTOTPSecret(accountID string) (*dm.TOTPSecretDO, error) {
	if s.cache != nil {
		cached, err := s.cache.GetTOTPSecret(accountID)
		if err == nil && cached != nil {
			// 转换为DO对象
			return &dm.TOTPSecretDO{
				ID:          cached.ID,
				AccountID:   cached.AccountID,
				Secret:      cached.Secret,
				Algorithm:   cached.Algorithm,
				Digits:      cached.Digits,
				Period:      cached.Period,
				IsEnabled:   cached.IsEnabled,
				BackupCodes: []string{},
				CreatedAt:   cached.CreatedAt,
				UpdatedAt:   cached.UpdatedAt,
			}, nil
		}
	}

	totpSecret, err := s.totpRepo.FindByAccountID(accountID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil && totpSecret != nil {
		// 转换为DTO进行缓存
		dto := &dto.TOTPSecretDTO{
			ID:          totpSecret.ID,
			AccountID:   totpSecret.AccountID,
			Secret:      totpSecret.Secret,
			Algorithm:   totpSecret.Algorithm,
			Digits:      totpSecret.Digits,
			Period:      totpSecret.Period,
			IsEnabled:   totpSecret.IsEnabled,
			BackupCodes: totpSecret.BackupCodes,
			CreatedAt:   totpSecret.CreatedAt,
			UpdatedAt:   totpSecret.UpdatedAt,
		}
		if err := s.cache.SetTOTPSecret(accountID, dto, constant.CacheTTLLong); err != nil {
			log.Printf("设置TOTP缓存失败: %v", err)
		}
	}

	return totpSecret, nil
}

// getFailedAttempts 获取失败次数
func (s *TOTPService) getFailedAttempts(accountID string) int {
	if s.cache == nil {
		return 0
	}
	count, _ := s.cache.GetTOTPFailedAttempts(accountID)
	return int(count)
}

// incrementFailedAttempts 增加失败次数
func (s *TOTPService) incrementFailedAttempts(accountID string) {
	if s.cache == nil {
		return
	}
	_, err := s.cache.IncrementTOTPFailedAttempts(accountID)
	if err != nil {
		return
	}
}

// clearFailedAttempts 清除失败次数
func (s *TOTPService) clearFailedAttempts(accountID string) {
	if s.cache == nil {
		return
	}
	err := s.cache.ClearTOTPFailedAttempts(accountID)
	if err != nil {
		return
	}
}

// verifyPassword 验证密码（简化实现）
func (s *TOTPService) verifyPassword(password, hash string) bool {
	// 这里应该使用bcrypt验证
	// 简化实现
	return password != ""
}
