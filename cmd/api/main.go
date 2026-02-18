package main

import (
	"auth-perm/config"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// 启动并运行应用
	if err := c.Invoke(func(engine *gin.Engine) {

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

		// 优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		log.Println("Server exited")
	}); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
