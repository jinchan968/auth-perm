package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-perm/config"
	"auth-perm/internal/container"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config/app.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 确保日志目录存在
	if err := config.EnsureLogDir(cfg.Log.File); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 构建依赖注入容器
	c, err := container.BuildContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to build container: %v", err)
	}

	// 创建顶层可取消 context，用于控制所有后台 goroutine（含定时任务）的生命周期
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel() // 兜底：确保任何异常退出路径都能释放 context

	// 启动并运行应用
	if err := c.Invoke(func(engine *gin.Engine, scheduler container.Scheduler) {

		// 启动 Todo 定时调度器，绑定 appCtx——服务关闭时随之停止
		go scheduler.Start(appCtx)

		// 创建HTTP服务器
		server := &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      engine,
			ReadTimeout:  time.Second * 30,
			WriteTimeout: time.Second * 30,
			IdleTimeout:  time.Second * 60,
		}

		// 启动服务器
		go func() {
			log.Printf("Starting server on %s:%d", cfg.Server.Host, cfg.Server.Port)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("Failed to start server: %v", err)
			}
		}()

		// 等待中断信号
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down server...")

		// 先取消 appCtx，通知所有后台 goroutine 退出
		appCancel()

		// 再优雅关闭 HTTP server
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		log.Println("Server exited")
	}); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
