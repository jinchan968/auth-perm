package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strings"
	"time"

	commonConstant "auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	authRepo "auth-perm/internal/domain/auth/repo"
	"auth-perm/internal/domain/auth/vo"

	"gorm.io/gorm"
)

// RegistrationInvitationService 管理注册邀请码
type RegistrationInvitationService struct {
	repo *authRepo.RegistrationInvitationRepo
}

func NewRegistrationInvitationService(repo *authRepo.RegistrationInvitationRepo) *RegistrationInvitationService {
	return &RegistrationInvitationService{repo: repo}
}

func (s *RegistrationInvitationService) Create(ctx context.Context, requestedTenantID, createdByAccountID, actorTenantID string, allowCrossTenant bool, expiresAt *time.Time, baseURL string) (*vo.CreateInvitationResponse, error) {
	tenantID, err := resolveInvitationTenantScope(requestedTenantID, actorTenantID, allowCrossTenant)
	if err != nil {
		return nil, err
	}
	if createdByAccountID == "" {
		return nil, errors.NewValidationError("创建人不能为空")
	}

	expiry := time.Now().Add(authConstant.DefaultInvitationTTL)
	if expiresAt != nil {
		if expiresAt.Before(time.Now()) {
			return nil, errors.NewValidationError("过期时间不能早于当前时间")
		}
		expiry = *expiresAt
	}

	code, err := generateInvitationCode()
	if err != nil {
		return nil, err
	}
	invitation := dm.NewRegistrationInvitation(hashInvitationCode(code), previewInvitationCode(code), tenantID, createdByAccountID, expiry)
	if err := s.repo.Create(ctx, invitation); err != nil {
		return nil, errors.WrapBizError(err, "创建邀请码失败")
	}

	resp := s.toResponse(invitation)
	return &vo.CreateInvitationResponse{
		InvitationResponse: *resp,
		InviteCode:         code,
		InviteURL:          buildInviteURL(baseURL, code),
	}, nil
}

func (s *RegistrationInvitationService) List(ctx context.Context, requestedTenantID, actorTenantID, status string, page, pageSize int, allowCrossTenant bool) (*vo.InvitationListResponse, error) {
	tenantID, err := resolveInvitationTenantScope(requestedTenantID, actorTenantID, allowCrossTenant)
	if err != nil {
		return nil, err
	}
	items, total, err := s.repo.List(ctx, tenantID, status, page, pageSize)
	if err != nil {
		return nil, err
	}
	data := make([]*vo.InvitationResponse, 0, len(items))
	for _, item := range items {
		data = append(data, s.toResponse(item))
	}
	return &vo.InvitationListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

func (s *RegistrationInvitationService) Invalidate(ctx context.Context, id, accountID, actorTenantID string, allowCrossTenant bool) error {
	if id == "" {
		return errors.NewValidationError("邀请码ID不能为空")
	}
	if accountID == "" {
		return errors.NewValidationError("操作人不能为空")
	}
	invitation, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !allowCrossTenant && invitation.TenantID != actorTenantID {
		return errors.NewBusinessError("无权操作其他租户的邀请码")
	}
	if s.effectiveStatus(invitation) != authConstant.InvitationStatusActive {
		return errors.NewBusinessError("邀请码不可失效")
	}
	return s.repo.Invalidate(ctx, id, accountID)
}

func (s *RegistrationInvitationService) findForRegistrationWithTx(ctx context.Context, tx *gorm.DB, code string) (*dm.RegistrationInvitationDO, error) {
	code = normalizeInvitationCode(code)
	if code == "" {
		return nil, errors.NewValidationError("邀请码不能为空")
	}
	invitation, err := s.repo.FindByCodeHashForUpdate(ctx, tx, hashInvitationCode(code))
	if err != nil {
		return nil, err
	}
	if err := s.validateInvitationState(invitation); err != nil {
		return nil, err
	}
	return invitation, nil
}

func (s *RegistrationInvitationService) toResponse(invitation *dm.RegistrationInvitationDO) *vo.InvitationResponse {
	usedBy := ""
	if invitation.UsedByAccountID != nil {
		usedBy = *invitation.UsedByAccountID
	}
	invalidatedBy := ""
	if invitation.InvalidatedByAccountID != nil {
		invalidatedBy = *invitation.InvalidatedByAccountID
	}
	return &vo.InvitationResponse{
		ID:                     invitation.ID,
		CodePreview:            invitation.CodePreview,
		TenantID:               invitation.TenantID,
		Status:                 s.effectiveStatus(invitation),
		ExpiresAt:              invitation.ExpiresAt,
		UsedAt:                 invitation.UsedAt,
		UsedByAccountID:        usedBy,
		CreatedByAccountID:     invitation.CreatedByAccountID,
		InvalidatedAt:          invitation.InvalidatedAt,
		InvalidatedByAccountID: invalidatedBy,
		CreatedAt:              invitation.CreatedAt,
		UpdatedAt:              invitation.UpdatedAt,
	}
}

func (s *RegistrationInvitationService) effectiveStatus(invitation *dm.RegistrationInvitationDO) string {
	if invitation.Status == authConstant.InvitationStatusActive && !invitation.ExpiresAt.After(time.Now()) {
		return authConstant.InvitationStatusExpired
	}
	return invitation.Status
}

func (s *RegistrationInvitationService) validateInvitationState(invitation *dm.RegistrationInvitationDO) error {
	switch s.effectiveStatus(invitation) {
	case authConstant.InvitationStatusActive:
		return nil
	case authConstant.InvitationStatusUsed:
		return errors.NewBusinessError("邀请码已被使用")
	case authConstant.InvitationStatusInvalidated:
		return errors.NewBusinessError("邀请码已失效")
	case authConstant.InvitationStatusExpired:
		return errors.NewBusinessError("邀请码已过期")
	default:
		return errors.NewBusinessError("邀请码不可用")
	}
}

func resolveInvitationTenantScope(requestedTenantID, actorTenantID string, allowCrossTenant bool) (string, error) {
	requestedTenantID = strings.TrimSpace(requestedTenantID)
	actorTenantID = strings.TrimSpace(actorTenantID)

	if allowCrossTenant {
		if requestedTenantID == "" {
			return commonConstant.DefaultTenantID, nil
		}
		return requestedTenantID, nil
	}

	if actorTenantID == "" {
		return "", errors.NewValidationError("租户ID不能为空")
	}
	if requestedTenantID != "" && requestedTenantID != actorTenantID {
		return "", errors.NewBusinessError("无权操作其他租户的邀请码")
	}
	return actorTenantID, nil
}

func generateInvitationCode() (string, error) {
	buf := make([]byte, authConstant.InvitationCodeBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.WrapBizError(err, "生成邀请码失败")
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(buf), "="), nil
}

func normalizeInvitationCode(code string) string {
	return strings.TrimSpace(code)
}

func hashInvitationCode(code string) string {
	sum := sha256.Sum256([]byte(normalizeInvitationCode(code)))
	return hex.EncodeToString(sum[:])
}

func previewInvitationCode(code string) string {
	code = normalizeInvitationCode(code)
	if len(code) <= 8 {
		return code
	}
	return code[:4] + "..." + code[len(code)-4:]
}

func buildInviteURL(baseURL, code string) string {
	baseURL = normalizeInviteBaseURL(baseURL)
	if baseURL == "" {
		return "/register?invite_code=" + code
	}
	return baseURL + "/register?invite_code=" + code
}

func normalizeInviteBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}
