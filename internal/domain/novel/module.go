package novel

import (
	"log"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"auth-perm/internal/domain/novel/handler"
	"auth-perm/internal/domain/novel/repo"
	"auth-perm/internal/domain/novel/service"
)

func RegisterNovelDomain(container *dig.Container) error {
	if err := container.Provide(func(db *gorm.DB) *repo.NovelRepo {
		return repo.NewNovelRepo(db)
	}); err != nil {
		return err
	}

	if err := container.Provide(func(novelRepo *repo.NovelRepo) *service.NovelService {
		return service.NewNovelService(novelRepo)
	}); err != nil {
		return err
	}

	if err := container.Provide(func(svc *service.NovelService) *handler.NovelHandler {
		return handler.NewNovelHandler(svc)
	}); err != nil {
		return err
	}

	log.Println("Novel domain registered successfully")
	return nil
}
