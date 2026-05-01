package constant

// TenantPlan 租户套餐
type TenantPlan string

const (
	TenantPlanFree       TenantPlan = "free"
	TenantPlanBasic      TenantPlan = "basic"
	TenantPlanPro        TenantPlan = "pro"
	TenantPlanEnterprise TenantPlan = "enterprise"
)

// IsValid 检查套餐是否有效
func (p TenantPlan) IsValid() bool {
	switch p {
	case TenantPlanFree, TenantPlanBasic, TenantPlanPro, TenantPlanEnterprise:
		return true
	}
	return false
}
