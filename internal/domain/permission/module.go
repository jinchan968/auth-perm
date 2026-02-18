package permission

import (
	"log"

	"go.uber.org/dig"
	"gorm.io/gorm"

	authService "auth-perm/internal/domain/auth/service"
	permissionRepo "auth-perm/internal/domain/permission/repo"
	"auth-perm/internal/domain/permission/service"
	
)

// RegisterPermissionDomain 注册权限领域的所有依赖
func RegisterPermissionDomain(container *dig.Container) error {
	// -----------------------------------------------------
	// 第一步：注册数据仓储层 (Repository Layer)
	// -----------------------------------------------------
	if err := container.Provide(func(db *gorm.DB) *permissionRepo.PermissionRepo {
		return permissionRepo.NewPermissionRepo(db)
	}); err != nil {
		return err
	}

	// 注册权限资源仓储
	if err := container.Provide(func(db *gorm.DB) *permissionRepo.PermissionResourceRepo {
		return permissionRepo.NewPermissionResourceRepo(db)
	}); err != nil {
		return err
	}

	// 注册组织仓储
	if err := container.Provide(func(db *gorm.DB) *permissionRepo.OrganizationRepo {
		return permissionRepo.NewOrganizationRepo(db)
	}); err != nil {
		return err
	}

	// -----------------------------------------------------
	// 第二步：注册领域服务层 (Service Layer)
	// -----------------------------------------------------
	if err := container.Provide(func(
		authService *authService.AuthService,
		permissionRepo *permissionRepo.PermissionRepo,
		permissionResourceRepo *permissionRepo.PermissionResourceRepo,
		cache *authService.CacheService,
		codeGen code_gen.CodeGenerator,
	) *service.PermissionService {
		// 直接使用 auth 服务的实例
		return service.NewPermissionServiceWithResourceRepo(authService, permissionRepo, permissionResourceRepo, cache, codeGen)
	}); err != nil {
		return err
	}

	// 注册组织服务
	if err := container.Provide(func(orgRepo *permissionRepo.OrganizationRepo) *service.OrganizationService {
		return service.NewOrganizationService(orgRepo)
	}); err != nil {
		return err
	}

	// 记录注册成功日志
	log.Println("Permission domain registered successfully")

	return nil
}
