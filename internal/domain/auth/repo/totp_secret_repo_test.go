package repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"auth-perm/internal/domain/auth/dm"
)

func TestTOTPSecretRepository_FindByAccountID(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&dm.TOTPSecretDO{})
	assert.NoError(t, err)

	// 创建仓储实例
	repo := NewTOTPSecretRepository(db)

	// 测试：查找不存在的账户
	totpSecret, err := repo.FindByAccountID("non-existent")
	assert.NoError(t, err)
	assert.Nil(t, totpSecret)

	// 创建测试数据
	testSecret := &dm.TOTPSecretDO{
		ID:        "test-id",
		AccountID: "test-account",
		Secret:    "ABCDEFGH",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Save(testSecret)
	assert.NoError(t, err)

	// 测试：查找存在的账户
	foundSecret, err := repo.FindByAccountID("test-account")
	assert.NoError(t, err)
	assert.NotNil(t, foundSecret)
	assert.Equal(t, "test-account", foundSecret.AccountID)
	assert.Equal(t, "ABCDEFGH", foundSecret.Secret)
	assert.Equal(t, true, foundSecret.IsEnabled)
}

func TestTOTPSecretRepository_Save(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&dm.TOTPSecretDO{})
	assert.NoError(t, err)

	// 创建仓储实例
	repo := NewTOTPSecretRepository(db)

	// 测试：保存新的TOTP密钥
	newSecret := &dm.TOTPSecretDO{
		ID:        "new-id",
		AccountID: "new-account",
		Secret:    "NEWKEY123",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		IsEnabled: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Save(newSecret)
	assert.NoError(t, err)

	// 验证保存成功
	savedSecret, err := repo.FindByAccountID("new-account")
	assert.NoError(t, err)
	assert.NotNil(t, savedSecret)
	assert.Equal(t, "new-account", savedSecret.AccountID)

	// 测试：更新已存在的TOTP密钥
	updatedSecret := &dm.TOTPSecretDO{
		ID:        "updated-id",
		AccountID: "new-account",
		Secret:    "UPDATEDKEY",
		Algorithm: "SHA256",
		Digits:    8,
		Period:    60,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Save(updatedSecret)
	assert.NoError(t, err)

	// 验证更新成功
	foundSecret, err := repo.FindByAccountID("new-account")
	assert.NoError(t, err)
	assert.NotNil(t, foundSecret)
	assert.Equal(t, "UPDATEDKEY", foundSecret.Secret)
	assert.Equal(t, "SHA256", foundSecret.Algorithm)
	assert.Equal(t, 8, foundSecret.Digits)
	assert.Equal(t, true, foundSecret.IsEnabled)
}

func TestTOTPSecretRepository_Delete(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&dm.TOTPSecretDO{})
	assert.NoError(t, err)

	// 创建仓储实例
	repo := NewTOTPSecretRepository(db)

	// 创建测试数据
	testSecret := &dm.TOTPSecretDO{
		ID:        "delete-test-id",
		AccountID: "delete-account",
		Secret:    "DELETEKEY",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Save(testSecret)
	assert.NoError(t, err)

	// 验证创建成功
	foundSecret, err := repo.FindByAccountID("delete-account")
	assert.NoError(t, err)
	assert.NotNil(t, foundSecret)

	// 删除
	err = repo.Delete("delete-account")
	assert.NoError(t, err)

	// 验证删除成功（软删除）
	foundSecret, err = repo.FindByAccountID("delete-account")
	assert.NoError(t, err)
	assert.Nil(t, foundSecret)
}

func TestTOTPSecretRepository_FindByID(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&dm.TOTPSecretDO{})
	assert.NoError(t, err)

	// 创建仓储实例
	repo := NewTOTPSecretRepository(db)

	// 测试：查找不存在的ID
	totpSecret, err := repo.FindByID("non-existent-id")
	assert.NoError(t, err)
	assert.Nil(t, totpSecret)

	// 创建测试数据
	testSecret := &dm.TOTPSecretDO{
		ID:        "find-by-id-test",
		AccountID: "find-account",
		Secret:    "FINDBYIDKEY",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Save(testSecret)
	assert.NoError(t, err)

	// 测试：查找存在的ID
	foundSecret, err := repo.FindByID("find-by-id-test")
	assert.NoError(t, err)
	assert.NotNil(t, foundSecret)
	assert.Equal(t, "find-by-id-test", foundSecret.ID)
	assert.Equal(t, "find-account", foundSecret.AccountID)
	assert.Equal(t, "FINDBYIDKEY", foundSecret.Secret)
}
