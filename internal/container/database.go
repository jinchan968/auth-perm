package container

import (
	"auth-perm/config"
	"auth-perm/internal/common/constant"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase 创建数据库连接
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()

	// 配置GORM日志
	var gormLogger logger.Interface
	if cfg.Server.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// 打开数据库连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt:            true,
		SkipDefaultTransaction: false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(constant.CacheTTLMedium)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected successfully: %s:%d/%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	return db, nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}

// AutoMigrate 自动迁移数据库
// 注意：生产环境建议使用迁移工具而不是AutoMigrate
// AutoMigrate仅用于开发环境快速同步数据库结构
func AutoMigrate(db *gorm.DB) error {
	log.Println("Warning: AutoMigrate is intended for development only.")
	log.Println("For production environments, please use migration scripts (goose).")
	log.Println("To create a new migration: make migrate-create NAME=your_migration_name")
	log.Println("To run migrations: make migrate-up")
	log.Println("To rollback migrations: make migrate-down")

	// 开发环境下可以临时启用，但生产环境必须禁用
	if os.Getenv("GIN_MODE") == "release" {
		log.Println("AutoMigrate is disabled in production mode.")
		return nil
	}

	// 开发环境下的自动迁移（谨慎使用）
	// 实际项目中建议完全禁用此功能
	log.Println("Auto migrate is disabled by default. Please use migration scripts.")

	return nil
}
