package config

import (
	"fmt"
	"log"
	"os"

	"auth-perm/internal/common/errors"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server     ServerConfig     `yaml:"server" mapstructure:"server"`
	Database   DatabaseConfig   `yaml:"database" mapstructure:"database"`
	Redis      RedisConfig      `yaml:"redis" mapstructure:"redis"`
	Cache      CacheConfig      `yaml:"cache" mapstructure:"cache"`
	Token      TokenConfig      `yaml:"token"  mapstructure:"token"`
	Log        LogConfig        `yaml:"log" mapstructure:"log"`
	Tenant     TenantConfig     `yaml:"tenant" mapstructure:"tenant"`
	OAuth      OAuthConfig      `yaml:"oauth" mapstructure:"oauth"`
	SMTP       SMTPConfig       `yaml:"smtp" mapstructure:"smtp"`
	Monitoring MonitoringConfig `yaml:"monitoring" mapstructure:"monitoring"`
	RSS        RSSConfig        `yaml:"rss" mapstructure:"rss"`
	LLM        LLMConfig        `yaml:"llm" mapstructure:"llm"`
	Stock      StockConfig      `yaml:"stock" mapstructure:"stock"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host       string `mapstructure:"host" yaml:"host" env:"SERVER_HOST" default:"localhost"`
	Port       int    `mapstructure:"port" yaml:"port" env:"SERVER_PORT" default:"8080"`
	Mode       string `mapstructure:"mode" yaml:"mode" env:"GIN_MODE" default:"debug"`
	SuperAdmin string `mapstructure:"super_admin" yaml:"super_admin" env:"SUPER_ADMIN" default:""`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host" yaml:"host" env:"DB_HOST" required:"true"`
	Port         int    `mapstructure:"port" yaml:"port" env:"DB_PORT" default:"5432"`
	User         string `mapstructure:"user" yaml:"user" env:"DB_USER" required:"true"`
	Password     string `mapstructure:"password" yaml:"password" env:"DB_PASSWORD" required:"true"`
	DBName       string `mapstructure:"dbname" yaml:"dbname" env:"DB_NAME" required:"true"`
	SSLMode      string `mapstructure:"sslmode" yaml:"sslmode" env:"DB_SSLMODE" default:"disable"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"10"`
	MaxOpenConns int    `mapstructure:"max_open_conns" yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" default:"100"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host" yaml:"host" env:"REDIS_HOST" default:"localhost"`
	Port     int    `mapstructure:"port" yaml:"port" env:"REDIS_PORT" default:"6379"`
	Password string `mapstructure:"password" yaml:"password" env:"REDIS_PASSWORD" default:""`
	DB       int    `mapstructure:"db" yaml:"db" env:"REDIS_DB" default:"0"`
	PoolSize int    `mapstructure:"pool_size" yaml:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Type string `yaml:"type" mapstructure:"type" default:"redis"` // redis or memory
}

// TokenConfig 令牌配置
type TokenConfig struct {
	Secret           string `mapstructure:"secret" yaml:"secret" env:"TOKEN_SECRET" required:"true"`
	ExpiresIn        string `mapstructure:"expires_in" yaml:"expires_in" env:"TOKEN_EXPIRES_IN" default:"24h"`
	SessionSecret    string `mapstructure:"session_secret" yaml:"session_secret" env:"SESSION_SECRET" required:"true"`
	SessionExpiresIn string `mapstructure:"session_expires_in" yaml:"session_expires_in" env:"SESSION_EXPIRES_IN" default:"168h"`
	CookieSecure     bool   `mapstructure:"cookie_secure" yaml:"cookie_secure" env:"COOKIE_SECURE" default:"false"`
	CookieHTTPOnly   bool   `mapstructure:"cookie_http_only" yaml:"cookie_http_only" env:"COOKIE_HTTP_ONLY" default:"true"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level" yaml:"level" env:"LOG_LEVEL" default:"info"`
	File       string `mapstructure:"file" yaml:"file" env:"LOG_FILE" default:"logs/app.log"`
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size" env:"LOG_MAX_SIZE" default:"100"`
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups" env:"LOG_MAX_BACKUPS" default:"3"`
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age" env:"LOG_MAX_AGE" default:"28"`
	Compress   bool   `mapstructure:"compress" yaml:"compress" env:"LOG_COMPRESS" default:"true"`
}

