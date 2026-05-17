# 部署指南

本文中 `{#xxx}` 为占位符，部署时替换为实际值：

| 占位符 | 说明 |
|--------|------|
| `{#userName}` | VPS SSH 用户名（`whoami` 查看） |
| `{#vpsIp}` | VPS 公网 IP |
| `{#projectName}` | Cloudflare Pages 项目名（自动分配 `xxx.pages.dev` 域名） |

## 架构概览

```
┌──────────────────────────────────────────────────────────┐
│  Cloudflare Pages                                        │
│  ui (Next.js)  ← 仅部署 UI，不部署 newshock               │
│  https://{#projectName}.pages.dev                        │
└───────────────┬──────────────────────────────────────────┘
                │ /api/* (Next.js rewrite → VPS)
                ▼
┌──────────────────────────────────────────────────────────┐
│  Google Cloud VPS (e2-small / 1 vCPU, 2GB RAM)           │
│                                                          │
│  ┌─────────────────┐   ┌─────────────────────┐          │
│  │ auth-perm-api    │   │ auth-perm-worker     │          │
│  │ :8080 (Gin API)  │   │ (定时任务调度器)       │          │
│  │                  │   │ STOCK_SCHEDULER_     │          │
│  │                  │   │ ENABLED=false        │          │
│  └────────┬─────────┘   └──────────┬──────────┘          │
│           │                        │                      │
└───────────┼────────────────────────┼──────────────────────┘
            │                        │
            ▼                        ▼
┌───────────────────────┐  ┌───────────────────┐
│  Supabase PostgreSQL   │  │  Upstash Redis     │
│  (免费 500MB)          │  │  (免费 256MB)      │
│  db.xxx.supabase.co   │  │  us1-xxx.upstash.io│
│  sslmode=require       │  │  TLS required      │
└───────────────────────┘  └───────────────────┘
```

---

## 前置准备

| 平台 | 需要注册/获取 |
|------|-------------|
| Google Cloud | 账号 + VPS 实例（e2-micro 免费层或 e2-small） |
| Cloudflare Pages | 账号 + GitHub 仓库授权 |
| Supabase | 账号 + 创建项目 → 获取数据库连接串 |
| Upstash | 账号 + 创建 Redis 数据库 → 获取连接信息 |

---

## 一、Supabase PostgreSQL

### 1.1 创建项目

