package workflow

import (
	"auth-perm/internal/domain/workflow/handler"
	"auth-perm/internal/domain/workflow/repo"
	"auth-perm/internal/domain/workflow/service"
	"auth-perm/internal/infra/opencode"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

func RegisterWorkflowDomain(container *dig.Container) error {
	container.Provide(func(db *gorm.DB) *repo.WorkflowRepo {
		return repo.NewWorkflowRepo(db)
	})
	container.Provide(func(db *gorm.DB) *repo.WorkflowRunRepo {
		return repo.NewWorkflowRunRepo(db)
	})
	container.Provide(func(db *gorm.DB) *repo.WorkflowRunNodeRepo {
		return repo.NewWorkflowRunNodeRepo(db)
	})

	container.Provide(func(
		wr *repo.WorkflowRepo,
		rr *repo.WorkflowRunRepo,
		nr *repo.WorkflowRunNodeRepo,
		oc *opencode.Client,
	) *service.WorkflowService {
		return service.NewWorkflowService(wr, rr, nr, oc)
	})

	container.Provide(func(svc *service.WorkflowService) *handler.WorkflowHandler {
		return handler.NewWorkflowHandler(svc)
	})
	container.Provide(func(svc *service.WorkflowService) *handler.WorkflowWSHandler {
		return handler.NewWorkflowWSHandler(svc)
	})

	return nil
}