// TenantConfig 多租户配置
type TenantConfig struct {
	DefaultTenantID   string `mapstructure:"default_tenant_id" yaml:"default_tenant_id" env:"DEFAULT_TENANT_ID" default:"default"`
	EnableMultiTenant bool   `mapstructure:"enable_multi_tenant" yaml:"enable_multi_tenant" env:"ENABLE_MULTI_TENANT" default:"true"`
}

// OAuthConfig OAuth配置
type OAuthConfig struct {
	GitHub GitHubOAuthConfig `yaml:"github" mapstructure:"github"`
	Google GoogleOAuthConfig `yaml:"google" mapstructure:"google"`
	WeChat WeChatOAuthConfig `yaml:"wechat" mapstructure:"wechat"`
}

// GitHubOAuthConfig GitHub OAuth配置
type GitHubOAuthConfig struct {
	ClientID     string `yaml:"client_id" env:"GITHUB_CLIENT_ID"`
	ClientSecret string `yaml:"client_secret" env:"GITHUB_CLIENT_SECRET"`
	RedirectURL  string `yaml:"redirect_url" env:"GITHUB_REDIRECT_URL"`
}

// GoogleOAuthConfig Google OAuth配置
type GoogleOAuthConfig struct {
	ClientID     string `yaml:"client_id" env:"GOOGLE_CLIENT_ID"`
	ClientSecret string `yaml:"client_secret" env:"GOOGLE_CLIENT_SECRET"`
	RedirectURL  string `yaml:"redirect_url" env:"GOOGLE_REDIRECT_URL"`
}

// WeChatOAuthConfig 微信OAuth配置
type WeChatOAuthConfig struct {
	AppID       string `yaml:"app_id" env:"WECHAT_APP_ID"`
	AppSecret   string `yaml:"app_secret" env:"WECHAT_APP_SECRET"`
	RedirectURL string `yaml:"redirect_url" env:"WECHAT_REDIRECT_URL"`
}

// SMTPConfig SMTP邮件配置
type SMTPConfig struct {
	Host     string `yaml:"host" mapstructure:"host" env:"SMTP_HOST"`
	Port     int    `yaml:"port" mapstructure:"port" env:"SMTP_PORT" default:"587"`
	Username string `yaml:"username" mapstructure:"username" env:"SMTP_USERNAME"`
	Password string `yaml:"password" mapstructure:"password" env:"SMTP_PASSWORD"`
	From     string `yaml:"from" mapstructure:"from" env:"SMTP_FROM_EMAIL"`
	FromName string `yaml:"from_name" mapstructure:"from_name" env:"SMTP_FROM_NAME" default:"Auth-Perm"`
	UseTLS   bool   `yaml:"use_tls" mapstructure:"use_tls" env:"SMTP_USE_TLS" default:"true"`
}

// RSSConfig RSS 采集配置
type RSSConfig struct {
	Feeds         []FeedConfig `yaml:"feeds" mapstructure:"feeds"`
	FetchInterval int          `yaml:"fetch_interval" mapstructure:"fetch_interval" default:"30"`
	ScoreInterval int          `yaml:"score_interval" mapstructure:"score_interval" default:"60"`
	TenantID      string       `yaml:"tenant_id" mapstructure:"tenant_id"`
	UserAgent     string       `yaml:"user_agent" mapstructure:"user_agent" default:"NewshockBot/1.0"`
}

// LLMConfig 大模型配置（OpenAI 兼容接口）
type LLMConfig struct {
	BaseURL string `yaml:"base_url" mapstructure:"base_url" env:"LLM_BASE_URL"`
	APIKey  string `yaml:"api_key" mapstructure:"api_key" env:"LLM_API_KEY"`
	Model   string `yaml:"model" mapstructure:"model" default:"gpt-4o-mini"`
}

