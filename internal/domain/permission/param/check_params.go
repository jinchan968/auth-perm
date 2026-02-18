package param

import "fmt"

// CheckPermissionParams 检查权限参数
type CheckPermissionParams struct {
	AccountID      string
	PermissionCode string
}

// NewCheckPermissionParams 创建权限检查参数
func NewCheckPermissionParams(accountID, permissionCode string) *CheckPermissionParams {
	return &CheckPermissionParams{
		AccountID:      accountID,
		PermissionCode: permissionCode,
	}
}

// Validate 验证检查权限参数
func (p *CheckPermissionParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if p.PermissionCode == "" {
		return errEmpty("权限代码")
	}
	return nil
}

// CheckOrgPermissionParams 检查组织权限参数
type CheckOrgPermissionParams struct {
	AccountID      string
	OrgID          string
	PermissionCode string
}

// NewCheckOrgPermissionParams 创建检查组织权限参数
func NewCheckOrgPermissionParams(accountID, orgID, permissionCode string) *CheckOrgPermissionParams {
	return &CheckOrgPermissionParams{
		AccountID:      accountID,
		OrgID:          orgID,
		PermissionCode: permissionCode,
	}
}

// Validate 验证检查组织权限参数
func (p *CheckOrgPermissionParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if p.OrgID == "" {
		return errEmpty("组织ID")
	}
	if p.PermissionCode == "" {
		return errEmpty("权限代码")
	}
	return nil
}

// CheckPermissionsParams 检查多个权限参数
type CheckPermissionsParams struct {
	AccountID   string
	Permissions []string
	// CheckType: "any" 表示任意一个，"all" 表示全部
	CheckType string
}

// NewCheckAnyPermissionsParams 创建任意权限检查参数
func NewCheckAnyPermissionsParams(accountID string, permissions []string) *CheckPermissionsParams {
	return &CheckPermissionsParams{
		AccountID:   accountID,
		Permissions: permissions,
		CheckType:   "any",
	}
}

// NewCheckAllPermissionsParams 创建所有权限检查参数
func NewCheckAllPermissionsParams(accountID string, permissions []string) *CheckPermissionsParams {
	return &CheckPermissionsParams{
		AccountID:   accountID,
		Permissions: permissions,
		CheckType:   "all",
	}
}

// Validate 验证检查多个权限参数
func (p *CheckPermissionsParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if len(p.Permissions) == 0 {
		return errEmpty("权限列表")
	}
	if p.CheckType != "any" && p.CheckType != "all" {
		return &ValidationError{Field: "检查类型", Message: "必须是 'any' 或 'all'"}
	}
	return nil
}

// CheckRolesParams 检查多个角色参数
type CheckRolesParams struct {
	AccountID string
	RoleCodes []string
	// CheckType: "any" 表示任意一个，"all" 表示全部
	CheckType string
}

// NewCheckAnyRolesParams 创建任意角色检查参数
func NewCheckAnyRolesParams(accountID string, roleCodes []string) *CheckRolesParams {
	return &CheckRolesParams{
		AccountID: accountID,
		RoleCodes: roleCodes,
		CheckType: "any",
	}
}

// NewCheckAllRolesParams 创建所有角色检查参数
func NewCheckAllRolesParams(accountID string, roleCodes []string) *CheckRolesParams {
	return &CheckRolesParams{
		AccountID: accountID,
		RoleCodes: roleCodes,
		CheckType: "all",
	}
}

// Validate 验证检查多个角色参数
func (p *CheckRolesParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if len(p.RoleCodes) == 0 {
		return errEmpty("角色列表")
	}
	if p.CheckType != "any" && p.CheckType != "all" {
		return &ValidationError{Field: "检查类型", Message: "必须是 'any' 或 'all'"}
	}
	return nil
}

// GetUserDataParams 获取账户数据参数
type GetUserDataParams struct {
	AccountID string
}

// NewGetUserPermissionsParams 创建获取账户权限参数
func NewGetUserPermissionsParams(accountID string) *GetUserDataParams {
	return &GetUserDataParams{AccountID: accountID}
}

// NewGetUserRolesParams 创建获取账户角色参数
func NewGetUserRolesParams(accountID string) *GetUserDataParams {
	return &GetUserDataParams{AccountID: accountID}
}

// Validate 验证获取账户数据参数
func (p *GetUserDataParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	return nil
}

// IsAdminParams 检查管理员参数
type IsAdminParams struct {
	AccountID string
	OrgID     string
}

// NewIsSystemAdminParams 创建系统管理员检查参数
func NewIsSystemAdminParams(accountID string) *IsAdminParams {
	return &IsAdminParams{AccountID: accountID}
}

// NewIsOrgAdminParams 创建组织管理员检查参数
func NewIsOrgAdminParams(accountID, orgID string) *IsAdminParams {
	return &IsAdminParams{AccountID: accountID, OrgID: orgID}
}

// Validate 验证管理员参数
func (p *IsAdminParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// errEmpty 空字段错误
func errEmpty(field string) error {
	return &ValidationError{Field: field, Message: "不能为空"}
}
