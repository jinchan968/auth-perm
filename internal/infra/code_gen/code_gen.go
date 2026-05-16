package code_gen

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

// CodeGenerator Code生成器接口
type CodeGenerator interface {
	// GenerateCode 生成下一个自增Code
	// currentMaxCode: 当前最大的code，如果为空则从000001开始
	// prefix: 前缀（如P、T、R等）
	// tenantID: 租户ID（可选，用于区分不同租户的计数器）
	GenerateCode(currentMaxCode, prefix string, tenantID ...string) (string, error)
}

// RedisCodeGenerator 基于Redis的Code生成器
type RedisCodeGenerator struct {
	client *redis.Client
	lock   *sync.Mutex
}

// NewRedisCodeGenerator 创建Redis Code生成器
func NewRedisCodeGenerator(client *redis.Client) *RedisCodeGenerator {
	return &RedisCodeGenerator{
		client: client,
		lock:   &sync.Mutex{},
	}
}

// buildKey 构建Redis key
// 格式: code_gen:{type}:{tenantID} 或 code_gen:{type} (对于租户级别的code)
func buildKey(prefix string, tenantID string) string {
	// 统一使用大写前缀
	prefix = strings.ToUpper(prefix)
	if tenantID != "" {
		return fmt.Sprintf("code_gen:%s:%s", prefix, tenantID)
	}
	return fmt.Sprintf("code_gen:%s", prefix)
}

// GenerateCode 生成下一个自增Code
// 使用Hash结构存储计数器，field为"counter"
func (g *RedisCodeGenerator) GenerateCode(currentMaxCode, prefix string, tenantID ...string) (string, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	ctx := context.Background()
	tid := ""
	if len(tenantID) > 0 {
		tid = tenantID[0]
	}

	// 使用Hash结构，field为"counter"
	key := buildKey(prefix, tid)
	field := "counter"

	// 如果currentMaxCode不为空，解析数字并设置初始值
	if currentMaxCode != "" {
		num := extractNumber(currentMaxCode)
		if num > 0 {
			// 获取当前值
			currentVal, err := g.client.HGet(ctx, key, field).Result()
			// key不存在是正常的，不需要报错
			if err != nil && err != redis.Nil {
				return "", err
			}
			// 如果当前值小于num，则设置
			if currentVal == "" {
				g.client.HSet(ctx, key, field, num)
			} else {
				currentNum, _ := strconv.Atoi(currentVal)
				if num > currentNum {
					g.client.HSet(ctx, key, field, num)
				}
			}
		}
	}

	// 原子自增
	newNum, err := g.client.HIncrBy(ctx, key, field, 1).Result()
	if err != nil {
		return "", err
	}

	// 检查是否需要循环
	if newNum > 999999 {
		return "", fmt.Errorf("code overflow for prefix %s and tenant %s", prefix, tid)
	}

	// 格式化输出
	return formatCode(prefix, int(newNum)), nil
}

// MemoryCodeGenerator 基于内存的Code生成器（降级方案）
type MemoryCodeGenerator struct {
	mu       sync.Mutex
	counters map[string]int // key: prefix+tenantID -> current number
}

// NewMemoryCodeGenerator 创建内存Code生成器
func NewMemoryCodeGenerator() *MemoryCodeGenerator {
	return &MemoryCodeGenerator{
		counters: make(map[string]int),
	}
}

// makeKey 创建内存计数器的key
func makeKey(prefix, tenantID string) string {
	if tenantID != "" {
		return prefix + ":" + tenantID
	}
	return prefix
}

// GenerateCode 生成下一个自增Code
func (g *MemoryCodeGenerator) GenerateCode(currentMaxCode, prefix string, tenantID ...string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	tid := ""
	if len(tenantID) > 0 {
		tid = tenantID[0]
	}

	key := makeKey(prefix, tid)

	// 如果currentMaxCode不为空，同步计数器
	if currentMaxCode != "" {
		num := extractNumber(currentMaxCode)
		// 只有当currentMaxCode比计数器大时才更新
		if num > g.counters[key] {
			g.counters[key] = num
		}
	}

	// 自增
	g.counters[key]++

	// 循环处理：如果超过999999，则回到1
	if g.counters[key] > 999999 {
		return "", fmt.Errorf("code overflow for prefix %s and tenant %s", prefix, tid)
	}

	// 格式化输出
	return formatCode(prefix, g.counters[key]), nil
}

// extractNumber 从code中提取数字部分
// 例如: "P000025" -> 25, "T000001" -> 1
func extractNumber(code string) int {
	// 找出数字部分的起始位置（第一个数字字符的位置）
	startIdx := -1
	for i := len(code) - 1; i >= 0; i-- {
		if code[i] >= '0' && code[i] <= '9' {
			startIdx = i
		} else if startIdx != -1 {
			// 已经找到数字，现在遇到了非数字字符
			break
		}
	}

	if startIdx == -1 {
		return 0
	}

	// 提取数字部分并转换
	numStr := code[startIdx:]
	// 去掉前导零
	numStr = strings.TrimLeft(numStr, "0")
	if numStr == "" {
		return 0
	}

	num, _ := strconv.Atoi(numStr)
	return num
}

// formatCode 格式化code为 prefix + 6位数字
func formatCode(prefix string, num int) string {
	// 统一转大写
	prefix = strings.ToUpper(prefix)
	return fmt.Sprintf("%s%06d", prefix, num)
}

// NewCodeGenerator 创建Code生成器
// 如果redisClient不为nil，使用Redis实现；否则使用内存实现
func NewCodeGenerator(redisClient *redis.Client) CodeGenerator {
	if redisClient != nil {
		return NewRedisCodeGenerator(redisClient)
	}
	return NewMemoryCodeGenerator()
}

// GenerateCodeWithDB 生成下一个自增Code（带DB回退）
// 支持自动从DB查询当前最大code
// tenantID: 可选的租户ID，用于Redis key区分
func GenerateCodeWithDB(gen CodeGenerator, prefix string, dbQueryFunc func() (string, error), tenantID ...string) (string, error) {
	// 先尝试从DB获取当前最大code
	currentMaxCode := ""
	if dbQueryFunc != nil {
		maxCode, err := dbQueryFunc()
		if err == nil && maxCode != "" {
			currentMaxCode = maxCode
		}
	}
	// 使用基础生成器生成
	return gen.GenerateCode(currentMaxCode, prefix, tenantID...)
}

// CleanOldKeys 清理旧格式的Redis key（迁移用）
// 旧格式: code_gen:P, code_gen:T, code_gen:R, code_gen:p, code_gen:t, code_gen:r
// 新格式: code_gen:P:{tenant_id}, code_gen:T, code_gen:R
func (g *RedisCodeGenerator) CleanOldKeys(ctx context.Context) error {
	oldKeys := []string{
		"code_gen:P", "code_gen:T", "code_gen:R",
		"code_gen:p", "code_gen:t", "code_gen:r",
	}
	for _, key := range oldKeys {
		exists, err := g.client.Exists(ctx, key).Result()
		if err == nil && exists > 0 {
			g.client.Del(ctx, key)
		}
	}
	return nil
}
