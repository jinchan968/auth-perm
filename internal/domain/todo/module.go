package todo

import (
	"log"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"auth-perm/internal/domain/todo/handler"
	"auth-perm/internal/domain/todo/repo"
	"auth-perm/internal/domain/todo/service"
)

// RegisterTodoDomain 注册 todo 领域的所有依赖
func RegisterTodoDomain(container *dig.Container) error {
	// 仓储层
	if err := container.Provide(func(db *gorm.DB) *repo.TodoRepo {
		return repo.NewTodoRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.TodoCategoryRepo {
		return repo.NewTodoCategoryRepo(db)
	}); err != nil {
		return err
	}

	// 服务层
	if err := container.Provide(func(todoRepo *repo.TodoRepo, categoryRepo *repo.TodoCategoryRepo) *service.TodoService {
		return service.NewTodoService(todoRepo, categoryRepo)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(todoRepo *repo.TodoRepo) *service.TodoScheduler {
		return service.NewTodoScheduler(todoRepo)
	}); err != nil {
		return err
	}

	// HTTP 处理器
	if err := container.Provide(func(svc *service.TodoService) *handler.TodoHandler {
		return handler.NewTodoHandler(svc)
	}); err != nil {
		return err
	}

	log.Println("Todo domain registered successfully")
	return nil
}
