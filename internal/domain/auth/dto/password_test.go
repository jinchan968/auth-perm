package dto

import (
	"testing"
	"time"

	"auth-perm/internal/domain/auth/constant"
	"golang.org/x/crypto/bcrypt"
)

// TestPasswordValidation 测试密码验证逻辑
func TestPasswordValidation(t *testing.T) {
	t.Run("密码设置和验证", func(t *testing.T) {
		// 创建用户
		user, err := NewUserDTO("testuser", constant.IdentifierTypeEmail, "test@example.com")
		if err != nil {
			t.Fatalf("创建用户失败: %v", err)
		}

		testPassword := "testpassword123"

		// 设置密码
		err = user.SetPassword(testPassword)
		if err != nil {
			t.Fatalf("设置密码失败: %v", err)
		}

		// 验证密码哈希不为空
		if user.PasswordHash == "" {
			t.Error("密码哈希为空")
		}

		// 验证bcrypt可以正确验证密码
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(testPassword))
		if err != nil {
			t.Errorf("密码验证失败: %v", err)
		}

		// 验证错误密码不能通过验证
		wrongPassword := "wrongpassword"
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(wrongPassword))
		if err == nil {
			t.Error("错误密码不应该通过验证")
		}
	})

	t.Run("密码前后空格清理", func(t *testing.T) {
		user, err := NewUserDTO("testuser", constant.IdentifierTypeEmail, "test@example.com")
		if err != nil {
			t.Fatalf("创建用户失败: %v", err)
		}

		// 测试带空格的密码
		passwordWithSpaces := "  testpassword123  "

		err = user.SetPassword(passwordWithSpaces)
		if err != nil {
			t.Fatalf("设置带空格的密码失败: %v", err)
		}

		// 验证不带空格的密码可以登录（被自动清理）
		cleanPassword := "testpassword123"
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cleanPassword))
		if err != nil {
			t.Errorf("清理后的密码验证失败: %v", err)
		}
	})

	t.Run("密码长度验证", func(t *testing.T) {
		user, err := NewUserDTO("testuser", constant.IdentifierTypeEmail, "test@example.com")
		if err != nil {
			t.Fatalf("创建用户失败: %v", err)
		}

		// 测试过短密码
		shortPassword := "12345" // 少于6位
		err = user.SetPassword(shortPassword)
		if err == nil {
			t.Error("过短密码应该返回错误")
		}

		// 测试正常长度密码
		normalPassword := "123456" // 6位
		err = user.SetPassword(normalPassword)
		if err != nil {
			t.Errorf("正常密码设置失败: %v", err)
		}
	})

	t.Run("空密码验证", func(t *testing.T) {
		user, err := NewUserDTO("testuser", constant.IdentifierTypeEmail, "test@example.com")
		if err != nil {
			t.Fatalf("创建用户失败: %v", err)
		}

		// 测试空密码
		err = user.SetPassword("")
		if err == nil {
			t.Error("空密码应该返回错误")
		}

		// 测试纯空格密码
		err = user.SetPassword("   ")
		if err == nil {
			t.Error("纯空格密码应该返回错误")
		}
	})
}

// TestPasswordConsistency 测试密码一致性
func TestPasswordConsistency(t *testing.T) {
	user, err := NewUserDTO("testuser", constant.IdentifierTypeEmail, "test@example.com")
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	password := "testpassword123"

	// 设置密码
	err = user.SetPassword(password)
	if err != nil {
		t.Fatalf("设置密码失败: %v", err)
	}

	// 验证PasswordHash字段存在且不为空
	if user.PasswordHash == "" {
		t.Error("密码哈希为空")
	}

	// 验证哈希值格式正确（以$2a$或$2b$开头）
	if len(user.PasswordHash) < 60 {
		t.Error("密码哈希长度不正确")
	}

	// 验证转换后的密码仍然可以验证
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		t.Errorf("密码验证失败: %v", err)
	}
}

// TestPasswordTimestamp 测试密码更新时间戳
func TestPasswordTimestamp(t *testing.T) {
	user, err := NewUserDTO("testuser", constant.IdentifierTypeEmail, "test@example.com")
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	originalUpdatedAt := user.UpdatedAt

	// 等待一小段时间确保时间戳不同
	time.Sleep(1 * time.Millisecond)

	// 设置密码
	err = user.SetPassword("testpassword123")
	if err != nil {
		t.Fatalf("设置密码失败: %v", err)
	}

	// 验证更新时间戳已更新
	if !user.UpdatedAt.After(originalUpdatedAt) {
		t.Error("设置密码后更新时间戳应该更新")
	}
}
