package multimodal

import (
	"go.uber.org/dig"

	"auth-perm/config"
	"auth-perm/internal/domain/multimodal/handler"
	"auth-perm/internal/domain/multimodal/service"
	"auth-perm/internal/infra/opencode"
)

// RegisterMultimodalDomain 注册多模态领域的所有依赖
func RegisterMultimodalDomain(container *dig.Container) error {
	if err := container.Provide(func(oc *opencode.Client, cfg *config.Config) *service.MultimodalService {
		return service.NewMultimodalService(oc, &cfg.ImageGen)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(svc *service.MultimodalService) *handler.MultimodalHandler {
		return handler.NewMultimodalHandler(svc)
	}); err != nil {
		return err
	}
	return nil
}