1. 登录 [supabase.com](https://supabase.com) → New Project
2. 设置数据库密码，选择离 VPS 最近的区域（如 `ap-southeast-1`）
3. 创建后进入 Settings → Database → Connection string → 选择 **Session pooler** 模式
4. 复制连接串，格式：
   ```
   postgresql://postgres.[ref]:[password]@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
   ```

### 1.2 初始化数据库

在 VPS 上执行 migrations：
```bash
# 安装 goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# 执行迁移
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgresql://postgres.xxx:password@aws-0-xxx.pooler.supabase.com:6543/postgres?sslmode=require" goose -dir migrations up
```

### 1.3 注意事项

- Supabase 免费层提供 **500MB 数据库**，A 股全量数据约 5k 条 ticker + 概念/日线会快速增长，需监控容量
- Session pooler 端口为 **6543**（常规连接为 5432）
- 必须使用 `sslmode=require`
- 免费层项目 **1 周无活动会暂停**，需定期访问或升级

---

## 二、Upstash Redis

### 2.1 创建数据库

1. 登录 [upstash.com](https://upstash.com) → Create Database
2. 选择区域（与 Supabase 同区域），类型选 Redis
3. 创建后进入 Details → 复制 **UPSTASH_REDIS_REST_URL** 和密码

### 2.2 连接信息

Upstash 的 REST URL 格式：
```
https://proud-crane-12345.upstash.io
```

从中提取：
- `REDIS_HOST`: `proud-crane-12345.upstash.io`
- `REDIS_PORT`: `6379`（Upstash Redis 端口）
- `REDIS_PASSWORD`: 详情页显示的密码
- `REDIS_USE_TLS`: `true`

### 2.3 注意事项

- 免费层 **256MB**，限流 **1000 次/天**
- **必须开启 TLS**（`REDIS_USE_TLS=true`）
- 限流触发后返回 429，业务的滑动窗口限流 + 会话缓存 + 权限缓存可能较快耗尽日配额
- 如额度不足，用内存缓存替代（`CACHE_TYPE=memory`）

---

## 三、Google Cloud VPS

### 3.1 创建实例

1. Compute Engine → Create Instance
2. 机型：`e2-micro`（免费层）或 `e2-small`（推荐，~$12/月）
3. 系统：Ubuntu 22.04 LTS
4. 防火墙：允许 HTTP(80) + HTTPS(443)，以及自定义 TCP 8080
5. 创建后 SSH 登录

### 3.2 安装依赖

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 Go（如果需要在 VPS 上编译）
wget https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 如果本地编译交叉编译则跳过 Go 安装
```

### 3.3 部署 API 服务

#### 方式 A：本地交叉编译上传（推荐）

```bash
# 本地执行
make build-prod

# 上传到 VPS
scp bin/auth-perm-linux {#userName}@{#vpsIp}:/opt/auth-perm/
scp bin/auth-perm-worker-linux {#userName}@{#vpsIp}:/opt/auth-perm/
scp config/app.yaml {#userName}@{#vpsIp}:/opt/auth-perm/config/
scp -r migrations/ {#userName}@{#vpsIp}:/opt/auth-perm/

# 创建日志目录 + 赋权（程序运行需要写 logs/）
ssh {#userName}@{#vpsIp} "chmod +x /opt/auth-perm/auth-perm-linux /opt/auth-perm/auth-perm-worker-linux && mkdir -p /opt/auth-perm/logs && chown -R {#userName}:{#userName} /opt/auth-perm"
```

#### 方式 B：VPS 上编译

```bash
git clone https://github.com/xxx/auth-perm.git
cd auth-perm
go build -o bin/auth-perm-api cmd/api/main.go
```

### 3.4 API 环境变量

在 VPS 上创建 `/opt/auth-perm/.env`：

```bash
# ===== 数据库 =====
DB_HOST=aws-0-ap-southeast-1.pooler.supabase.com
DB_PORT=6543
DB_USER=postgres.xxxxxxxx
DB_PASSWORD=your-supabase-db-password
DB_NAME=postgres
DB_SSLMODE=require

# ===== Redis =====
REDIS_HOST=your-redis.upstash.io
REDIS_PORT=6379
REDIS_PASSWORD=your-upstash-password
REDIS_USE_TLS=true

# ===== 服务端口 =====
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SUPER_ADMIN=your-admin-email@example.com

# ===== 前端 CORS（Cloudflare Pages 域名，逗号分隔）=====
CORS_ORIGINS=https://{#projectName}.pages.dev

# ===== LLM（可选）=====
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=sk-xxx
LLM_MODEL=gpt-4o-mini
```

### 3.5 创建 systemd 服务

```bash
sudo cat > /etc/systemd/system/auth-perm-api.service << 'EOF'
[Unit]
Description=auth-perm API service
After=network.target

[Service]
Type=simple
User={#userName}
WorkingDirectory=/opt/auth-perm
EnvironmentFile=/opt/auth-perm/.env
ExecStart=/opt/auth-perm/auth-perm-linux
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable auth-perm-api
sudo systemctl start auth-perm-api
sudo systemctl status auth-perm-api
```

### 3.6 Nginx 反代（安全 + HTTPS）

e2-micro（0.25 vCPU + 1GB RAM）同时运行 Go API + Nginx 完全可行，Nginx 内存占用约 10-20MB。

```bash
sudo apt install nginx -y

sudo cat > /etc/nginx/sites-available/auth-perm << 'NEOF'
server {
    listen 80;
    server_name _;  # 先用 IP，后续绑定域名后改为你的域名

    # Cloudflare Pages 发起的 API 请求
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
    }

    # health check
    location /health {
        proxy_pass http://127.0.0.1:8080;
    }
}
NEOF

sudo ln -s /etc/nginx/sites-available/auth-perm /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

配置后，8080 端口无需对外开放。GCP 防火墙只保留 80（HTTP）、443（HTTPS）、22（SSH）。

> VPS 自带一个临时公网 IP，后续绑定自定义域名后更新 `server_name` 并配置 Let's Encrypt HTTPS 即可。

#### 前端 URL 更新

Nginx 部署后，前端通过 80 端口访问 API，`NEXT_PUBLIC_API_URL` 改为：

```
NEXT_PUBLIC_API_URL=http://{#vpsIp}/api/v1
```

（注意不再带 8080 端口）

---

## 四、部署 Worker（定时任务）

### 4.1 构建

```bash
# 本地交叉编译
CGO_ENABLED=0 GOOS=linux go build -o bin/auth-perm-worker-linux cmd/worker/main.go

# 或使用 Makefile
make build-prod  # 已包含 worker 构建
```

### 4.2 Worker 环境变量

在 VPS 上创建 `/opt/auth-perm/worker.env`（与 API 共用数据库/Redis，额外加定时任务开关）：

```bash
# ===== 基础配置（同上）=====
DB_HOST=aws-0-ap-southeast-1.pooler.supabase.com
DB_PORT=6543
DB_USER=postgres.xxxxxxxx
DB_PASSWORD=your-supabase-db-password
DB_NAME=postgres
DB_SSLMODE=require

REDIS_HOST=your-redis.upstash.io
REDIS_PORT=6379
REDIS_PASSWORD=your-upstash-password
REDIS_USE_TLS=true

# ===== 服务端口 =====
SERVER_HOST=0.0.0.0
SERVER_PORT=8081

# ===== 定时任务开关 =====
# 设为 false 禁用所有财经数据下载（RSS/股票列表/K线/F10/新闻/概念/评分/Polymarket）
STOCK_SCHEDULER_ENABLED=false

# ===== 仅保留 TODO 定时任务（如不需要也可以不部署 worker）=====
```

### 4.3 Worker systemd 服务

```bash
sudo cat > /etc/systemd/system/auth-perm-worker.service << 'EOF'
[Unit]
Description=auth-perm Worker scheduler
After=network.target

[Service]
Type=simple
User={#userName}
WorkingDirectory=/opt/auth-perm
EnvironmentFile=/opt/auth-perm/worker.env
ExecStart=/opt/auth-perm/auth-perm-worker-linux
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable auth-perm-worker
sudo systemctl start auth-perm-worker
```

---

## 五、Cloudflare Pages（UI 前端）

### 5.1 适配说明

项目已安装 `@cloudflare/next-on-pages` 适配器，使 Next.js SSR + 动态路由 `[id]` 在 Cloudflare Pages 上正常运行。

### 5.2 配置

1. 登录 [Cloudflare Dashboard](https://dash.cloudflare.com) → Workers & Pages → Create → Pages → Connect to Git
2. 选择 GitHub 仓库 → 设置构建配置：

| 配置项 | 值 |
|--------|-----|
| Framework preset | Next.js |
| Root directory | `ui` |
| Build command | `npm install -g pnpm && pnpm install && pnpm build && npx @cloudflare/next-on-pages` |
| Build output directory | `.vercel/output/static` |
| Deploy command | 留空 |

### 5.3 环境变量

Cloudflare Pages → Settings → Environment Variables：

| 变量名 | 值 |
|--------|-----|
| `NEXT_PUBLIC_API_URL` | `http://{#vpsIp}/api/v1` |

### 5.4 默认域名

Cloudflare Pages 自动分配 `https://{#projectName}.pages.dev` 域名，需添加到 VPS CORS 白名单：

```bash
CORS_ORIGINS=https://{#projectName}.pages.dev
```

---

## 六、整体验证

### 6.1 后端健康检查

```bash
# Nginx 代理后直接通过 80 端口
curl http://{#vpsIp}/health
# 预期: {"status":"healthy"}
```

### 6.2 前端登录验证

1. 浏览器打开 `https://{#projectName}.pages.dev`
2. 使用超级管理员邮箱注册/登录
3. 检查能否正常加载首页

### 6.3 数据库连接测试

```bash
# SSH 到 VPS 测试
psql "postgresql://postgres.xxx:password@aws-0-xxx.pooler.supabase.com:6543/postgres?sslmode=require" -c "SELECT 1"
```

### 6.4 Redis 连接测试

```bash
redis-cli -h your-redis.upstash.io -p 6379 -a password --tls PING
# 预期: PONG
```

---

## 七、环境变量速查

### VPS 上 API 和 Worker 共用

| 变量 | 说明 | 必填 |
|------|------|------|
| `DB_HOST` | Supabase 数据库主机 | ✅ |
| `DB_PORT` | 数据库端口（Supabase pooler: 6543） | ✅ |
| `DB_USER` | 数据库用户名 | ✅ |
| `DB_PASSWORD` | 数据库密码 | ✅ |
| `DB_NAME` | 数据库名（postgres） | ✅ |
| `DB_SSLMODE` | SSL 模式（require） | ✅ |
| `REDIS_HOST` | Upstash Redis 主机 | ✅ |
| `REDIS_PORT` | Redis 端口（6379） | ✅ |
| `REDIS_PASSWORD` | Redis 密码 | ✅ |
| `REDIS_USE_TLS` | Redis TLS（Upstash 必须 true） | ✅ |
| `SERVER_HOST` | 监听地址（0.0.0.0） | - |
| `SERVER_PORT` | 监听端口（8080/8081） | - |
| `SUPER_ADMIN` | 超级管理员邮箱 | ✅ |
| `CORS_ORIGINS` | 额外允许的跨域源，逗号分隔 | ✅ |
| `STOCK_SCHEDULER_ENABLED` | 是否启用财经定时任务（worker 应设为 false） | worker 必设 |
| `LLM_BASE_URL` | LLM API 地址 | 可选 |
| `LLM_API_KEY` | LLM API 密钥 | 可选 |

### Cloudflare Pages

| 变量 | 说明 |
|------|------|
| `NEXT_PUBLIC_API_URL` | `http://{#vpsIp}/api/v1` |

---

## 八、安全注意事项

1. **VPS 防火墙**：只开放 80/443（如有 Nginx）和 22（SSH）。API 端口 8080 不建议直接暴露公网，通过 Nginx 反代 `location /api/` 更安全
2. **Supabase RLS**：启用 Row Level Security，确保租户数据隔离
3. **Redis 密码**：Upstash 强制密码认证，已满足
4. **环境变量**：`.env` 文件权限 600，不提交到 Git
5. **CORS**：仅添加实际使用的域名，避免 `*`

## 九、故障排查

| 现象 | 检查 |
|------|------|
| `status=217/USER` | systemd `User=` 与实际用户名不匹配，`whoami` 确认后修改 `/etc/systemd/system/auth-perm-api.service` |
| `status=203/EXEC` | 二进制不存在或没有执行权限。检查 `ls -l`、`file`（是否 ARM 而非 x86）、`chmod +x` |
| `status=1/FAILURE` + `mkdir logs: permission denied` | 工作目录属主不对，执行 `chown -R {#userName}:{#userName} /opt/auth-perm` |
| `status=1/FAILURE` + 数据库连接错误 | `.env` 中 `DB_*` 是否正确；Supabase 免费项目是否因 1 周未活动被暂停 |
| API 起不来 | `sudo journalctl -u auth-perm-api -n 50 --no-pager` 看具体错误 |
| 前端 401 | 检查 `CORS_ORIGINS` 是否包含 Vercel 域名 |
| 前端 CORS 错误 | 确认 `CORS_ORIGINS=https://{#projectName}.pages.dev` 已在 VPS `.env` 中设置并重启了服务 |
| 数据库连接失败 | `DB_SSLMODE=require` 是否设置；Supabase 免费项目是否暂停；`DB_PORT` 是否用 6543 (pooler) |
| Redis 连接失败 | `REDIS_USE_TLS=true` 是否设置；Upstash 日请求配额是否耗尽 |
| `X-Forwarded-For` inet 报错 | Cloudflare 代理 IP 列表逗号分隔，已通过 `extractClientIP()` 取第一个 IP 修复 |

## 十、CI/CD 自动部署（GitHub Actions）

push main 分支时自动构建、上传、重启 VPS 服务，无需手动 scp。

### 10.1 配置 Secrets

GitHub 仓库 → Settings → Secrets and variables → Actions → 添加三个 secrets：

| Secret | 值 |
|--------|-----|
| `GCP_HOST` | VPS 公网 IP |
| `GCP_USER` | VPS SSH 用户名 |
| `GCP_SSH_KEY` | SSH 私钥（`~/.ssh/id_rsa` 或 `~/.ssh/id_ed25519` 的内容） |

### 10.2 工作流

`.github/workflows/deploy.yml` 已配置，触发条件：main 分支 push，且改动涉及 `cmd/**`、`internal/**`、`config/**`、`go.*`。

流程：Setup Go → 交叉编译 linux/amd64 → scp 上传 → ssh 重启 systemd 服务。

### 10.3 首次使用

确保 VPS 已配置 SSH key：
```bash
# 本机
ssh-copy-id {#userName}@{#vpsIp}

# 确认免密登录
ssh {#userName}@{#vpsIp} "echo ok"
```
