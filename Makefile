.PHONY: help build test run clean docker-build docker-run migrate-up migrate-down migrate-create migrate-status lint fmt vet

GOBASE=$(shell pwd)
INTERNAL_PATH=GOBASE/internal

# 默认目标
help: ## 显示帮助信息
	@echo "Auth-Perm 开发工具"
	@echo ""
	@echo "可用命令："
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 构建应用
build: ## 构建应用
	@echo "构建应用..."
	go build -o .bin/auth-perm cmd/api/main.go

# 运行测试
test: ## 运行所有测试
	@echo "运行测试..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 运行特定测试
test-unit: ## 运行单元测试
	@echo "运行单元测试..."
	go test -v -short ./...

test-integration: ## 运行集成测试
	@echo "运行集成测试..."
	go test -v -tags=integration ./...

# 运行应用
run: ## 运行应用
	@echo "启动应用..."
	go run cmd/api/main.go

# 清理构建文件
clean: ## 清理构建文件
	@echo "清理构建文件..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -testcache

# 代码检查
lint: ## 运行代码检查
	@echo "运行代码检查..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "安装 golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

# 代码格式化
fmt: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
vet: ## 运行 go vet
	@echo "运行 go vet..."
	go vet ./...

# 依赖管理
deps: ## 下载依赖
	@echo "下载依赖..."
	go mod download
	go mod tidy

# 依赖更新
deps-update: ## 更新依赖
	@echo "更新依赖..."
	go get -u ./...
	go mod tidy

# 数据库迁移
migrate-install: ## 安装 goose 工具
	@echo "安装 goose 工具..."
	GOPROXY=https://goproxy.cn,direct go install github.com/pressly/goose/v3/cmd/goose@latest

migrate-up: ## 运行数据库迁移（向上）
	@echo "运行数据库迁移..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "错误: 请设置 DB_URL 环境变量"; \
		echo "示例: export DB_URL='postgres://user:pass@localhost:5432/dbname'"; \
		exit 1; \
	fi
	~/go/bin/goose postgres "$(DB_URL)" -dir migrations up

migrate-down: ## 回滚数据库迁移（向下）
	@echo "回滚数据库迁移..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "错误: 请设置 DB_URL 环境变量"; \
		exit 1; \
	fi
	~/go/bin/goose postgres "$(DB_URL)" -dir migrations down

migrate-create: ## 创建新的迁移文件（使用 NAME=my_migration_name）
	@if [ -z "$(NAME)" ]; then \
		echo "错误: 请指定迁移名称"; \
		echo "示例: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	@echo "创建迁移文件: $(NAME)"
	~/go/bin/goose postgres "$(DB_URL)" -dir migrations create $(NAME) sql

migrate-status: ## 查看迁移状态
	@echo "查看迁移状态..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "错误: 请设置 DB_URL 环境变量"; \
		exit 1; \
	fi
	~/go/bin/goose postgres "$(DB_URL)" -dir migrations status

# Docker相关
docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t auth-perm:latest .

docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	docker-compose up -d

docker-stop: ## 停止 Docker 容器
	@echo "停止 Docker 容器..."
	docker-compose down

docker-logs: ## 查看 Docker 日志
	@echo "查看 Docker 日志..."
	docker-compose logs -f

# 生成Mock文件
generate-mocks: ## 生成 Mock 文件
	@echo "生成 Mock 文件..."
	@echo "安装 mockgen..."
	go install github.com/golang/mock/mockgen@latest
	@echo "生成用户仓储 Mock..."
	mockgen -source=internal/domain/auth/repo/user_repo.go -destination=internal/domain/auth/repo/mocks/user_repo_mock.go
	@echo "生成用户服务 Mock..."
	mockgen -source=internal/domain/auth/service/auth_service.go -destination=internal/domain/auth/service/mocks/auth_service_mock.go

# 生成代码
generate: ## 生成代码（Mock, Stringer等）
	@echo "生成代码..."
	go generate ./...

# 性能测试
benchmark: ## 运行性能测试
	@echo "运行性能测试..."
	go test -bench=. -benchmem ./...

# 竞态检测
race: ## 运行竞态检测
	@echo "运行竞态检测..."
	go test -race ./...

# 内存泄漏检测
memprofile: ## 生成内存性能分析
	@echo "生成内存性能分析..."
	go test -memprofile=mem.prof -bench=. ./...

# CPU性能分析
cpuprofile: ## 生成CPU性能分析
	@echo "生成CPU性能分析..."
	go test -cpuprofile=cpu.prof -bench=. ./...

# 安全检查
security: ## 运行安全检查
	@echo "运行安全检查..."
	@if ! command -v gosec &> /dev/null; then \
		echo "安装 gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...

# 文档生成
docs: ## 生成文档
	@echo "生成文档..."
	@if ! command -v godoc &> /dev/null; then \
		echo "安装 godoc..."; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "启动文档服务器: http://localhost:6060"
	godoc -http=:6060

# 安装开发工具
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/tools/cmd/godoc@latest

# 检查所有
check-all: fmt vet lint security test ## 运行所有检查
	@echo "所有检查完成！"

# 快速开始（开发环境）
dev-setup: install-tools deps migrate-up ## 快速设置开发环境
	@echo "开发环境设置完成！"

# 生产构建
build-prod: ## 生产环境构建
	@echo "生产环境构建..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/auth-perm-linux cmd/api/main.go
	CGO_ENABLED=0 GOOS=darwin go build -a -installsuffix cgo -o bin/auth-perm-darwin cmd/api/main.go
	CGO_ENABLED=0 GOOS=windows go build -a -installsuffix cgo -o bin/auth-perm.exe cmd/api/main.go

# 创建发布包
release: build-prod ## 创建发布包
	@echo "创建发布包..."
	mkdir -p release
	cp bin/* release/
	cp -r configs release/
	cp -r migrations release/
	cp -r scripts release/
	cp README.md LICENSE release/
	tar -czf release/auth-perm-linux.tar.gz -C release auth-perm-linux configs migrations scripts README.md LICENSE
	tar -czf release/auth-perm-darwin.tar.gz -C release auth-perm-darwin configs migrations scripts README.md LICENSE
	zip -j release/auth-perm-windows.zip release/auth-perm.exe configs/* migrations/* scripts/* README.md LICENSE
	@echo "发布包已创建在 release/ 目录"

# 监控应用
monitor: ## 监控应用（需要安装 monitoring tools）
	@echo "监控应用..."
	@if ! command -v httperf &> /dev/null; then \
		echo "请安装 httperf 进行性能测试"; \
	else \
		httperf --hog --server=localhost --port=8080 --uri=/health --num-conns=100 --rate=10; \
	fi

# 清理所有
clean-all: clean ## 清理所有生成的文件
	@echo "清理所有生成文件..."
	rm -rf release/
	rm -rf vendor/
	go clean -modcache