// StockConfig A股数据采集配置
type StockConfig struct {
	SyncInterval      int    `yaml:"sync_interval" mapstructure:"sync_interval" default:"24"`            // 股票列表同步间隔（小时）
	DailySyncInterval int    `yaml:"daily_sync_interval" mapstructure:"daily_sync_interval" default:"4"` // 日线数据同步间隔（小时）
	StartupDelayMin   int    `yaml:"startup_delay_min" mapstructure:"startup_delay_min" default:"5"`     // 启动后首次同步延迟（分钟），股票列表用此值，日线/概念延迟加倍
	TenantID          string `yaml:"tenant_id" mapstructure:"tenant_id"`
	TushareToken      string `yaml:"tushare_token" mapstructure:"tushare_token" env:"TUSHARE_TOKEN"` // Tushare Pro API token
}

// FeedConfig 单个 RSS 源配置
type FeedConfig struct {
	URL     string `yaml:"url" mapstructure:"url"`
	Source  string `yaml:"source" mapstructure:"source"`
	Channel string `yaml:"channel" mapstructure:"channel"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	EnableHealthCheck bool `mapstructure:"enable_health_check" yaml:"enable_health_check" env:"ENABLE_HEALTH_CHECK" default:"true"`
	EnableMetrics     bool `mapstructure:"enable_metrics" yaml:"enable_metrics" env:"ENABLE_METRICS" default:"true"`
	EnableAlerting    bool `mapstructure:"enable_alerting" yaml:"enable_alerting" env:"ENABLE_ALERTING" default:"true"`

	// 报警阈值
	MaxDBConnections  int `mapstructure:"max_db_connections" yaml:"max_db_connections" env:"MAX_DB_CONNECTIONS" default:"80"`
	MaxRedisMemoryMB  int `mapstructure:"max_redis_memory_mb" yaml:"max_redis_memory_mb" env:"MAX_REDIS_MEMORY_MB" default:"512"`
	MaxResponseTimeMS int `mapstructure:"max_response_time_ms" yaml:"max_response_time_ms" env:"MAX_RESPONSE_TIME_MS" default:"1000"`

	// 报警Webhook
	AlertWebhookURL string `mapstructure:"alert_webhook_url" yaml:"alert_webhook_url" env:"ALERT_WEBHOOK_URL"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// 设置配置文件路径
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 自动读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTH_PERM")

	// 绑定环境变量
	if err := viper.BindEnv("server.host", "SERVER_HOST"); err != nil {
		return nil, errors.NewInternalErrorWithDetails("绑定环境变量失败", err.Error(), err)
	}
	if err := viper.BindEnv("server.port", "SERVER_PORT"); err != nil {
		return nil, errors.NewInternalErrorWithDetails("绑定环境变量失败", err.Error(), err)
	}
	if err := viper.BindEnv("server.super_admin", "SUPER_ADMIN"); err != nil {
		return nil, errors.NewInternalErrorWithDetails("绑定环境变量失败", err.Error(), err)
	}
	if err := viper.BindEnv("llm.base_url", "LLM_BASE_URL"); err != nil {
		return nil, errors.NewInternalErrorWithDetails("绑定环境变量失败", err.Error(), err)
	}
	if err := viper.BindEnv("llm.api_key", "LLM_API_KEY"); err != nil {
		return nil, errors.NewInternalErrorWithDetails("绑定环境变量失败", err.Error(), err)
	}
	if err := viper.BindEnv("llm.model", "LLM_MODEL"); err != nil {
		return nil, errors.NewInternalErrorWithDetails("绑定环境变量失败", err.Error(), err)
	}

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，尝试使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.NewInternalErrorWithDetails("读取配置文件失败", err.Error(), err)
		}
		log.Printf("Config file not found, using default values and environment variables")
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(config); err != nil {
		return nil, errors.NewInternalErrorWithDetails("解析配置失败", err.Error(), err)
	}

	// 兜底默认值（viper 不认 struct tag 的 default）
	if config.Server.SuperAdmin == "" {
		config.Server.SuperAdmin = "admin"
	}
	if config.RSS.FetchInterval <= 0 {
		config.RSS.FetchInterval = 30
	}
	if config.RSS.ScoreInterval <= 0 {
		config.RSS.ScoreInterval = 60
	}
	if config.RSS.UserAgent == "" {
		config.RSS.UserAgent = "NewshockBot/1.0"
	}
	if config.LLM.Model == "" {
		config.LLM.Model = "gpt-4o-mini"
	}

	// 验证必需的配置项
	if err := validateConfig(config); err != nil {
		return nil, errors.WrapBizError(err, "验证配置失败")
	}

	return config, nil
}

