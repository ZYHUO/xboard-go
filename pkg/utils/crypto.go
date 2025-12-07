package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GenerateUUID 生成 UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateToken 生成随机 token
func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MD5 计算 MD5
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256 计算 SHA256
func SHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// GetServerKey 生成服务器密钥 (用于 Shadowsocks 2022)
func GetServerKey(createdAt int64, size int) string {
	key := fmt.Sprintf("%d", createdAt)
	hash := sha256.Sum256([]byte(key))
	return base64.StdEncoding.EncodeToString(hash[:size])
}

// UUIDToBase64 将 UUID 转换为 Base64 (用于 Shadowsocks 2022)
// 对于需要 32 字节密钥的加密方式，使用 SHA256 扩展 UUID
func UUIDToBase64(uuidStr string, size int) string {
	// 移除 UUID 中的连字符
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")
	
	if size <= 16 {
		// 16 字节密钥：直接使用 UUID 的字节
		bytes, _ := hex.DecodeString(uuidStr)
		if len(bytes) > size {
			bytes = bytes[:size]
		}
		return base64.StdEncoding.EncodeToString(bytes)
	}
	
	// 32 字节密钥：使用 SHA256 哈希 UUID 来生成足够长度的密钥
	hash := sha256.Sum256([]byte(uuidStr))
	return base64.StdEncoding.EncodeToString(hash[:size])
}

// RandomPort 从端口范围中随机选择一个端口
func RandomPort(portRange string) int {
	parts := strings.Split(portRange, "-")
	if len(parts) != 2 {
		return 0
	}
	var start, end int
	fmt.Sscanf(parts[0], "%d", &start)
	fmt.Sscanf(parts[1], "%d", &end)
	if start >= end {
		return start
	}
	b := make([]byte, 4)
	rand.Read(b)
	return start + int(b[0])%(end-start+1)
}
