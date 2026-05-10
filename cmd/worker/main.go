package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"auth-perm/config"
	"auth-perm/internal/container"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Printf(".env file not found or failed to load, using system env: %v", err)
	}

	// 加载配置
	cfg, err := config.LoadConfig("config/app.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 确保日志目录存在
	if err := config.EnsureLogDir(cfg.Log.File); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// 构建 Worker 容器（只包含基础设施 + 调度器，不含 HTTP 处理器）
	c, err := container.BuildWorkerContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to build worker container: %v", err)
	}

	// 创建顶层可取消 context，用于控制所有调度器的生命周期
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// 启动并运行定时任务
	if err := c.Invoke(func(scheduler container.Scheduler) {
		log.Println("Starting worker scheduler...")
		go scheduler.Start(appCtx)

		// 阻塞等待中断信号
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down worker...")
		appCancel() // 通知所有调度器退出
		log.Println("Worker exited")
	}); err != nil {
		log.Fatalf("Failed to run worker: %v", err)
	}
}