// validateConfig FUTURE: 配置验证 - 在实现配置验证时使用
func validateConfig(config *Config) error {
	var validationErrors []string

	// 调试：输出实际加载的配置值
	log.Printf("调试配置 - 数据库最大连接数: %d, 最大空闲连接数: %d", config.Database.MaxOpenConns, config.Database.MaxIdleConns)

	// 验证数据库配置
	if config.Database.Host == "" {
		validationErrors = append(validationErrors, "数据库主机配置不能为空")
	}
	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		validationErrors = append(validationErrors, fmt.Sprintf("数据库端口配置无效: %d (期望范围: 1-65535)", config.Database.Port))
	}
	if config.Database.User == "" {
		validationErrors = append(validationErrors, "数据库用户配置不能为空")
	}
	if config.Database.Password == "" {
		validationErrors = append(validationErrors, "数据库密码配置不能为空")
	}
	if config.Database.DBName == "" {
		validationErrors = append(validationErrors, "数据库名称配置不能为空")
	}
	if config.Database.MaxIdleConns < 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("数据库最大空闲连接数配置无效: %d (期望: >= 0)", config.Database.MaxIdleConns))
	}
	if config.Database.MaxOpenConns <= 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("数据库最大连接数配置无效: %d (期望: >= 1，建议值: 10-200)", config.Database.MaxOpenConns))
	}

	// 验证Redis配置
	if config.Redis.Host == "" {
		validationErrors = append(validationErrors, "Redis主机配置不能为空")
	}
	if config.Redis.Port <= 0 || config.Redis.Port > 65535 {
		validationErrors = append(validationErrors, "Redis端口配置无效")
	}
	if config.Redis.DB < 0 {
		validationErrors = append(validationErrors, "Redis数据库编号配置无效")
	}
	if config.Redis.PoolSize <= 0 {
		validationErrors = append(validationErrors, "Redis连接池大小配置无效")
	}

	// 验证令牌配置
	if config.Token.Secret == "" {
		validationErrors = append(validationErrors, "令牌密钥配置不能为空")
	} else if len(config.Token.Secret) < 32 {
		validationErrors = append(validationErrors, "令牌密钥长度不能少于32个字符")
	}
	if config.Token.SessionSecret == "" {
		validationErrors = append(validationErrors, "会话密钥配置不能为空")
	} else if len(config.Token.SessionSecret) < 32 {
		validationErrors = append(validationErrors, "会话密钥长度不能少于32个字符")
	}

	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		validationErrors = append(validationErrors, "服务器端口配置无效")
	}
	if config.Server.Mode != "debug" && config.Server.Mode != "release" && config.Server.Mode != "test" {
		validationErrors = append(validationErrors, "Gin模式配置无效（debug/release/test）")
	}

	// 验证日志配置
	if config.Log.Level != "debug" && config.Log.Level != "info" && config.Log.Level != "warn" && config.Log.Level != "error" {
		validationErrors = append(validationErrors, "日志级别配置无效（debug/info/warn/error）")
	}
	if config.Log.MaxSize <= 0 {
		validationErrors = append(validationErrors, "日志文件最大大小配置无效")
	}
	if config.Log.MaxBackups < 0 {
		validationErrors = append(validationErrors, "日志备份文件数量配置无效")
	}
	if config.Log.MaxAge < 0 {
		validationErrors = append(validationErrors, "日志保留天数配置无效")
	}

	// 验证多租户配置
	if config.Tenant.DefaultTenantID == "" {
		validationErrors = append(validationErrors, "默认租户ID配置不能为空")
	}

	// 验证监控配置
	if config.Monitoring.MaxDBConnections <= 0 {
		validationErrors = append(validationErrors, "数据库连接告警阈值配置无效")
	}
	if config.Monitoring.MaxRedisMemoryMB <= 0 {
		validationErrors = append(validationErrors, "Redis内存告警阈值配置无效")
	}
	if config.Monitoring.MaxResponseTimeMS <= 0 {
		validationErrors = append(validationErrors, "响应时间告警阈值配置无效")
	}

	// 验证缓存配置
	if config.Cache.Type != "redis" && config.Cache.Type != "memory" {
		validationErrors = append(validationErrors, "缓存类型配置无效（redis/memory）")
	}

	// 如果有验证错误，返回第一个错误
	if len(validationErrors) > 0 {
		return errors.NewInternalErrorF("配置验证失败: %s", validationErrors[0])
	}

	return nil
}

