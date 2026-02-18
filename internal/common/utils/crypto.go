package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken 对令牌进行SHA-256哈希
// 用于安全存储令牌哈希值，避免明文存储
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
