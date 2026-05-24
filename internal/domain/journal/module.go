package journal

import (
	"go.uber.org/dig"
	"gorm.io/gorm"

	"auth-perm/internal/domain/journal/handler"
	"auth-perm/internal/domain/journal/repo"
	"auth-perm/internal/domain/journal/service"
	"auth-perm/internal/infra/opencode"
)

// RegisterJournalDomain 注册 journal 领域的所有依赖
func RegisterJournalDomain(container *dig.Container) error {
	// 仓储层
	if err := container.Provide(func(db *gorm.DB) *repo.JournalRepo {
		return repo.NewJournalRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.TagRepo {
		return repo.NewTagRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.TemplateRepo {
		return repo.NewTemplateRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.AIPredictionRepo {
		return repo.NewAIPredictionRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.AICallQuotaRepo {
		return repo.NewAICallQuotaRepo(db)
	}); err != nil {
		return err
	}

	// 服务层
	if err := container.Provide(func(entryRepo *repo.JournalRepo, tagRepo *repo.TagRepo) *service.JournalService {
		return service.NewJournalService(entryRepo, tagRepo)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(templateRepo *repo.TemplateRepo) *service.TemplateService {
		return service.NewTemplateService(templateRepo)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(oc *opencode.Client, pr *repo.AIPredictionRepo, qr *repo.AICallQuotaRepo) *service.AIPredictionService {
		return service.NewAIPredictionService(oc, pr, qr)
	}); err != nil {
		return err
	}

	// HTTP 处理器
	if err := container.Provide(func(svc *service.JournalService) *handler.JournalHandler {
		return handler.NewJournalHandler(svc)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(svc *service.TemplateService) *handler.TemplateHandler {
		return handler.NewTemplateHandler(svc)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(svc *service.AIPredictionService) *handler.AIPredictionHandler {
		return handler.NewAIPredictionHandler(svc)
	}); err != nil {
		return err
	}

	return nil
}
