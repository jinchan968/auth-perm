package constant

// TenantStatus 租户状态
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// IsValid 检查状态是否有效
func (s TenantStatus) IsValid() bool {
	switch s {
	case TenantStatusActive, TenantStatusSuspended, TenantStatusDeleted:
		return true
	}
	return false
}

// IsActive 检查租户是否活跃
func (s TenantStatus) IsActive() bool {
	return s == TenantStatusActive
}