// ValidateProductionConfig FUTURE: 生产环境配置验证 - 在实现生产环境检查时使用
func ValidateProductionConfig(config *Config) error {
	var validationErrors []string

	// 生产环境必须使用release模式
	if config.Server.Mode != "release" {
		validationErrors = append(validationErrors, "生产环境必须使用release模式")
	}

	// 生产环境必须启用Cookie安全
	if !config.Token.CookieSecure {
		validationErrors = append(validationErrors, "生产环境必须启用Cookie安全")
	}

	// 生产环境应该使用SSL
	if config.Database.SSLMode == "disable" {
		validationErrors = append(validationErrors, "生产环境建议启用数据库SSL")
	}

	// 验证敏感配置是否为默认值
	if config.Token.Secret == "your_very_long_and_complex_secret_key_here_at_least_32_characters" {
		validationErrors = append(validationErrors, "生产环境必须修改默认的令牌密钥")
	}
	if config.Token.SessionSecret == "your_very_long_and_complex_session_secret_here_at_least_32_characters" {
		validationErrors = append(validationErrors, "生产环境必须修改默认的会话密钥")
	}

	// 检查是否有空的敏感配置
	if config.Database.Password == "" {
		validationErrors = append(validationErrors, "数据库密码不能为空")
	}

	// 如果有验证错误，返回所有错误
	if len(validationErrors) > 0 {
		return errors.NewBusinessErrorF("生产环境配置验证失败:\n%s", formatValidationErrors(validationErrors))
	}

	return nil
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(errors []string) string {
	result := ""
	for i, err := range errors {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("  - %s", err)
	}
	return result
}

// ValidateDevelopmentConfig FUTURE: 开发环境配置验证 - 在实现开发环境检查时使用
func ValidateDevelopmentConfig(config *Config) error {
	var warnings []string

	// 开发环境应该使用debug模式
	if config.Server.Mode != "debug" {
		warnings = append(warnings, "开发环境建议使用debug模式")
	}

	// 开发环境可以禁用Cookie安全
	if config.Token.CookieSecure {
		warnings = append(warnings, "开发环境可以禁用Cookie安全（COOKIE_SECURE=false）")
	}

	// 如果有警告，返回警告
	if len(warnings) > 0 {
		log.Printf("开发环境配置建议:\n%s", formatValidationErrors(warnings))
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetAddr 获取Redis地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment 是否为开发环境
func (c *ServerConfig) IsDevelopment() bool {
	return c.Mode == "debug"
}

// IsProduction 是否为生产环境
func (c *ServerConfig) IsProduction() bool {
	return c.Mode == "release"
}

// EnsureLogDir 确保日志目录存在
func EnsureLogDir(logFile string) error {
	logDir := "logs"
	if logFile != "" {
		if idx := lastIndex(logFile, "/"); idx != -1 {
			logDir = logFile[:idx]
		}
	}

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return os.MkdirAll(logDir, 0755)
	}
	return nil
}

// lastIndex FUTURE: 最后索引查找 - 在实现字符串处理时使用
func lastIndex(s, sep string) int {
	for i := len(s) - len(sep); i >= 0; i-- {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
