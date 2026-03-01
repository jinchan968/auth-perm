package service

import (
	"context"
	"log"
	"time"

	"auth-perm/internal/domain/todo/repo"
)

// TodoScheduler 待办定时任务调度器
// 职责：定期将截止时间已过且仍活跃的待办优先级升级为 urgent
type TodoScheduler struct {
	todoRepo *repo.TodoRepo
	interval time.Duration
}

// NewTodoScheduler 创建调度器，interval 默认 1 分钟
func NewTodoScheduler(todoRepo *repo.TodoRepo) *TodoScheduler {
	return &TodoScheduler{
		todoRepo: todoRepo,
		interval: time.Minute,
	}
}

// Start 启动调度器（阻塞，应在独立 goroutine 中调用）
// ctx 取消时自动停止
func (s *TodoScheduler) Start(ctx context.Context) {
	log.Println("[TodoScheduler] 启动，扫描间隔:", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 启动时立即执行一次
	s.escalate(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("[TodoScheduler] 已停止")
			return
		case <-ticker.C:
			s.escalate(ctx)
		}
	}
}

// escalate 执行一次升级
func (s *TodoScheduler) escalate(ctx context.Context) {
	affected, err := s.todoRepo.EscalateOverdue(ctx)
	if err != nil {
		log.Printf("[TodoScheduler] 升级过期待办失败: %v", err)
		return
	}
	if affected > 0 {
		log.Printf("[TodoScheduler] 已将 %d 条过期待办升级为 urgent", affected)
	}
}
