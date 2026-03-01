package tenant

import (
	"auth-perm/internal/infra/code_gen"
	"log"

	"go.uber.org/dig"
	"gorm.io/gorm"

	tenantRepo "auth-perm/internal/domain/tenant/repo"
	"auth-perm/internal/domain/tenant/service"
)

// RegisterTenantDomain 注册租户领域的所有依赖
func RegisterTenantDomain(container *dig.Container) error {
	// -----------------------------------------------------
	// 第一步：注册数据仓储层 (Repository Layer)
	// -----------------------------------------------------
	if err := container.Provide(func(db *gorm.DB) *tenantRepo.TenantRepo {
		return tenantRepo.NewTenantRepo(db)
	}); err != nil {
		return err
	}

	// -----------------------------------------------------
	// 第二步：注册领域服务层 (Service Layer)
	// SessionInvalidator 由 auth 域提供，用于租户删除时失效会话
	// -----------------------------------------------------
	if err := container.Provide(func(
		tenantRepo *tenantRepo.TenantRepo,
		codeGen code_gen.CodeGenerator,
		sessionInvalidator service.SessionInvalidator,
	) *service.TenantService {
		return service.NewTenantService(tenantRepo, codeGen, sessionInvalidator)
	}); err != nil {
		return err
	}

	// 记录注册成功日志
	log.Println("Tenant domain registered successfully")

	return nil
}
